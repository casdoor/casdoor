// Copyright 2026 Casdoor Inc. All Rights Reserved.
//
// This file is part of the commercial and enterprise editions of Casdoor and is licensed under the Casdoor Enterprise
// License, not the Apache License 2.0. You may not use this file except in compliance with that license, see
// ee/LICENSE.

package controllers

import (
	"strings"

	"github.com/beego/beego/v2/server/web/pagination"
	"github.com/casdoor/casdoor/ee/mcptools"
	"github.com/casdoor/casdoor/util"
)

// GetMcpChanges
// @Title GetMcpChanges
// @Tag MCP API
// @Description get the changes that agents made through Casdoor's own MCP server in an organization
// @Param   owner    query string true  "The organization"
// @Param   agent    query string false "The agent ID"
// @Param   user     query string false "The user ID"
// @Param   kind     query string false "The kind of object, e.g. application"
// @Param   objectId query string false "The object ID"
// @Param   state    query string false "Applied or Rolled back"
// @Param   traceId  query string false "The trace ID"
// @Param   p        query string false "The page"
// @Param   pageSize query string false "The page size"
// @Success 200 {array} mcptools.McpChange The Response object
// @router /get-mcp-changes [get]
func (c *ApiController) GetMcpChanges() {
	owner := c.Ctx.Input.Query("owner")
	if !c.isAdminOfOwner(owner) {
		c.ResponseError(c.T("auth:Unauthorized operation"))
		return
	}

	filter := &mcptools.McpChangeFilter{
		Agent:    c.Ctx.Input.Query("agent"),
		User:     c.Ctx.Input.Query("user"),
		Kind:     c.Ctx.Input.Query("kind"),
		ObjectId: strings.TrimSpace(c.Ctx.Input.Query("objectId")),
		State:    c.Ctx.Input.Query("state"),
		TraceId:  strings.TrimSpace(c.Ctx.Input.Query("traceId")),
	}
	limit := util.ParseInt(c.Ctx.Input.Query("pageSize"))
	if limit <= 0 || limit > 500 {
		limit = 50
	}

	count, err := mcptools.GetMcpChangeCount(owner, filter)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	paginator := pagination.SetPaginator(c.Ctx, limit, count)
	changes, err := mcptools.GetPaginationMcpChanges(owner, filter, paginator.Offset(), limit)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	c.ResponseOk(changes, paginator.Nums())
}

// RollbackMcpChange
// @Title RollbackMcpChange
// @Tag MCP API
// @Description roll back a change that an agent made through Casdoor's own MCP server, or with wholeCall all the
// changes of its tool call
// @Param   id        query string true  "The ID of the change"
// @Param   wholeCall query string false "true to roll back all the changes of the change's tool call"
// @Success 200 {object} controllers.Response The Response object
// @router /rollback-mcp-change [post]
func (c *ApiController) RollbackMcpChange() {
	id := c.Ctx.Input.Query("id")
	owner, _, err := util.GetOwnerAndNameFromIdWithError(id)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	if !c.isAdminOfOwner(owner) {
		c.ResponseError(c.T("auth:Unauthorized operation"))
		return
	}

	if c.Ctx.Input.Query("wholeCall") == "true" {
		change, err := mcptools.GetMcpChange(id)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}
		if change == nil {
			c.ResponseError(c.T("general:The object does not exist"))
			return
		}
		err = mcptools.RollbackMcpCall(c.Ctx, c.GetSessionUsername(), owner, change.TraceId)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}
		c.ResponseOk()
		return
	}

	err = mcptools.RollbackMcpChange(c.Ctx, c.GetSessionUsername(), id)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	c.ResponseOk()
}
