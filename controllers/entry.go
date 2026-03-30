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

	"github.com/casdoor/casdoor/object"
)

// GetEntries
// @Title GetEntries
// @Tag Entry API
// @Description get entries by agent name
// @Param   agentName     query    string  false       "The name of the agent"
// @Param   owner         query    string  false       "The owner of the entries"
// @Success 200 {array} object.Entry The Response object
// @router /get-entries [get]
func (c *ApiController) GetEntries() {
	agentName := c.Ctx.Input.Query("agentName")
	owner := c.Ctx.Input.Query("owner")
	entries, err := object.GetEntries(owner, agentName)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(entries)
}

// UpdateEntry
// @Title UpdateEntry
// @Tag Entry API
// @Description update entry
// @Param   body    body   object.Entry  true        "The details of the entry"
// @Success 200 {object} controllers.Response The Response object
// @router /update-entry [post]
func (c *ApiController) UpdateEntry() {
	var entry object.Entry
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &entry)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if entry.Id == 0 {
		c.ResponseError(c.T("general:Missing parameter"))
		return
	}

	c.Data["json"] = wrapActionResponse(object.UpdateEntry(entry.Id, &entry))
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

	if entry.Id == 0 {
		c.ResponseError(c.T("general:Missing parameter"))
		return
	}

	c.Data["json"] = wrapActionResponse(object.DeleteEntry(&entry))
	c.ServeJSON()
}
