// Copyright 2024 The Casdoor Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package object

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/url"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/casdoor/casdoor/i18n"
	"github.com/casdoor/casdoor/util"
	"github.com/golang-jwt/jwt/v5"
	"github.com/xorm-io/core"
)

const (
	hourSeconds          = int(time.Hour / time.Second)
	InvalidRequest       = "invalid_request"
	InvalidClient        = "invalid_client"
	InvalidGrant         = "invalid_grant"
	UnauthorizedClient   = "unauthorized_client"
	UnsupportedGrantType = "unsupported_grant_type"
	InvalidScope         = "invalid_scope"
	InvalidTarget        = "invalid_target"
	EndpointError        = "endpoint_error"
	DeviceAuthExpiresIn  = 120
	DeviceAuthInterval   = 5

	DeviceAuthStatusPending     = "pending"
	DeviceAuthStatusApproved    = "approved"
	DeviceAuthStatusDenied      = "denied"
	DeviceAuthStatusTokenIssued = "token_issued"
)

// DeviceAuthMap stores the transient state of the OAuth 2.0 Device Authorization Grant (RFC 8628).
// It defaults to an in-memory store; call InitDeviceAuthStore() at startup to switch to Redis
// when redisEndpoint is configured, enabling correct behaviour across multiple replicas.
var DeviceAuthMap deviceAuthStore = &memoryDeviceAuthStore{}

type Code struct {
	Message string `xorm:"varchar(100)" json:"message"`
	Code    string `xorm:"varchar(100)" json:"code"`
	// Token holds the tokens a hybrid flow returns together with the code
	Token *Token `xorm:"-" json:"-"`
}

// ResponseType is a parsed OAuth 2.0 / OIDC response_type: a space-separated set of
// "code", "token" and "id_token", see https://openid.net/specs/oauth-v2-multiple-response-types-1_0.html
type ResponseType struct {
	Code    bool
	Token   bool
	IdToken bool
}

func ParseResponseType(responseType string) (ResponseType, bool) {
	res := ResponseType{}
	values := strings.Fields(responseType)
	if len(values) == 0 {
		return res, false
	}

	for _, value := range values {
		switch value {
		case "code":
			if res.Code {
				return res, false
			}
			res.Code = true
		case "token":
			if res.Token {
				return res, false
			}
			res.Token = true
		case "id_token":
			if res.IdToken {
				return res, false
			}
			res.IdToken = true
		default:
			return res, false
		}
	}
	return res, true
}

func (rt ResponseType) IsCodeOnly() bool {
	return rt.Code && !rt.Token && !rt.IdToken
}

// IsNonceRequired tells whether the request has to carry a nonce: OIDC requires it whenever a token
// comes back from the authorization endpoint, i.e. for the implicit and hybrid flows
func (rt ResponseType) IsNonceRequired(scope string) bool {
	if !util.InSlice(strings.Fields(scope), "openid") {
		return false
	}
	return rt.IdToken || (rt.Code && rt.Token)
}

// IsSessionAuthFresh tells whether signing in with an existing session satisfies the prompt and
// max_age parameters of an authorization request: prompt=login and a max_age older than the time the
// user entered credentials (authTime, 0 if unknown) require entering them again
func IsSessionAuthFresh(prompt string, maxAge string, authTime int64) bool {
	if util.InSlice(strings.Fields(prompt), "login") {
		return false
	}

	seconds, err := strconv.ParseInt(maxAge, 10, 64)
	if err != nil {
		return true
	}
	return authTime != 0 && time.Now().Unix()-authTime <= seconds
}

// checkResponseTypeGrant checks the parts of the response type returning tokens from the
// authorization endpoint against the grant types the application allows
func checkResponseTypeGrant(rt ResponseType, grantTypes []string) string {
	if rt.Token && !IsGrantTypeValid("token", grantTypes) {
		return "token"
	}
	if rt.IdToken && !IsGrantTypeValid("id_token", grantTypes) {
		return "id_token"
	}
	return ""
}

type TokenWrapper struct {
	AccessToken  string `json:"access_token"`
	IdToken      string `json:"id_token,omitempty"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
}

type TokenError struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description,omitempty"`
}

// DPoPConfirmation holds the DPoP key confirmation claim (RFC 9449).
type DPoPConfirmation struct {
	JKT string `json:"jkt"`
}

type IntrospectionResponse struct {
	Active    bool              `json:"active"`
	Scope     string            `json:"scope,omitempty"`
	ClientId  string            `json:"client_id,omitempty"`
	Username  string            `json:"username,omitempty"`
	TokenType string            `json:"token_type,omitempty"`
	Exp       int64             `json:"exp,omitempty"`
	Iat       int64             `json:"iat,omitempty"`
	Nbf       int64             `json:"nbf,omitempty"`
	Sub       string            `json:"sub,omitempty"`
	Aud       []string          `json:"aud,omitempty"`
	Iss       string            `json:"iss,omitempty"`
	Jti       string            `json:"jti,omitempty"`
	Cnf       *DPoPConfirmation `json:"cnf,omitempty"` // RFC 9449 DPoP key binding
}

type DeviceAuthCache struct {
	UserSignIn    bool
	UserName      string
	ApplicationId string
	ClientId      string
	Scope         string
	RequestAt     time.Time
	Status        string
	CancelToken   string
	ExpiresIn     int
}

func InitCleanupDeviceAuthMap() {
	InitDeviceAuthStore()
	util.SafeGoroutine(func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			now := time.Now()
			DeviceAuthMap.Range(func(key, value any) bool {
				cache := value.(DeviceAuthCache)
				expiresIn := cache.ExpiresIn
				if expiresIn == 0 {
					expiresIn = DeviceAuthExpiresIn
				}
				if cache.RequestAt.Add(time.Duration(expiresIn) * time.Second).Before(now) {
					DeviceAuthMap.Delete(key)
				}
				return true
			})
		}
	})
}

type DeviceAuthResponse struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationUri string `json:"verification_uri"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
}

// validateResourceURI validates that the resource parameter is a valid absolute URI
// according to RFC 8707 Section 2
func validateResourceURI(resource string) error {
	if resource == "" {
		return nil // empty resource is allowed (backward compatibility)
	}

	parsedURL, err := url.Parse(resource)
	if err != nil {
		return fmt.Errorf("resource must be a valid URI")
	}

	// RFC 8707: The resource parameter must be an absolute URI
	if !parsedURL.IsAbs() {
		return fmt.Errorf("resource must be an absolute URI")
	}

	return nil
}

// pkceChallenge returns the base64-URL-encoded SHA256 hash of verifier, per RFC 7636
func pkceChallenge(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(sum[:])
}

// IsGrantTypeValid checks if grantType is allowed in the current application.
// authorization_code is allowed by default.
func IsGrantTypeValid(method string, grantTypes []string) bool {
	if method == "authorization_code" {
		return true
	}
	for _, m := range grantTypes {
		if m == method {
			return true
		}
	}
	return false
}

func isScopeSubset(scope string, grantedScope string) bool {
	granted := strings.Fields(grantedScope)
	for _, s := range strings.Fields(scope) {
		if !util.InSlice(granted, s) {
			return false
		}
	}
	return true
}

// isRegexScope returns true if the scope string contains regex metacharacters.
func isRegexScope(scope string) bool {
	return strings.ContainsAny(scope, ".*+?^${}()|[]\\")
}

// IsScopeValidAndExpand expands any regex patterns in the space-separated scope string
// against the application's configured scopes. Literal scopes are kept as-is
// after verifying they exist in the allowed list. Regex scopes are matched
// against every allowed scope name; all matches replace the pattern.
// If the application has no defined scopes, the original scope string is
// returned unchanged (backward-compatible behaviour).
// Returns the expanded scope string and whether the scope is valid.
func IsScopeValidAndExpand(scope string, application *Application) (string, bool) {
	if len(application.Scopes) == 0 || scope == "" {
		return scope, true
	}

	allowedNames := make([]string, 0, len(application.Scopes))
	allowedSet := make(map[string]bool, len(application.Scopes))
	for _, s := range application.Scopes {
		allowedNames = append(allowedNames, s.Name)
		allowedSet[s.Name] = true
	}

	seen := make(map[string]bool)
	var expanded []string

	for _, s := range strings.Fields(scope) {
		// Try exact match first.
		if allowedSet[s] {
			if !seen[s] {
				seen[s] = true
				expanded = append(expanded, s)
			}
			continue
		}

		// Not an exact match – if it looks like a regex, try pattern matching.
		if !isRegexScope(s) {
			return "", false
		}

		// Treat as regex pattern – must be a valid regex and match ≥ 1 scope.
		re, err := regexp.Compile("^" + s + "$")
		if err != nil {
			return "", false
		}

		matched := false
		for _, name := range allowedNames {
			if re.MatchString(name) {
				matched = true
				if !seen[name] {
					seen[name] = true
					expanded = append(expanded, name)
				}
			}
		}
		if !matched {
			return "", false
		}
	}

	return strings.Join(expanded, " "), true
}

// IsScopeValid checks whether all space-separated scopes in the scope string
// are defined in the application's Scopes list (including regex expansion).
// If the application has no defined scopes, every scope is considered valid
// (backward-compatible behaviour).
func IsScopeValid(scope string, application *Application) bool {
	_, ok := IsScopeValidAndExpand(scope, application)
	return ok
}

func ExpireToken(token *Token) (bool, error) {
	token.ExpiresIn = 0
	affected, err := ormer.Engine.ID(core.PK{token.Owner, token.Name}).Cols("expires_in").Update(token)
	if err != nil {
		return false, err
	}

	// the tokens refreshed without rotation share the refresh token, it ends with all of them
	if token.RefreshTokenHash != "" {
		_, err = ormer.Engine.Where("refresh_token_hash = ? and expires_in > 0", token.RefreshTokenHash).Cols("expires_in").Update(&Token{ExpiresIn: 0})
		if err != nil {
			return false, err
		}
	}

	return affected != 0, nil
}

func CheckOAuthLogin(clientId string, responseType string, redirectUri string, scope string, state string, nonce string, lang string) (string, *Application, error) {
	rt, ok := ParseResponseType(responseType)
	if !ok {
		return fmt.Sprintf(i18n.Translate(lang, "token:Grant_type: %s is not supported in this application"), responseType), nil, nil
	}

	application, err := GetApplicationByClientId(clientId)
	if err != nil {
		return "", nil, err
	}

	if application == nil {
		return i18n.Translate(lang, "token:Invalid client_id"), nil, nil
	}

	if !application.IsRedirectUriValid(redirectUri) {
		return fmt.Sprintf(i18n.Translate(lang, "token:Redirect URI: %s doesn't exist in the allowed Redirect URI list"), redirectUri), application, nil
	}

	if !IsScopeValid(scope, application) {
		return i18n.Translate(lang, "token:Invalid scope"), application, nil
	}

	if grantType := checkResponseTypeGrant(rt, application.GrantTypes); grantType != "" {
		return fmt.Sprintf(i18n.Translate(lang, "token:Grant_type: %s is not supported in this application"), grantType), application, nil
	}

	if nonce == "" && rt.IsNonceRequired(scope) {
		return i18n.Translate(lang, "token:The nonce parameter is required for this response type"), application, nil
	}

	// Mask application for /api/get-app-login
	application.ClientSecret = ""
	return "", application, nil
}

func checkOAuthCodeUser(user *User, application *Application, lang string) (string, error) {
	if user.IsDeleted {
		return i18n.Translate(lang, "check:The user has been deleted and cannot be used to sign in, please contact the administrator"), nil
	}

	isUserOfApplication, err := IsUserOfApplication(user, application)
	if err != nil {
		return "", err
	}
	if !isUserOfApplication {
		return i18n.Translate(lang, "auth:Unauthorized operation"), nil
	}
	return "", nil
}

func GetOAuthCode(userId string, clientId string, provider string, signinMethod string, responseType string, redirectUri string, scope string, state string, nonce string, challenge string, resource string, sessionId string, authTime int64, host string, lang string) (*Code, error) {
	user, err := GetUser(userId)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return &Code{
			Message: fmt.Sprintf("general:The user: %s doesn't exist", userId),
			Code:    "",
		}, nil
	}
	if user.IsForbidden {
		return &Code{
			Message: "error: the user is forbidden to sign in, please contact the administrator",
			Code:    "",
		}, nil
	}

	msg, application, err := CheckOAuthLogin(clientId, responseType, redirectUri, scope, state, nonce, lang)
	if err != nil {
		return nil, err
	}

	if msg != "" {
		return &Code{
			Message: msg,
			Code:    "",
		}, nil
	}

	msg, err = checkOAuthCodeUser(user, application, lang)
	if err != nil {
		return nil, err
	}
	if msg != "" {
		return &Code{
			Message: msg,
			Code:    "",
		}, nil
	}

	// Expand regex/wildcard scopes to concrete scope names.
	expandedScope, ok := IsScopeValidAndExpand(scope, application)
	if !ok {
		return &Code{
			Message: i18n.Translate(lang, "token:Invalid scope"),
			Code:    "",
		}, nil
	}
	scope = expandedScope

	// Validate resource parameter (RFC 8707)
	if err := validateResourceURI(resource); err != nil {
		return &Code{
			Message: err.Error(),
			Code:    "",
		}, nil
	}

	err = ExtendUserWithRolesAndPermissions(user)
	if err != nil {
		return nil, err
	}

	// the code is generated first: an ID token returned together with it carries its c_hash
	rt, _ := ParseResponseType(responseType)
	code := util.GenerateAuthorizationCode()
	options := jwtTokenOptions{AuthTime: authTime, WithAtHash: rt.IsCodeOnly() || rt.Token}
	if rt.IdToken {
		options.Code = code
	}
	accessToken, refreshToken, idToken, tokenName, err := generateJwtTokenWithOptions(application, user, provider, signinMethod, nonce, scope, resource, host, options)
	if err != nil {
		return nil, err
	}

	if challenge == "null" {
		challenge = ""
	}

	token := &Token{
		Owner:         application.Owner,
		Name:          tokenName,
		CreatedTime:   util.GetCurrentTime(),
		Application:   application.Name,
		Organization:  user.Owner,
		User:          user.Name,
		Code:          code,
		AccessToken:   accessToken,
		RefreshToken:  refreshToken,
		IdToken:       idToken,
		ExpiresIn:     int(application.ExpireInHours * float64(hourSeconds)),
		Scope:         scope,
		TokenType:     "Bearer",
		CodeChallenge: challenge,
		CodeIsUsed:    false,
		CodeExpireIn:  time.Now().Add(time.Minute * 5).Unix(),
		Resource:      resource,
		SessionId:     sessionId,
	}
	_, err = AddToken(token)
	if err != nil {
		return nil, err
	}

	return &Code{
		Message: "",
		Code:    token.Code,
		Token:   token,
	}, nil
}

func RefreshToken(application *Application, grantType string, refreshToken string, scope string, clientId string, clientSecret string, resource string, host string, dpopProof string) (interface{}, error) {
	if grantType != "refresh_token" {
		return &TokenError{
			Error:            UnsupportedGrantType,
			ErrorDescription: "grant_type should be refresh_token",
		}, nil
	}

	var err error
	if application == nil {
		application, err = GetApplicationByClientId(clientId)
		if err != nil {
			return nil, err
		}

		if application == nil {
			return &TokenError{
				Error:            InvalidClient,
				ErrorDescription: "client_id is invalid",
			}, nil
		}
	}

	// check whether the refresh token is valid, and has not expired.
	token, err := GetTokenByRefreshToken(refreshToken)
	if err != nil || token == nil {
		return &TokenError{
			Error:            InvalidGrant,
			ErrorDescription: "refresh token is invalid or revoked",
		}, nil
	}

	if tokenError := checkRefreshClientSecret(application, token, clientSecret); tokenError != nil {
		return tokenError, nil
	}

	// check if the token has been invalidated (e.g., by SSO logout)
	if token.ExpiresIn <= 0 {
		return &TokenError{
			Error:            InvalidGrant,
			ErrorDescription: "refresh token is expired",
		}, nil
	}

	// The refresh token must belong to the authenticated client, exactly as the
	// authorization_code exchange checks it. Without this a client may present a
	// refresh token issued to a different application and, together with the
	// audience restore below, mint a token carrying another grant's resource.
	if application.Name != token.Application {
		return &TokenError{
			Error:            InvalidGrant,
			ErrorDescription: fmt.Sprintf("the token is for wrong application (client_id), application.Name: [%s], token.Application: [%s]", application.Name, token.Application),
		}, nil
	}

	// RFC 8707: the refreshed token must keep the audience of the original grant.
	// The client MAY repeat the resource parameter; when it does, it has to match.
	if resource != "" && resource != token.Resource {
		return &TokenError{
			Error:            InvalidGrant,
			ErrorDescription: fmt.Sprintf("resource parameter does not match the original grant, expected: [%s], got: [%s]", token.Resource, resource),
		}, nil
	}
	resource = token.Resource

	dpopJkt, tokenError := getRefreshDPoPJkt(token, dpopProof, host)
	if tokenError != nil {
		return tokenError, nil
	}

	cert, err := getCertByApplication(application)
	if err != nil {
		return nil, err
	}
	if cert == nil {
		return &TokenError{
			Error:            InvalidGrant,
			ErrorDescription: fmt.Sprintf("cert: %s cannot be found", application.Cert),
		}, nil
	}

	var oldTokenScope string
	var oldTokenSubject string
	if application.TokenFormat == "JWT-Standard" {
		oldToken, err := ParseStandardJwtToken(refreshToken, cert)
		if err != nil {
			return &TokenError{
				Error:            InvalidGrant,
				ErrorDescription: fmt.Sprintf("parse refresh token error: %s", err.Error()),
			}, nil
		}
		oldTokenScope = oldToken.Scope
		oldTokenSubject = oldToken.Subject
	} else {
		oldToken, err := ParseJwtToken(refreshToken, cert)
		if err != nil {
			return &TokenError{
				Error:            InvalidGrant,
				ErrorDescription: fmt.Sprintf("parse refresh token error: %s", err.Error()),
			}, nil
		}
		oldTokenScope = oldToken.Scope
		oldTokenSubject = oldToken.Subject
	}

	if scope == "" {
		scope = oldTokenScope
	} else if !isScopeSubset(scope, oldTokenScope) {
		return &TokenError{
			Error:            InvalidScope,
			ErrorDescription: "the requested scope exceeds the scope of the original grant",
		}, nil
	}

	// generate a new token
	user, err := getUser(token.Organization, token.User)
	if err != nil {
		return nil, err
	}
	// a token issued before its user was renamed still has the old name, the
	// subject is the user's ID, which a rename keeps
	if user == nil && oldTokenSubject != "" {
		user, err = GetUserByUserId(token.Organization, oldTokenSubject)
		if err != nil {
			return nil, err
		}
	}
	if user == nil {
		return &TokenError{
			Error:            InvalidGrant,
			ErrorDescription: fmt.Sprintf("the user: %s doesn't exist", util.GetId(token.Organization, token.User)),
		}, nil
	}

	if tokenError := getInactiveUserTokenError(user); tokenError != nil {
		return tokenError, nil
	}

	err = ExtendUserWithRolesAndPermissions(user)
	if err != nil {
		return nil, err
	}

	newAccessToken, newRefreshToken, newIdToken, tokenName, err := generateJwtToken(application, user, "", "", "", scope, resource, host)
	if err != nil {
		return &TokenError{
			Error:            EndpointError,
			ErrorDescription: fmt.Sprintf("generate jwt token error: %s", err.Error()),
		}, nil
	}

	// without rotation the refresh token is handed back as is and keeps its own expiry, so the
	// processes sharing it (e.g. the parallel jobs of a CLI) don't revoke it for one another
	if application.DisableRefreshRotation {
		newRefreshToken = refreshToken
	}

	newToken := &Token{
		Owner:        application.Owner,
		Name:         tokenName,
		CreatedTime:  util.GetCurrentTime(),
		Application:  application.Name,
		Organization: user.Owner,
		User:         user.Name,
		Code:         util.GenerateAuthorizationCode(),
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		IdToken:      newIdToken,
		ExpiresIn:    int(application.ExpireInHours * float64(hourSeconds)),
		Scope:        scope,
		TokenType:    "Bearer",
		Resource:     resource,
		GrantType:    token.GrantType,
		// the refreshed token stays bound to the login session that minted the original one
		SessionId: token.SessionId,
	}
	_, err = AddToken(newToken)
	if err != nil {
		return nil, err
	}

	// Apply DPoP binding to the refreshed token if a DPoP proof was provided.
	if dpopJkt != "" {
		newToken.TokenType = "DPoP"
		newToken.DPoPJkt = dpopJkt
		if err = updateTokenDPoP(newToken); err != nil {
			return nil, err
		}
	}

	// the access token of the earlier refresh may still be in use by another sharer of the
	// refresh token, it expires on its own
	if !application.DisableRefreshRotation {
		_, err = DeleteToken(token)
		if err != nil {
			return nil, err
		}
	}

	tokenWrapper := &TokenWrapper{
		AccessToken:  newToken.AccessToken,
		IdToken:      newToken.IdToken,
		RefreshToken: newToken.RefreshToken,
		TokenType:    newToken.TokenType,
		ExpiresIn:    newToken.ExpiresIn,
		Scope:        newToken.Scope,
	}
	return tokenWrapper, nil
}

// clientAuthenticatedGrantTypes mark the tokens of a client that authenticated with its secret,
// their refresh tokens are of no use without the secret either
var clientAuthenticatedGrantTypes = []string{"authorization_code", "urn:ietf:params:oauth:grant-type:token-exchange"}

func checkRefreshClientSecret(application *Application, token *Token, clientSecret string) *TokenError {
	if clientSecret == "" && !util.InSlice(clientAuthenticatedGrantTypes, token.GrantType) {
		return nil
	}

	if application.ClientSecret != clientSecret {
		return &TokenError{
			Error:            InvalidClient,
			ErrorDescription: "client_secret is invalid",
		}
	}
	return nil
}

func ValidateJwtAssertion(clientAssertion string, application *Application, host string) (bool, *Claims, error) {
	_, originBackend := getOriginFromHost(host)

	clientCert, err := getCert(application.Owner, application.ClientCert)
	if err != nil {
		return false, nil, err
	}
	if clientCert == nil {
		return false, nil, fmt.Errorf("client certificate is not configured for application: [%s]", application.GetId())
	}

	claims, err := ParseJwtToken(clientAssertion, clientCert)
	if err != nil {
		return false, nil, err
	}

	if !slices.Contains(application.RedirectUris, claims.Issuer) {
		return false, nil, nil
	}

	if !slices.Contains(claims.Audience, fmt.Sprintf("%s/api/login/oauth/access_token", originBackend)) {
		return false, nil, nil
	}

	return true, claims, nil
}

func ValidateClientAssertion(clientAssertion string, clientId string, host string) (bool, *Application, error) {
	// the subject of an external workload token is not a client ID, so the client_id of the request wins
	if clientId == "" {
		token, err := ParseJwtTokenWithoutValidation(clientAssertion)
		if err != nil {
			return false, nil, err
		}

		clientId, err = token.Claims.GetSubject()
		if err != nil {
			return false, nil, err
		}
	}

	application, err := GetApplicationByClientId(clientId)
	if err != nil {
		return false, nil, err
	}
	if application == nil {
		return false, nil, fmt.Errorf("application not found for client: [%s]", clientId)
	}

	credential, _, err := validateFederatedToken(application, clientAssertion, host)
	if err != nil {
		return false, application, err
	}
	if credential != nil {
		return true, application, nil
	}

	ok, _, err := ValidateJwtAssertion(clientAssertion, application, host)
	if err != nil {
		return false, application, err
	}
	if !ok {
		return false, application, nil
	}

	return true, application, nil
}

// mintImplicitToken mints a token for an already-authenticated user.
// Callers must verify user identity before calling this function.
func mintImplicitToken(application *Application, username string, scope string, nonce string, host string, clientIp string, lang string) (*Token, *TokenError, error) {
	user, err := GetUserByFieldsForSharedApp(application, application.Organization, username)
	if err != nil {
		return nil, nil, err
	}
	if user != nil {
		if tokenError := checkGrantUserSignin(application, user, clientIp, lang); tokenError != nil {
			return nil, tokenError, nil
		}
	}
	return mintTokenForUser(application, user, scope, nonce, host)
}

// GetDeviceCodeToken takes the full user ID recorded by the browser approval: looking a bare name
// up again in the application's organization may resolve to another organization's namesake.
func GetDeviceCodeToken(application *Application, userId string, scope string, nonce string, host string) (*Token, *TokenError, error) {
	user, err := GetUser(userId)
	if err != nil {
		return nil, nil, err
	}
	return mintTokenForUser(application, user, scope, nonce, host)
}

func mintTokenForUser(application *Application, user *User, scope string, nonce string, host string) (*Token, *TokenError, error) {
	expandedScope, ok := IsScopeValidAndExpand(scope, application)
	if !ok {
		return nil, &TokenError{
			Error:            InvalidScope,
			ErrorDescription: "the requested scope is invalid or not defined in the application",
		}, nil
	}
	scope = expandedScope

	if user == nil {
		return nil, &TokenError{
			Error:            InvalidGrant,
			ErrorDescription: "the user does not exist",
		}, nil
	}
	if user.IsForbidden {
		return nil, &TokenError{
			Error:            InvalidGrant,
			ErrorDescription: "the user is forbidden to sign in, please contact the administrator",
		}, nil
	}

	token, err := GetTokenByUser(application, user, scope, nonce, "", host)
	if err != nil {
		return nil, nil, err
	}
	return token, nil, nil
}

func getUnverifiedSubjectTokenAzp(subjectToken string) (string, *TokenError) {
	claims := jwt.MapClaims{}
	_, _, err := jwt.NewParser().ParseUnverified(subjectToken, claims)
	if err != nil {
		return "", &TokenError{Error: InvalidGrant, ErrorDescription: fmt.Sprintf("invalid subject_token: %s", err.Error())}
	}

	azp, _ := claims["azp"].(string)
	if azp == "" {
		return "", &TokenError{Error: InvalidGrant, ErrorDescription: "subject_token is missing the azp claim"}
	}
	return azp, nil
}

// parseAndValidateSubjectToken validates a subject_token for RFC 8693 token exchange.
// It uses the ISSUING application's certificate (not the requesting client's) and
// enforces audience binding to prevent cross-client token reuse.
func parseAndValidateSubjectToken(subjectToken string, requestingClientId string) (owner, name, scope string, tokenErr *TokenError, err error) {
	azp, tokenErr := getUnverifiedSubjectTokenAzp(subjectToken)
	if tokenErr != nil {
		return "", "", "", tokenErr, nil
	}

	issuingApp, err := GetApplicationByClientId(azp)
	if err != nil {
		return "", "", "", nil, err
	}
	if issuingApp == nil {
		return "", "", "", &TokenError{Error: InvalidGrant, ErrorDescription: fmt.Sprintf("subject_token issuing application not found: %s", azp)}, nil
	}

	cert, err := getCertByApplication(issuingApp)
	if err != nil {
		return "", "", "", nil, err
	}
	if cert == nil {
		return "", "", "", &TokenError{Error: EndpointError, ErrorDescription: fmt.Sprintf("cert for issuing application %s cannot be found", azp)}, nil
	}

	var audience []string
	if issuingApp.TokenFormat == "JWT-Standard" {
		standardClaims, err := ParseStandardJwtToken(subjectToken, cert)
		if err != nil {
			return "", "", "", &TokenError{Error: InvalidGrant, ErrorDescription: fmt.Sprintf("invalid subject_token: %s", err.Error())}, nil
		}
		if standardClaims.UserStandard == nil {
			return "", "", "", &TokenError{Error: InvalidGrant, ErrorDescription: "subject_token has no user"}, nil
		}
		owner, name, scope = standardClaims.Owner, standardClaims.Name, standardClaims.Scope
		audience = standardClaims.Audience
	} else {
		claims, err := ParseJwtToken(subjectToken, cert)
		if err != nil {
			return "", "", "", &TokenError{Error: InvalidGrant, ErrorDescription: fmt.Sprintf("invalid subject_token: %s", err.Error())}, nil
		}
		if claims.User == nil {
			return "", "", "", &TokenError{Error: InvalidGrant, ErrorDescription: "subject_token has no user"}, nil
		}
		owner, name, scope = claims.Owner, claims.Name, claims.Scope
		audience = claims.Audience
	}

	// Audience binding: requesting client must be the issuer itself or appear in token's aud.
	// Prevents an attacker from exchanging App A's token to obtain an App B token (RFC 8693 §2.1).
	if issuingApp.ClientId != requestingClientId && !util.InSlice(audience, requestingClientId) {
		return "", "", "", &TokenError{Error: InvalidGrant, ErrorDescription: fmt.Sprintf("subject_token audience does not include the requesting client '%s'", requestingClientId)}, nil
	}

	return owner, name, scope, nil, nil
}

// createGuestUserToken creates a new guest user and returns a token for them.
func createGuestUserToken(application *Application, clientSecret string, verifier string, lang string) (*Token, *TokenError, error) {
	if clientSecret != "" && application.ClientSecret != clientSecret {
		return nil, &TokenError{
			Error:            InvalidClient,
			ErrorDescription: "client_secret is invalid",
		}, nil
	}

	guestUsername := generateGuestUsername()
	guestPassword := util.GenerateId()

	organization, err := GetOrganization(util.GetId("admin", application.Organization))
	if err != nil {
		return nil, &TokenError{
			Error:            EndpointError,
			ErrorDescription: fmt.Sprintf("failed to get organization: %s", err.Error()),
		}, nil
	}
	if organization == nil {
		return nil, &TokenError{
			Error:            InvalidClient,
			ErrorDescription: fmt.Sprintf("organization: %s does not exist", application.Organization),
		}, nil
	}

	initScore, err := organization.GetInitScore()
	if err != nil {
		return nil, &TokenError{
			Error:            EndpointError,
			ErrorDescription: fmt.Sprintf("failed to get init score: %s", err.Error()),
		}, nil
	}

	newUserId, idErr := GenerateIdForNewUser(application)
	if idErr != nil {
		newUserId = util.GenerateId()
	}

	guestUser := &User{
		Owner:             application.Organization,
		Name:              guestUsername,
		CreatedTime:       util.GetCurrentTime(),
		Id:                newUserId,
		Type:              "normal-user",
		Password:          guestPassword,
		Tag:               "guest-user",
		DisplayName:       fmt.Sprintf("Guest_%s", guestUsername[:8]),
		Avatar:            "",
		Address:           []string{},
		Email:             "",
		Phone:             "",
		Score:             initScore,
		IsAdmin:           false,
		IsForbidden:       false,
		IsDeleted:         false,
		SignupApplication: application.Name,
		Properties:        map[string]string{},
		RegisterType:      "Guest Signup",
		RegisterSource:    fmt.Sprintf("%s/%s", application.Organization, application.Name),
	}

	affected, err := AddUser(guestUser, lang)
	if err != nil {
		return nil, &TokenError{
			Error:            EndpointError,
			ErrorDescription: fmt.Sprintf("failed to create guest user: %s", err.Error()),
		}, nil
	}
	if !affected {
		return nil, &TokenError{
			Error:            EndpointError,
			ErrorDescription: "failed to create guest user",
		}, nil
	}

	err = ExtendUserWithRolesAndPermissions(guestUser)
	if err != nil {
		return nil, &TokenError{
			Error:            EndpointError,
			ErrorDescription: fmt.Sprintf("failed to extend user: %s", err.Error()),
		}, nil
	}

	accessToken, refreshToken, idToken, tokenName, err := generateJwtToken(application, guestUser, "", "", "", "", "", "")
	if err != nil {
		return nil, &TokenError{
			Error:            EndpointError,
			ErrorDescription: fmt.Sprintf("failed to generate token: %s", err.Error()),
		}, nil
	}

	token := &Token{
		Owner:         application.Owner,
		Name:          tokenName,
		CreatedTime:   util.GetCurrentTime(),
		Application:   application.Name,
		Organization:  guestUser.Owner,
		User:          guestUser.Name,
		Code:          util.GenerateAuthorizationCode(),
		AccessToken:   accessToken,
		RefreshToken:  refreshToken,
		IdToken:       idToken,
		ExpiresIn:     int(application.ExpireInHours * float64(hourSeconds)),
		Scope:         "",
		TokenType:     "Bearer",
		CodeChallenge: "",
		CodeIsUsed:    true,
		CodeExpireIn:  0,
	}

	_, err = AddToken(token)
	if err != nil {
		return nil, &TokenError{
			Error:            EndpointError,
			ErrorDescription: fmt.Sprintf("failed to add token: %s", err.Error()),
		}, nil
	}

	return token, nil, nil
}

// generateGuestUsername generates a unique username for guest users.
func generateGuestUsername() string {
	return fmt.Sprintf("guest_%s", util.GenerateUUID())
}
