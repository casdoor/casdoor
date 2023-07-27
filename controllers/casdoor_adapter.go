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

// GetCasdoorAdapters
// @Title GetCasdoorAdapters
// @Tag Adapter API
// @Description get adapters
// @Param   owner     query    string  true        "The owner of adapters"
// @Success 200 {array} object.CasdoorAdapter The Response object
// @router /get-adapters [get]
func (c *ApiController) GetCasdoorAdapters() {
	owner := c.Input().Get("owner")
	limit := c.Input().Get("pageSize")
	page := c.Input().Get("p")
	field := c.Input().Get("field")
	value := c.Input().Get("value")
	sortField := c.Input().Get("sortField")
	sortOrder := c.Input().Get("sortOrder")

	if limit == "" || page == "" {
		adapters, err := object.GetCasdoorAdapters(owner)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		c.ResponseOk(adapters)
	} else {
		limit := util.ParseInt(limit)
		count, err := object.GetCasdoorAdapterCount(owner, field, value)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		paginator := pagination.SetPaginator(c.Ctx, limit, count)
		adapters, err := object.GetPaginationCasdoorAdapters(owner, paginator.Offset(), limit, field, value, sortField, sortOrder)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		c.ResponseOk(adapters, paginator.Nums())
	}
}

// GetCasdoorAdapter
// @Title GetCasdoorAdapter
// @Tag Adapter API
// @Description get adapter
// @Param   id     query    string  true        "The id ( owner/name ) of the adapter"
// @Success 200 {object} object.CasdoorAdapter The Response object
// @router /get-adapter [get]
func (c *ApiController) GetCasdoorAdapter() {
	id := c.Input().Get("id")

	adapter, err := object.GetCasdoorAdapter(id)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(adapter)
}

// UpdateCasdoorAdapter
// @Title UpdateCasdoorAdapter
// @Tag Adapter API
// @Description update adapter
// @Param   id     query    string  true        "The id ( owner/name ) of the adapter"
// @Param   body    body   object.CasdoorAdapter  true        "The details of the adapter"
// @Success 200 {object} controllers.Response The Response object
// @router /update-adapter [post]
func (c *ApiController) UpdateCasdoorAdapter() {
	id := c.Input().Get("id")

	var casbinAdapter object.CasdoorAdapter
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &casbinAdapter)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = wrapActionResponse(object.UpdateCasdoorAdapter(id, &casbinAdapter))
	c.ServeJSON()
}

// AddCasdoorAdapter
// @Title AddCasdoorAdapter
// @Tag Adapter API
// @Description add adapter
// @Param   body    body   object.CasdoorAdapter  true        "The details of the adapter"
// @Success 200 {object} controllers.Response The Response object
// @router /add-adapter [post]
func (c *ApiController) AddCasdoorAdapter() {
	var casbinAdapter object.CasdoorAdapter
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &casbinAdapter)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = wrapActionResponse(object.AddCasdoorAdapter(&casbinAdapter))
	c.ServeJSON()
}

// DeleteCasdoorAdapter
// @Title DeleteCasdoorAdapter
// @Tag Adapter API
// @Description delete adapter
// @Param   body    body   object.CasdoorAdapter  true        "The details of the adapter"
// @Success 200 {object} controllers.Response The Response object
// @router /delete-adapter [post]
func (c *ApiController) DeleteCasdoorAdapter() {
	var casbinAdapter object.CasdoorAdapter
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &casbinAdapter)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = wrapActionResponse(object.DeleteCasdoorAdapter(&casbinAdapter))
	c.ServeJSON()
}

func (c *ApiController) SyncPolicies() {
	id := c.Input().Get("id")
	adapter, err := object.GetCasdoorAdapter(id)
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
	adapter, err := object.GetCasdoorAdapter(id)
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
	c.Data["json"] = wrapActionResponse(affected)
	c.ServeJSON()
}

func (c *ApiController) AddPolicy() {
	id := c.Input().Get("id")
	adapter, err := object.GetCasdoorAdapter(id)
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
	c.Data["json"] = wrapActionResponse(affected)
	c.ServeJSON()
}

func (c *ApiController) RemovePolicy() {
	id := c.Input().Get("id")
	adapter, err := object.GetCasdoorAdapter(id)
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
	c.Data["json"] = wrapActionResponse(affected)
	c.ServeJSON()
}
