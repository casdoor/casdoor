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
	"strings"

	"github.com/casdoor/casdoor/util"
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

	Application string `xorm:"varchar(100)" json:"application"`

	Code        string `xorm:"varchar(100)" json:"code"`
	AccessToken string `xorm:"varchar(100)" json:"accessToken"`
	ExpiresIn   int    `json:"expiresIn"`
	Scope       string `xorm:"varchar(100)" json:"scope"`
	TokenType   string `xorm:"varchar(100)" json:"tokenType"`
}

type TokenWrapper struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
}

func GetTokens(owner string) []*Token {
	tokens := []*Token{}
	err := adapter.engine.Desc("created_time").Find(&tokens, &Token{Owner: owner})
	if err != nil {
		panic(err)
	}

	return tokens
}

func getToken(owner string, name string) *Token {
	token := Token{Owner: owner, Name: name}
	existed, err := adapter.engine.Get(&token)
	if err != nil {
		panic(err)
	}

	if existed {
		return &token
	} else {
		return nil
	}
}

func getTokenByCode(code string) *Token {
	token := Token{}
	existed, err := adapter.engine.Where("code=?", code).Get(&token)
	if err != nil {
		panic(err)
	}

	if existed {
		return &token
	} else {
		return nil
	}
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

	_, err := adapter.engine.ID(core.PK{owner, name}).AllCols().Update(token)
	if err != nil {
		panic(err)
	}

	//return affected != 0
	return true
}

func AddToken(token *Token) bool {
	affected, err := adapter.engine.Insert(token)
	if err != nil {
		panic(err)
	}

	return affected != 0
}

func DeleteToken(token *Token) bool {
	affected, err := adapter.engine.ID(core.PK{token.Owner, token.Name}).Delete(&Token{})
	if err != nil {
		panic(err)
	}

	return affected != 0
}

func GetOAuthCode(clientId string, responseType string, redirectUri string, scope string, state string) *Code {
	application := getApplicationByClientId(clientId)
	if application == nil {
		return &Code{
			Message: "invalid client_id",
			Code:    "",
		}
	}

	if responseType != "code" {
		return &Code{
			Message: "response_type should be \"code\"",
			Code:    "",
		}
	}

	validUri := false
	for _, url := range application.RedirectUrls {
		if strings.Contains(redirectUri, url) {
			validUri = true
			break
		}
	}
	if !validUri {
		return &Code{
			Message: "redirect_uri doesn't exist in the allowed Redirect URL list",
			Code:    "",
		}
	}

	token := &Token{
		Owner:       application.Owner,
		Name:        util.GenerateId(),
		CreatedTime: util.GetCurrentTime(),
		Application: application.Name,
		Code:        util.GenerateClientId(),
		AccessToken: "",
		ExpiresIn:   7200,
		Scope:       scope,
		TokenType:   "Bearer",
	}
	AddToken(token)

	return &Code{
		Message: "",
		Code:    token.Code,
	}
}

func GetOAuthToken(grantType string, clientId string, clientSecret string, code string) *TokenWrapper {
	application := getApplicationByClientId(clientId)
	if application == nil {
		return &TokenWrapper{
			AccessToken: "invalid client_id",
			TokenType:   "",
			ExpiresIn:   0,
			Scope:       "",
		}
	}

	if grantType != "authorization_code" {
		return &TokenWrapper{
			AccessToken: "grant_type should be \"authorization_code\"",
			TokenType:   "",
			ExpiresIn:   0,
			Scope:       "",
		}
	}

	token := getTokenByCode(code)
	if token == nil {
		return &TokenWrapper{
			AccessToken: "invalid code",
			TokenType:   "",
			ExpiresIn:   0,
			Scope:       "",
		}
	}

	if application.Name != token.Application {
		return &TokenWrapper{
			AccessToken: "token is for wrong application (client_id)",
			TokenType:   "",
			ExpiresIn:   0,
			Scope:       "",
		}
	}

	if application.ClientSecret != clientSecret {
		return &TokenWrapper{
			AccessToken: "invalid client_secret",
			TokenType:   "",
			ExpiresIn:   0,
			Scope:       "",
		}
	}

	tokenWrapper := &TokenWrapper{
		AccessToken: token.AccessToken,
		TokenType:   token.TokenType,
		ExpiresIn:   token.ExpiresIn,
		Scope:       token.Scope,
	}

	return tokenWrapper
}
