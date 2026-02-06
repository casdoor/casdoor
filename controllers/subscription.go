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
	"fmt"

	"github.com/beego/beego/v2/core/utils/pagination"
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
	owner := c.Ctx.Input.Query("owner")
	limit := c.Ctx.Input.Query("pageSize")
	page := c.Ctx.Input.Query("p")
	field := c.Ctx.Input.Query("field")
	value := c.Ctx.Input.Query("value")
	sortField := c.Ctx.Input.Query("sortField")
	sortOrder := c.Ctx.Input.Query("sortOrder")

	if limit == "" || page == "" {
		var subscriptions []*object.Subscription
		var err error

		if c.IsAdmin() {
			// If field is "user", filter by that user even for admins
			if field == "user" && value != "" {
				subscriptions, err = object.GetSubscriptionsByUser(owner, value)
			} else {
				subscriptions, err = object.GetSubscriptions(owner)
			}
		} else {
			user := c.GetSessionUsername()
			_, userName, userErr := util.GetOwnerAndNameFromIdWithError(user)
			if userErr != nil {
				c.ResponseError(userErr.Error())
				return
			}
			subscriptions, err = object.GetSubscriptionsByUser(owner, userName)
		}

		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		c.ResponseOk(subscriptions)
	} else {
		limit := util.ParseInt(limit)
		if !c.IsAdmin() {
			user := c.GetSessionUsername()
			_, userName, userErr := util.GetOwnerAndNameFromIdWithError(user)
			if userErr != nil {
				c.ResponseError(userErr.Error())
				return
			}
			field = "user"
			value = userName
		}
		count, err := object.GetSubscriptionCount(owner, field, value)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		paginator := pagination.NewPaginator(c.Ctx.Request, limit, count)
		subscription, err := object.GetPaginationSubscriptions(owner, paginator.Offset(), limit, field, value, sortField, sortOrder)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		c.ResponseOk(subscription, paginator.Nums())
	}
}

// GetSubscription
// @Title GetSubscription
// @Tag Subscription API
// @Description get subscription
// @Param   id     query    string  true        "The id ( owner/name ) of the subscription"
// @Success 200 {object} object.Subscription The Response object
// @router /get-subscription [get]
func (c *ApiController) GetSubscription() {
	id := c.Ctx.Input.Query("id")

	subscription, err := object.GetSubscription(id)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(subscription)
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
	id := c.Ctx.Input.Query("id")

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

	// Check if plan restricts user to one subscription
	if subscription.Plan != "" {
		plan, err := object.GetPlan(util.GetId(subscription.Owner, subscription.Plan))
		if err != nil {
			c.ResponseError(err.Error())
			return
		}
		if plan != nil && plan.IsOneTimeSubscription {
			hasSubscription, err := object.HasActiveSubscriptionForPlan(subscription.Owner, subscription.User, subscription.Plan)
			if err != nil {
				c.ResponseError(err.Error())
				return
			}
			if hasSubscription {
				c.ResponseError(fmt.Sprintf("User already has an active subscription for plan: %s", subscription.Plan))
				return
			}
		}
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
