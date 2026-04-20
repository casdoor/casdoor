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

// GetEntries
// @Title GetEntries
// @Tag Entry API
// @Description get entries
// @Param   owner     query    string  true        "The owner of entries"
// @Success 200 {array} object.Entry The Response object
// @router /get-entries [get]
func (c *ApiController) GetEntries() {
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
		entries, err := object.GetEntries(owner)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}
		c.ResponseOk(entries)
		return
	}

	limitInt := util.ParseInt(limit)
	count, err := object.GetEntryCount(owner, field, value)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	paginator := pagination.SetPaginator(c.Ctx, limitInt, count)
	entries, err := object.GetPaginationEntries(owner, paginator.Offset(), limitInt, field, value, sortField, sortOrder)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(entries, paginator.Nums())
}

// GetEntry
// @Title GetEntry
// @Tag Entry API
// @Description get entry
// @Param   id     query    string  true        "The id ( owner/name ) of the entry"
// @Success 200 {object} object.Entry The Response object
// @router /get-entry [get]
func (c *ApiController) GetEntry() {
	id := c.Ctx.Input.Query("id")

	entry, err := object.GetEntry(id)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(entry)
}

// GetOpenClawSessionGraph
// @Title GetOpenClawSessionGraph
// @Tag Entry API
// @Description get OpenClaw session graph
// @Param   id     query    string  true        "The id ( owner/name ) of the entry"
// @Success 200 {object} object.OpenClawSessionGraph The Response object
// @router /get-openclaw-session-graph [get]
func (c *ApiController) GetOpenClawSessionGraph() {
	id := c.Ctx.Input.Query("id")

	graph, err := object.GetOpenClawSessionGraph(id)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(graph)
}

// GetOpenClawSessionTranscript
// @Title GetOpenClawSessionTranscript
// @Tag Entry API
// @Description get OpenClaw raw session transcript
// @Param   id     query    string  true        "The id ( owner/name ) of the entry"
// @Success 200 {object} object.OpenClawSessionTranscriptPreview The raw JSONL transcript preview
// @router /get-openclaw-session-transcript [get]
func (c *ApiController) GetOpenClawSessionTranscript() {
	id := c.Ctx.Input.Query("id")

	preview, err := object.GetOpenClawSessionTranscript(id, c.GetAcceptLanguage())
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	if preview == nil {
		c.ResponseError("OpenClaw raw session transcript is not found")
		return
	}

	c.ResponseOk(preview)
}

// UpdateEntry
// @Title UpdateEntry
// @Tag Entry API
// @Description update entry
// @Param   id     query    string  true        "The id ( owner/name ) of the entry"
// @Param   body    body   object.Entry  true        "The details of the entry"
// @Success 200 {object} controllers.Response The Response object
// @router /update-entry [post]
func (c *ApiController) UpdateEntry() {
	id := c.Ctx.Input.Query("id")

	var entry object.Entry
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &entry)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = wrapActionResponse(object.UpdateEntry(id, &entry))
	c.ServeJSON()
}

// AddEntry
// @Title AddEntry
// @Tag Entry API
// @Description add entry
// @Param   body    body   object.Entry  true        "The details of the entry"
// @Success 200 {object} controllers.Response The Response object
// @router /add-entry [post]
func (c *ApiController) AddEntry() {
	var entry object.Entry
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &entry)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = wrapActionResponse(object.AddEntry(&entry))
	c.ServeJSON()
}

// DeleteEntry
// @Title DeleteEntry
// @Tag Entry API
// @Description delete entry
// @Param   body    body   object.Entry  true        "The details of the entry"
// @Success 200 {object} controllers.Response The Response object
// @router /delete-entry [post]
func (c *ApiController) DeleteEntry() {
	var entry object.Entry
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &entry)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = wrapActionResponse(object.DeleteEntry(&entry))
	c.ServeJSON()
}
