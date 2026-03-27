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

func parseEnforceRequestBody(requestBody []byte, requestTokens []string) ([]interface{}, error) {
	if len(requestBody) == 0 {
		return nil, fmt.Errorf("the request body should not be empty")
	}

	var rawBody interface{}
	if err := json.Unmarshal(requestBody, &rawBody); err != nil {
		return nil, err
	}

	switch typedBody := rawBody.(type) {
	case []interface{}:
		return typedBody, nil
	case map[string]interface{}:
		if len(requestTokens) == 0 {
			return nil, fmt.Errorf("the request model is missing request tokens")
		}

		request := make([]interface{}, 0, len(requestTokens))
		for _, token := range requestTokens {
			value, ok := typedBody[token]
			if !ok {
				return nil, fmt.Errorf("the request body is missing %q", token)
			}
			request = append(request, value)
		}
		return request, nil
	default:
		return nil, fmt.Errorf("the request body should be a JSON array or object")
	}
}

func parseBatchEnforceRequestBody(requestBody []byte, requestTokens []string) ([][]interface{}, error) {
	if len(requestBody) == 0 {
		return nil, fmt.Errorf("the request body should not be empty")
	}

	var rawBody interface{}
	if err := json.Unmarshal(requestBody, &rawBody); err != nil {
		return nil, err
	}

	items, ok := rawBody.([]interface{})
	if !ok {
		return nil, fmt.Errorf("the request body should be a JSON array")
	}

	requests := make([][]interface{}, 0, len(items))
	for i, item := range items {
		switch typedItem := item.(type) {
		case []interface{}:
			requests = append(requests, typedItem)
		case map[string]interface{}:
			if len(requestTokens) == 0 {
				return nil, fmt.Errorf("the request model is missing request tokens")
			}

			request := make([]interface{}, 0, len(requestTokens))
			for _, token := range requestTokens {
				value, ok := typedItem[token]
				if !ok {
					return nil, fmt.Errorf("the batch request at index %d is missing %q", i, token)
				}
				request = append(request, value)
			}
			requests = append(requests, request)
		default:
			return nil, fmt.Errorf("the batch request at index %d should be a JSON array or object", i)
		}
	}

	return requests, nil
}

// Enforce
// @Title Enforce
// @Tag Enforcer API
// @Description Call Casbin Enforce API
// @Param   body    body   []interface{}  true   "Casbin request (array of strings or JSON objects for ABAC)"
// @Param   permissionId    query   string  false   "permission id"
// @Param   modelId    query   string  false   "model id"
// @Param   resourceId    query   string  false   "resource id"
// @Param   owner    query   string  false   "owner"
// @Success 200 {object} controllers.Response The Response object
// @router /enforce [post]
func (c *ApiController) Enforce() {
	permissionId := c.Ctx.Input.Query("permissionId")
	modelId := c.Ctx.Input.Query("modelId")
	resourceId := c.Ctx.Input.Query("resourceId")
	enforcerId := c.Ctx.Input.Query("enforcerId")
	owner := c.Ctx.Input.Query("owner")

	params := []string{permissionId, modelId, resourceId, enforcerId, owner}
	nonEmpty := 0
	for _, param := range params {
		if param != "" {
			nonEmpty++
		}
	}
	if nonEmpty > 1 {
		c.ResponseError("Only one of the parameters (permissionId, modelId, resourceId, enforcerId, owner) should be provided")
		return
	}

	if enforcerId != "" {
		enforcer, err := object.GetInitializedEnforcer(enforcerId)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		requestTokens, err := object.GetRequestTokensFromModel(enforcer.GetModel())
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		request, err := parseEnforceRequestBody(c.Ctx.Input.RequestBody, requestTokens)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		res := []bool{}
		keyRes := []string{}

		// convert any JSON-encoded string elements to anonymous structs for ABAC support
		interfaceRequest := util.ConvertInterfaceArray(request)

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

		requestTokens, err := object.GetRequestTokensByPermission(permission)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		request, err := parseEnforceRequestBody(c.Ctx.Input.RequestBody, requestTokens)
		if err != nil {
			c.ResponseError(err.Error())
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

	var err error
	permissions := []*object.Permission{}
	if modelId != "" {
		owner, modelName, err := util.GetOwnerAndNameFromIdWithError(modelId)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}
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

		requestTokens, err := object.GetRequestTokensByPermission(firstPermission)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		request, err := parseEnforceRequestBody(c.Ctx.Input.RequestBody, requestTokens)
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
// @Param   body    body   [][]interface{}  true   "array of Casbin requests (each request is an array of strings or JSON objects for ABAC)"
// @Param   permissionId    query   string  false   "permission id"
// @Param   modelId    query   string  false   "model id"
// @Param   owner    query   string  false   "owner"
// @Success 200 {object} controllers.Response The Response object
// @router /batch-enforce [post]
func (c *ApiController) BatchEnforce() {
	permissionId := c.Ctx.Input.Query("permissionId")
	modelId := c.Ctx.Input.Query("modelId")
	enforcerId := c.Ctx.Input.Query("enforcerId")
	owner := c.Ctx.Input.Query("owner")

	params := []string{permissionId, modelId, enforcerId, owner}
	nonEmpty := 0
	for _, param := range params {
		if param != "" {
			nonEmpty++
		}
	}
	if nonEmpty > 1 {
		c.ResponseError("Only one of the parameters (permissionId, modelId, enforcerId, owner) should be provided")
		return
	}

	if enforcerId != "" {
		enforcer, err := object.GetInitializedEnforcer(enforcerId)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		requestTokens, err := object.GetRequestTokensFromModel(enforcer.GetModel())
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		requests, err := parseBatchEnforceRequestBody(c.Ctx.Input.RequestBody, requestTokens)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		res := [][]bool{}
		keyRes := []string{}

		// convert any JSON-encoded string elements to anonymous structs for ABAC support
		interfaceRequests := util.ConvertInterfaceArray2d(requests)

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

		requestTokens, err := object.GetRequestTokensByPermission(permission)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		requests, err := parseBatchEnforceRequestBody(c.Ctx.Input.RequestBody, requestTokens)
		if err != nil {
			c.ResponseError(err.Error())
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

	var err error
	permissions := []*object.Permission{}
	if modelId != "" {
		owner, modelName, err := util.GetOwnerAndNameFromIdWithError(modelId)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}
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

		requestTokens, err := object.GetRequestTokensByPermission(firstPermission)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		requests, err := parseBatchEnforceRequestBody(c.Ctx.Input.RequestBody, requestTokens)
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

// GetAllObjects
// @Title GetAllObjects
// @Tag Enforcer API
// @Description Get all objects for a user (Casbin API)
// @Param   userId    query   string  false   "user id like built-in/admin"
// @Success 200 {object} controllers.Response The Response object
// @router /get-all-objects [get]
func (c *ApiController) GetAllObjects() {
	userId := c.Ctx.Input.Query("userId")
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

// GetAllActions
// @Title GetAllActions
// @Tag Enforcer API
// @Description Get all actions for a user (Casbin API)
// @Param   userId    query   string  false   "user id like built-in/admin"
// @Success 200 {object} controllers.Response The Response object
// @router /get-all-actions [get]
func (c *ApiController) GetAllActions() {
	userId := c.Ctx.Input.Query("userId")
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

// GetAllRoles
// @Title GetAllRoles
// @Tag Enforcer API
// @Description Get all roles for a user (Casbin API)
// @Param   userId    query   string  false   "user id like built-in/admin"
// @Success 200 {object} controllers.Response The Response object
// @router /get-all-roles [get]
func (c *ApiController) GetAllRoles() {
	userId := c.Ctx.Input.Query("userId")
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
