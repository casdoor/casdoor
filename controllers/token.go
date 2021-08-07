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

package controllers

import (
	"encoding/json"

	"github.com/casbin/casdoor/object"
)

// GetTokens
// @Title GetTokens
// @Description get tokens
// @Param   owner     query    string  true        "The owner of tokens"
// @Success 200 {array} object.Token The Response object
// @router /get-tokens [get]
func (c *ApiController) GetTokens() {
	owner := c.Input().Get("owner")

	c.Data["json"] = object.GetTokens(owner)
	c.ServeJSON()
}

// GetToken
// @Title GetToken
// @Description get token
// @Param   id     query    string  true        "The id of token"
// @Success 200 {object} object.Token The Response object
// @router /get-token [get]
func (c *ApiController) GetToken() {
	id := c.Input().Get("id")

	c.Data["json"] = object.GetToken(id)
	c.ServeJSON()
}

// UpdateToken
// @Title UpdateToken
// @Description update token
// @Param   id     query    string  true        "The id of token"
// @Param   body    body   object.Token  true        "Details of the token"
// @Success 200 {object} controllers.Response The Response object
// @router /update-token [post]
func (c *ApiController) UpdateToken() {
	id := c.Input().Get("id")

	var token object.Token
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &token)
	if err != nil {
		panic(err)
	}

	c.Data["json"] = wrapActionResponse(object.UpdateToken(id, &token))
	c.ServeJSON()
}

// AddToken
// @Title AddToken
// @Description add token
// @Param   body    body   object.Token  true        "Details of the token"
// @Success 200 {object} controllers.Response The Response object
// @router /add-token [post]
func (c *ApiController) AddToken() {
	var token object.Token
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &token)
	if err != nil {
		panic(err)
	}

	c.Data["json"] = wrapActionResponse(object.AddToken(&token))
	c.ServeJSON()
}

// DeleteToken
// @Title DeleteToken
// @Description delete token
// @Param   body    body   object.Token  true        "Details of the token"
// @Success 200 {object} controllers.Response The Response object
// @router /delete-token [post]
func (c *ApiController) DeleteToken() {
	var token object.Token
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &token)
	if err != nil {
		panic(err)
	}

	c.Data["json"] = wrapActionResponse(object.DeleteToken(&token))
	c.ServeJSON()
}

// GetOAuthToken
// @Title GetOAuthToken
// @Description get oAuth token
// @Param   grant_type     query    string  true        "oAuth grant type"
// @Param   client_id     query    string  true        "oAuth client id"
// @Param   client_secret     query    string  true        "oAuth client secret"
// @Param   code     query    string  true        "oAuth code"
// @Success 200 {object} object.TokenWrapper The Response object
// @router /login/oauth/access_token [post]
func (c *ApiController) GetOAuthToken() {
	grantType := c.Input().Get("grant_type")
	clientId := c.Input().Get("client_id")
	clientSecret := c.Input().Get("client_secret")
	code := c.Input().Get("code")

	c.Data["json"] = object.GetOAuthToken(grantType, clientId, clientSecret, code)
	c.ServeJSON()
}
