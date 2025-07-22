// Copyright 2021 The Casdoor Authors. All Rights Reserved.
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
)

// GetSyncers
// @Title GetSyncers
// @Tag Syncer API
// @Description get syncers
// @Param   owner     query    string  true        "The owner of syncers"
// @Success 200 {array} object.Syncer The Response object
// @router /get-syncers [get]
func (c *ApiController) GetSyncers() {
	owner := c.Input().Get("owner")
	limit := c.Input().Get("pageSize")
	page := c.Input().Get("p")
	field := c.Input().Get("field")
	value := c.Input().Get("value")
	sortField := c.Input().Get("sortField")
	sortOrder := c.Input().Get("sortOrder")
	organization := c.Input().Get("organization")

	if limit == "" || page == "" {
		syncers, err := object.GetMaskedSyncers(object.GetOrganizationSyncers(owner, organization))
		if err != nil {
			c.ResponseErr(err)
			return
		}

		c.ResponseOk(syncers)
	} else {
		limit := util.ParseInt(limit)
		count, err := object.GetSyncerCount(owner, organization, field, value)
		if err != nil {
			c.ResponseErr(err)
			return
		}

		paginator := pagination.SetPaginator(c.Ctx, limit, count)
		syncers, err := object.GetMaskedSyncers(object.GetPaginationSyncers(owner, organization, paginator.Offset(), limit, field, value, sortField, sortOrder))
		if err != nil {
			c.ResponseErr(err)
			return
		}

		c.ResponseOk(syncers, paginator.Nums())
	}
}

// GetSyncer
// @Title GetSyncer
// @Tag Syncer API
// @Description get syncer
// @Param   id     query    string  true        "The id ( owner/name ) of the syncer"
// @Success 200 {object} object.Syncer The Response object
// @router /get-syncer [get]
func (c *ApiController) GetSyncer() {
	id := c.Input().Get("id")

	syncer, err := object.GetMaskedSyncer(object.GetSyncer(id))
	if err != nil {
		c.ResponseErr(err)
		return
	}

	c.ResponseOk(syncer)
}

// UpdateSyncer
// @Title UpdateSyncer
// @Tag Syncer API
// @Description update syncer
// @Param   id     query    string  true        "The id ( owner/name ) of the syncer"
// @Param   body    body   object.Syncer  true        "The details of the syncer"
// @Success 200 {object} controllers.Response The Response object
// @router /update-syncer [post]
func (c *ApiController) UpdateSyncer() {
	id := c.Input().Get("id")

	var syncer object.Syncer
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &syncer)
	if err != nil {
		c.ResponseErr(err)
		return
	}

	c.Data["json"] = wrapActionResponse(object.UpdateSyncer(id, &syncer))
	c.ServeJSON()
}

// AddSyncer
// @Title AddSyncer
// @Tag Syncer API
// @Description add syncer
// @Param   body    body   object.Syncer  true        "The details of the syncer"
// @Success 200 {object} controllers.Response The Response object
// @router /add-syncer [post]
func (c *ApiController) AddSyncer() {
	var syncer object.Syncer
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &syncer)
	if err != nil {
		c.ResponseErr(err)
		return
	}

	c.Data["json"] = wrapActionResponse(object.AddSyncer(&syncer))
	c.ServeJSON()
}

// DeleteSyncer
// @Title DeleteSyncer
// @Tag Syncer API
// @Description delete syncer
// @Param   body    body   object.Syncer  true        "The details of the syncer"
// @Success 200 {object} controllers.Response The Response object
// @router /delete-syncer [post]
func (c *ApiController) DeleteSyncer() {
	var syncer object.Syncer
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &syncer)
	if err != nil {
		c.ResponseErr(err)
		return
	}

	c.Data["json"] = wrapActionResponse(object.DeleteSyncer(&syncer))
	c.ServeJSON()
}

// RunSyncer
// @Title RunSyncer
// @Tag Syncer API
// @Description run syncer
// @Param   body    body   object.Syncer  true        "The details of the syncer"
// @Success 200 {object} controllers.Response The Response object
// @router /run-syncer [get]
func (c *ApiController) RunSyncer() {
	id := c.Input().Get("id")
	syncer, err := object.GetSyncer(id)
	if err != nil {
		c.ResponseErr(err)
		return
	}

	err = object.RunSyncer(syncer)
	if err != nil {
		c.ResponseErr(err)
		return
	}

	c.ResponseOk()
}

func (c *ApiController) TestSyncerDb() {
	var syncer object.Syncer
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &syncer)
	if err != nil {
		c.ResponseErr(err)
		return
	}

	err = object.TestSyncerDb(syncer)
	if err != nil {
		c.ResponseErr(err)
		return
	}

	c.ResponseOk()
}
