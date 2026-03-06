// Copyright 2025 The Casdoor Authors. All Rights Reserved.
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
	"fmt"
	"net/http"
	"strings"

	"github.com/casdoor/casdoor/form"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

// GetKerberosLogin
// @Title GetKerberosLogin
// @Tag Login API
// @Description perform Integrated Windows Authentication (IWA) via Kerberos SPNEGO
// @Param   applicationName   query   string  true    "name of the application"
// @Param   redirectUri       query   string  false   "redirect URI after successful login"
// @Success 200 {object} controllers.Response The Response object
// @router /api/kerberos-login [get]
func (c *ApiController) GetKerberosLogin() {
	applicationName := c.Ctx.Input.Query("application")
	responseType := c.Ctx.Input.Query("responseType")
	clientId := c.Ctx.Input.Query("clientId")
	redirectUri := c.Ctx.Input.Query("redirectUri")

	application, err := object.GetApplication(fmt.Sprintf("admin/%s", applicationName))
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	if application == nil {
		c.ResponseError(fmt.Sprintf(c.T("auth:The application: %s does not exist"), applicationName))
		return
	}

	if !application.IsKerberosEnabled() {
		c.ResponseError(c.T("auth:The login method: login with Kerberos is not enabled for the application"))
		return
	}

	// Get organization to access Kerberos config
	org, err := object.GetOrganization(util.GetId("admin", application.Organization))
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	if org == nil {
		c.ResponseError(fmt.Sprintf(c.T("auth:The organization: %s does not exist"), application.Organization))
		return
	}

	// Check for SPNEGO Authorization header
	authHeader := c.Ctx.Request.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Negotiate ") {
		// No Kerberos token yet - send 401 challenge to trigger browser negotiation
		c.Ctx.ResponseWriter.Header().Set("WWW-Authenticate", "Negotiate")
		c.Ctx.ResponseWriter.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Extract the SPNEGO token from the Authorization header
	tokenBase64 := strings.TrimPrefix(authHeader, "Negotiate ")

	// Validate the Kerberos token and get the username
	username, err := object.CheckKerberosToken(org, tokenBase64)
	if err != nil {
		c.ResponseError(fmt.Sprintf(c.T("auth:Kerberos authentication failed: %s"), err.Error()))
		return
	}

	// Find the user in the organization
	user, err := object.GetUserByFields(application.Organization, username)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	if user == nil {
		c.ResponseError(fmt.Sprintf(c.T("general:The user: %s doesn't exist"), username))
		return
	}

	// Build the AuthForm for the login flow
	authForm := form.AuthForm{
		Type:        responseType,
		Application: applicationName,
		Organization: application.Organization,
		Username:    username,
		SigninMethod: "Kerberos",
		ClientId:    clientId,
		RedirectUri: redirectUri,
	}

	organization, err := object.GetOrganizationByUser(user)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if checkMfaEnable(c, user, organization, "kerberos") {
		return
	}

	resp := c.HandleLoggedIn(application, user, &authForm)
	c.Ctx.Input.SetParam("recordUserId", user.GetId())

	c.Data["json"] = resp
	c.ServeJSON()
}
