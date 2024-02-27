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
	"fmt"

	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

// Enforce
// @Title Enforce
// @Tag Enforcer API
// @Description Call Casbin Enforce API
// @Param   body    body   []string  true   "Casbin request"
// @Param   permissionId    query   string  false   "permission id"
// @Param   modelId    query   string  false   "model id"
// @Param   resourceId    query   string  false   "resource id"
// @Param   owner    query   string  false   "owner"
// @Success 200 {object} controllers.Response The Response object
// @router /enforce [post]
func (c *ApiController) Enforce() {
	permissionId := c.Input().Get("permissionId")
	modelId := c.Input().Get("modelId")
	resourceId := c.Input().Get("resourceId")
	enforcerId := c.Input().Get("enforcerId")
	owner := c.Input().Get("owner")

	if len(c.Ctx.Input.RequestBody) == 0 {
		c.ResponseError("The request body should not be empty")
		return
	}

	var request []string
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &request)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if enforcerId != "" {
		enforcer, err := object.GetInitializedEnforcer(enforcerId)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		res := []bool{}
		keyRes := []string{}

		// type transformation
		interfaceRequest := util.StringToInterfaceArray(request)

		enforceResult, err := enforcer.Enforce(interfaceRequest...)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		res = append(res, enforceResult)
		keyRes = append(keyRes, enforcer.GetModelAndAdapter())

		c.ResponseOk(res, keyRes)
		return
	}

	if permissionId != "" {
		permission, err := object.GetPermission(permissionId)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}
		if permission == nil {
			c.ResponseError(fmt.Sprintf(c.T("permission:The permission: \"%s\" doesn't exist"), permissionId))
			return
		}

		res := []bool{}
		keyRes := []string{}

		enforceResult, err := object.Enforce(permission, request)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		res = append(res, enforceResult)
		keyRes = append(keyRes, permission.GetModelAndAdapter())

		c.ResponseOk(res, keyRes)
		return
	}

	permissions := []*object.Permission{}
	if modelId != "" {
		owner, modelName := util.GetOwnerAndNameFromId(modelId)
		permissions, err = object.GetPermissionsByModel(owner, modelName)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}
	} else if resourceId != "" {
		permissions, err = object.GetPermissionsByResource(resourceId)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}
	} else if owner != "" {
		permissions, err = object.GetPermissions(owner)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}
	} else {
		c.ResponseError(c.T("general:Missing parameter"))
		return
	}

	res := []bool{}
	keyRes := []string{}
	listPermissionIdMap := object.GroupPermissionsByModelAdapter(permissions)
	for key, permissionIds := range listPermissionIdMap {
		firstPermission, err := object.GetPermission(permissionIds[0])
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		enforceResult, err := object.Enforce(firstPermission, request, permissionIds...)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		res = append(res, enforceResult)
		keyRes = append(keyRes, key)
	}

	c.ResponseOk(res, keyRes)
}

// BatchEnforce
// @Title BatchEnforce
// @Tag Enforcer API
// @Description Call Casbin BatchEnforce API
// @Param   body    body   []string  true   "array of casbin requests"
// @Param   permissionId    query   string  false   "permission id"
// @Param   modelId    query   string  false   "model id"
// @Param   owner    query   string  false   "owner"
// @Success 200 {object} controllers.Response The Response object
// @router /batch-enforce [post]
func (c *ApiController) BatchEnforce() {
	permissionId := c.Input().Get("permissionId")
	modelId := c.Input().Get("modelId")
	enforcerId := c.Input().Get("enforcerId")
	owner := c.Input().Get("owner")

	var requests [][]string
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &requests)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if enforcerId != "" {
		enforcer, err := object.GetInitializedEnforcer(enforcerId)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		res := [][]bool{}
		keyRes := []string{}

		// type transformation
		interfaceRequests := util.StringToInterfaceArray2d(requests)

		enforceResult, err := enforcer.BatchEnforce(interfaceRequests)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		res = append(res, enforceResult)
		keyRes = append(keyRes, enforcer.GetModelAndAdapter())

		c.ResponseOk(res, keyRes)
		return
	}

	if permissionId != "" {
		permission, err := object.GetPermission(permissionId)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}
		if permission == nil {
			c.ResponseError(fmt.Sprintf(c.T("permission:The permission: \"%s\" doesn't exist"), permissionId))
			return
		}

		res := [][]bool{}
		keyRes := []string{}

		enforceResult, err := object.BatchEnforce(permission, requests)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		res = append(res, enforceResult)
		keyRes = append(keyRes, permission.GetModelAndAdapter())

		c.ResponseOk(res, keyRes)
		return
	}

	permissions := []*object.Permission{}
	if modelId != "" {
		owner, modelName := util.GetOwnerAndNameFromId(modelId)
		permissions, err = object.GetPermissionsByModel(owner, modelName)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}
	} else if owner != "" {
		permissions, err = object.GetPermissions(owner)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}
	} else {
		c.ResponseError(c.T("general:Missing parameter"))
		return
	}

	res := [][]bool{}
	keyRes := []string{}
	listPermissionIdMap := object.GroupPermissionsByModelAdapter(permissions)
	for _, permissionIds := range listPermissionIdMap {
		firstPermission, err := object.GetPermission(permissionIds[0])
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		enforceResult, err := object.BatchEnforce(firstPermission, requests, permissionIds...)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		res = append(res, enforceResult)
		keyRes = append(keyRes, firstPermission.GetModelAndAdapter())
	}

	c.ResponseOk(res, keyRes)
}

func (c *ApiController) GetAllObjects() {
	userId := c.Input().Get("userId")
	if userId == "" {
		userId = c.GetSessionUsername()
		if userId == "" {
			c.ResponseError(c.T("general:Please login first"))
			return
		}
	}

	objects, err := object.GetAllObjects(userId)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(objects)
}

func (c *ApiController) GetAllActions() {
	userId := c.Input().Get("userId")
	if userId == "" {
		userId = c.GetSessionUsername()
		if userId == "" {
			c.ResponseError(c.T("general:Please login first"))
			return
		}
	}

	actions, err := object.GetAllActions(userId)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(actions)
}

func (c *ApiController) GetAllRoles() {
	userId := c.Input().Get("userId")
	if userId == "" {
		userId = c.GetSessionUsername()
		if userId == "" {
			c.ResponseError(c.T("general:Please login first"))
			return
		}
	}

	roles, err := object.GetAllRoles(userId)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(roles)
}
