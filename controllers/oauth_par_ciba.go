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
	"net/http"

	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

func (c *ApiController) getOAuthClientCredentials() (string, string, string, string) {
	clientId := c.Ctx.Input.Query("client_id")
	clientSecret := c.Ctx.Input.Query("client_secret")
	if clientId == "" && clientSecret == "" {
		clientId, clientSecret, _ = c.getClientBasicAuth()
	}
	return clientId, clientSecret, c.Ctx.Input.Query("client_assertion"), c.Ctx.Input.Query("client_assertion_type")
}

func (c *ApiController) responseOAuthJson(status int, data interface{}) {
	c.Ctx.Output.Header("Cache-Control", "no-store")
	c.Ctx.Output.Header("Pragma", "no-cache")
	if tokenError, ok := data.(*object.TokenError); ok {
		if tokenError.Error == object.InvalidClient {
			status = http.StatusUnauthorized
		} else {
			status = http.StatusBadRequest
		}
	}
	c.Ctx.Output.SetStatus(status)
	c.Data["json"] = data
	c.ServeJSON()
}

// PushAuthorizationRequest
// @Title PushAuthorizationRequest
// @Tag Token API
// @Description push an authorization request and get the request_uri for the authorization endpoint (RFC 9126)
// @Param   client_id     formData    string  true        "OAuth client id"
// @Param   client_secret     formData    string  false        "OAuth client secret, can be empty for a public client that uses PKCE"
// @Param   response_type     formData    string  true        "the parameters of the authorization request"
// @Success 201 {object} object.ParResponse The Response object
// @Success 400 {object} object.TokenError The Response object
// @Success 401 {object} object.TokenError The Response object
// @router /login/oauth/par [post]
func (c *ApiController) PushAuthorizationRequest() {
	clientId, clientSecret, clientAssertion, clientAssertionType := c.getOAuthClientCredentials()
	isPublicClient := clientSecret == "" && clientAssertion == ""

	application, tokenError, err := object.AuthenticateOAuthClient(clientId, clientSecret, clientAssertion, clientAssertionType, c.Ctx.Request.Host, true)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	if tokenError != nil {
		c.responseOAuthJson(http.StatusBadRequest, tokenError)
		return
	}

	res, tokenError, err := object.PushAuthorizationRequest(application, c.Ctx.Request.Form, isPublicClient, c.GetAcceptLanguage())
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	if tokenError != nil {
		c.responseOAuthJson(http.StatusBadRequest, tokenError)
		return
	}

	c.responseOAuthJson(http.StatusCreated, res)
}

// BackchannelAuthenticate
// @Title BackchannelAuthenticate
// @Tag Token API
// @Description start a client-initiated backchannel authentication (OpenID Connect CIBA, poll mode), the user approves it on their own device
// @Param   client_id     formData    string  true        "OAuth client id"
// @Param   client_secret     formData    string  true        "OAuth client secret"
// @Param   scope     formData    string  true        "must contain openid"
// @Param   login_hint     formData    string  false        "the username, email or phone of the user"
// @Param   id_token_hint     formData    string  false        "an ID token previously issued to the client for the user"
// @Param   binding_message     formData    string  false        "a short message shown to the user on both devices"
// @Param   requested_expiry     formData    int  false        "seconds the request stays valid, at most 1800"
// @Success 200 {object} object.CibaResponse The Response object
// @Success 400 {object} object.TokenError The Response object
// @Success 401 {object} object.TokenError The Response object
// @router /login/oauth/bc-authorize [post]
func (c *ApiController) BackchannelAuthenticate() {
	clientId, clientSecret, clientAssertion, clientAssertionType := c.getOAuthClientCredentials()

	application, tokenError, err := object.AuthenticateOAuthClient(clientId, clientSecret, clientAssertion, clientAssertionType, c.Ctx.Request.Host, false)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	if tokenError != nil {
		c.responseOAuthJson(http.StatusBadRequest, tokenError)
		return
	}

	res, tokenError, err := object.StartCibaRequest(application,
		c.Ctx.Input.Query("scope"),
		c.Ctx.Input.Query("login_hint"),
		c.Ctx.Input.Query("id_token_hint"),
		c.Ctx.Input.Query("login_hint_token"),
		c.Ctx.Input.Query("binding_message"),
		c.Ctx.Input.Query("requested_expiry"),
		util.GetClientIpFromRequest(c.Ctx.Request),
		c.Ctx.Request.Host)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	if tokenError != nil {
		c.responseOAuthJson(http.StatusBadRequest, tokenError)
		return
	}

	c.responseOAuthJson(http.StatusOK, res)
}

// GetCibaRequest
// @Title GetCibaRequest
// @Tag Token API
// @Description get a backchannel authentication request of the signed-in user, for the approval page
// @Param   requestId     query    string  true        "the id in the link sent to the user"
// @Success 200 {object} object.CibaRequest The Response object
// @router /login/oauth/ciba-request [get]
func (c *ApiController) GetCibaRequest() {
	userId, ok := c.RequireSignedIn()
	if !ok {
		return
	}

	request, err := object.GetCibaRequest(c.Ctx.Input.Query("requestId"), userId)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(request)
}

// AnswerCibaRequest
// @Title AnswerCibaRequest
// @Tag Token API
// @Description approve or deny a backchannel authentication request of the signed-in user
// @Param   requestId     query    string  true        "the id in the link sent to the user"
// @Param   approve     query    bool  true        "true to approve, false to deny"
// @Success 200 {object} controllers.Response The Response object
// @router /login/oauth/ciba-answer [post]
func (c *ApiController) AnswerCibaRequest() {
	userId, ok := c.RequireSignedIn()
	if !ok {
		return
	}

	status, err := object.AnswerCibaRequest(c.Ctx.Input.Query("requestId"), userId, c.Ctx.Input.Query("approve") == "true")
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(status)
}
