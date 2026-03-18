// Copyright 2023 The casbin Authors. All Rights Reserved.
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

// GetGlobalSites
// @Title GetGlobalSites
// @Tag Site API
// @Description get global sites
// @Success 200 {array} object.Site The Response object
// @router /get-global-sites [get]
func (c *ApiController) GetGlobalSites() {
	sites, err := object.GetGlobalSites()
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(object.GetMaskedSites(sites, util.GetHostname()))
}

// GetSites
// @Title GetSites
// @Tag Site API
// @Description get sites
// @Param   owner     query    string  true        "The owner of sites"
// @Success 200 {array} object.Site The Response object
// @router /get-sites [get]
func (c *ApiController) GetSites() {
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
		sites, err := object.GetSites(owner)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}
		c.ResponseOk(object.GetMaskedSites(sites, util.GetHostname()))
		return
	}

	limitInt := util.ParseInt(limit)
	count, err := object.GetSiteCount(owner, field, value)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	paginator := pagination.SetPaginator(c.Ctx, limitInt, count)
	sites, err := object.GetPaginationSites(owner, paginator.Offset(), limitInt, field, value, sortField, sortOrder)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(object.GetMaskedSites(sites, util.GetHostname()), paginator.Nums())
}

// GetSite
// @Title GetSite
// @Tag Site API
// @Description get site
// @Param   id     query    string  true        "The id ( owner/name ) of the site"
// @Success 200 {object} object.Site The Response object
// @router /get-site [get]
func (c *ApiController) GetSite() {
	id := c.Ctx.Input.Query("id")

	site, err := object.GetSite(id)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(object.GetMaskedSite(site, util.GetHostname()))
}

// UpdateSite
// @Title UpdateSite
// @Tag Site API
// @Description update site
// @Param   id     query    string  true        "The id ( owner/name ) of the site"
// @Param   body    body   object.Site  true        "The details of the site"
// @Success 200 {object} controllers.Response The Response object
// @router /update-site [post]
func (c *ApiController) UpdateSite() {
	id := c.Ctx.Input.Query("id")

	var site object.Site
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &site)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = wrapActionResponse(object.UpdateSite(id, &site))
	c.ServeJSON()
}

// AddSite
// @Title AddSite
// @Tag Site API
// @Description add site
// @Param   body    body   object.Site  true        "The details of the site"
// @Success 200 {object} controllers.Response The Response object
// @router /add-site [post]
func (c *ApiController) AddSite() {
	var site object.Site
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &site)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = wrapActionResponse(object.AddSite(&site))
	c.ServeJSON()
}

// DeleteSite
// @Title DeleteSite
// @Tag Site API
// @Description delete site
// @Param   body    body   object.Site  true        "The details of the site"
// @Success 200 {object} controllers.Response The Response object
// @router /delete-site [post]
func (c *ApiController) DeleteSite() {
	var site object.Site
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &site)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = wrapActionResponse(object.DeleteSite(&site))
	c.ServeJSON()
}
