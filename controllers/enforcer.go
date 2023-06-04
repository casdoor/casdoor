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
	"encoding/json"

	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

// Enforce
// @Title Enforce
// @Tag Enforce API
// @Description perform enforce
// @Param   body    body   object.CasbinRequest  true   "casbin request"
// @Param   permissionId    query   string  false   "permission id"
// @Param   modelId    query   string  false   "model id
// @Param   resourceId    query   string  false   "resource id
// @Success 200 {object} controllers.Response The Response object
// @router /enforce [post]
func (c *ApiController) Enforce() {
	permissionId := c.Input().Get("permissionId")
	modelId := c.Input().Get("modelId")
	resourceId := c.Input().Get("resourceId")

	if util.IsAllStringEmpty(permissionId, modelId, resourceId) {
		c.ResponseError(c.T("general:Missing parameter"))
		return
	}

	var request object.CasbinRequest
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &request)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if permissionId != "" {
		c.ResponseOk(object.Enforce(permissionId, &request))
		return
	}

	permissions := make([]*object.Permission, 0)
	res := []bool{}

	if modelId != "" {
		owner, modelName := util.GetOwnerAndNameFromId(modelId)
		permissions, err = object.GetPermissionsByModel(owner, modelName)
		if err != nil {
			c.ResponseError("Permissions for the model could not be found")
		}
	}

	if resourceId != "" {
		permissions, err = object.GetPermissionsByResource(resourceId)
		if err != nil {
			c.ResponseError("Permissions for the resource could not be found")
		}
	}

	for _, permission := range permissions {
		res = append(res, object.Enforce(permission.GetId(), &request))
	}
	c.ResponseOk(res)
}

// BatchEnforce
// @Title BatchEnforce
// @Tag Enforce API
// @Description perform enforce
// @Param   body    body   object.CasbinRequest  true   "casbin request array"
// @Param   permissionId    query   string  false   "permission id"
// @Param   modelId    query   string  false   "model id
// @Success 200 {object} controllers.Response The Response object
// @router /batch-enforce [post]
func (c *ApiController) BatchEnforce() {
	permissionId := c.Input().Get("permissionId")
	modelId := c.Input().Get("modelId")

	if util.IsAllStringEmpty(permissionId, modelId) {
		c.ResponseError(c.T("general:Missing parameter"))
		return
	}

	var requests []object.CasbinRequest
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &requests)
	if err != nil {
		c.ResponseError(err.Error())
	}

	if permissionId != "" {
		c.ResponseOk(object.BatchEnforce(permissionId, &requests))
	}

	if modelId != "" {
		owner, modelName := util.GetOwnerAndNameFromId(modelId)
		permissions, err := object.GetPermissionsByModel(owner, modelName)
		if err != nil {
			c.ResponseError("Permissions for the model could not be found")
		}

		res := [][]bool{}
		for _, permission := range permissions {
			res = append(res, object.BatchEnforce(permission.GetId(), &requests))
		}

		c.ResponseOk(res)
	}
}

func (c *ApiController) GetAllObjects() {
	userId := c.GetSessionUsername()
	if userId == "" {
		c.ResponseError(c.T("general:Please login first"))
		return
	}

	c.ResponseOk(object.GetAllObjects(userId))
}

func (c *ApiController) GetAllActions() {
	userId := c.GetSessionUsername()
	if userId == "" {
		c.ResponseError(c.T("general:Please login first"))
		return
	}

	c.ResponseOk(object.GetAllActions(userId))
}

func (c *ApiController) GetAllRoles() {
	userId := c.GetSessionUsername()
	if userId == "" {
		c.ResponseError(c.T("general:Please login first"))
		return
	}

	c.ResponseOk(object.GetAllRoles(userId))
}
