// Copyright 2021 The Casdoor Authors. All Rights Reserved.
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
	"time"

	"github.com/casdoor/casdoor/i18n"
	"github.com/casdoor/casdoor/idp"
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

type Code struct {
	Message string `xorm:"varchar(100)" json:"message"`
	Code    string `xorm:"varchar(100)" json:"code"`
}

type Token struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`

	Application  string `xorm:"varchar(100)" json:"application"`
	Organization string `xorm:"varchar(100)" json:"organization"`
	User         string `xorm:"varchar(100)" json:"user"`

	Code          string `xorm:"varchar(100) index" json:"code"`
	AccessToken   string `xorm:"mediumtext" json:"accessToken"`
	RefreshToken  string `xorm:"mediumtext" json:"refreshToken"`
	ExpiresIn     int    `json:"expiresIn"`
	Scope         string `xorm:"varchar(100)" json:"scope"`
	TokenType     string `xorm:"varchar(100)" json:"tokenType"`
	CodeChallenge string `xorm:"varchar(100)" json:"codeChallenge"`
	CodeIsUsed    bool   `json:"codeIsUsed"`
	CodeExpireIn  int64  `json:"codeExpireIn"`
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

type IntrospectionResponse struct {
	Active    bool     `json:"active"`
	Scope     string   `json:"scope,omitempty"`
	ClientId  string   `json:"client_id,omitempty"`
	Username  string   `json:"username,omitempty"`
	TokenType string   `json:"token_type,omitempty"`
	Exp       int64    `json:"exp,omitempty"`
	Iat       int64    `json:"iat,omitempty"`
	Nbf       int64    `json:"nbf,omitempty"`
	Sub       string   `json:"sub,omitempty"`
	Aud       []string `json:"aud,omitempty"`
	Iss       string   `json:"iss,omitempty"`
	Jti       string   `json:"jti,omitempty"`
}

func GetTokenCount(owner, organization, field, value string) (int64, error) {
	session := GetSession(owner, -1, -1, field, value, "", "")
	return session.Count(&Token{Organization: organization})
}

func GetTokens(owner string, organization string) ([]*Token, error) {
	tokens := []*Token{}
	err := adapter.Engine.Desc("created_time").Find(&tokens, &Token{Owner: owner, Organization: organization})
	return tokens, err
}

func GetPaginationTokens(owner, organization string, offset, limit int, field, value, sortField, sortOrder string) ([]*Token, error) {
	tokens := []*Token{}
	session := GetSession(owner, offset, limit, field, value, sortField, sortOrder)
	err := session.Find(&tokens, &Token{Organization: organization})
	return tokens, err
}

func getToken(owner string, name string) (*Token, error) {
	if owner == "" || name == "" {
		return nil, nil
	}

	token := Token{Owner: owner, Name: name}
	existed, err := adapter.Engine.Get(&token)
	if err != nil {
		return nil, err
	}

	if existed {
		return &token, nil
	}

	return nil, nil
}

func getTokenByCode(code string) (*Token, error) {
	token := Token{Code: code}
	existed, err := adapter.Engine.Get(&token)
	if err != nil {
		return nil, err
	}

	if existed {
		return &token, nil
	}

	return nil, nil
}

func updateUsedByCode(token *Token) bool {
	affected, err := adapter.Engine.Where("code=?", token.Code).Cols("code_is_used").Update(token)
	if err != nil {
		panic(err)
	}

	return affected != 0
}

func GetToken(id string) (*Token, error) {
	owner, name := util.GetOwnerAndNameFromId(id)
	return getToken(owner, name)
}

func (token *Token) GetId() string {
	return fmt.Sprintf("%s/%s", token.Owner, token.Name)
}

func UpdateToken(id string, token *Token) (bool, error) {
	owner, name := util.GetOwnerAndNameFromId(id)
	if t, err := getToken(owner, name); err != nil {
		return false, err
	} else if t == nil {
		return false, nil
	}

	affected, err := adapter.Engine.ID(core.PK{owner, name}).AllCols().Update(token)
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func AddToken(token *Token) (bool, error) {
	affected, err := adapter.Engine.Insert(token)
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func DeleteToken(token *Token) (bool, error) {
	affected, err := adapter.Engine.ID(core.PK{token.Owner, token.Name}).Delete(&Token{})
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func ExpireTokenByAccessToken(accessToken string) (bool, *Application, *Token, error) {
	token := Token{AccessToken: accessToken}
	existed, err := adapter.Engine.Get(&token)
	if err != nil {
		return false, nil, nil, err
	}

	if !existed {
		return false, nil, nil, nil
	}

	token.ExpiresIn = 0
	affected, err := adapter.Engine.ID(core.PK{token.Owner, token.Name}).Cols("expires_in").Update(&token)
	if err != nil {
		return false, nil, nil, err
	}

	application, err := getApplication(token.Owner, token.Application)
	if err != nil {
		return false, nil, nil, err
	}

	return affected != 0, application, &token, nil
}

func GetTokenByAccessToken(accessToken string) (*Token, error) {
	// Check if the accessToken is in the database
	token := Token{AccessToken: accessToken}
	existed, err := adapter.Engine.Get(&token)
	if err != nil {
		return nil, err
	}

	if !existed {
		return nil, nil
	}

	return &token, nil
}

func GetTokenByTokenAndApplication(token string, application string) (*Token, error) {
	tokenResult := Token{}
	existed, err := adapter.Engine.Where("(refresh_token = ? or access_token = ? ) and application = ?", token, token, application).Get(&tokenResult)
	if err != nil {
		return nil, err
	}

	if !existed {
		return nil, nil
	}

	return &tokenResult, nil
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

	// Mask application for /api/get-app-login
	application.ClientSecret = ""
	return "", application, nil
}

func GetOAuthCode(userId string, clientId string, responseType string, redirectUri string, scope string, state string, nonce string, challenge string, host string, lang string) (*Code, error) {
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

	err = ExtendUserWithRolesAndPermissions(user)
	if err != nil {
		return nil, err
	}
	accessToken, refreshToken, tokenName, err := generateJwtToken(application, user, nonce, scope, host)
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
		ExpiresIn:     application.ExpireInHours * hourSeconds,
		Scope:         scope,
		TokenType:     "Bearer",
		CodeChallenge: challenge,
		CodeIsUsed:    false,
		CodeExpireIn:  time.Now().Add(time.Minute * 5).Unix(),
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

func GetOAuthToken(grantType string, clientId string, clientSecret string, code string, verifier string, scope string, username string, password string, host string, refreshToken string, tag string, avatar string, lang string) (interface{}, error) {
	application, err := GetApplicationByClientId(clientId)
	if err != nil {
		return nil, err
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
		token, tokenError, err = GetAuthorizationCodeToken(application, clientSecret, code, verifier)
	case "password": //	Resource Owner Password Credentials Grant
		token, tokenError, err = GetPasswordToken(application, username, password, scope, host)
	case "client_credentials": // Client Credentials Grant
		token, tokenError, err = GetClientCredentialsToken(application, clientSecret, scope, host)
	case "refresh_token":
		refreshToken2, err := RefreshToken(grantType, refreshToken, scope, clientId, clientSecret, host)
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

	token.CodeIsUsed = true

	go updateUsedByCode(token)

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

func RefreshToken(grantType string, refreshToken string, scope string, clientId string, clientSecret string, host string) (interface{}, error) {
	// check parameters
	if grantType != "refresh_token" {
		return &TokenError{
			Error:            UnsupportedGrantType,
			ErrorDescription: "grant_type should be refresh_token",
		}, nil
	}
	application, err := GetApplicationByClientId(clientId)
	if err != nil {
		return nil, err
	}

	if application == nil {
		return &TokenError{
			Error:            InvalidClient,
			ErrorDescription: "client_id is invalid",
		}, nil
	}
	if clientSecret != "" && application.ClientSecret != clientSecret {
		return &TokenError{
			Error:            InvalidClient,
			ErrorDescription: "client_secret is invalid",
		}, nil
	}
	// check whether the refresh token is valid, and has not expired.
	token := Token{RefreshToken: refreshToken}
	existed, err := adapter.Engine.Get(&token)
	if err != nil || !existed {
		return &TokenError{
			Error:            InvalidGrant,
			ErrorDescription: "refresh token is invalid, expired or revoked",
		}, nil
	}

	cert, err := getCertByApplication(application)
	if err != nil {
		return nil, err
	}

	_, err = ParseJwtToken(refreshToken, cert)
	if err != nil {
		return &TokenError{
			Error:            InvalidGrant,
			ErrorDescription: fmt.Sprintf("parse refresh token error: %s", err.Error()),
		}, nil
	}
	// generate a new token
	user, err := getUser(application.Organization, token.User)
	if err != nil {
		return nil, err
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
	newAccessToken, newRefreshToken, tokenName, err := generateJwtToken(application, user, "", scope, host)
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
		ExpiresIn:    application.ExpireInHours * hourSeconds,
		Scope:        scope,
		TokenType:    "Bearer",
	}
	_, err = AddToken(newToken)
	if err != nil {
		return nil, err
	}

	_, err = DeleteToken(&token)
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

// PkceChallenge: base64-URL-encoded SHA256 hash of verifier, per rfc 7636
func pkceChallenge(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	challenge := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(sum[:])
	return challenge
}

// IsGrantTypeValid
// Check if grantType is allowed in the current application
// authorization_code is allowed by default
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

// GetAuthorizationCodeToken
// Authorization code flow
func GetAuthorizationCodeToken(application *Application, clientSecret string, code string, verifier string) (*Token, *TokenError, error) {
	if code == "" {
		return nil, &TokenError{
			Error:            InvalidRequest,
			ErrorDescription: "authorization code should not be empty",
		}, nil
	}

	token, err := getTokenByCode(code)
	if err != nil {
		return nil, nil, err
	}

	if token == nil {
		return nil, &TokenError{
			Error:            InvalidGrant,
			ErrorDescription: "authorization code is invalid",
		}, nil
	}
	if token.CodeIsUsed {
		// anti replay attacks
		return nil, &TokenError{
			Error:            InvalidGrant,
			ErrorDescription: "authorization code has been used",
		}, nil
	}

	if token.CodeChallenge != "" && pkceChallenge(verifier) != token.CodeChallenge {
		return nil, &TokenError{
			Error:            InvalidGrant,
			ErrorDescription: "verifier is invalid",
		}, nil
	}

	if application.ClientSecret != clientSecret {
		// when using PKCE, the Client Secret can be empty,
		// but if it is provided, it must be accurate.
		if token.CodeChallenge == "" {
			return nil, &TokenError{
				Error:            InvalidClient,
				ErrorDescription: "client_secret is invalid",
			}, nil
		} else {
			if clientSecret != "" {
				return nil, &TokenError{
					Error:            InvalidClient,
					ErrorDescription: "client_secret is invalid",
				}, nil
			}
		}
	}

	if application.Name != token.Application {
		return nil, &TokenError{
			Error:            InvalidGrant,
			ErrorDescription: "the token is for wrong application (client_id)",
		}, nil
	}

	if time.Now().Unix() > token.CodeExpireIn {
		// code must be used within 5 minutes
		return nil, &TokenError{
			Error:            InvalidGrant,
			ErrorDescription: "authorization code has expired",
		}, nil
	}
	return token, nil, nil
}

// GetPasswordToken
// Resource Owner Password Credentials flow
func GetPasswordToken(application *Application, username string, password string, scope string, host string) (*Token, *TokenError, error) {
	user, err := getUser(application.Organization, username)
	if err != nil {
		return nil, nil, err
	}

	if user == nil {
		return nil, &TokenError{
			Error:            InvalidGrant,
			ErrorDescription: "the user does not exist",
		}, nil
	}
	msg := CheckPassword(user, password, "en")
	if msg != "" {
		return nil, &TokenError{
			Error:            InvalidGrant,
			ErrorDescription: "invalid username or password",
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

	accessToken, refreshToken, tokenName, err := generateJwtToken(application, user, "", scope, host)
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
		ExpiresIn:    application.ExpireInHours * hourSeconds,
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

// GetClientCredentialsToken
// Client Credentials flow
func GetClientCredentialsToken(application *Application, clientSecret string, scope string, host string) (*Token, *TokenError, error) {
	if application.ClientSecret != clientSecret {
		return nil, &TokenError{
			Error:            InvalidClient,
			ErrorDescription: "client_secret is invalid",
		}, nil
	}
	nullUser := &User{
		Owner: application.Owner,
		Id:    application.GetId(),
		Name:  application.Name,
		Type:  "application",
	}

	accessToken, _, tokenName, err := generateJwtToken(application, nullUser, "", scope, host)
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
		ExpiresIn:    application.ExpireInHours * hourSeconds,
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

// GetTokenByUser
// Implicit flow
func GetTokenByUser(application *Application, user *User, scope string, host string) (*Token, error) {
	err := ExtendUserWithRolesAndPermissions(user)
	if err != nil {
		return nil, err
	}

	accessToken, refreshToken, tokenName, err := generateJwtToken(application, user, "", scope, host)
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
		ExpiresIn:    application.ExpireInHours * hourSeconds,
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

// GetWechatMiniProgramToken
// Wechat Mini Program flow
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
	user, err := getUserByWechatId(openId, unionId)
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

		user = &User{
			Owner:             application.Organization,
			Id:                util.GenerateId(),
			Name:              name,
			Avatar:            avatar,
			SignupApplication: application.Name,
			WeChat:            openId,
			Type:              "normal-user",
			CreatedTime:       util.GetCurrentTime(),
			IsAdmin:           false,
			IsGlobalAdmin:     false,
			IsForbidden:       false,
			IsDeleted:         false,
			Properties: map[string]string{
				UserPropertiesWechatOpenId:  openId,
				UserPropertiesWechatUnionId: unionId,
			},
		}
		_, err = AddUser(user)
		if err != nil {
			return nil, nil, err
		}
	}

	err = ExtendUserWithRolesAndPermissions(user)
	if err != nil {
		return nil, nil, err
	}

	accessToken, refreshToken, tokenName, err := generateJwtToken(application, user, "", "", host)
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
		ExpiresIn:    application.ExpireInHours * 60,
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
