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

// GetMarketplaceIndex
// @Title GetMarketplaceIndex
// @Tag Integration API
// @Description get the catalog of the Casdoor Marketplace
// @Param   owner     query    string  true        "The organization of the admin"
// @Param   refresh   query    string  false       "1 to skip the cached copy"
// @Success 200 {object} object.MarketplaceIndex The Response object
// @router /get-marketplace-index [get]
func (c *ApiController) GetMarketplaceIndex() {
	refresh := c.Ctx.Input.Query("refresh") == "1"

	index, err := object.GetMarketplaceIndex(refresh)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(index)
}

// GetMarketplaceBundle
// @Title GetMarketplaceBundle
// @Tag Integration API
// @Description get the manifest and content files of an integration in the Casdoor Marketplace
// @Param   owner     query    string  true        "The organization of the admin"
// @Param   integrationId query string  true        "The id of the integration"
// @Param   version   query    string  false       "The version, the latest one if empty"
// @Success 200 {object} controllers.Response The Response object
// @router /get-marketplace-bundle [get]
func (c *ApiController) GetMarketplaceBundle() {
	integrationId := c.Ctx.Input.Query("integrationId")
	version := c.Ctx.Input.Query("version")

	bundle, err := object.GetMarketplaceBundle(integrationId, version)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(bundle)
}

// GetIntegrations
// @Title GetIntegrations
// @Tag Integration API
// @Description get installed integrations
// @Param   owner     query    string  true        "The owner of integrations"
// @Success 200 {array} object.Integration The Response object
// @router /get-integrations [get]
func (c *ApiController) GetIntegrations() {
	owner := c.Ctx.Input.Query("owner")
	limit := c.Ctx.Input.Query("pageSize")
	page := c.Ctx.Input.Query("p")
	field := c.Ctx.Input.Query("field")
	value := c.Ctx.Input.Query("value")
	sortField := c.Ctx.Input.Query("sortField")
	sortOrder := c.Ctx.Input.Query("sortOrder")

	if limit == "" || page == "" {
		integrations, err := object.GetIntegrations(owner)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		c.ResponseOk(integrations)
	} else {
		limit := util.ParseInt(limit)
		count, err := object.GetIntegrationCount(owner, field, value)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		paginator := pagination.NewPaginator(c.Ctx.Request, limit, count)
		integrations, err := object.GetPaginationIntegrations(owner, paginator.Offset(), limit, field, value, sortField, sortOrder)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		c.ResponseOk(integrations, paginator.Nums())
	}
}

// GetIntegration
// @Title GetIntegration
// @Tag Integration API
// @Description get an installed integration
// @Param   id     query    string  true        "The id ( owner/name ) of the integration"
// @Success 200 {object} object.Integration The Response object
// @router /get-integration [get]
func (c *ApiController) GetIntegration() {
	id := c.Ctx.Input.Query("id")

	integration, err := object.GetIntegration(id)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(integration)
}

// InstallIntegration
// @Title InstallIntegration
// @Tag Integration API
// @Description install an integration from the Casdoor Marketplace
// @Param   body    body   object.IntegrationInstallRequest  true        "The integration, its variables and where to install it"
// @Success 200 {object} controllers.Response The Response object
// @router /install-integration [post]
func (c *ApiController) InstallIntegration() {
	var req object.IntegrationInstallRequest
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &req)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if !c.requireOrganizationPermission(req.Owner) {
		return
	}

	item, err := object.GetMarketplaceIntegration(req.IntegrationId)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	switch item.Type {
	case "app":
		count, err := object.GetApplicationCount("", "", "")
		if err != nil {
			c.ResponseError(err.Error())
			return
		}
		if err = checkQuotaForApplication(int(count)); err != nil {
			c.ResponseError(err.Error())
			return
		}
	case "provider":
		count, err := object.GetProviderCount("", "", "")
		if err != nil {
			c.ResponseError(err.Error())
			return
		}
		if err = checkQuotaForProvider(int(count)); err != nil {
			c.ResponseError(err.Error())
			return
		}
	}

	integration, err := object.InstallIntegration(&req, c.IsGlobalAdmin(), c.Ctx.Request.Host, c.GetAcceptLanguage())
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(integration)
}

// DeleteIntegration
// @Title DeleteIntegration
// @Tag Integration API
// @Description uninstall an integration: delete the application or provider it created, or restore the theme it replaced
// @Param   body    body   object.Integration  true        "The details of the integration"
// @Success 200 {object} controllers.Response The Response object
// @router /delete-integration [post]
func (c *ApiController) DeleteIntegration() {
	var integration object.Integration
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &integration)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if !c.requireOrganizationPermission(integration.Owner) {
		return
	}

	c.Data["json"] = wrapActionResponse(object.UninstallIntegration(&integration, c.IsGlobalAdmin(), c.GetAcceptLanguage()))
	c.ServeJSON()
}
