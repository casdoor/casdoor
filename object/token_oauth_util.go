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
	"strings"
	"sync"
	"time"

	"github.com/casdoor/casdoor/i18n"
	"github.com/casdoor/casdoor/util"
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
	EndpointError        = "endpoint_error"
)

var DeviceAuthMap = sync.Map{}

type Code struct {
	Message string `xorm:"varchar(100)" json:"message"`
	Code    string `xorm:"varchar(100)" json:"code"`
}

type TokenWrapper struct {
	AccessToken  string `json:"access_token"`
	IdToken      string `json:"id_token"`
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
	Scope         string
	RequestAt     time.Time
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

func ExpireTokenByAccessToken(accessToken string) (bool, *Application, *Token, error) {
	token, err := GetTokenByAccessToken(accessToken)
	if err != nil {
		return false, nil, nil, err
	}
	if token == nil {
		return false, nil, nil, nil
	}

	token.ExpiresIn = 0
	affected, err := ormer.Engine.ID(core.PK{token.Owner, token.Name}).Cols("expires_in").Update(token)
	if err != nil {
		return false, nil, nil, err
	}

	application, err := getApplication(token.Owner, token.Application)
	if err != nil {
		return false, nil, nil, err
	}

	return affected != 0, application, token, nil
}

func CheckOAuthLogin(clientId string, responseType string, redirectUri string, scope string, state string, lang string) (string, *Application, error) {
	if responseType != "code" && responseType != "token" && responseType != "id_token" {
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

	// Mask application for /api/get-app-login
	application.ClientSecret = ""
	return "", application, nil
}

func GetOAuthCode(userId string, clientId string, provider string, signinMethod string, responseType string, redirectUri string, scope string, state string, nonce string, challenge string, resource string, host string, lang string) (*Code, error) {
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

	msg, application, err := CheckOAuthLogin(clientId, responseType, redirectUri, scope, state, lang)
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
	accessToken, refreshToken, tokenName, err := generateJwtToken(application, user, provider, signinMethod, nonce, scope, resource, host)
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
		Code:          util.GenerateClientId(),
		AccessToken:   accessToken,
		RefreshToken:  refreshToken,
		ExpiresIn:     int(application.ExpireInHours * float64(hourSeconds)),
		Scope:         scope,
		TokenType:     "Bearer",
		CodeChallenge: challenge,
		CodeIsUsed:    false,
		CodeExpireIn:  time.Now().Add(time.Minute * 5).Unix(),
		Resource:      resource,
	}
	_, err = AddToken(token)
	if err != nil {
		return nil, err
	}

	return &Code{
		Message: "",
		Code:    token.Code,
	}, nil
}

func RefreshToken(application *Application, grantType string, refreshToken string, scope string, clientId string, clientSecret string, host string, dpopProof string) (interface{}, error) {
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

	if clientSecret != "" && application.ClientSecret != clientSecret {
		return &TokenError{
			Error:            InvalidClient,
			ErrorDescription: "client_secret is invalid",
		}, nil
	}

	// check whether the refresh token is valid, and has not expired.
	token, err := GetTokenByRefreshToken(refreshToken)
	if err != nil || token == nil {
		return &TokenError{
			Error:            InvalidGrant,
			ErrorDescription: "refresh token is invalid or revoked",
		}, nil
	}

	// check if the token has been invalidated (e.g., by SSO logout)
	if token.ExpiresIn <= 0 {
		return &TokenError{
			Error:            InvalidGrant,
			ErrorDescription: "refresh token is expired",
		}, nil
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
	if application.TokenFormat == "JWT-Standard" {
		oldToken, err := ParseStandardJwtToken(refreshToken, cert)
		if err != nil {
			return &TokenError{
				Error:            InvalidGrant,
				ErrorDescription: fmt.Sprintf("parse refresh token error: %s", err.Error()),
			}, nil
		}
		oldTokenScope = oldToken.Scope
	} else {
		oldToken, err := ParseJwtToken(refreshToken, cert)
		if err != nil {
			return &TokenError{
				Error:            InvalidGrant,
				ErrorDescription: fmt.Sprintf("parse refresh token error: %s", err.Error()),
			}, nil
		}
		oldTokenScope = oldToken.Scope
	}

	if scope == "" {
		scope = oldTokenScope
	}

	// generate a new token
	user, err := getUser(application.Organization, token.User)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return "", fmt.Errorf("The user: %s doesn't exist", util.GetId(application.Organization, token.User))
	}

	if user.IsForbidden {
		return &TokenError{
			Error:            InvalidGrant,
			ErrorDescription: "the user is forbidden to sign in, please contact the administrator",
		}, nil
	}

	err = ExtendUserWithRolesAndPermissions(user)
	if err != nil {
		return nil, err
	}

	newAccessToken, newRefreshToken, tokenName, err := generateJwtToken(application, user, "", "", "", scope, "", host)
	if err != nil {
		return &TokenError{
			Error:            EndpointError,
			ErrorDescription: fmt.Sprintf("generate jwt token error: %s", err.Error()),
		}, nil
	}

	newToken := &Token{
		Owner:        application.Owner,
		Name:         tokenName,
		CreatedTime:  util.GetCurrentTime(),
		Application:  application.Name,
		Organization: user.Owner,
		User:         user.Name,
		Code:         util.GenerateClientId(),
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    int(application.ExpireInHours * float64(hourSeconds)),
		Scope:        scope,
		TokenType:    "Bearer",
	}
	_, err = AddToken(newToken)
	if err != nil {
		return nil, err
	}

	// Apply DPoP binding to the refreshed token if a DPoP proof was provided.
	if dpopProof != "" {
		dpopHtu := GetDPoPHtu(host, "/api/login/oauth/access_token")
		jkt, err := ValidateDPoPProof(dpopProof, "POST", dpopHtu, "")
		if err != nil {
			return &TokenError{
				Error:            "invalid_dpop_proof",
				ErrorDescription: err.Error(),
			}, nil
		}
		newToken.TokenType = "DPoP"
		newToken.DPoPJkt = jkt
		if err = updateTokenDPoP(newToken); err != nil {
			return nil, err
		}
	}

	_, err = DeleteToken(token)
	if err != nil {
		return nil, err
	}

	tokenWrapper := &TokenWrapper{
		AccessToken:  newToken.AccessToken,
		IdToken:      newToken.AccessToken,
		RefreshToken: newToken.RefreshToken,
		TokenType:    newToken.TokenType,
		ExpiresIn:    newToken.ExpiresIn,
		Scope:        newToken.Scope,
	}
	return tokenWrapper, nil
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

func ValidateClientAssertion(clientAssertion string, host string) (bool, *Application, error) {
	token, err := ParseJwtTokenWithoutValidation(clientAssertion)
	if err != nil {
		return false, nil, err
	}

	clientId, err := token.Claims.GetSubject()
	if err != nil {
		return false, nil, err
	}

	application, err := GetApplicationByClientId(clientId)
	if err != nil {
		return false, nil, err
	}
	if application == nil {
		return false, nil, fmt.Errorf("application not found for client: [%s]", clientId)
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
func mintImplicitToken(application *Application, username string, scope string, nonce string, host string) (*Token, *TokenError, error) {
	expandedScope, ok := IsScopeValidAndExpand(scope, application)
	if !ok {
		return nil, &TokenError{
			Error:            InvalidScope,
			ErrorDescription: "the requested scope is invalid or not defined in the application",
		}, nil
	}
	scope = expandedScope

	user, err := GetUserByFields(application.Organization, username)
	if err != nil {
		return nil, nil, err
	}
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

	token, err := GetTokenByUser(application, user, scope, nonce, host)
	if err != nil {
		return nil, nil, err
	}
	return token, nil, nil
}

// parseAndValidateSubjectToken validates a subject_token for RFC 8693 token exchange.
// It uses the ISSUING application's certificate (not the requesting client's) and
// enforces audience binding to prevent cross-client token reuse.
func parseAndValidateSubjectToken(subjectToken string, requestingClientId string) (owner, name, scope string, tokenErr *TokenError, err error) {
	unverifiedToken, err := ParseJwtTokenWithoutValidation(subjectToken)
	if err != nil {
		return "", "", "", &TokenError{Error: InvalidGrant, ErrorDescription: fmt.Sprintf("invalid subject_token: %s", err.Error())}, nil
	}

	unverifiedClaims, ok := unverifiedToken.Claims.(*Claims)
	if !ok || unverifiedClaims.Azp == "" {
		return "", "", "", &TokenError{Error: InvalidGrant, ErrorDescription: "subject_token is missing the azp claim"}, nil
	}

	issuingApp, err := GetApplicationByClientId(unverifiedClaims.Azp)
	if err != nil {
		return "", "", "", nil, err
	}
	if issuingApp == nil {
		return "", "", "", &TokenError{Error: InvalidGrant, ErrorDescription: fmt.Sprintf("subject_token issuing application not found: %s", unverifiedClaims.Azp)}, nil
	}

	cert, err := getCertByApplication(issuingApp)
	if err != nil {
		return "", "", "", nil, err
	}
	if cert == nil {
		return "", "", "", &TokenError{Error: EndpointError, ErrorDescription: fmt.Sprintf("cert for issuing application %s cannot be found", unverifiedClaims.Azp)}, nil
	}

	if issuingApp.TokenFormat == "JWT-Standard" {
		standardClaims, err := ParseStandardJwtToken(subjectToken, cert)
		if err != nil {
			return "", "", "", &TokenError{Error: InvalidGrant, ErrorDescription: fmt.Sprintf("invalid subject_token: %s", err.Error())}, nil
		}
		return standardClaims.Owner, standardClaims.Name, standardClaims.Scope, nil, nil
	}

	claims, err := ParseJwtToken(subjectToken, cert)
	if err != nil {
		return "", "", "", &TokenError{Error: InvalidGrant, ErrorDescription: fmt.Sprintf("invalid subject_token: %s", err.Error())}, nil
	}

	// Audience binding: requesting client must be the issuer itself or appear in token's aud.
	// Prevents an attacker from exchanging App A's token to obtain an App B token (RFC 8693 §2.1).
	if issuingApp.ClientId != requestingClientId {
		audienceMatched := false
		for _, aud := range claims.Audience {
			if aud == requestingClientId {
				audienceMatched = true
				break
			}
		}
		if !audienceMatched {
			return "", "", "", &TokenError{Error: InvalidGrant, ErrorDescription: fmt.Sprintf("subject_token audience does not include the requesting client '%s'", requestingClientId)}, nil
		}
	}

	return claims.Owner, claims.Name, claims.Scope, nil, nil
}

// createGuestUserToken creates a new guest user and returns a token for them.
func createGuestUserToken(application *Application, clientSecret string, verifier string) (*Token, *TokenError, error) {
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

	affected, err := AddUser(guestUser, "en")
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

	accessToken, refreshToken, tokenName, err := generateJwtToken(application, guestUser, "", "", "", "", "", "")
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
		Code:          util.GenerateClientId(),
		AccessToken:   accessToken,
		RefreshToken:  refreshToken,
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
