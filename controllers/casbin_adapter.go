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

func (c *ApiController) GetCasbinAdapters() {
	owner := c.Input().Get("owner")
	limit := c.Input().Get("pageSize")
	page := c.Input().Get("p")
	field := c.Input().Get("field")
	value := c.Input().Get("value")
	sortField := c.Input().Get("sortField")
	sortOrder := c.Input().Get("sortOrder")
	if limit == "" || page == "" {
		adapters := object.GetCasbinAdapters(owner)
		c.ResponseOk(adapters)
	} else {
		limit := util.ParseInt(limit)
		paginator := pagination.SetPaginator(c.Ctx, limit, int64(object.GetCasbinAdapterCount(owner, field, value)))
		adapters := object.GetPaginationCasbinAdapters(owner, paginator.Offset(), limit, field, value, sortField, sortOrder)
		c.ResponseOk(adapters, paginator.Nums())
	}
}

func (c *ApiController) GetCasbinAdapter() {
	id := c.Input().Get("id")
	adapter := object.GetCasbinAdapter(id)
	c.ResponseOk(adapter)
}

func (c *ApiController) UpdateCasbinAdapter() {
	id := c.Input().Get("id")

	var casbinAdapter object.CasbinAdapter
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &casbinAdapter)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	response := wrapActionResponse(object.UpdateCasbinAdapter(id, &casbinAdapter))
	c.ResponseOk(response)
}

func (c *ApiController) AddCasbinAdapter() {
	var casbinAdapter object.CasbinAdapter
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &casbinAdapter)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	response := wrapActionResponse(object.AddCasbinAdapter(&casbinAdapter))
	c.ResponseOk(response)
}

func (c *ApiController) DeleteCasbinAdapter() {
	var casbinAdapter object.CasbinAdapter
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &casbinAdapter)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	response := wrapActionResponse(object.DeleteCasbinAdapter(&casbinAdapter))
	c.ResponseOk(response)
}

func (c *ApiController) SyncPolicies() {
	id := c.Input().Get("id")
	adapter := object.GetCasbinAdapter(id)

	policies, err := object.SyncPolicies(adapter)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(policies)
}

func (c *ApiController) UpdatePolicy() {
	id := c.Input().Get("id")
	adapter := object.GetCasbinAdapter(id)
	var policies []xormadapter.CasbinRule
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &policies)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	affected, err := object.UpdatePolicy(util.CasbinToSlice(policies[0]), util.CasbinToSlice(policies[1]), adapter)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	response := wrapActionResponse(affected)
	c.ResponseOk(response)
}

func (c *ApiController) AddPolicy() {
	id := c.Input().Get("id")
	adapter := object.GetCasbinAdapter(id)
	var policy xormadapter.CasbinRule
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &policy)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	affected, err := object.AddPolicy(util.CasbinToSlice(policy), adapter)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	response := wrapActionResponse(affected)
	c.ResponseOk(response)
}

func (c *ApiController) RemovePolicy() {
	id := c.Input().Get("id")
	adapter := object.GetCasbinAdapter(id)
	var policy xormadapter.CasbinRule
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &policy)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	affected, err := object.RemovePolicy(util.CasbinToSlice(policy), adapter)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	response := wrapActionResponse(affected)
	c.ResponseOk(response)
}
