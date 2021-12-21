// Copyright 2021 The casbin Authors. All Rights Reserved.
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
	"github.com/casbin/casdoor/object"
	"github.com/casbin/casdoor/util"
)

// GetProviders
// @Title GetProviders
// @Tag Provider API
// @Description get providers
// @Param   owner     query    string  true        "The owner of providers"
// @Success 200 {array} object.Provider The Response object
// @router /get-providers [get]
func (c *ApiController) GetProviders() {
	owner := c.Input().Get("owner")
	limit := c.Input().Get("pageSize")
	page := c.Input().Get("p")
	field := c.Input().Get("field")
	value := c.Input().Get("value")
	sortField := c.Input().Get("sortField")
	sortOrder := c.Input().Get("sortOrder")
	if limit == "" || page == "" {
		c.Data["json"] = object.GetMaskedProviders(object.GetProviders(owner))
		c.ServeJSON()
	} else {
		limit := util.ParseInt(limit)
		paginator := pagination.SetPaginator(c.Ctx, limit, int64(object.GetProviderCount(owner, field, value)))
		providers := object.GetMaskedProviders(object.GetPaginationProviders(owner, paginator.Offset(), limit, field, value, sortField, sortOrder))
		c.ResponseOk(providers, paginator.Nums())
	}
}

// @Title GetProvider
// @Tag Provider API
// @Description get provider
// @Param   id    query    string  true        "The id of the provider"
// @Success 200 {object} object.Provider The Response object
// @router /get-provider [get]
func (c *ApiController) GetProvider() {
	id := c.Input().Get("id")

	c.Data["json"] = object.GetMaskedProvider(object.GetProvider(id))
	c.ServeJSON()
}

// @Title UpdateProvider
// @Tag Provider API
// @Description update provider
// @Param   id    query    string  true        "The id of the provider"
// @Param   body    body   object.Provider  true        "The details of the provider"
// @Success 200 {object} controllers.Response The Response object
// @router /update-provider [post]
func (c *ApiController) UpdateProvider() {
	id := c.Input().Get("id")

	var provider object.Provider
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &provider)
	if err != nil {
		panic(err)
	}

	c.Data["json"] = wrapActionResponse(object.UpdateProvider(id, &provider))
	c.ServeJSON()
}

// @Title AddProvider
// @Tag Provider API
// @Description add provider
// @Param   body    body   object.Provider  true        "The details of the provider"
// @Success 200 {object} controllers.Response The Response object
// @router /add-provider [post]
func (c *ApiController) AddProvider() {
	var provider object.Provider
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &provider)
	if err != nil {
		panic(err)
	}

	c.Data["json"] = wrapActionResponse(object.AddProvider(&provider))
	c.ServeJSON()
}

// @Title DeleteProvider
// @Tag Provider API
// @Description delete provider
// @Param   body    body   object.Provider  true        "The details of the provider"
// @Success 200 {object} controllers.Response The Response object
// @router /delete-provider [post]
func (c *ApiController) DeleteProvider() {
	var provider object.Provider
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &provider)
	if err != nil {
		panic(err)
	}

	c.Data["json"] = wrapActionResponse(object.DeleteProvider(&provider))
	c.ServeJSON()
}
