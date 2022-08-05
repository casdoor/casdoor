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

	"github.com/astaxie/beego/utils/pagination"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

// GetPermissions
// @Title GetPermissions
// @Tag Permission API
// @Description get permissions
// @Param   owner     query    string  true        "The owner of permissions"
// @Success 200 {array} object.Permission The Response object
// @router /get-permissions [get]
func (c *ApiController) GetPermissions() {
	owner := c.Input().Get("owner")
	limit := c.Input().Get("pageSize")
	page := c.Input().Get("p")
	field := c.Input().Get("field")
	value := c.Input().Get("value")
	sortField := c.Input().Get("sortField")
	sortOrder := c.Input().Get("sortOrder")
	if limit == "" || page == "" {
		c.Data["json"] = object.GetPermissions(owner)
		c.ServeJSON()
	} else {
		limit := util.ParseInt(limit)
		paginator := pagination.SetPaginator(c.Ctx, limit, int64(object.GetPermissionCount(owner, field, value)))
		permissions := object.GetPaginationPermissions(owner, paginator.Offset(), limit, field, value, sortField, sortOrder)
		c.ResponseOk(permissions, paginator.Nums())
	}
}

// @Title GetPermission
// @Tag Permission API
// @Description get permission
// @Param   id    query    string  true        "The id of the permission"
// @Success 200 {object} object.Permission The Response object
// @router /get-permission [get]
func (c *ApiController) GetPermission() {
	id := c.Input().Get("id")

	c.Data["json"] = object.GetPermission(id)
	c.ServeJSON()
}

// @Title UpdatePermission
// @Tag Permission API
// @Description update permission
// @Param   id    query    string  true        "The id of the permission"
// @Param   body    body   object.Permission  true        "The details of the permission"
// @Success 200 {object} controllers.Response The Response object
// @router /update-permission [post]
func (c *ApiController) UpdatePermission() {
	id := c.Input().Get("id")

	var permission object.Permission
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &permission)
	if err != nil {
		panic(err)
	}

	c.Data["json"] = wrapActionResponse(object.UpdatePermission(id, &permission))
	c.ServeJSON()
}

// @Title AddPermission
// @Tag Permission API
// @Description add permission
// @Param   body    body   object.Permission  true        "The details of the permission"
// @Success 200 {object} controllers.Response The Response object
// @router /add-permission [post]
func (c *ApiController) AddPermission() {
	var permission object.Permission
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &permission)
	if err != nil {
		panic(err)
	}

	c.Data["json"] = wrapActionResponse(object.AddPermission(&permission))
	c.ServeJSON()
}

// @Title DeletePermission
// @Tag Permission API
// @Description delete permission
// @Param   body    body   object.Permission  true        "The details of the permission"
// @Success 200 {object} controllers.Response The Response object
// @router /delete-permission [post]
func (c *ApiController) DeletePermission() {
	var permission object.Permission
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &permission)
	if err != nil {
		panic(err)
	}

	c.Data["json"] = wrapActionResponse(object.DeletePermission(&permission))
	c.ServeJSON()
}

func (c *ApiController) Enforce() {
	userId := c.GetSessionUsername()
	if userId == "" {
		c.ResponseError("Please sign in first")
	}

	var permissionRule object.PermissionRule
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &permissionRule)
	if err != nil {
		panic(err)
	}

	c.Data["json"] = object.Enforce(userId, &permissionRule)
	c.ServeJSON()
}

func (c *ApiController) BatchEnforce() {
	userId := c.GetSessionUsername()
	if userId == "" {
		c.ResponseError("Please sign in first")
	}

	var permissionRules []object.PermissionRule
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &permissionRules)
	if err != nil {
		panic(err)
	}

	c.Data["json"] = object.BatchEnforce(userId, permissionRules)
	c.ServeJSON()
}

func (c *ApiController) GetAllObjects() {
	userId := c.GetSessionUsername()
	if userId == "" {
		c.ResponseError("Please sign in first")
	}

	c.Data["json"] = object.GetAllObjects(userId)
	c.ServeJSON()
}

func (c *ApiController) GetAllActions() {
	userId := c.GetSessionUsername()
	if userId == "" {
		c.ResponseError("Please sign in first")
	}

	c.Data["json"] = object.GetAllActions(userId)
	c.ServeJSON()
}

func (c *ApiController) GetAllRoles() {
	userId := c.GetSessionUsername()
	if userId == "" {
		c.ResponseError("Please sign in first")
	}

	c.Data["json"] = object.GetAllRoles(userId)
	c.ServeJSON()
}
