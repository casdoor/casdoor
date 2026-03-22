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

// GetKeys
// @Title GetKeys
// @Tag Key API
// @Description get keys
// @Param   owner     query    string  true        "The owner of keys"
// @Success 200 {array} object.Key The Response object
// @router /get-keys [get]
func (c *ApiController) GetKeys() {
	owner := c.Ctx.Input.Query("owner")
	limit := c.Ctx.Input.Query("pageSize")
	page := c.Ctx.Input.Query("p")
	field := c.Ctx.Input.Query("field")
	value := c.Ctx.Input.Query("value")
	sortField := c.Ctx.Input.Query("sortField")
	sortOrder := c.Ctx.Input.Query("sortOrder")

	if limit == "" || page == "" {
		keys, err := object.GetKeys(owner)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}
		maskedKeys, err := object.GetMaskedKeys(keys, true, nil)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		c.ResponseOk(maskedKeys)
	} else {
		limit := util.ParseInt(limit)
		count, err := object.GetKeyCount(owner, field, value)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		paginator := pagination.NewPaginator(c.Ctx.Request, limit, count)
		keys, err := object.GetPaginationKeys(owner, paginator.Offset(), limit, field, value, sortField, sortOrder)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}
		maskedKeys, err := object.GetMaskedKeys(keys, true, nil)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		c.ResponseOk(maskedKeys, paginator.Nums())
	}
}

// GetGlobalKeys
// @Title GetGlobalKeys
// @Tag Key API
// @Description get global keys
// @Success 200 {array} object.Key The Response object
// @router /get-global-keys [get]
func (c *ApiController) GetGlobalKeys() {
	limit := c.Ctx.Input.Query("pageSize")
	page := c.Ctx.Input.Query("p")
	field := c.Ctx.Input.Query("field")
	value := c.Ctx.Input.Query("value")
	sortField := c.Ctx.Input.Query("sortField")
	sortOrder := c.Ctx.Input.Query("sortOrder")

	if limit == "" || page == "" {
		keys, err := object.GetGlobalKeys()
		if err != nil {
			c.ResponseError(err.Error())
			return
		}
		maskedKeys, err := object.GetMaskedKeys(keys, true, nil)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		c.ResponseOk(maskedKeys)
	} else {
		limit := util.ParseInt(limit)
		count, err := object.GetGlobalKeyCount(field, value)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		paginator := pagination.NewPaginator(c.Ctx.Request, limit, count)
		keys, err := object.GetPaginationGlobalKeys(paginator.Offset(), limit, field, value, sortField, sortOrder)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}
		maskedKeys, err := object.GetMaskedKeys(keys, true, nil)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		c.ResponseOk(maskedKeys, paginator.Nums())
	}
}

// GetKey
// @Title GetKey
// @Tag Key API
// @Description get key
// @Param   id     query    string  true        "The id ( owner/name ) of the key"
// @Success 200 {object} object.Key The Response object
// @router /get-key [get]
func (c *ApiController) GetKey() {
	id := c.Ctx.Input.Query("id")

	key, err := object.GetKey(id)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(key)
}

// UpdateKey
// @Title UpdateKey
// @Tag Key API
// @Description update key
// @Param   id     query    string  true        "The id ( owner/name ) of the key"
// @Param   body    body   object.Key  true        "The details of the key"
// @Success 200 {object} controllers.Response The Response object
// @router /update-key [post]
func (c *ApiController) UpdateKey() {
	id := c.Ctx.Input.Query("id")

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

	var key object.Key
	err = json.Unmarshal(c.Ctx.Input.RequestBody, &key)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if !c.IsGlobalAdmin() && oldKey.Owner != key.Owner {
		c.ResponseError(c.T("auth:Unauthorized operation"))
		return
	}

	c.Data["json"] = wrapActionResponse(object.UpdateKey(id, &key))
	c.ServeJSON()
}

// AddKey
// @Title AddKey
// @Tag Key API
// @Description add key
// @Param   body    body   object.Key  true        "The details of the key"
// @Success 200 {object} controllers.Response The Response object
// @router /add-key [post]
func (c *ApiController) AddKey() {
	var key object.Key
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &key)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = wrapActionResponse(object.AddKey(&key))
	c.ServeJSON()
}

// DeleteKey
// @Title DeleteKey
// @Tag Key API
// @Description delete key
// @Param   body    body   object.Key  true        "The details of the key"
// @Success 200 {object} controllers.Response The Response object
// @router /delete-key [post]
func (c *ApiController) DeleteKey() {
	var key object.Key
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &key)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = wrapActionResponse(object.DeleteKey(&key))
	c.ServeJSON()
}
