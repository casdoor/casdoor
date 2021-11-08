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

	"github.com/astaxie/beego/utils/pagination"
	"github.com/casbin/casdoor/object"
	"github.com/casbin/casdoor/util"
)

// GetTokens
// @Title GetTokens
// @Description get tokens
// @Param   owner     query    string  true        "The owner of tokens"
// @Param   pageSize     query    string  true        "The size of each page"
// @Param   p     query    string  true        "The number of the page"
// @Success 200 {array} object.Token The Response object
// @router /get-tokens [get]
func (c *ApiController) GetTokens() {
	owner := c.Input().Get("owner")
	limit := c.Input().Get("pageSize")
	page := c.Input().Get("p")
	if limit == "" || page == "" {
		c.Data["json"] = object.GetTokens(owner)
		c.ServeJSON()
	} else {
		limit := util.ParseInt(limit)
		paginator := pagination.SetPaginator(c.Ctx, limit, int64(object.GetTokenCount(owner)))
		tokens := object.GetPaginationTokens(owner, paginator.Offset(), limit)
		c.ResponseOk(tokens, paginator.Nums())
	}
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

// GetOAuthCode
// @Title GetOAuthCode
// @Description get OAuth code
// @Param   user_id     query    string  true        "The id of user"
// @Param   client_id     query    string  true        "OAuth client id"
// @Param   response_type     query    string  true        "OAuth response type"
// @Param   redirect_uri     query    string  true        "OAuth redirect URI"
// @Param   scope     query    string  true        "OAuth scope"
// @Param   state     query    string  true        "OAuth state"
// @Success 200 {object} object.TokenWrapper The Response object
// @router /login/oauth/code [post]
func (c *ApiController) GetOAuthCode() {
	userId := c.Input().Get("user_id")
	clientId := c.Input().Get("client_id")
	responseType := c.Input().Get("response_type")
	redirectUri := c.Input().Get("redirect_uri")
	scope := c.Input().Get("scope")
	state := c.Input().Get("state")

	c.Data["json"] = object.GetOAuthCode(userId, clientId, responseType, redirectUri, scope, state)
	c.ServeJSON()
}

// GetOAuthToken
// @Title GetOAuthToken
// @Description get OAuth access token
// @Param   grant_type     query    string  true        "OAuth grant type"
// @Param   client_id     query    string  true        "OAuth client id"
// @Param   client_secret     query    string  true        "OAuth client secret"
// @Param   code     query    string  true        "OAuth code"
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
