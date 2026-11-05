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

// GetRadiusClients
// @Title GetRadiusClients
// @Tag RadiusClient API
// @Description get radius clients
// @Param   owner     query    string  true        "The owner of radius clients"
// @Success 200 {array} object.RadiusClient The Response object
// @router /get-radius-clients [get]
func (c *ApiController) GetRadiusClients() {
	owner := c.Ctx.Input.Query("owner")
	limit := c.Ctx.Input.Query("pageSize")
	page := c.Ctx.Input.Query("p")
	field := c.Ctx.Input.Query("field")
	value := c.Ctx.Input.Query("value")
	sortField := c.Ctx.Input.Query("sortField")
	sortOrder := c.Ctx.Input.Query("sortOrder")

	if limit == "" || page == "" {
		radiusClients, err := object.GetRadiusClients(owner)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		c.ResponseOk(radiusClients)
	} else {
		limit := util.ParseInt(limit)
		count, err := object.GetRadiusClientCount(owner, field, value)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		paginator := pagination.NewPaginator(c.Ctx.Request, limit, count)
		radiusClients, err := object.GetPaginationRadiusClients(owner, paginator.Offset(), limit, field, value, sortField, sortOrder)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		c.ResponseOk(radiusClients, paginator.Nums())
	}
}

// GetRadiusClient
// @Title GetRadiusClient
// @Tag RadiusClient API
// @Description get radius client
// @Param   id     query    string  true        "The id ( owner/name ) of the radius client"
// @Success 200 {object} object.RadiusClient The Response object
// @router /get-radius-client [get]
func (c *ApiController) GetRadiusClient() {
	id := c.Ctx.Input.Query("id")

	radiusClient, err := object.GetRadiusClient(id)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(radiusClient)
}

// UpdateRadiusClient
// @Title UpdateRadiusClient
// @Tag RadiusClient API
// @Description update radius client
// @Param   id     query    string  true        "The id ( owner/name ) of the radius client"
// @Param   body    body   object.RadiusClient  true        "The details of the radius client"
// @Success 200 {object} controllers.Response The Response object
// @router /update-radius-client [post]
func (c *ApiController) UpdateRadiusClient() {
	id := c.Ctx.Input.Query("id")

	var radiusClient object.RadiusClient
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &radiusClient)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = wrapActionResponse(object.UpdateRadiusClient(id, &radiusClient))
	c.ServeJSON()
}

// AddRadiusClient
// @Title AddRadiusClient
// @Tag RadiusClient API
// @Description add radius client
// @Param   body    body   object.RadiusClient  true        "The details of the radius client"
// @Success 200 {object} controllers.Response The Response object
// @router /add-radius-client [post]
func (c *ApiController) AddRadiusClient() {
	var radiusClient object.RadiusClient
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &radiusClient)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = wrapActionResponse(object.AddRadiusClient(&radiusClient))
	c.ServeJSON()
}

// DeleteRadiusClient
// @Title DeleteRadiusClient
// @Tag RadiusClient API
// @Description delete radius client
// @Param   body    body   object.RadiusClient  true        "The details of the radius client"
// @Success 200 {object} controllers.Response The Response object
// @router /delete-radius-client [post]
func (c *ApiController) DeleteRadiusClient() {
	var radiusClient object.RadiusClient
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &radiusClient)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = wrapActionResponse(object.DeleteRadiusClient(&radiusClient))
	c.ServeJSON()
}
