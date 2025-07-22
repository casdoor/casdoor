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

	"github.com/beego/beego/utils/pagination"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

// GetRoles
// @Title GetRoles
// @Tag Role API
// @Description get roles
// @Param   owner     query    string  true        "The owner of roles"
// @Success 200 {array} object.Role The Response object
// @router /get-roles [get]
func (c *ApiController) GetRoles() {
	owner := c.Input().Get("owner")
	limit := c.Input().Get("pageSize")
	page := c.Input().Get("p")
	field := c.Input().Get("field")
	value := c.Input().Get("value")
	sortField := c.Input().Get("sortField")
	sortOrder := c.Input().Get("sortOrder")

	if limit == "" || page == "" {
		roles, err := object.GetRoles(owner)
		if err != nil {
			c.ResponseErr(err)
			return
		}

		c.ResponseOk(roles)
	} else {
		limit := util.ParseInt(limit)
		count, err := object.GetRoleCount(owner, field, value)
		if err != nil {
			c.ResponseErr(err)
			return
		}

		paginator := pagination.SetPaginator(c.Ctx, limit, count)
		roles, err := object.GetPaginationRoles(owner, paginator.Offset(), limit, field, value, sortField, sortOrder)
		if err != nil {
			c.ResponseErr(err)
			return
		}

		c.ResponseOk(roles, paginator.Nums())
	}
}

// GetRole
// @Title GetRole
// @Tag Role API
// @Description get role
// @Param   id     query    string  true        "The id ( owner/name ) of the role"
// @Success 200 {object} object.Role The Response object
// @router /get-role [get]
func (c *ApiController) GetRole() {
	id := c.Input().Get("id")

	role, err := object.GetRole(id)
	if err != nil {
		c.ResponseErr(err)
		return
	}

	c.ResponseOk(role)
}

// UpdateRole
// @Title UpdateRole
// @Tag Role API
// @Description update role
// @Param   id     query    string  true        "The id ( owner/name ) of the role"
// @Param   body    body   object.Role  true        "The details of the role"
// @Success 200 {object} controllers.Response The Response object
// @router /update-role [post]
func (c *ApiController) UpdateRole() {
	id := c.Input().Get("id")

	var role object.Role
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &role)
	if err != nil {
		c.ResponseErr(err)
		return
	}

	c.Data["json"] = wrapActionResponse(object.UpdateRole(id, &role))
	c.ServeJSON()
}

// AddRole
// @Title AddRole
// @Tag Role API
// @Description add role
// @Param   body    body   object.Role  true        "The details of the role"
// @Success 200 {object} controllers.Response The Response object
// @router /add-role [post]
func (c *ApiController) AddRole() {
	var role object.Role
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &role)
	if err != nil {
		c.ResponseErr(err)
		return
	}

	c.Data["json"] = wrapActionResponse(object.AddRole(&role))
	c.ServeJSON()
}

// DeleteRole
// @Title DeleteRole
// @Tag Role API
// @Description delete role
// @Param   body    body   object.Role  true        "The details of the role"
// @Success 200 {object} controllers.Response The Response object
// @router /delete-role [post]
func (c *ApiController) DeleteRole() {
	var role object.Role
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &role)
	if err != nil {
		c.ResponseErr(err)
		return
	}

	c.Data["json"] = wrapActionResponse(object.DeleteRole(&role))
	c.ServeJSON()
}
