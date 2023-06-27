// Copyright 2023 The Casdoor Authors. All Rights Reserved.
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

// GetPricings
// @Title GetPricings
// @Tag Pricing API
// @Description get pricings
// @Param   owner     query    string  true        "The owner of pricings"
// @Success 200 {array} object.Pricing The Response object
// @router /get-pricings [get]
func (c *ApiController) GetPricings() {
	owner := c.Input().Get("owner")
	limit := c.Input().Get("pageSize")
	page := c.Input().Get("p")
	field := c.Input().Get("field")
	value := c.Input().Get("value")
	sortField := c.Input().Get("sortField")
	sortOrder := c.Input().Get("sortOrder")

	if limit == "" || page == "" {
		pricings, err := object.GetPricings(owner)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		c.Data["json"] = pricings
		c.ServeJSON()
	} else {
		limit := util.ParseInt(limit)
		count, err := object.GetPricingCount(owner, field, value)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		paginator := pagination.SetPaginator(c.Ctx, limit, count)
		pricing, err := object.GetPaginatedPricings(owner, paginator.Offset(), limit, field, value, sortField, sortOrder)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		c.ResponseOk(pricing, paginator.Nums())
	}
}

// GetPricing
// @Title GetPricing
// @Tag Pricing API
// @Description get pricing
// @Param   id     query    string  true        "The id ( owner/name ) of the pricing"
// @Success 200 {object} object.pricing The Response object
// @router /get-pricing [get]
func (c *ApiController) GetPricing() {
	id := c.Input().Get("id")

	pricing, err := object.GetPricing(id)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = pricing
	c.ServeJSON()
}

// UpdatePricing
// @Title UpdatePricing
// @Tag Pricing API
// @Description update pricing
// @Param   id     query    string  true        "The id ( owner/name ) of the pricing"
// @Param   body    body   object.Pricing  true        "The details of the pricing"
// @Success 200 {object} controllers.Response The Response object
// @router /update-pricing [post]
func (c *ApiController) UpdatePricing() {
	id := c.Input().Get("id")

	var pricing object.Pricing
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &pricing)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = wrapActionResponse(object.UpdatePricing(id, &pricing))
	c.ServeJSON()
}

// AddPricing
// @Title AddPricing
// @Tag Pricing API
// @Description add pricing
// @Param   body    body   object.Pricing  true        "The details of the pricing"
// @Success 200 {object} controllers.Response The Response object
// @router /add-pricing [post]
func (c *ApiController) AddPricing() {
	var pricing object.Pricing
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &pricing)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = wrapActionResponse(object.AddPricing(&pricing))
	c.ServeJSON()
}

// DeletePricing
// @Title DeletePricing
// @Tag Pricing API
// @Description delete pricing
// @Param   body    body   object.Pricing  true        "The details of the pricing"
// @Success 200 {object} controllers.Response The Response object
// @router /delete-pricing [post]
func (c *ApiController) DeletePricing() {
	var pricing object.Pricing
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &pricing)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = wrapActionResponse(object.DeletePricing(&pricing))
	c.ServeJSON()
}
