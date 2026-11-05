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

	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

// requireApplicationGrantAdmin returns the application whose grants the request manages. Only a
// global admin or an admin of the application's own organization may manage them: the admin of
// the organization a principal belongs to may not grant it access to a foreign application.
func (c *ApiController) requireApplicationGrantAdmin(organization string, applicationName string) (*object.Application, bool) {
	if _, ok := c.RequireSignedInUser(); !ok {
		return nil, false
	}
	if organization == "" || applicationName == "" || !c.IsAdminOfOrganization(organization) {
		c.ResponseError(c.T("auth:Unauthorized operation"))
		return nil, false
	}

	application, err := object.GetApplication(util.GetId("admin", applicationName))
	if err != nil {
		c.ResponseError(err.Error())
		return nil, false
	}
	if application == nil {
		c.ResponseError(fmt.Sprintf(c.T("auth:The application: %s does not exist"), applicationName))
		return nil, false
	}
	if application.Organization != organization {
		c.ResponseError(c.T("auth:Unauthorized operation"))
		return nil, false
	}
	return application, true
}

// GetApplicationGrants
// @Title GetApplicationGrants
// @Tag Application Grant API
// @Description get the cross-organization access grants of an application
// @Param   owner          query    string  true    "The organization of the application"
// @Param   application    query    string  true    "The name of the application"
// @Success 200 {array} object.ApplicationGrant The Response object
// @router /get-application-grants [get]
func (c *ApiController) GetApplicationGrants() {
	owner := c.Ctx.Input.Query("owner")
	applicationName := c.Ctx.Input.Query("application")

	application, ok := c.requireApplicationGrantAdmin(owner, applicationName)
	if !ok {
		return
	}

	grants, err := object.GetApplicationGrants(application.Organization, application.Name)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(grants)
}

// AddApplicationGrant
// @Title AddApplicationGrant
// @Tag Application Grant API
// @Description grant a user or group of another organization access to an application
// @Param   body    body   object.ApplicationGrant  true        "The details of the grant"
// @Success 200 {object} controllers.Response The Response object
// @router /add-application-grant [post]
func (c *ApiController) AddApplicationGrant() {
	var grant object.ApplicationGrant
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &grant)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	application, ok := c.requireApplicationGrantAdmin(grant.Owner, grant.Application)
	if !ok {
		return
	}

	err = object.CheckApplicationGrant(&grant, application, c.GetAcceptLanguage())
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	grant.CreatedBy = c.GetSessionUsername()

	c.Data["json"] = wrapActionResponse(object.AddApplicationGrant(&grant))
	c.ServeJSON()
}

// DeleteApplicationGrant
// @Title DeleteApplicationGrant
// @Tag Application Grant API
// @Description revoke a cross-organization access grant of an application
// @Param   body    body   object.ApplicationGrant  true        "The grant, identified by its owner and name"
// @Success 200 {object} controllers.Response The Response object
// @router /delete-application-grant [post]
func (c *ApiController) DeleteApplicationGrant() {
	var request object.ApplicationGrant
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &request)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	grant, err := object.GetApplicationGrant(util.GetId(request.Owner, request.Name))
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	if grant == nil {
		if _, ok := c.RequireSignedInUser(); ok {
			c.ResponseError(c.T("auth:Unauthorized operation"))
		}
		return
	}

	// authorized by the stored grant, not by what the request claims about it
	if _, ok := c.requireApplicationGrantAdmin(grant.Owner, grant.Application); !ok {
		return
	}

	c.Data["json"] = wrapActionResponse(object.DeleteApplicationGrant(grant))
	c.ServeJSON()
}

// GetApplicationGrantPrincipals
// @Title GetApplicationGrantPrincipals
// @Tag Application Grant API
// @Description find the users or groups of another organization that may be granted access to an application. A global admin may search by part of the name, an admin of the application's organization only finds the principal they name exactly.
// @Param   owner          query    string  true    "The organization of the application"
// @Param   application    query    string  true    "The name of the application"
// @Param   principalType  query    string  true    "User or Group"
// @Param   organization   query    string  true    "The organization to search"
// @Param   query          query    string  true    "The name to search for"
// @Success 200 {array} object.ApplicationGrantPrincipal The Response object
// @router /get-application-grant-principals [get]
func (c *ApiController) GetApplicationGrantPrincipals() {
	owner := c.Ctx.Input.Query("owner")
	applicationName := c.Ctx.Input.Query("application")
	principalType := c.Ctx.Input.Query("principalType")
	organization := c.Ctx.Input.Query("organization")
	query := c.Ctx.Input.Query("query")

	application, ok := c.requireApplicationGrantAdmin(owner, applicationName)
	if !ok {
		return
	}

	principals, err := object.SearchApplicationGrantPrincipals(application, principalType, organization, query, !c.IsGlobalAdmin())
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(principals)
}
