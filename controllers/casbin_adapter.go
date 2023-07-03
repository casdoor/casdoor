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

	"github.com/beego/beego/utils/pagination"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
	xormadapter "github.com/casdoor/xorm-adapter/v3"
)

// GetCasbinAdapters
// @Title GetCasbinAdapters
// @Tag Adapter API
// @Description get adapters
// @Param   owner     query    string  true        "The owner of adapters"
// @Success 200 {array} object.Adapter The Response object
// @router /get-adapters [get]
func (c *ApiController) GetCasbinAdapters() {
	owner := c.Input().Get("owner")
	limit := c.Input().Get("pageSize")
	page := c.Input().Get("p")
	field := c.Input().Get("field")
	value := c.Input().Get("value")
	sortField := c.Input().Get("sortField")
	sortOrder := c.Input().Get("sortOrder")

	if limit == "" || page == "" {
		adapters, err := object.GetCasbinAdapters(owner)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		c.ResponseOk(adapters)
	} else {
		limit := util.ParseInt(limit)
		count, err := object.GetCasbinAdapterCount(owner, field, value)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		paginator := pagination.SetPaginator(c.Ctx, limit, count)
		adapters, err := object.GetPaginationCasbinAdapters(owner, paginator.Offset(), limit, field, value, sortField, sortOrder)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		c.ResponseOk(adapters, paginator.Nums())
	}
}

// GetCasbinAdapter
// @Title GetCasbinAdapter
// @Tag Adapter API
// @Description get adapter
// @Param   id     query    string  true        "The id ( owner/name ) of the adapter"
// @Success 200 {object} object.Adapter The Response object
// @router /get-adapter [get]
func (c *ApiController) GetCasbinAdapter() {
	id := c.Input().Get("id")

	adapter, err := object.GetCasbinAdapter(id)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(adapter)
}

// UpdateCasbinAdapter
// @Title UpdateCasbinAdapter
// @Tag Adapter API
// @Description update adapter
// @Param   id     query    string  true        "The id ( owner/name ) of the adapter"
// @Param   body    body   object.Adapter  true        "The details of the adapter"
// @Success 200 {object} controllers.Response The Response object
// @router /update-adapter [post]
func (c *ApiController) UpdateCasbinAdapter() {
	id := c.Input().Get("id")

	var casbinAdapter object.CasbinAdapter
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &casbinAdapter)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	resp := wrapActionResponse(object.UpdateCasbinAdapter(id, &casbinAdapter))
	c.ResponseOk(resp)
}

// AddCasbinAdapter
// @Title AddCasbinAdapter
// @Tag Adapter API
// @Description add adapter
// @Param   body    body   object.Adapter  true        "The details of the adapter"
// @Success 200 {object} controllers.Response The Response object
// @router /add-adapter [post]
func (c *ApiController) AddCasbinAdapter() {
	var casbinAdapter object.CasbinAdapter
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &casbinAdapter)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	resp := wrapActionResponse(object.AddCasbinAdapter(&casbinAdapter))
	c.ResponseOk(resp)
}

// DeleteCasbinAdapter
// @Title DeleteCasbinAdapter
// @Tag Adapter API
// @Description delete adapter
// @Param   body    body   object.Adapter  true        "The details of the adapter"
// @Success 200 {object} controllers.Response The Response object
// @router /delete-adapter [post]
func (c *ApiController) DeleteCasbinAdapter() {
	var casbinAdapter object.CasbinAdapter
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &casbinAdapter)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	resp := wrapActionResponse(object.DeleteCasbinAdapter(&casbinAdapter))
	c.ResponseOk(resp)
}

func (c *ApiController) SyncPolicies() {
	id := c.Input().Get("id")
	adapter, err := object.GetCasbinAdapter(id)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	policies, err := object.SyncPolicies(adapter)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(policies)
}

func (c *ApiController) UpdatePolicy() {
	id := c.Input().Get("id")
	adapter, err := object.GetCasbinAdapter(id)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	var policies []xormadapter.CasbinRule
	err = json.Unmarshal(c.Ctx.Input.RequestBody, &policies)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	affected, err := object.UpdatePolicy(util.CasbinToSlice(policies[0]), util.CasbinToSlice(policies[1]), adapter)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	resp := wrapActionResponse(affected)
	c.ResponseOk(resp)
}

func (c *ApiController) AddPolicy() {
	id := c.Input().Get("id")
	adapter, err := object.GetCasbinAdapter(id)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	var policy xormadapter.CasbinRule
	err = json.Unmarshal(c.Ctx.Input.RequestBody, &policy)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	affected, err := object.AddPolicy(util.CasbinToSlice(policy), adapter)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	resp := wrapActionResponse(affected)
	c.ResponseOk(resp)
}

func (c *ApiController) RemovePolicy() {
	id := c.Input().Get("id")
	adapter, err := object.GetCasbinAdapter(id)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	var policy xormadapter.CasbinRule
	err = json.Unmarshal(c.Ctx.Input.RequestBody, &policy)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	affected, err := object.RemovePolicy(util.CasbinToSlice(policy), adapter)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	resp := wrapActionResponse(affected)
	c.ResponseOk(resp)
}
