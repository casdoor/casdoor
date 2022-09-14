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

	"github.com/astaxie/beego/utils/pagination"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
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
		c.Data["json"] = object.GetCasbinAdapters(owner)
		c.ServeJSON()
	} else {
		limit := util.ParseInt(limit)
		paginator := pagination.SetPaginator(c.Ctx, limit, int64(object.GetCasbinAdapterCount(owner, field, value)))
		adapters := object.GetPaginationCasbinAdapters(owner, paginator.Offset(), limit, field, value, sortField, sortOrder)
		c.ResponseOk(adapters, paginator.Nums())
	}
}

func (c *ApiController) GetCasbinAdapter() {
	id := c.Input().Get("id")
	c.Data["json"] = object.GetCasbinAdapter(id)
	c.ServeJSON()
}

func (c *ApiController) UpdateCasbinAdapter() {
	id := c.Input().Get("id")

	var casbinAdapter object.CasbinAdapter
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &casbinAdapter)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = wrapActionResponse(object.UpdateCasbinAdapter(id, &casbinAdapter))
	c.ServeJSON()
}

func (c *ApiController) AddCasbinAdapter() {
	var casbinAdapter object.CasbinAdapter
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &casbinAdapter)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = wrapActionResponse(object.AddCasbinAdapter(&casbinAdapter))
	c.ServeJSON()
}

func (c *ApiController) DeleteCasbinAdapter() {
	var casbinAdapter object.CasbinAdapter
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &casbinAdapter)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = wrapActionResponse(object.DeleteCasbinAdapter(&casbinAdapter))
	c.ServeJSON()
}

func (c *ApiController) SyncPolicies() {
	id := c.Input().Get("id")
	adapter := object.GetCasbinAdapter(id)

	c.Data["json"] = object.SyncPolicies(adapter)
	c.ServeJSON()
}
