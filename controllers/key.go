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

	"github.com/beego/beego/v2/core/utils/pagination"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

func (c *ApiController) GetKeys() {
	owner := c.Ctx.Input.Query("owner")
	limit := c.Ctx.Input.Query("pageSize")
	page := c.Ctx.Input.Query("p")
	field := c.Ctx.Input.Query("field")
	value := c.Ctx.Input.Query("value")
	sortField := c.Ctx.Input.Query("sortField")
	sortOrder := c.Ctx.Input.Query("sortOrder")
	keyType := c.Ctx.Input.Query("type")
	organization := c.Ctx.Input.Query("organization")
	application := c.Ctx.Input.Query("application")
	user := c.Ctx.Input.Query("user")

	if c.IsGlobalAdmin() {
		if limit == "" || page == "" {
			keys, err := object.GetMaskedKeys(object.GetKeys(owner, keyType, organization, application, user))
			if err != nil {
				c.ResponseError(err.Error())
				return
			}

			c.ResponseOk(keys)
		} else {
			limit := util.ParseInt(limit)
			count, err := object.GetKeyCount(owner, keyType, organization, application, user, field, value)
			if err != nil {
				c.ResponseError(err.Error())
				return
			}

			paginator := pagination.NewPaginator(c.Ctx.Request, limit, count)
			keys, err := object.GetMaskedKeys(object.GetPaginationKeys(owner, keyType, organization, application, user, paginator.Offset(), limit, field, value, sortField, sortOrder))
			if err != nil {
				c.ResponseError(err.Error())
				return
			}

			c.ResponseOk(keys, paginator.Nums())
		}
		return
	}

	keys, err := object.GetPaginationKeys(owner, keyType, organization, application, user, -1, -1, field, value, sortField, sortOrder)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	keys, err = c.filterAuthorizedKeys(keys)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	keys, err = object.GetMaskedKeys(keys)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if limit == "" || page == "" {
		c.ResponseOk(keys)
	} else {
		pageSize := util.ParseInt(limit)
		paginator := pagination.NewPaginator(c.Ctx.Request, pageSize, int64(len(keys)))
		start := paginator.Offset()
		if start > len(keys) {
			start = len(keys)
		}
		end := start + pageSize
		if end > len(keys) {
			end = len(keys)
		}

		c.ResponseOk(keys[start:end], paginator.Nums())
	}
}

func (c *ApiController) GetKey() {
	id := c.Ctx.Input.Query("id")
	key, err := object.GetKey(id)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	if key == nil {
		c.ResponseOk(nil)
		return
	}

	ok, err := c.canManageKey(key)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	if !ok {
		c.ResponseError(c.T("auth:Unauthorized operation"))
		return
	}

	key, err = object.GetMaskedKey(key)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(key)
}

func (c *ApiController) AddKey() {
	var key object.Key
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &key)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if err = object.CheckKey(&key, c.GetAcceptLanguage()); err != nil {
		c.ResponseError(err.Error())
		return
	}

	ok, err := c.canManageKey(&key)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	if !ok {
		c.ResponseError(c.T("auth:Unauthorized operation"))
		return
	}

	if !key.IsEnabled {
		key.IsEnabled = true
	}

	rawSecret := object.GenerateKeySecret()
	key.SetSecret(rawSecret)

	affected, err := object.AddKey(&key)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	if !affected {
		c.Data["json"] = wrapActionResponse(false)
		c.ServeJSON()
		return
	}

	key.SecretHash = ""
	c.ResponseOk(map[string]interface{}{
		"key":    &key,
		"apiKey": rawSecret,
	})
}

func (c *ApiController) UpdateKey() {
	id := c.Ctx.Input.Query("id")

	var key object.Key
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &key)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	oldKey, err := object.GetKey(id)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	if oldKey == nil {
		c.Data["json"] = wrapActionResponse(false)
		c.ServeJSON()
		return
	}

	ok, err := c.canManageKey(oldKey)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	if !ok {
		c.ResponseError(c.T("auth:Unauthorized operation"))
		return
	}

	key.SecretHash = oldKey.SecretHash
	key.SecretPreview = oldKey.SecretPreview
	key.CreatedTime = oldKey.CreatedTime

	if err = object.CheckKey(&key, c.GetAcceptLanguage()); err != nil {
		c.ResponseError(err.Error())
		return
	}
	ok, err = c.canManageKey(&key)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	if !ok {
		c.ResponseError(c.T("auth:Unauthorized operation"))
		return
	}

	affected, err := object.UpdateKey(id, &key)
	c.Data["json"] = wrapActionResponse(affected, err)
	c.ServeJSON()
}

func (c *ApiController) DeleteKey() {
	var req object.Key
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &req)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	key, err := object.GetKey(req.GetId())
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	if key == nil {
		c.Data["json"] = wrapActionResponse(false)
		c.ServeJSON()
		return
	}

	ok, err := c.canManageKey(key)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	if !ok {
		c.ResponseError(c.T("auth:Unauthorized operation"))
		return
	}

	affected, err := object.DeleteKey(key)
	c.Data["json"] = wrapActionResponse(affected, err)
	c.ServeJSON()
}

func (c *ApiController) RotateKey() {
	var req object.Key
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &req)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	key, err := object.GetKey(req.GetId())
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	if key == nil {
		c.Data["json"] = wrapActionResponse(false)
		c.ServeJSON()
		return
	}

	ok, err := c.canManageKey(key)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	if !ok {
		c.ResponseError(c.T("auth:Unauthorized operation"))
		return
	}

	rawSecret := object.GenerateKeySecret()
	key.SetSecret(rawSecret)
	key.LastUsedTime = ""

	affected, err := object.UpdateKey(key.GetId(), key)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	if !affected {
		c.Data["json"] = wrapActionResponse(false)
		c.ServeJSON()
		return
	}

	key.SecretHash = ""
	c.ResponseOk(map[string]interface{}{
		"key":    key,
		"apiKey": rawSecret,
	})
}
