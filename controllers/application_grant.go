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
	"strings"

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
	// an application's client credentials act as an admin of its organization; access to other
	// applications is only granted by people
	if object.IsAppUser(c.GetSessionUsername()) {
		c.ResponseError(c.T("auth:Unauthorized operation"))
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

// UpdateApplicationGrant
// @Title UpdateApplicationGrant
// @Tag Application Grant API
// @Description change the application roles, expiry or description of a grant; whom and which application it is for can't change
// @Param   body    body   object.ApplicationGrant  true        "The grant, identified by its owner and name"
// @Success 200 {object} controllers.Response The Response object
// @router /update-application-grant [post]
func (c *ApiController) UpdateApplicationGrant() {
	// a field left out keeps its value, so a partial update can't clear an expiry by accident
	var request struct {
		Owner       string    `json:"owner"`
		Name        string    `json:"name"`
		Roles       *[]string `json:"roles"`
		ExpireTime  *string   `json:"expireTime"`
		Description *string   `json:"description"`
	}
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

	application, ok := c.requireApplicationGrantAdmin(grant.Owner, grant.Application)
	if !ok {
		return
	}

	update := object.ApplicationGrant{Roles: grant.Roles, ExpireTime: grant.ExpireTime, Description: grant.Description}
	if request.Roles != nil {
		update.Roles = *request.Roles
	}
	if request.ExpireTime != nil {
		update.ExpireTime = *request.ExpireTime
	}
	if request.Description != nil {
		update.Description = *request.Description
	}

	err = object.CheckApplicationGrantUpdate(grant, &update, application, c.GetAcceptLanguage())
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	grant.UpdatedBy = c.GetSessionUsername()

	c.Data["json"] = wrapActionResponse(object.UpdateApplicationGrant(grant))
	c.ServeJSON()
}

// UpdateApplicationRoles
// @Title UpdateApplicationRoles
// @Tag Application Grant API
// @Description replace the application roles of an application. A role still assigned by a grant can't be removed.
// @Param   body    body   object.Application  true        "owner (the application's organization), name and appRoles"
// @Success 200 {object} controllers.Response The Response object
// @router /update-application-roles [post]
func (c *ApiController) UpdateApplicationRoles() {
	var request struct {
		Owner    string            `json:"owner"`
		Name     string            `json:"name"`
		AppRoles []*object.AppRole `json:"appRoles"`
	}
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &request)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	application, ok := c.requireApplicationGrantAdmin(request.Owner, request.Name)
	if !ok {
		return
	}

	c.Data["json"] = wrapActionResponse(object.UpdateApplicationAppRoles(application, request.AppRoles, c.GetAcceptLanguage()))
	c.ServeJSON()
}

// GetApplicationAccessPreview
// @Title GetApplicationAccessPreview
// @Tag Application Grant API
// @Description explain how a user may sign in to an application and which groups, roles and claims the application receives, optionally with application roles added as if granted
// @Param   owner          query    string  true    "The organization of the application"
// @Param   application    query    string  true    "The name of the application"
// @Param   user           query    string  true    "The user, organization/name"
// @Param   roles          query    string  false   "Application roles to add, comma separated"
// @Param   draft          query    string  false   "true to preview a grant about to be made for the user"
// @Success 200 {object} object.ApplicationAccessPreview The Response object
// @router /get-application-access-preview [get]
func (c *ApiController) GetApplicationAccessPreview() {
	owner := c.Ctx.Input.Query("owner")
	applicationName := c.Ctx.Input.Query("application")
	userId := c.Ctx.Input.Query("user")
	roles := []string{}
	for _, name := range strings.Split(c.Ctx.Input.Query("roles"), ",") {
		if name = strings.TrimSpace(name); name != "" {
			roles = append(roles, name)
		}
	}

	application, ok := c.requireApplicationGrantAdmin(owner, applicationName)
	if !ok {
		return
	}

	// without a user: only what the roles add, e.g. to preview a group grant
	if userId == "" {
		entitlements, err := object.GetApplicationEntitlements(nil, application, roles)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}
		c.ResponseOk(&object.ApplicationAccessPreview{
			Entitlements: entitlements, Groups: entitlements.Groups, Roles: entitlements.RoleNames,
			EmitsGroups: object.EmitsTokenField(application, "Groups"), EmitsRoles: object.EmitsTokenField(application, "Roles"),
		})
		return
	}

	user, err := object.GetUser(userId)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	draft := c.Ctx.Input.Query("draft") == "true"
	// a user of another organization is only explained when a grant covers them, or as the
	// principal of a grant being made; otherwise one answer, so the preview can't probe other
	// organizations
	visible := user != nil && !user.IsDeleted
	if visible && user.Owner != application.Organization && !draft {
		visible, err = object.IsApplicationGrantPrincipal(user, application)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}
	}
	if !visible {
		c.ResponseError(fmt.Sprintf(c.T("application:The user: %s doesn't exist or has no grant for this application"), userId))
		return
	}

	preview, err := object.GetApplicationAccessPreview(user, application, roles, draft, c.GetAcceptLanguage())
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(preview)
}
