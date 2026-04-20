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
	"fmt"

	"github.com/beego/beego/v2/server/web/pagination"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

// GetServers
// @Title GetServers
// @Tag Server API
// @Description get servers
// @Param   owner     query    string  true        "The owner of servers"
// @Success 200 {array} object.Server The Response object
// @router /get-servers [get]
func (c *ApiController) GetServers() {
	owner := c.Ctx.Input.Query("owner")
	if owner == "admin" {
		owner = ""
	}

	limit := c.Ctx.Input.Query("pageSize")
	page := c.Ctx.Input.Query("p")
	field := c.Ctx.Input.Query("field")
	value := c.Ctx.Input.Query("value")
	sortField := c.Ctx.Input.Query("sortField")
	sortOrder := c.Ctx.Input.Query("sortOrder")

	if limit == "" || page == "" {
		servers, err := object.GetServers(owner)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}
		c.ResponseOk(servers)
		return
	}

	limitInt := util.ParseInt(limit)
	count, err := object.GetServerCount(owner, field, value)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	paginator := pagination.SetPaginator(c.Ctx, limitInt, count)
	servers, err := object.GetPaginationServers(owner, paginator.Offset(), limitInt, field, value, sortField, sortOrder)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(servers, paginator.Nums())
}

// GetServer
// @Title GetServer
// @Tag Server API
// @Description get server
// @Param   id     query    string  true        "The id ( owner/name ) of the server"
// @Success 200 {object} object.Server The Response object
// @router /get-server [get]
func (c *ApiController) GetServer() {
	id := c.Ctx.Input.Query("id")

	server, err := object.GetServer(id)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(server)
}

// UpdateServer
// @Title UpdateServer
// @Tag Server API
// @Description update server
// @Param   id     query    string  true        "The id ( owner/name ) of the server"
// @Param   body    body   object.Server  true        "The details of the server"
// @Success 200 {object} controllers.Response The Response object
// @router /update-server [post]
func (c *ApiController) UpdateServer() {
	id := c.Ctx.Input.Query("id")

	var server object.Server
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &server)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = wrapActionResponse(object.UpdateServer(id, &server))
	c.ServeJSON()
}

// SyncMcpTool
// @Title SyncMcpTool
// @Tag Server API
// @Description sync MCP tools for a server and return sync errors directly
// @Param   id     query    string  true        "The id ( owner/name ) of the server"
// @Param   isCleared  query    bool    false       "Whether to clear all tools instead of syncing"
// @Param   body    body   object.Server  true        "The details of the server"
// @Success 200 {object} controllers.Response The Response object
// @router /sync-mcp-tool [post]
func (c *ApiController) SyncMcpTool() {
	id := c.Ctx.Input.Query("id")
	isCleared := c.Ctx.Input.Query("isCleared") == "1"

	var server object.Server
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &server)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = wrapActionResponse(object.SyncMcpTool(id, &server, isCleared))
	c.ServeJSON()
}

// AddServer
// @Title AddServer
// @Tag Server API
// @Description add server
// @Param   body    body   object.Server  true        "The details of the server"
// @Success 200 {object} controllers.Response The Response object
// @router /add-server [post]
func (c *ApiController) AddServer() {
	var server object.Server
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &server)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = wrapActionResponse(object.AddServer(&server))
	c.ServeJSON()
}

// DeleteServer
// @Title DeleteServer
// @Tag Server API
// @Description delete server
// @Param   body    body   object.Server  true        "The details of the server"
// @Success 200 {object} controllers.Response The Response object
// @router /delete-server [post]
func (c *ApiController) DeleteServer() {
	var server object.Server
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &server)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = wrapActionResponse(object.DeleteServer(&server))
	c.ServeJSON()
}

// GetMcpAccessToken
// @Title GetMcpAccessToken
// @Tag Server API
// @Description get an access token for the current session user to use with an MCP server
// @Param   owner            query  string  true  "The owner of the application"
// @Param   applicationName  query  string  true  "The name of the application"
// @Success 200 {object} controllers.Response The Response object
// @router /get-mcp-access-token [get]
func (c *ApiController) GetMcpAccessToken() {
	owner := c.Ctx.Input.Query("owner")
	applicationName := c.Ctx.Input.Query("applicationName")

	user := c.getCurrentUser()
	if user == nil {
		c.ResponseError("user is not logged in")
		return
	}

	application, err := object.GetApplication(fmt.Sprintf("%s/%s", owner, applicationName))
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	if application == nil {
		c.ResponseError(fmt.Sprintf("application %s/%s does not exist", owner, applicationName))
		return
	}

	token, err := object.GetTokenByUser(application, user, "read", "", c.Ctx.Request.Host)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(token.AccessToken)
}
