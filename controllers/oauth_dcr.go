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
	"net/http"
	"strings"

	"github.com/casdoor/casdoor/object"
)

func (c *ApiController) getRegistrationBaseUri() string {
	scheme := "https"
	if c.Ctx.Request.TLS == nil {
		scheme = "http"
	}
	return fmt.Sprintf("%s://%s/api/oauth/register", scheme, c.Ctx.Request.Host)
}

func (c *ApiController) getBearerToken() string {
	auth := c.Ctx.Request.Header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}
	return ""
}

// DynamicClientRegister
// @Title DynamicClientRegister
// @Tag OAuth API
// @Description Register a new OAuth 2.0 client dynamically (RFC 7591)
// @Param   organization     query    string  false        "The organization name (defaults to built-in)"
// @Param   body    body   object.DynamicClientRegistrationRequest  true        "Client registration request"
// @Success 201 {object} object.DynamicClientRegistrationResponse
// @Failure 400 {object} object.DcrError
// @router /api/oauth/register [post]
func (c *ApiController) DynamicClientRegister() {
	var req object.DynamicClientRegistrationRequest
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &req)
	if err != nil {
		c.Ctx.Output.Status = http.StatusBadRequest
		c.Data["json"] = object.DcrError{
			Error:            "invalid_client_metadata",
			ErrorDescription: "invalid request body: " + err.Error(),
		}
		c.ServeJSON()
		return
	}

	// Get organization from query parameter or default to built-in
	organization := c.Ctx.Input.Query("organization")
	if organization == "" {
		organization = "built-in"
	}

	response, dcrErr, err := object.RegisterDynamicClient(&req, organization, c.getRegistrationBaseUri())
	if err != nil {
		c.Ctx.Output.Status = http.StatusInternalServerError
		c.Data["json"] = object.DcrError{
			Error:            "server_error",
			ErrorDescription: err.Error(),
		}
		c.ServeJSON()
		return
	}
	if dcrErr != nil {
		c.Ctx.Output.Status = http.StatusBadRequest
		c.Data["json"] = dcrErr
		c.ServeJSON()
		return
	}

	// Return 201 Created
	c.Ctx.Output.Status = http.StatusCreated
	c.Data["json"] = response
	c.ServeJSON()
}

// DynamicClientRead
// @Title DynamicClientRead
// @Tag OAuth API
// @Description Read a dynamically registered client's metadata (RFC 7592)
// @Param   clientId     path    string  true        "The client_id"
// @Success 200 {object} object.DynamicClientRegistrationResponse
// @Failure 401 {object} object.DcrError
// @Failure 404 {object} object.DcrError
// @router /api/oauth/register/:clientId [get]
func (c *ApiController) DynamicClientRead() {
	clientId := c.Ctx.Input.Param(":clientId")
	app, dcrErr := object.GetDynamicClientByToken(clientId, c.getBearerToken())
	if dcrErr != nil {
		if dcrErr.Error == "invalid_token" {
			c.Ctx.Output.Status = http.StatusUnauthorized
		} else {
			c.Ctx.Output.Status = http.StatusInternalServerError
		}
		c.Data["json"] = dcrErr
		c.ServeJSON()
		return
	}

	c.Data["json"] = object.GetDynamicClientRegistrationResponse(app, c.getRegistrationBaseUri())
	c.ServeJSON()
}

// DynamicClientUpdate
// @Title DynamicClientUpdate
// @Tag OAuth API
// @Description Update a dynamically registered client's metadata (RFC 7592)
// @Param   clientId     path    string  true        "The client_id"
// @Param   body    body   object.DynamicClientRegistrationRequest  true        "Updated client metadata"
// @Success 200 {object} object.DynamicClientRegistrationResponse
// @Failure 400 {object} object.DcrError
// @Failure 401 {object} object.DcrError
// @router /api/oauth/register/:clientId [put]
func (c *ApiController) DynamicClientUpdate() {
	clientId := c.Ctx.Input.Param(":clientId")
	app, dcrErr := object.GetDynamicClientByToken(clientId, c.getBearerToken())
	if dcrErr != nil {
		if dcrErr.Error == "invalid_token" {
			c.Ctx.Output.Status = http.StatusUnauthorized
		} else {
			c.Ctx.Output.Status = http.StatusInternalServerError
		}
		c.Data["json"] = dcrErr
		c.ServeJSON()
		return
	}

	var req object.DynamicClientRegistrationRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		c.Ctx.Output.Status = http.StatusBadRequest
		c.Data["json"] = object.DcrError{
			Error:            "invalid_client_metadata",
			ErrorDescription: "invalid request body: " + err.Error(),
		}
		c.ServeJSON()
		return
	}

	response, dcrErr, err := object.UpdateDynamicClient(app, &req)
	if err != nil {
		c.Ctx.Output.Status = http.StatusInternalServerError
		c.Data["json"] = object.DcrError{Error: "server_error", ErrorDescription: err.Error()}
		c.ServeJSON()
		return
	}
	if dcrErr != nil {
		c.Ctx.Output.Status = http.StatusBadRequest
		c.Data["json"] = dcrErr
		c.ServeJSON()
		return
	}

	response.RegistrationClientUri = fmt.Sprintf("%s/%s", c.getRegistrationBaseUri(), clientId)
	c.Data["json"] = response
	c.ServeJSON()
}

// DynamicClientDelete
// @Title DynamicClientDelete
// @Tag OAuth API
// @Description Delete a dynamically registered client (RFC 7592)
// @Param   clientId     path    string  true        "The client_id"
// @Success 204
// @Failure 401 {object} object.DcrError
// @router /api/oauth/register/:clientId [delete]
func (c *ApiController) DynamicClientDelete() {
	clientId := c.Ctx.Input.Param(":clientId")
	app, dcrErr := object.GetDynamicClientByToken(clientId, c.getBearerToken())
	if dcrErr != nil {
		if dcrErr.Error == "invalid_token" {
			c.Ctx.Output.Status = http.StatusUnauthorized
		} else {
			c.Ctx.Output.Status = http.StatusInternalServerError
		}
		c.Data["json"] = dcrErr
		c.ServeJSON()
		return
	}

	if dcrErr = object.DeleteDynamicClient(app); dcrErr != nil {
		c.Ctx.Output.Status = http.StatusInternalServerError
		c.Data["json"] = dcrErr
		c.ServeJSON()
		return
	}

	c.Ctx.Output.Status = http.StatusNoContent
	c.ServeJSON()
}
