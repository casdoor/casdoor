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

	"github.com/beego/beego/v2/server/web/pagination"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

// GetAgents
// @Title GetAgents
// @Tag Agent API
// @Description get agents
// @Param   owner     query    string  true        "The owner of agents"
// @Success 200 {array} object.Agent The Response object
// @router /get-agents [get]
func (c *ApiController) GetAgents() {
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
		agents, err := object.GetAgents(owner)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}
		c.ResponseOk(agents)
		return
	}

	limitInt := util.ParseInt(limit)
	count, err := object.GetAgentCount(owner, field, value)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	paginator := pagination.SetPaginator(c.Ctx, limitInt, count)
	agents, err := object.GetPaginationAgents(owner, paginator.Offset(), limitInt, field, value, sortField, sortOrder)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(agents, paginator.Nums())
}

// GetAgent
// @Title GetAgent
// @Tag Agent API
// @Description get agent
// @Param   id     query    string  true        "The id ( owner/name ) of the agent"
// @Success 200 {object} object.Agent The Response object
// @router /get-agent [get]
func (c *ApiController) GetAgent() {
	id := c.Ctx.Input.Query("id")

	agent, err := object.GetAgent(id)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(agent)
}

// UpdateAgent
// @Title UpdateAgent
// @Tag Agent API
// @Description update agent
// @Param   id     query    string  true        "The id ( owner/name ) of the agent"
// @Param   body    body   object.Agent  true        "The details of the agent"
// @Success 200 {object} controllers.Response The Response object
// @router /update-agent [post]
func (c *ApiController) UpdateAgent() {
	id := c.Ctx.Input.Query("id")

	var agent object.Agent
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &agent)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = wrapActionResponse(object.UpdateAgent(id, &agent))
	c.ServeJSON()
}

// AddAgent
// @Title AddAgent
// @Tag Agent API
// @Description add agent
// @Param   body    body   object.Agent  true        "The details of the agent"
// @Success 200 {object} controllers.Response The Response object
// @router /add-agent [post]
func (c *ApiController) AddAgent() {
	var agent object.Agent
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &agent)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = wrapActionResponse(object.AddAgent(&agent))
	c.ServeJSON()
}

// DeleteAgent
// @Title DeleteAgent
// @Tag Agent API
// @Description delete agent
// @Param   body    body   object.Agent  true        "The details of the agent"
// @Success 200 {object} controllers.Response The Response object
// @router /delete-agent [post]
func (c *ApiController) DeleteAgent() {
	var agent object.Agent
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &agent)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = wrapActionResponse(object.DeleteAgent(&agent))
	c.ServeJSON()
}
