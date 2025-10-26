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
	"encoding/json"

	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

// GetUserOrganizations
// @Title GetUserOrganizations
// @Tag User API
// @Description get organizations for a user
// @Param   id     query    string  true        "User ID"
// @Success 200 {array} object.UserOrganization The Response object
// @router /get-user-organizations [get]
func (c *ApiController) GetUserOrganizations() {
	userId := c.Input().Get("id")
	owner, name := util.GetOwnerAndNameFromId(userId)

	userOrganizations, err := object.GetUserOrganizations(owner, name)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(userOrganizations)
}

// AddUserToOrganization
// @Title AddUserToOrganization
// @Tag User API
// @Description add user to an organization
// @Param   body    body   object.UserOrganization  true        "The details of the user organization"
// @Success 200 {object} controllers.Response The Response object
// @router /add-user-to-organization [post]
func (c *ApiController) AddUserToOrganization() {
	var userOrganization object.UserOrganization
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &userOrganization)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	// Check if the user exists
	user, err := object.GetUser(util.GetId(userOrganization.Owner, userOrganization.Name))
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	if user == nil {
		c.ResponseError(c.T("general:The user does not exist"))
		return
	}

	// Check if the organization exists
	organization, err := object.GetOrganization(util.GetId("admin", userOrganization.Organization))
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	if organization == nil {
		c.ResponseError(c.T("general:The organization does not exist"))
		return
	}

	// Check if the relationship already exists
	existing, err := object.GetUserOrganization(userOrganization.Owner, userOrganization.Name, userOrganization.Organization)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	if existing != nil {
		c.ResponseError(c.T("general:The user is already a member of this organization"))
		return
	}

	userOrganization.CreatedTime = util.GetCurrentTime()

	success, err := object.AddUserOrganization(&userOrganization)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(success)
}

// RemoveUserFromOrganization
// @Title RemoveUserFromOrganization
// @Tag User API
// @Description remove user from an organization
// @Param   owner        query    string  true        "User owner"
// @Param   name         query    string  true        "User name"
// @Param   organization query    string  true        "Organization name"
// @Success 200 {object} controllers.Response The Response object
// @router /remove-user-from-organization [post]
func (c *ApiController) RemoveUserFromOrganization() {
	owner := c.Input().Get("owner")
	name := c.Input().Get("name")
	organization := c.Input().Get("organization")

	// Check if trying to remove from primary organization
	userOrg, err := object.GetUserOrganization(owner, name, organization)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	if userOrg == nil {
		c.ResponseError(c.T("general:The user organization relationship does not exist"))
		return
	}

	if userOrg.IsDefault {
		c.ResponseError(c.T("general:Cannot remove user from their primary organization"))
		return
	}

	success, err := object.DeleteUserOrganization(owner, name, organization)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(success)
}

// SetOrganizationContext
// @Title SetOrganizationContext
// @Tag User API
// @Description set the active organization context for the current user session
// @Param   organization query    string  true        "Organization name"
// @Success 200 {object} controllers.Response The Response object
// @router /set-organization-context [post]
func (c *ApiController) SetOrganizationContext() {
	organization := c.Input().Get("organization")

	userId := c.GetSessionUsername()
	if userId == "" {
		c.ResponseError(c.T("general:Please login first"))
		return
	}

	owner, name := util.GetOwnerAndNameFromId(userId)

	// Verify user is member of the organization
	userOrg, err := object.GetUserOrganization(owner, name, organization)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	if userOrg == nil {
		c.ResponseError(c.T("general:User is not a member of this organization"))
		return
	}

	// Store the organization context in session
	sessionData := c.GetSessionData()
	if sessionData == nil {
		sessionData = &SessionData{}
	}
	sessionData.OrganizationContext = organization
	c.SetSessionData(sessionData)

	c.ResponseOk(organization)
}

// GetOrganizationContext
// @Title GetOrganizationContext
// @Tag User API
// @Description get the active organization context for the current user session
// @Success 200 {object} controllers.Response The Response object
// @router /get-organization-context [get]
func (c *ApiController) GetOrganizationContext() {
	sessionData := c.GetSessionData()
	organizationContext := ""
	
	if sessionData != nil && sessionData.OrganizationContext != "" {
		organizationContext = sessionData.OrganizationContext
	} else {
		// Default to user's primary organization
		userId := c.GetSessionUsername()
		if userId != "" {
			owner, _ := util.GetOwnerAndNameFromId(userId)
			organizationContext = owner
		}
	}

	c.ResponseOk(organizationContext)
}
