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

package controllers

import (
	"encoding/json"
	"time"

	"github.com/beego/beego/utils/pagination"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

// GetTokens
// @Title GetTokens
// @Tag Token API
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
	field := c.Input().Get("field")
	value := c.Input().Get("value")
	sortField := c.Input().Get("sortField")
	sortOrder := c.Input().Get("sortOrder")
	organization := c.Input().Get("organization")
	if limit == "" || page == "" {
		token, err := object.GetTokens(owner, organization)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		c.ResponseOk(token)
	} else {
		limit := util.ParseInt(limit)
		count, err := object.GetTokenCount(owner, organization, field, value)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		paginator := pagination.SetPaginator(c.Ctx, limit, count)
		tokens, err := object.GetPaginationTokens(owner, organization, paginator.Offset(), limit, field, value, sortField, sortOrder)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		c.ResponseOk(tokens, paginator.Nums())
	}
}

// GetToken
// @Title GetToken
// @Tag Token API
// @Description get token
// @Param   id     query    string  true        "The id ( owner/name ) of token"
// @Success 200 {object} object.Token The Response object
// @router /get-token [get]
func (c *ApiController) GetToken() {
	id := c.Input().Get("id")
	token, err := object.GetToken(id)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(token)
}

// UpdateToken
// @Title UpdateToken
// @Tag Token API
// @Description update token
// @Param   id     query    string  true        "The id ( owner/name ) of token"
// @Param   body    body   object.Token  true        "Details of the token"
// @Success 200 {object} controllers.Response The Response object
// @router /update-token [post]
func (c *ApiController) UpdateToken() {
	id := c.Input().Get("id")

	var token object.Token
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &token)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = wrapActionResponse(object.UpdateToken(id, &token))
	c.ServeJSON()
}

// AddToken
// @Title AddToken
// @Tag Token API
// @Description add token
// @Param   body    body   object.Token  true        "Details of the token"
// @Success 200 {object} controllers.Response The Response object
// @router /add-token [post]
func (c *ApiController) AddToken() {
	var token object.Token
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &token)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = wrapActionResponse(object.AddToken(&token))
	c.ServeJSON()
}

// DeleteToken
// @Tag Token API
// @Title DeleteToken
// @Description delete token
// @Param   body    body   object.Token  true        "Details of the token"
// @Success 200 {object} controllers.Response The Response object
// @router /delete-token [post]
func (c *ApiController) DeleteToken() {
	var token object.Token
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &token)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = wrapActionResponse(object.DeleteToken(&token))
	c.ServeJSON()
}

// GetOAuthToken
// @Title GetOAuthToken
// @Tag Token API
// @Description get OAuth access token
// @Param   grant_type     query    string  true        "OAuth grant type"
// @Param   client_id     query    string  true        "OAuth client id"
// @Param   client_secret     query    string  true        "OAuth client secret"
// @Param   code     query    string  true        "OAuth code"
// @Success 200 {object} object.TokenWrapper The Response object
// @Success 400 {object} object.TokenError The Response object
// @Success 401 {object} object.TokenError The Response object
// @router /login/oauth/access_token [post]
func (c *ApiController) GetOAuthToken() {
	clientId := c.Input().Get("client_id")
	clientSecret := c.Input().Get("client_secret")
	grantType := c.Input().Get("grant_type")
	code := c.Input().Get("code")
	verifier := c.Input().Get("code_verifier")
	scope := c.Input().Get("scope")
	nonce := c.Input().Get("nonce")
	username := c.Input().Get("username")
	password := c.Input().Get("password")
	tag := c.Input().Get("tag")
	avatar := c.Input().Get("avatar")
	refreshToken := c.Input().Get("refresh_token")
	deviceCode := c.Input().Get("device_code")

	if clientId == "" && clientSecret == "" {
		clientId, clientSecret, _ = c.Ctx.Request.BasicAuth()
	}

	if grantType == "urn:ietf:params:oauth:grant-type:device_code" {
		clientId, clientSecret, _ = c.Ctx.Request.BasicAuth()
	}

	if len(c.Ctx.Input.RequestBody) != 0 && grantType != "urn:ietf:params:oauth:grant-type:device_code" {
		// If clientId is empty, try to read data from RequestBody
		var tokenRequest TokenRequest
		err := json.Unmarshal(c.Ctx.Input.RequestBody, &tokenRequest)
		if err == nil {
			if clientId == "" {
				clientId = tokenRequest.ClientId
			}
			if clientSecret == "" {
				clientSecret = tokenRequest.ClientSecret
			}
			if grantType == "" {
				grantType = tokenRequest.GrantType
			}
			if code == "" {
				code = tokenRequest.Code
			}
			if verifier == "" {
				verifier = tokenRequest.Verifier
			}
			if scope == "" {
				scope = tokenRequest.Scope
			}
			if nonce == "" {
				nonce = tokenRequest.Nonce
			}
			if username == "" {
				username = tokenRequest.Username
			}
			if password == "" {
				password = tokenRequest.Password
			}
			if tag == "" {
				tag = tokenRequest.Tag
			}
			if avatar == "" {
				avatar = tokenRequest.Avatar
			}
			if refreshToken == "" {
				refreshToken = tokenRequest.RefreshToken
			}
		}
	}

	if deviceCode != "" {
		deviceAuthCache, ok := object.DeviceAuthMap.Load(deviceCode)
		if !ok {
			c.Data["json"] = object.TokenError{
				Error:            "expired_token",
				ErrorDescription: "token is expired",
			}
			c.ServeJSON()
			return
		}

		deviceAuthCacheCast := deviceAuthCache.(object.DeviceAuthCache)
		if !deviceAuthCacheCast.UserSignIn {
			c.Data["json"] = object.TokenError{
				Error:            "authorization_pending",
				ErrorDescription: "authorization pending",
			}
			c.ServeJSON()
			return
		}

		if deviceAuthCacheCast.RequestAt.Add(time.Second * 120).Before(time.Now()) {
			c.Data["json"] = object.TokenError{
				Error:            "expired_token",
				ErrorDescription: "token is expired",
			}
			c.ServeJSON()
			return
		}
		object.DeviceAuthMap.Delete(deviceCode)

		username = deviceAuthCacheCast.UserName
	}

	host := c.Ctx.Request.Host
	token, err := object.GetOAuthToken(grantType, clientId, clientSecret, code, verifier, scope, nonce, username, password, host, refreshToken, tag, avatar, c.GetAcceptLanguage())
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = token
	c.SetTokenErrorHttpStatus()
	c.ServeJSON()
}

// RefreshToken
// @Title RefreshToken
// @Tag Token API
// @Description refresh OAuth access token
// @Param   grant_type     query    string  true        "OAuth grant type"
// @Param	refresh_token	query	string	true		"OAuth refresh token"
// @Param   scope     query    string  true        "OAuth scope"
// @Param   client_id     query    string  true        "OAuth client id"
// @Param   client_secret     query    string  false        "OAuth client secret"
// @Success 200 {object} object.TokenWrapper The Response object
// @Success 400 {object} object.TokenError The Response object
// @Success 401 {object} object.TokenError The Response object
// @router /login/oauth/refresh_token [post]
func (c *ApiController) RefreshToken() {
	grantType := c.Input().Get("grant_type")
	refreshToken := c.Input().Get("refresh_token")
	scope := c.Input().Get("scope")
	clientId := c.Input().Get("client_id")
	clientSecret := c.Input().Get("client_secret")
	host := c.Ctx.Request.Host

	if clientId == "" {
		// If clientID is empty, try to read data from RequestBody
		var tokenRequest TokenRequest
		if err := json.Unmarshal(c.Ctx.Input.RequestBody, &tokenRequest); err == nil {
			clientId = tokenRequest.ClientId
			clientSecret = tokenRequest.ClientSecret
			grantType = tokenRequest.GrantType
			scope = tokenRequest.Scope
			refreshToken = tokenRequest.RefreshToken
		}
	}

	refreshToken2, err := object.RefreshToken(grantType, refreshToken, scope, clientId, clientSecret, host)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = refreshToken2
	c.SetTokenErrorHttpStatus()
	c.ServeJSON()
}

func (c *ApiController) ResponseTokenError(errorMsg string) {
	c.Data["json"] = &object.TokenError{
		Error: errorMsg,
	}
	c.SetTokenErrorHttpStatus()
	c.ServeJSON()
}

// IntrospectToken
// @Title IntrospectToken
// @Tag Login API
// @Description The introspection endpoint is an OAuth 2.0 endpoint that takes a
// parameter representing an OAuth 2.0 token and returns a JSON document
// representing the meta information surrounding the
// token, including whether this token is currently active.
// This endpoint only support Basic Authorization.
//
// @Param token formData string true "access_token's value or refresh_token's value"
// @Param token_type_hint formData string true "the token type access_token or refresh_token"
// @Success 200 {object} object.IntrospectionResponse The Response object
// @Success 400 {object} object.TokenError The Response object
// @Success 401 {object} object.TokenError The Response object
// @router /login/oauth/introspect [post]
func (c *ApiController) IntrospectToken() {
	tokenValue := c.Input().Get("token")
	clientId, clientSecret, ok := c.Ctx.Request.BasicAuth()
	if !ok {
		clientId = c.Input().Get("client_id")
		clientSecret = c.Input().Get("client_secret")
		if clientId == "" || clientSecret == "" {
			c.ResponseTokenError(object.InvalidRequest)
			return
		}
	}

	application, err := object.GetApplicationByClientId(clientId)
	if err != nil {
		c.ResponseTokenError(err.Error())
		return
	}

	if application == nil || application.ClientSecret != clientSecret {
		c.ResponseTokenError(c.T("token:Invalid application or wrong clientSecret"))
		return
	}

	respondWithInactiveToken := func() {
		c.Data["json"] = &object.IntrospectionResponse{Active: false}
		c.ServeJSON()
	}

	tokenTypeHint := c.Input().Get("token_type_hint")
	var token *object.Token
	if tokenTypeHint != "" {
		token, err = object.GetTokenByTokenValue(tokenValue, tokenTypeHint)
		if err != nil {
			c.ResponseTokenError(err.Error())
			return
		}
		if token == nil || token.ExpiresIn <= 0 {
			respondWithInactiveToken()
			return
		}

		if token.ExpiresIn <= 0 {
			c.Data["json"] = &object.IntrospectionResponse{Active: false}
			c.ServeJSON()
			return
		}
	}

	var introspectionResponse object.IntrospectionResponse

	if application.TokenFormat == "JWT-Standard" {
		jwtToken, err := object.ParseStandardJwtTokenByApplication(tokenValue, application)
		if err != nil {
			// and token revoked case. but we not implement
			// TODO: 2022-03-03 add token revoked check, when we implemented the Token Revocation(rfc7009) Specs.
			// refs: https://tools.ietf.org/html/rfc7009
			respondWithInactiveToken()
			return
		}

		introspectionResponse = object.IntrospectionResponse{
			Active:    true,
			Scope:     jwtToken.Scope,
			ClientId:  clientId,
			Username:  jwtToken.Name,
			TokenType: jwtToken.TokenType,
			Exp:       jwtToken.ExpiresAt.Unix(),
			Iat:       jwtToken.IssuedAt.Unix(),
			Nbf:       jwtToken.NotBefore.Unix(),
			Sub:       jwtToken.Subject,
			Aud:       jwtToken.Audience,
			Iss:       jwtToken.Issuer,
			Jti:       jwtToken.ID,
		}
	} else {
		jwtToken, err := object.ParseJwtTokenByApplication(tokenValue, application)
		if err != nil {
			// and token revoked case. but we not implement
			// TODO: 2022-03-03 add token revoked check, when we implemented the Token Revocation(rfc7009) Specs.
			// refs: https://tools.ietf.org/html/rfc7009
			respondWithInactiveToken()
			return
		}

		introspectionResponse = object.IntrospectionResponse{
			Active:   true,
			ClientId: clientId,
			Exp:      jwtToken.ExpiresAt.Unix(),
			Iat:      jwtToken.IssuedAt.Unix(),
			Nbf:      jwtToken.NotBefore.Unix(),
			Sub:      jwtToken.Subject,
			Aud:      jwtToken.Audience,
			Iss:      jwtToken.Issuer,
			Jti:      jwtToken.ID,
		}

		if jwtToken.Scope != "" {
			introspectionResponse.Scope = jwtToken.Scope
		}
		if jwtToken.Name != "" {
			introspectionResponse.Username = jwtToken.Name
		}
		if jwtToken.TokenType != "" {
			introspectionResponse.TokenType = jwtToken.TokenType
		}
	}

	if tokenTypeHint == "" {
		token, err = object.GetTokenByTokenValue(tokenValue, introspectionResponse.TokenType)
		if err != nil {
			c.ResponseTokenError(err.Error())
			return
		}
		if token == nil || token.ExpiresIn <= 0 {
			respondWithInactiveToken()
			return
		}
	}

	if token != nil {
		introspectionResponse.TokenType = token.TokenType
	}

	c.Data["json"] = introspectionResponse
	c.ServeJSON()
}
