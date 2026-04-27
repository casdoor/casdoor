// Copyright 2026 The Casdoor Authors. All Rights Reserved.
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

package controllers

import (
	"encoding/json"
	"fmt"

	"github.com/casdoor/casdoor/form"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

// NativeSsoComplete
// @Title NativeSsoComplete
// @Tag Native SSO
// @Description Complete OAuth authorization after Native SSO token exchange
// @router /native-sso-complete [post]
func (c *ApiController) NativeSsoComplete() {
	var request struct {
		AccessToken string `json:"accessToken"`
	}
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &request); err != nil {
		c.ResponseError(err.Error())
		return
	}

	token, err := object.GetTokenByAccessToken(request.AccessToken)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	if token == nil {
		c.ResponseError(c.T("token:Token not found, invalid accessToken"))
		return
	}
	if expired, _ := util.IsTokenExpired(token.CreatedTime, token.ExpiresIn); expired {
		c.ResponseError(c.T("token:Token expired"))
		return
	}
	if token.GrantType != "urn:ietf:params:oauth:grant-type:token-exchange" {
		c.ResponseError(c.T("auth:The access token was not issued by native SSO"))
		return
	}

	application, err := object.GetApplicationByClientId(c.Ctx.Input.Query("clientId"))
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	if application == nil || application.Owner != token.Owner || application.Name != token.Application {
		c.ResponseError(c.T("auth:The application does not match the native SSO token"))
		return
	}

	user, err := object.GetUser(util.GetId(token.Organization, token.User))
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	if user == nil {
		c.ResponseError(fmt.Sprintf(c.T("general:The user: %s doesn't exist"), util.GetId(token.Organization, token.User)))
		return
	}

	responseType := c.Ctx.Input.Query("responseType")
	if responseType == "" {
		responseType = ResponseTypeLogin
	}
	authForm := form.AuthForm{
		Type:         responseType,
		SigninMethod: "Native SSO",
	}
	resp := c.HandleLoggedIn(application, user, &authForm)

	c.Ctx.Input.SetParam("recordUserId", user.GetId())
	c.Data["json"] = resp
	c.ServeJSON()
}
