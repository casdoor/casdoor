// Copyright 2022 The Casdoor Authors. All Rights Reserved.
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
	"context"
	"fmt"
	"html"
	"net/http"
	"net/url"

	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

func (c *ApiController) GetSamlMeta() {
	host := c.Ctx.Request.Host
	paramApp := c.Ctx.Input.Query("application")
	application, err := object.GetApplication(paramApp)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if application == nil {
		c.ResponseError(fmt.Sprintf(c.T("saml:Application %s not found"), paramApp))
		return
	}

	enablePostBinding, err := c.GetBool("enablePostBinding", false)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	metadata, err := object.GetSamlMeta(application, host, enablePostBinding)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["xml"] = metadata
	c.ServeXML()
}

func (c *ApiController) HandleSamlRedirect() {
	host := c.Ctx.Request.Host

	owner := c.Ctx.Input.Param(":owner")
	application := c.Ctx.Input.Param(":application")

	relayState := c.Ctx.Input.Query("RelayState")
	samlRequest := c.Ctx.Input.Query("SAMLRequest")
	username := c.Ctx.Input.Query("username")
	loginHint := c.Ctx.Input.Query("login_hint")

	relayState = url.QueryEscape(relayState)
	targetURL := object.GetSamlRedirectAddress(owner, application, relayState, samlRequest, host, username, loginHint)

	c.Redirect(targetURL, http.StatusSeeOther)
}

// HandleSamlLogout
// @Title HandleSamlLogout
// @Tag Login API
// @Description the SAML SingleLogoutService of the application, it takes an SP's LogoutRequest over the HTTP-Redirect or HTTP-POST binding
// @Param   owner          path    string  true    "The owner of the application"
// @Param   application    path    string  true    "The name of the application"
// @Param   SAMLRequest    query   string  false   "The LogoutRequest of the SP"
// @Param   SAMLResponse   query   string  false   "The LogoutResponse of the SP"
// @Param   RelayState     query   string  false   "The relay state"
// @Success 200 {object} controllers.Response The Response object
// @router /saml/logout/:owner/:application [get,post]
func (c *ApiController) HandleSamlLogout() {
	applicationId := util.GetId(c.Ctx.Input.Param(":owner"), c.Ctx.Input.Param(":application"))
	application, err := object.GetApplication(applicationId)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	if application == nil {
		c.ResponseError(fmt.Sprintf(c.T("saml:Application %s not found"), applicationId))
		return
	}

	samlRequest := c.Ctx.Input.Query("SAMLRequest")
	if samlRequest == "" {
		// The SP answers a LogoutRequest of Casdoor
		err = object.CheckSamlLogoutResponse(application, c.Ctx.Input.Query("SAMLResponse"))
		if err != nil {
			c.ResponseError(err.Error())
			return
		}
		c.ResponseOk()
		return
	}

	logoutRequest, err := object.ParseSamlLogoutRequest(application, samlRequest)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	if application.SamlSingleLogoutUrl == "" {
		c.ResponseError(fmt.Sprintf("err: the SAML single logout URL of the application: %s is empty", applicationId))
		return
	}

	currentSessionId := c.Ctx.Input.CruSession.SessionID(context.Background())
	isCurrentSession, err := object.LogoutBySamlRequest(application, logoutRequest, c.GetSessionUsername(), currentSessionId, c.Ctx.Request.Host)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	if isCurrentSession {
		c.ClearUserSession()
		c.ClearTokenSession()
	}

	// The LogoutResponse goes back over the binding the LogoutRequest came with
	relayState := c.Ctx.Input.Query("RelayState")
	usePost := c.Ctx.Request.Method == http.MethodPost
	logoutResponse, err := object.GetSamlLogoutResponse(application, logoutRequest, relayState, c.Ctx.Request.Host, usePost)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if !usePost {
		c.Redirect(logoutResponse, http.StatusFound)
		return
	}

	relayStateInput := ""
	if relayState != "" {
		relayStateInput = fmt.Sprintf(`<input type="hidden" name="RelayState" value="%s"/>`, html.EscapeString(relayState))
	}
	c.Ctx.Output.Header("Content-Type", "text/html; charset=utf-8")
	_ = c.Ctx.Output.Body([]byte(fmt.Sprintf(`<!DOCTYPE html><html><body onload="document.forms[0].submit()"><form method="post" action="%s"><input type="hidden" name="SAMLResponse" value="%s"/>%s<noscript><button type="submit">Continue</button></noscript></form></body></html>`,
		html.EscapeString(application.SamlSingleLogoutUrl), html.EscapeString(logoutResponse), relayStateInput)))
}
