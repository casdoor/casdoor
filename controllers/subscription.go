// Copyright 2022 The Casdoor Authors. All Rights Reserved.
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

// GetSubscriptions
// @Title GetSubscriptions
// @Tag Subscription API
// @Description get subscriptions
// @Param   owner     query    string  true        "The owner of subscriptions"
// @Success 200 {array} object.Subscription The Response object
// @router /get-subscriptions [get]
func (c *ApiController) GetSubscriptions() {
	owner := c.Input().Get("owner")
	limit := c.Input().Get("pageSize")
	page := c.Input().Get("p")
	field := c.Input().Get("field")
	value := c.Input().Get("value")
	sortField := c.Input().Get("sortField")
	sortOrder := c.Input().Get("sortOrder")
	if limit == "" || page == "" {
		c.Data["json"] = object.GetSubscriptions(owner)
		c.ServeJSON()
	} else {
		limit := util.ParseInt(limit)
		paginator := pagination.SetPaginator(c.Ctx, limit, int64(object.GetSubscriptionCount(owner, field, value)))
		subscription := object.GetPaginationSubscriptions(owner, paginator.Offset(), limit, field, value, sortField, sortOrder)
		c.ResponseOk(subscription, paginator.Nums())
	}
}

// GetSubscription
// @Title GetSubscription
// @Tag Subscription API
// @Description get subscription
// @Param   id     query    string  true        "The id ( owner/name ) of the subscription"
// @Success 200 {object} object.subscription The Response object
// @router /get-subscription [get]
func (c *ApiController) GetSubscription() {
	id := c.Input().Get("id")

	subscription := object.GetSubscription(id)

	c.Data["json"] = subscription
	c.ServeJSON()
}

// UpdateSubscription
// @Title UpdateSubscription
// @Tag Subscription API
// @Description update subscription
// @Param   id     query    string  true        "The id ( owner/name ) of the subscription"
// @Param   body    body   object.Subscription  true        "The details of the subscription"
// @Success 200 {object} controllers.Response The Response object
// @router /update-subscription [post]
func (c *ApiController) UpdateSubscription() {
	id := c.Input().Get("id")

	var subscription object.Subscription
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &subscription)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = wrapActionResponse(object.UpdateSubscription(id, &subscription))
	c.ServeJSON()
}

// AddSubscription
// @Title AddSubscription
// @Tag Subscription API
// @Description add subscription
// @Param   body    body   object.Subscription  true        "The details of the subscription"
// @Success 200 {object} controllers.Response The Response object
// @router /add-subscription [post]
func (c *ApiController) AddSubscription() {
	var subscription object.Subscription
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &subscription)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = wrapActionResponse(object.AddSubscription(&subscription))
	c.ServeJSON()
}

// DeleteSubscription
// @Title DeleteSubscription
// @Tag Subscription API
// @Description delete subscription
// @Param   body    body   object.Subscription  true        "The details of the subscription"
// @Success 200 {object} controllers.Response The Response object
// @router /delete-subscription [post]
func (c *ApiController) DeleteSubscription() {
	var subscription object.Subscription
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &subscription)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = wrapActionResponse(object.DeleteSubscription(&subscription))
	c.ServeJSON()
}
