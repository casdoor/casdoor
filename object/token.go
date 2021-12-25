// Copyright 2021 The casbin Authors. All Rights Reserved.
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

	"github.com/casbin/casdoor/util"
	"xorm.io/core"
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

	Code         string `xorm:"varchar(100)" json:"code"`
	AccessToken  string `xorm:"mediumtext" json:"accessToken"`
	RefreshToken string `xorm:"mediumtext" json:"refreshToken"`
	ExpiresIn    int    `json:"expiresIn"`
	Scope        string `xorm:"varchar(100)" json:"scope"`
	TokenType    string `xorm:"varchar(100)" json:"tokenType"`
}

type TokenWrapper struct {
	AccessToken string `json:"access_token"`
	IdToken     string `json:"id_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
}

func GetTokenCount(owner, field, value string) int {
	session := adapter.Engine.Where("owner=?", owner)
	if field != "" && value != "" {
		session = session.And(fmt.Sprintf("%s like ?", util.SnakeString(field)), fmt.Sprintf("%%%s%%", value))
	}
	count, err := session.Count(&Token{})
	if err != nil {
		panic(err)
	}

	return int(count)
}

func GetTokens(owner string) []*Token {
	tokens := []*Token{}
	err := adapter.Engine.Desc("created_time").Find(&tokens, &Token{Owner: owner})
	if err != nil {
		panic(err)
	}

	return tokens
}

func GetPaginationTokens(owner string, offset, limit int, field, value, sortField, sortOrder string) []*Token {
	tokens := []*Token{}
	session := GetSession(owner, offset, limit, field, value, sortField, sortOrder)
	err := session.Find(&tokens)
	if err != nil {
		panic(err)
	}

	return tokens
}

func getToken(owner string, name string) *Token {
	if owner == "" || name == "" {
		return nil
	}

	token := Token{Owner: owner, Name: name}
	existed, err := adapter.Engine.Get(&token)
	if err != nil {
		panic(err)
	}

	if existed {
		return &token
	}

	return nil
}

func getTokenByCode(code string) *Token {
	token := Token{}
	existed, err := adapter.Engine.Where("code=?", code).Get(&token)
	if err != nil {
		panic(err)
	}

	if existed {
		return &token
	}

	return nil
}

func GetToken(id string) *Token {
	owner, name := util.GetOwnerAndNameFromId(id)
	return getToken(owner, name)
}

func UpdateToken(id string, token *Token) bool {
	owner, name := util.GetOwnerAndNameFromId(id)
	if getToken(owner, name) == nil {
		return false
	}

	affected, err := adapter.Engine.ID(core.PK{owner, name}).AllCols().Update(token)
	if err != nil {
		panic(err)
	}

	return affected != 0
}

func AddToken(token *Token) bool {
	affected, err := adapter.Engine.Insert(token)
	if err != nil {
		panic(err)
	}

	return affected != 0
}

func DeleteToken(token *Token) bool {
	affected, err := adapter.Engine.ID(core.PK{token.Owner, token.Name}).Delete(&Token{})
	if err != nil {
		panic(err)
	}

	return affected != 0
}

func CheckOAuthLogin(clientId string, responseType string, redirectUri string, scope string, state string) (string, *Application) {
	if responseType != "code" {
		return "response_type should be \"code\"", nil
	}

	application := GetApplicationByClientId(clientId)
	if application == nil {
		return "Invalid client_id", nil
	}

	validUri := false
	for _, tmpUri := range application.RedirectUris {
		if strings.Contains(redirectUri, tmpUri) {
			validUri = true
			break
		}
	}
	if !validUri {
		return fmt.Sprintf("Redirect URI: \"%s\" doesn't exist in the allowed Redirect URI list", redirectUri), application
	}

	// Mask application for /api/get-app-login
	application.ClientSecret = ""
	return "", application
}

func GetOAuthCode(userId string, clientId string, responseType string, redirectUri string, scope string, state string, nonce string) *Code {
	user := GetUser(userId)
	if user == nil {
		return &Code{
			Message: fmt.Sprintf("The user: %s doesn't exist", userId),
			Code:    "",
		}
	}

	msg, application := CheckOAuthLogin(clientId, responseType, redirectUri, scope, state)
	if msg != "" {
		return &Code{
			Message: msg,
			Code:    "",
		}
	}

	accessToken, refreshToken, err := generateJwtToken(application, user, nonce)
	if err != nil {
		panic(err)
	}

	token := &Token{
		Owner:        application.Owner,
		Name:         util.GenerateId(),
		CreatedTime:  util.GetCurrentTime(),
		Application:  application.Name,
		Organization: user.Owner,
		User:         user.Name,
		Code:         util.GenerateClientId(),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    application.ExpireInHours * 60,
		Scope:        scope,
		TokenType:    "Bearer",
	}
	AddToken(token)

	return &Code{
		Message: "",
		Code:    token.Code,
	}
}

func GetOAuthToken(grantType string, clientId string, clientSecret string, code string) *TokenWrapper {
	application := GetApplicationByClientId(clientId)
	if application == nil {
		return &TokenWrapper{
			AccessToken: "error: invalid client_id",
			TokenType:   "",
			ExpiresIn:   0,
			Scope:       "",
		}
	}

	if grantType != "authorization_code" {
		return &TokenWrapper{
			AccessToken: "error: grant_type should be \"authorization_code\"",
			TokenType:   "",
			ExpiresIn:   0,
			Scope:       "",
		}
	}

	if code == "" {
		return &TokenWrapper{
			AccessToken: "error: code should not be empty",
			TokenType:   "",
			ExpiresIn:   0,
			Scope:       "",
		}
	}

	token := getTokenByCode(code)
	if token == nil {
		return &TokenWrapper{
			AccessToken: "error: invalid code",
			TokenType:   "",
			ExpiresIn:   0,
			Scope:       "",
		}
	}

	if application.Name != token.Application {
		return &TokenWrapper{
			AccessToken: "error: the token is for wrong application (client_id)",
			TokenType:   "",
			ExpiresIn:   0,
			Scope:       "",
		}
	}

	if application.ClientSecret != clientSecret {
		return &TokenWrapper{
			AccessToken: "error: invalid client_secret",
			TokenType:   "",
			ExpiresIn:   0,
			Scope:       "",
		}
	}

	tokenWrapper := &TokenWrapper{
		AccessToken: token.AccessToken,
		IdToken:     token.AccessToken,
		TokenType:   token.TokenType,
		ExpiresIn:   token.ExpiresIn,
		Scope:       token.Scope,
	}

	return tokenWrapper
}

func RefreshToken(grantType string, refreshToken string, scope string, clientId string, clientSecret string) *Code {
	// check parameters
	if grantType != "refresh_token" {
		return &Code{
			Message: "error: grant_type should be \"refresh_token\"",
			Code:    "",
		}
	}
	application := GetApplicationByClientId(clientId)
	if application == nil {
		return &Code{
			Message: "error: invalid client_id",
			Code:    "",
		}
	}
	if application.ClientSecret != clientSecret {
		return &Code{
			Message: "error: invalid client_secret",
			Code:    "",
		}
	}
	// check whether the refresh token is valid, and has not expired.
	token := Token{RefreshToken: refreshToken}
	existed, err := adapter.Engine.Get(&token)
	if err != nil || !existed {
		return &Code{
			Message: "error: invalid refresh_token",
			Code:    "",
		}
	}
	claims, err := ParseJwtToken(refreshToken)
	if err != nil {
		return &Code{
			Message: "error: invalid refresh_token",
			Code:    "",
		}
	}
	if time.Now().Unix() > claims.ExpiresAt.Unix() {
		return &Code{
			Message: "error: expired refresh_token",
			Code:    "",
		}
	}
	// generate a new token
	user := getUser(application.Owner, token.User)
	newAccessToken, newRefreshToken, err := generateJwtToken(application, user, "")
	if err != nil {
		panic(err)
	}

	newToken := &Token{
		Owner:        application.Owner,
		Name:         util.GenerateId(),
		CreatedTime:  util.GetCurrentTime(),
		Application:  application.Name,
		Organization: user.Owner,
		User:         user.Name,
		Code:         util.GenerateClientId(),
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    application.ExpireInHours * 60,
		Scope:        scope,
		TokenType:    "Bearer",
	}
	AddToken(newToken)

	return &Code{
		Message: "",
		Code:    token.Code,
	}
}
