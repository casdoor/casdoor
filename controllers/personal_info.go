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

	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

// maskPersonalInfo masks the phone, email and ID card of other users for the console, under the organization's
// security policy. The applications calling the API with their credentials get the values.
func (c *ApiController) maskPersonalInfo(users ...*object.User) {
	requester := c.GetSessionUsername()
	if strings.HasPrefix(requester, "app/") {
		return
	}
	for _, user := range users {
		if user != nil && user.GetId() != requester {
			object.MaskPersonalInfo(user)
		}
	}
}

// GetUserPersonalInfo
// @Title GetUserPersonalInfo
// @Tag User API
// @Description show one masked field of a user to an administrator, the request is recorded
// @Param   id    query string true "The id ( owner/name ) of the user"
// @Param   field query string true "phone, email or idCard"
// @Success 200 {string} string The Response object
// @router /get-user-personal-info [get]
func (c *ApiController) GetUserPersonalInfo() {
	id := c.Ctx.Input.Query("id")
	field := c.Ctx.Input.Query("field")

	user, err := object.GetUser(id)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	if user == nil {
		c.ResponseError(fmt.Sprintf(c.T("general:The user: %s doesn't exist"), id))
		return
	}
	if !c.IsAdminOrSelf(user) {
		c.ResponseError(c.T("auth:Unauthorized operation"))
		return
	}

	value, err := object.GetPersonalInfoValue(user, field)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	c.ResponseOk(value)
}

// ExportOwnData
// @Title ExportOwnData
// @Tag User API
// @Description download what Casdoor keeps about the signed-in user, when the organization allows it
// @Success 200 {object} controllers.Response The Response object
// @router /export-own-data [get]
func (c *ApiController) ExportOwnData() {
	user := c.getCurrentUser()
	if user == nil {
		c.ResponseError(c.T("general:Please login first"))
		return
	}

	policy := object.GetSecurityPolicyByOwner(user.Owner)
	if policy == nil || !policy.AllowDataExport {
		c.ResponseError(c.T("auth:Unauthorized operation"))
		return
	}

	data, err := object.ExportUserData(user)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Ctx.Output.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s-%s.json", user.Owner, user.Name))
	c.ResponseOk(data)
}

// DeleteOwnAccount
// @Title DeleteOwnAccount
// @Tag User API
// @Description delete the signed-in user's own account, when the organization allows it. A user with a password has to enter it.
// @Param   password formData string false "the user's password"
// @Success 200 {object} controllers.Response The Response object
// @router /delete-own-account [post]
func (c *ApiController) DeleteOwnAccount() {
	user := c.getCurrentUser()
	if user == nil {
		c.ResponseError(c.T("general:Please login first"))
		return
	}
	if user.IsAdmin || (user.Owner == "built-in") {
		c.ResponseError(c.T("user:Administrators can't delete their own accounts"))
		return
	}

	policy := object.GetSecurityPolicyByOwner(user.Owner)
	if policy == nil || !policy.AllowAccountDeletion {
		c.ResponseError(c.T("auth:Unauthorized operation"))
		return
	}

	if user.Password != "" {
		password := c.Ctx.Request.Form.Get("password")
		if _, err := object.CheckUserPassword(user.Owner, user.Name, password, c.GetAcceptLanguage(), false, false, user.Ldap != ""); err != nil {
			c.ResponseError(err.Error())
			return
		}
	}

	affected, err := object.DeleteUser(user)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ClearUserSession()
	c.Ctx.Input.SetParam("recordUserId", util.GetId(user.Owner, user.Name))
	c.ResponseOk(affected)
}
