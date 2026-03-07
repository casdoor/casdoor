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
	"fmt"
	"strings"

	"github.com/casdoor/casdoor/form"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

// KerberosLogin
// @Title KerberosLogin
// @Tag Login API
// @Description Kerberos/SPNEGO login via Integrated Windows Authentication
// @Param   application     query    string  true        "application name"
// @Success 200 {object} controllers.Response The Response object
// @router /kerberos-login [get]
func (c *ApiController) KerberosLogin() {
	applicationName := c.Ctx.Input.Query("application")
	if applicationName == "" {
		c.ResponseError(c.T("general:Missing parameter") + ": application")
		return
	}

	application, err := object.GetApplication(fmt.Sprintf("admin/%s", applicationName))
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	if application == nil {
		c.ResponseError(fmt.Sprintf(c.T("auth:The application: %s does not exist"), applicationName))
		return
	}

	organization, err := object.GetOrganization(util.GetId("admin", application.Organization))
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	if organization == nil {
		c.ResponseError(fmt.Sprintf("The organization: %s does not exist", application.Organization))
		return
	}

	if organization.KerberosRealm == "" || organization.KerberosKeytab == "" {
		c.ResponseError("Kerberos is not configured for this organization")
		return
	}

	authHeader := c.Ctx.Input.Header("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Negotiate ") {
		c.Ctx.Output.Header("WWW-Authenticate", "Negotiate")
		c.Ctx.Output.SetStatus(401)
		c.Ctx.Output.Body([]byte("Kerberos authentication required"))
		return
	}

	spnegoToken := strings.TrimPrefix(authHeader, "Negotiate ")

	kerberosUsername, err := object.ValidateKerberosToken(organization, spnegoToken)
	if err != nil {
		c.Ctx.Output.Header("WWW-Authenticate", "Negotiate")
		c.ResponseError(fmt.Sprintf("Kerberos authentication failed: %s", err.Error()))
		return
	}

	user, err := object.GetUserByKerberosName(organization.Name, kerberosUsername)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	if user == nil {
		c.ResponseError(fmt.Sprintf(c.T("general:The user: %s doesn't exist"), kerberosUsername))
		return
	}

	application.OrganizationObj = organization

	authForm := &form.AuthForm{
		Type:         "code",
		Application:  applicationName,
		Organization: organization.Name,
	}

	resp := c.HandleLoggedIn(application, user, authForm)
	if resp != nil {
		c.Data["json"] = resp
		c.ServeJSON()
	}
}
