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

// GetWebhooks
// @Title GetWebhooks
// @Tag Webhook API
// @Description get webhooks
// @Param   owner     query    string  true        "The owner of webhooks"
// @Success 200 {array} object.Webhook The Response object
// @router /get-webhooks [get]
func (c *ApiController) GetWebhooks() {
	owner := c.Input().Get("owner")
	limit := c.Input().Get("pageSize")
	page := c.Input().Get("p")
	field := c.Input().Get("field")
	value := c.Input().Get("value")
	sortField := c.Input().Get("sortField")
	sortOrder := c.Input().Get("sortOrder")
	if limit == "" || page == "" {
		webhooks := object.GetWebhooks(owner)
		c.ResponseOk(webhooks)
	} else {
		limit := util.ParseInt(limit)
		paginator := pagination.SetPaginator(c.Ctx, limit, int64(object.GetWebhookCount(owner, field, value)))
		webhooks := object.GetPaginationWebhooks(owner, paginator.Offset(), limit, field, value, sortField, sortOrder)
		c.ResponseOk(webhooks, paginator.Nums())
	}
}

// GetWebhook
// @Title GetWebhook
// @Tag Webhook API
// @Description get webhook
// @Param   id     query    string  true        "The id ( owner/name ) of the webhook"
// @Success 200 {object} object.Webhook The Response object
// @router /get-webhook [get]
func (c *ApiController) GetWebhook() {
	id := c.Input().Get("id")

	webhook := object.GetWebhook(id)
	c.ResponseOk(webhook)
}

// UpdateWebhook
// @Title UpdateWebhook
// @Tag Webhook API
// @Description update webhook
// @Param   id     query    string  true        "The id ( owner/name ) of the webhook"
// @Param   body    body   object.Webhook  true        "The details of the webhook"
// @Success 200 {object} controllers.Response The Response object
// @router /update-webhook [post]
func (c *ApiController) UpdateWebhook() {
	id := c.Input().Get("id")

	var webhook object.Webhook
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &webhook)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	response := wrapActionResponse(object.UpdateWebhook(id, &webhook))
	c.ResponseOk(response)
}

// AddWebhook
// @Title AddWebhook
// @Tag Webhook API
// @Description add webhook
// @Param   body    body   object.Webhook  true        "The details of the webhook"
// @Success 200 {object} controllers.Response The Response object
// @router /add-webhook [post]
func (c *ApiController) AddWebhook() {
	var webhook object.Webhook
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &webhook)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	response := wrapActionResponse(object.AddWebhook(&webhook))
	c.ResponseOk(response)
}

// DeleteWebhook
// @Title DeleteWebhook
// @Tag Webhook API
// @Description delete webhook
// @Param   body    body   object.Webhook  true        "The details of the webhook"
// @Success 200 {object} controllers.Response The Response object
// @router /delete-webhook [post]
func (c *ApiController) DeleteWebhook() {
	var webhook object.Webhook
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &webhook)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	response := wrapActionResponse(object.DeleteWebhook(&webhook))
	c.ResponseOk(response)
}
