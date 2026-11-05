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
	"errors"

	"github.com/beego/beego/v2/core/utils/pagination"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

// GetRadiusClients
// @Title GetRadiusClients
// @Tag RADIUS Client API
// @Description get RADIUS clients
// @Param   owner     query    string  true        "The owner of RADIUS clients, always admin"
// @Success 200 {array} object.RadiusClient The Response object
// @router /get-radius-clients [get]
func (c *ApiController) GetRadiusClients() {
	if !c.requireGlobalAdmin() {
		return
	}

	limit := c.Ctx.Input.Query("pageSize")
	page := c.Ctx.Input.Query("p")
	field := c.Ctx.Input.Query("field")
	value := c.Ctx.Input.Query("value")
	sortField := c.Ctx.Input.Query("sortField")
	sortOrder := c.Ctx.Input.Query("sortOrder")

	if limit == "" || page == "" {
		clients, err := object.GetRadiusClients("admin")
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		c.ResponseOk(clients)
	} else {
		limit := util.ParseInt(limit)
		count, err := object.GetRadiusClientCount("admin", field, value)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		paginator := pagination.NewPaginator(c.Ctx.Request, limit, count)
		clients, err := object.GetPaginationRadiusClients("admin", paginator.Offset(), limit, field, value, sortField, sortOrder)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		c.ResponseOk(clients, paginator.Nums())
	}
}

// GetRadiusClient
// @Title GetRadiusClient
// @Tag RADIUS Client API
// @Description get RADIUS client
// @Param   id     query    string  true        "The id ( owner/name ) of the RADIUS client"
// @Success 200 {object} object.RadiusClient The Response object
// @router /get-radius-client [get]
func (c *ApiController) GetRadiusClient() {
	if !c.requireGlobalAdmin() {
		return
	}

	id := c.Ctx.Input.Query("id")
	client, err := object.GetRadiusClient(id)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(client)
}

// AddRadiusClient
// @Title AddRadiusClient
// @Tag RADIUS Client API
// @Description add RADIUS client
// @Param   body    body   object.RadiusClient  true        "The details of the RADIUS client"
// @Success 200 {object} controllers.Response The Response object
// @router /add-radius-client [post]
func (c *ApiController) AddRadiusClient() {
	if !c.requireGlobalAdmin() {
		return
	}

	currentTime := util.GetCurrentTime()
	client := object.RadiusClient{
		CreatedTime: currentTime,
		UpdatedTime: currentTime,
	}
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &client)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	err = checkRadiusClient(&client)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = wrapActionResponse(object.AddRadiusClient(&client))
	c.ServeJSON()
}

// UpdateRadiusClient
// @Title UpdateRadiusClient
// @Tag RADIUS Client API
// @Description update RADIUS client
// @Param   id     query    string  true        "The id ( owner/name ) of the RADIUS client"
// @Param   body    body   object.RadiusClient  true        "The details of the RADIUS client"
// @Success 200 {object} controllers.Response The Response object
// @router /update-radius-client [post]
func (c *ApiController) UpdateRadiusClient() {
	if !c.requireGlobalAdmin() {
		return
	}

	id := c.Ctx.Input.Query("id")

	var client object.RadiusClient
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &client)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	err = checkRadiusClient(&client)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = wrapActionResponse(object.UpdateRadiusClient(id, &client))
	c.ServeJSON()
}

// DeleteRadiusClient
// @Title DeleteRadiusClient
// @Tag RADIUS Client API
// @Description delete RADIUS client
// @Param   body    body   object.RadiusClient  true        "The details of the RADIUS client"
// @Success 200 {object} controllers.Response The Response object
// @router /delete-radius-client [post]
func (c *ApiController) DeleteRadiusClient() {
	if !c.requireGlobalAdmin() {
		return
	}

	var client object.RadiusClient
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &client)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = wrapActionResponse(object.DeleteRadiusClient(&client))
	c.ServeJSON()
}

func checkRadiusClient(client *object.RadiusClient) error {
	client.Owner = "admin"
	if client.Name == "" {
		return errors.New("name should not be empty")
	}
	if client.Secret == "" {
		return errors.New("secret should not be empty")
	}
	if client.Organization != "" {
		organization, err := object.GetOrganization(util.GetId("admin", client.Organization))
		if err != nil {
			return err
		}
		if organization == nil {
			return errors.New("organization does not exist: " + client.Organization)
		}
	}

	_, err := object.ParseRadiusClientIpRanges(client.IpRanges)
	return err
}
