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
	"fmt"
	"strings"
	"time"

	"github.com/casdoor/casdoor/idp"
	"github.com/casdoor/casdoor/util"
)

func GetOAuthToken(grantType string, clientId string, clientSecret string, code string, verifier string, scope string, nonce string, username string, password string, host string, refreshToken string, tag string, avatar string, lang string, subjectToken string, subjectTokenType string, assertion string, clientAssertion string, clientAssertionType string, audience string, resource string, dpopProof string) (interface{}, error) {
	var (
		application *Application
		err         error
		ok          bool
	)

	if clientAssertionType == "urn:ietf:params:oauth:client-assertion-type:jwt-bearer" {
		ok, application, err = ValidateClientAssertion(clientAssertion, host)
		if err != nil {
			return nil, err
		}

		if !ok || application == nil {
			return &TokenError{
				Error:            InvalidClient,
				ErrorDescription: "client_assertion is invalid",
			}, nil
		}

		clientSecret = application.ClientSecret
		clientId = application.ClientId
	} else {
		application, err = GetApplicationByClientId(clientId)
		if err != nil {
			return nil, err
		}
	}

	if application == nil {
		return &TokenError{
			Error:            InvalidClient,
			ErrorDescription: "client_id is invalid",
		}, nil
	}

	// Check if grantType is allowed in the current application
	if !IsGrantTypeValid(grantType, application.GrantTypes) && tag == "" {
		return &TokenError{
			Error:            UnsupportedGrantType,
			ErrorDescription: fmt.Sprintf("grant_type: %s is not supported in this application", grantType),
		}, nil
	}

	var token *Token
	var tokenError *TokenError
	switch grantType {
	case "authorization_code": // Authorization Code Grant
		token, tokenError, err = GetAuthorizationCodeToken(application, clientSecret, code, verifier, resource)
	case "password": // Resource Owner Password Credentials Grant
		token, tokenError, err = GetPasswordToken(application, username, password, scope, host)
	case "client_credentials": // Client Credentials Grant
		token, tokenError, err = GetClientCredentialsToken(application, clientSecret, scope, host)
	case "token", "id_token": // Implicit Grant
		token, tokenError, err = GetImplicitToken(application, username, password, scope, nonce, host)
	case "urn:ietf:params:oauth:grant-type:jwt-bearer":
		token, tokenError, err = GetJwtBearerToken(application, assertion, scope, nonce, host)
	case "urn:ietf:params:oauth:grant-type:device_code":
		// The user has already authenticated via browser in the device flow,
		// so we skip password verification and mint a token directly.
		token, tokenError, err = mintImplicitToken(application, username, scope, nonce, host)
	case "urn:ietf:params:oauth:grant-type:token-exchange": // Token Exchange Grant (RFC 8693)
		token, tokenError, err = GetTokenExchangeToken(application, clientSecret, subjectToken, subjectTokenType, audience, scope, host)
	case "refresh_token":
		refreshToken2, err := RefreshToken(application, grantType, refreshToken, scope, clientId, clientSecret, host, dpopProof)
		if err != nil {
			return nil, err
		}
		return refreshToken2, nil
	}

	if err != nil {
		return nil, err
	}

	if tag == "wechat_miniprogram" {
		// Wechat Mini Program
		token, tokenError, err = GetWechatMiniProgramToken(application, code, host, username, avatar, lang)
		if err != nil {
			return nil, err
		}
	}

	if tokenError != nil {
		return tokenError, nil
	}

	// Apply DPoP binding (RFC 9449) if a DPoP proof was supplied by the client.
	if dpopProof != "" {
		dpopHtu := GetDPoPHtu(host, "/api/login/oauth/access_token")
		jkt, dpopErr := ValidateDPoPProof(dpopProof, "POST", dpopHtu, "")
		if dpopErr != nil {
			return &TokenError{
				Error:            "invalid_dpop_proof",
				ErrorDescription: dpopErr.Error(),
			}, nil
		}
		token.TokenType = "DPoP"
		token.DPoPJkt = jkt
		if err = updateTokenDPoP(token); err != nil {
			return nil, err
		}
	}

	token.CodeIsUsed = true

	_, err = updateUsedByCode(token)
	if err != nil {
		return nil, err
	}

	tokenWrapper := &TokenWrapper{
		AccessToken:  token.AccessToken,
		IdToken:      token.AccessToken,
		RefreshToken: token.RefreshToken,
		TokenType:    token.TokenType,
		ExpiresIn:    token.ExpiresIn,
		Scope:        token.Scope,
	}

	return tokenWrapper, nil
}

// GetAuthorizationCodeToken handles the Authorization Code Grant flow.
func GetAuthorizationCodeToken(application *Application, clientSecret string, code string, verifier string, resource string) (*Token, *TokenError, error) {
	if code == "" {
		return nil, &TokenError{
			Error:            InvalidRequest,
			ErrorDescription: "authorization code should not be empty",
		}, nil
	}

	// Handle guest user creation
	if code == "guest-user" {
		if application.Organization == "built-in" {
			return nil, &TokenError{
				Error:            InvalidGrant,
				ErrorDescription: "guest signin is not allowed for built-in organization",
			}, nil
		}
		if !application.EnableGuestSignin {
			return nil, &TokenError{
				Error:            InvalidGrant,
				ErrorDescription: "guest signin is not enabled for this application",
			}, nil
		}
		if !application.EnableSignUp {
			return nil, &TokenError{
				Error:            InvalidGrant,
				ErrorDescription: "sign up is not enabled for this application",
			}, nil
		}
		return createGuestUserToken(application, clientSecret, verifier)
	}

	token, err := getTokenByCode(code)
	if err != nil {
		return nil, nil, err
	}

	if token == nil {
		return nil, &TokenError{
			Error:            InvalidGrant,
			ErrorDescription: fmt.Sprintf("authorization code: [%s] is invalid", code),
		}, nil
	}

	if token.CodeIsUsed {
		// anti replay attacks
		return nil, &TokenError{
			Error:            InvalidGrant,
			ErrorDescription: fmt.Sprintf("authorization code has been used for token: [%s]", token.GetId()),
		}, nil
	}

	if token.CodeChallenge != "" {
		challengeAnswer := pkceChallenge(verifier)
		if challengeAnswer != token.CodeChallenge {
			return nil, &TokenError{
				Error:            InvalidGrant,
				ErrorDescription: fmt.Sprintf("verifier is invalid, challengeAnswer: [%s], token.CodeChallenge: [%s]", challengeAnswer, token.CodeChallenge),
			}, nil
		}
	}

	if application.ClientSecret != clientSecret {
		// when using PKCE, the Client Secret can be empty,
		// but if it is provided, it must be accurate.
		if token.CodeChallenge == "" {
			return nil, &TokenError{
				Error:            InvalidClient,
				ErrorDescription: fmt.Sprintf("client_secret is invalid for application: [%s], token.CodeChallenge: empty", application.GetId()),
			}, nil
		} else {
			if clientSecret != "" {
				return nil, &TokenError{
					Error:            InvalidClient,
					ErrorDescription: fmt.Sprintf("client_secret is invalid for application: [%s], token.CodeChallenge: [%s]", application.GetId(), token.CodeChallenge),
				}, nil
			}
		}
	}

	if application.Name != token.Application {
		return nil, &TokenError{
			Error:            InvalidGrant,
			ErrorDescription: fmt.Sprintf("the token is for wrong application (client_id), application.Name: [%s], token.Application: [%s]", application.Name, token.Application),
		}, nil
	}

	// RFC 8707: Validate resource parameter matches the one in the authorization request
	if resource != token.Resource {
		return nil, &TokenError{
			Error:            InvalidGrant,
			ErrorDescription: fmt.Sprintf("resource parameter does not match authorization request, expected: [%s], got: [%s]", token.Resource, resource),
		}, nil
	}

	nowUnix := time.Now().Unix()
	if nowUnix > token.CodeExpireIn {
		// code must be used within 5 minutes
		return nil, &TokenError{
			Error:            InvalidGrant,
			ErrorDescription: fmt.Sprintf("authorization code has expired, nowUnix: [%s], token.CodeExpireIn: [%s]", time.Unix(nowUnix, 0).Format(time.RFC3339), time.Unix(token.CodeExpireIn, 0).Format(time.RFC3339)),
		}, nil
	}
	return token, nil, nil
}

// GetPasswordToken handles the Resource Owner Password Credentials Grant flow.
func GetPasswordToken(application *Application, username string, password string, scope string, host string) (*Token, *TokenError, error) {
	expandedScope, ok := IsScopeValidAndExpand(scope, application)
	if !ok {
		return nil, &TokenError{
			Error:            InvalidScope,
			ErrorDescription: "the requested scope is invalid or not defined in the application",
		}, nil
	}
	scope = expandedScope

	user, err := GetUserByFieldsForSharedApp(application, application.Organization, username)
	if err != nil {
		return nil, nil, err
	}
	if user == nil {
		return nil, &TokenError{
			Error:            InvalidGrant,
			ErrorDescription: "the user does not exist",
		}, nil
	}

	if user.Ldap != "" {
		err = CheckLdapUserPassword(user, password, "en")
	} else {
		// For OAuth users who don't have a password set, they cannot use password grant type
		if user.Password == "" {
			return nil, &TokenError{
				Error:            InvalidGrant,
				ErrorDescription: "OAuth users cannot use password grant type, please use authorization code flow",
			}, nil
		}
		err = CheckPassword(user, password, "en")
	}
	if err != nil {
		return nil, &TokenError{
			Error:            InvalidGrant,
			ErrorDescription: fmt.Sprintf("invalid username or password: %s", err.Error()),
		}, nil
	}

	if user.IsForbidden {
		return nil, &TokenError{
			Error:            InvalidGrant,
			ErrorDescription: "the user is forbidden to sign in, please contact the administrator",
		}, nil
	}

	err = ExtendUserWithRolesAndPermissions(user)
	if err != nil {
		return nil, nil, err
	}

	accessToken, refreshToken, tokenName, err := generateJwtToken(application, user, "", "", "", scope, "", host)
	if err != nil {
		return nil, &TokenError{
			Error:            EndpointError,
			ErrorDescription: fmt.Sprintf("generate jwt token error: %s", err.Error()),
		}, nil
	}
	token := &Token{
		Owner:        application.Owner,
		Name:         tokenName,
		CreatedTime:  util.GetCurrentTime(),
		Application:  application.Name,
		Organization: user.Owner,
		User:         user.Name,
		Code:         util.GenerateClientId(),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int(application.ExpireInHours * float64(hourSeconds)),
		Scope:        scope,
		TokenType:    "Bearer",
		CodeIsUsed:   true,
	}
	_, err = AddToken(token)
	if err != nil {
		return nil, nil, err
	}

	return token, nil, nil
}

// GetClientCredentialsToken handles the Client Credentials Grant flow.
func GetClientCredentialsToken(application *Application, clientSecret string, scope string, host string) (*Token, *TokenError, error) {
	if application.ClientSecret != clientSecret {
		return nil, &TokenError{
			Error:            InvalidClient,
			ErrorDescription: "client_secret is invalid",
		}, nil
	}
	expandedScope, ok := IsScopeValidAndExpand(scope, application)
	if !ok {
		return nil, &TokenError{
			Error:            InvalidScope,
			ErrorDescription: "the requested scope is invalid or not defined in the application",
		}, nil
	}
	scope = expandedScope
	nullUser := &User{
		Owner: application.Owner,
		Id:    application.GetId(),
		Name:  application.Name,
		Type:  "application",
	}

	accessToken, _, tokenName, err := generateJwtToken(application, nullUser, "", "", "", scope, "", host)
	if err != nil {
		return nil, &TokenError{
			Error:            EndpointError,
			ErrorDescription: fmt.Sprintf("generate jwt token error: %s", err.Error()),
		}, nil
	}
	token := &Token{
		Owner:        application.Owner,
		Name:         tokenName,
		CreatedTime:  util.GetCurrentTime(),
		Application:  application.Name,
		Organization: application.Organization,
		User:         nullUser.Name,
		Code:         util.GenerateClientId(),
		AccessToken:  accessToken,
		ExpiresIn:    int(application.ExpireInHours * float64(hourSeconds)),
		Scope:        scope,
		TokenType:    "Bearer",
		CodeIsUsed:   true,
	}
	_, err = AddToken(token)
	if err != nil {
		return nil, nil, err
	}

	return token, nil, nil
}

// GetImplicitToken handles the Implicit Grant flow (requires password verification).
func GetImplicitToken(application *Application, username string, password string, scope string, nonce string, host string) (*Token, *TokenError, error) {
	user, err := GetUserByFieldsForSharedApp(application, application.Organization, username)
	if err != nil {
		return nil, nil, err
	}
	if user == nil {
		return nil, &TokenError{
			Error:            InvalidGrant,
			ErrorDescription: "the user does not exist",
		}, nil
	}

	if user.Ldap != "" {
		err = CheckLdapUserPassword(user, password, "en")
	} else {
		if user.Password == "" {
			return nil, &TokenError{
				Error:            InvalidGrant,
				ErrorDescription: "OAuth users cannot use implicit grant type, please use authorization code flow",
			}, nil
		}
		err = CheckPassword(user, password, "en")
	}
	if err != nil {
		return nil, &TokenError{
			Error:            InvalidGrant,
			ErrorDescription: fmt.Sprintf("invalid username or password: %s", err.Error()),
		}, nil
	}

	return mintImplicitToken(application, username, scope, nonce, host)
}

// GetJwtBearerToken handles the JWT Bearer Grant flow (RFC 7523).
func GetJwtBearerToken(application *Application, assertion string, scope string, nonce string, host string) (*Token, *TokenError, error) {
	ok, claims, err := ValidateJwtAssertion(assertion, application, host)
	if err != nil || !ok {
		if err != nil {
			return nil, &TokenError{
				Error:            InvalidGrant,
				ErrorDescription: err.Error(),
			}, err
		}

		return nil, &TokenError{
			Error:            InvalidGrant,
			ErrorDescription: fmt.Sprintf("assertion (JWT) is invalid for application: [%s]", application.GetId()),
		}, nil
	}

	// JWT assertion has already been validated above; skip password re-verification
	return mintImplicitToken(application, claims.Subject, scope, nonce, host)
}

// GetTokenByUser mints a token for the given user (Implicit flow helper).
func GetTokenByUser(application *Application, user *User, scope string, nonce string, host string) (*Token, error) {
	err := ExtendUserWithRolesAndPermissions(user)
	if err != nil {
		return nil, err
	}

	accessToken, refreshToken, tokenName, err := generateJwtToken(application, user, "", "", nonce, scope, "", host)
	if err != nil {
		return nil, err
	}

	token := &Token{
		Owner:        application.Owner,
		Name:         tokenName,
		CreatedTime:  util.GetCurrentTime(),
		Application:  application.Name,
		Organization: user.Owner,
		User:         user.Name,
		Code:         util.GenerateClientId(),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int(application.ExpireInHours * float64(hourSeconds)),
		Scope:        scope,
		TokenType:    "Bearer",
		CodeIsUsed:   true,
	}
	_, err = AddToken(token)
	if err != nil {
		return nil, err
	}

	return token, nil
}

// GetWechatMiniProgramToken handles the WeChat Mini Program flow.
func GetWechatMiniProgramToken(application *Application, code string, host string, username string, avatar string, lang string) (*Token, *TokenError, error) {
	mpProvider := GetWechatMiniProgramProvider(application)
	if mpProvider == nil {
		return nil, &TokenError{
			Error:            InvalidClient,
			ErrorDescription: "the application does not support wechat mini program",
		}, nil
	}
	provider, err := GetProvider(util.GetId("admin", mpProvider.Name))
	if err != nil {
		return nil, nil, err
	}

	mpIdp := idp.NewWeChatMiniProgramIdProvider(provider.ClientId, provider.ClientSecret)
	session, err := mpIdp.GetSessionByCode(code)
	if err != nil {
		return nil, &TokenError{
			Error:            InvalidGrant,
			ErrorDescription: fmt.Sprintf("get wechat mini program session error: %s", err.Error()),
		}, nil
	}

	openId, unionId := session.Openid, session.Unionid
	if openId == "" && unionId == "" {
		return nil, &TokenError{
			Error:            InvalidRequest,
			ErrorDescription: "the wechat mini program session is invalid",
		}, nil
	}
	user, err := getUserByWechatId(application.Organization, openId, unionId)
	if err != nil {
		return nil, nil, err
	}

	if user == nil {
		if !application.EnableSignUp {
			return nil, &TokenError{
				Error:            InvalidGrant,
				ErrorDescription: "the application does not allow to sign up new account",
			}, nil
		}
		// Add new user
		var name string
		if CheckUsername(username, lang) == "" {
			name = username
		} else {
			name = fmt.Sprintf("wechat-%s", openId)
		}

		// Generate a unique user ID within the confines of the application
		newUserId, idErr := GenerateIdForNewUser(application)
		if idErr != nil {
			// If we fail to generate a unique user ID, we can fallback to a random ID
			newUserId = util.GenerateId()
		}

		user = &User{
			Owner:             application.Organization,
			Id:                newUserId,
			Name:              name,
			Avatar:            avatar,
			SignupApplication: application.Name,
			WeChat:            openId,
			Type:              "normal-user",
			CreatedTime:       util.GetCurrentTime(),
			IsAdmin:           false,
			IsForbidden:       false,
			IsDeleted:         false,
			Properties: map[string]string{
				UserPropertiesWechatOpenId:  openId,
				UserPropertiesWechatUnionId: unionId,
			},
		}
		_, err = AddUser(user, "en")
		if err != nil {
			return nil, nil, err
		}
	}

	err = ExtendUserWithRolesAndPermissions(user)
	if err != nil {
		return nil, nil, err
	}

	accessToken, refreshToken, tokenName, err := generateJwtToken(application, user, "", "", "", "", "", host)
	if err != nil {
		return nil, &TokenError{
			Error:            EndpointError,
			ErrorDescription: fmt.Sprintf("generate jwt token error: %s", err.Error()),
		}, nil
	}

	token := &Token{
		Owner:        application.Owner,
		Name:         tokenName,
		CreatedTime:  util.GetCurrentTime(),
		Application:  application.Name,
		Organization: user.Owner,
		User:         user.Name,
		Code:         session.SessionKey, // a trick, because miniprogram does not use the code, so use the code field to save the session_key
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int(application.ExpireInHours * float64(hourSeconds)),
		Scope:        "",
		TokenType:    "Bearer",
		CodeIsUsed:   true,
	}
	_, err = AddToken(token)
	if err != nil {
		return nil, nil, err
	}
	return token, nil, nil
}

// GetTokenExchangeToken handles the Token Exchange Grant flow (RFC 8693).
// Exchanges a subject token for a new token with different audience or scope.
func GetTokenExchangeToken(application *Application, clientSecret string, subjectToken string, subjectTokenType string, audience string, scope string, host string) (*Token, *TokenError, error) {
	if application.ClientSecret != clientSecret {
		return nil, &TokenError{
			Error:            InvalidClient,
			ErrorDescription: "client_secret is invalid",
		}, nil
	}

	if subjectToken == "" {
		return nil, &TokenError{
			Error:            InvalidRequest,
			ErrorDescription: "subject_token is required",
		}, nil
	}

	// RFC 8693 defines standard token type identifiers
	if subjectTokenType == "" {
		subjectTokenType = "urn:ietf:params:oauth:token-type:access_token" // Default to access_token
	}

	supportedTokenTypes := []string{
		"urn:ietf:params:oauth:token-type:access_token",
		"urn:ietf:params:oauth:token-type:jwt",
		"urn:ietf:params:oauth:token-type:id_token",
	}

	isValidTokenType := false
	for _, tokenType := range supportedTokenTypes {
		if subjectTokenType == tokenType {
			isValidTokenType = true
			break
		}
	}

	if !isValidTokenType {
		return nil, &TokenError{
			Error:            InvalidRequest,
			ErrorDescription: fmt.Sprintf("unsupported subject_token_type: %s", subjectTokenType),
		}, nil
	}

	subjectOwner, subjectName, subjectScope, tokenError, err := parseAndValidateSubjectToken(subjectToken, application.ClientId)
	if err != nil {
		return nil, nil, err
	}
	if tokenError != nil {
		return nil, tokenError, nil
	}

	user, err := getUser(subjectOwner, subjectName)
	if err != nil {
		return nil, nil, err
	}
	if user == nil {
		return nil, &TokenError{
			Error:            InvalidGrant,
			ErrorDescription: fmt.Sprintf("user from subject_token does not exist: %s", util.GetId(subjectOwner, subjectName)),
		}, nil
	}

	if user.IsForbidden {
		return nil, &TokenError{
			Error:            InvalidGrant,
			ErrorDescription: "the user is forbidden to sign in, please contact the administrator",
		}, nil
	}

	// If scope is not provided, use the scope from the subject token.
	// If scope is provided, it should be a subset of the subject token's scope (downscoping).
	if scope == "" {
		scope = subjectScope
	} else {
		if subjectScope != "" {
			subjectScopes := strings.Split(subjectScope, " ")
			requestedScopes := strings.Split(scope, " ")
			for _, requestedScope := range requestedScopes {
				if requestedScope == "" {
					continue
				}
				found := false
				for _, existingScope := range subjectScopes {
					if existingScope != "" && requestedScope == existingScope {
						found = true
						break
					}
				}
				if !found {
					return nil, &TokenError{
						Error:            InvalidScope,
						ErrorDescription: fmt.Sprintf("requested scope '%s' is not in subject token's scope", requestedScope),
					}, nil
				}
			}
		}
	}

	err = ExtendUserWithRolesAndPermissions(user)
	if err != nil {
		return nil, nil, err
	}

	accessToken, refreshToken, tokenName, err := generateJwtToken(application, user, "", "", "", scope, "", host)
	if err != nil {
		return nil, &TokenError{
			Error:            EndpointError,
			ErrorDescription: fmt.Sprintf("generate jwt token error: %s", err.Error()),
		}, nil
	}

	token := &Token{
		Owner:        application.Owner,
		Name:         tokenName,
		CreatedTime:  util.GetCurrentTime(),
		Application:  application.Name,
		Organization: user.Owner,
		User:         user.Name,
		Code:         util.GenerateClientId(),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int(application.ExpireInHours * float64(hourSeconds)),
		Scope:        scope,
		TokenType:    "Bearer",
		CodeIsUsed:   true,
	}

	_, err = AddToken(token)
	if err != nil {
		return nil, nil, err
	}

	return token, nil, nil
}

func GetAccessTokenByUser(user *User, host string) (string, error) {
	application, err := GetApplicationByUser(user)
	if err != nil {
		return "", err
	}
	if application == nil {
		return "", fmt.Errorf("the application for user %s is not found", user.Id)
	}

	token, err := GetTokenByUser(application, user, "profile", "", host)
	if err != nil {
		return "", err
	}

	return token.AccessToken, nil
}
