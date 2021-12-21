// Copyright 2021 The casbin Authors. All Rights Reserved.
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

	"github.com/astaxie/beego/utils/pagination"
	"github.com/casbin/casdoor/object"
	"github.com/casbin/casdoor/util"
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
		c.Data["json"] = object.GetWebhooks(owner)
		c.ServeJSON()
	} else {
		limit := util.ParseInt(limit)
		paginator := pagination.SetPaginator(c.Ctx, limit, int64(object.GetWebhookCount(owner, field, value)))
		webhooks := object.GetPaginationWebhooks(owner, paginator.Offset(), limit, field, value, sortField, sortOrder)
		c.ResponseOk(webhooks, paginator.Nums())
	}
}

// @Title GetWebhook
// @Tag Webhook API
// @Description get webhook
// @Param   id    query    string  true        "The id of the webhook"
// @Success 200 {object} object.Webhook The Response object
// @router /get-webhook [get]
func (c *ApiController) GetWebhook() {
	id := c.Input().Get("id")

	c.Data["json"] = object.GetWebhook(id)
	c.ServeJSON()
}

// @Title UpdateWebhook
// @Tag Webhook API
// @Description update webhook
// @Param   id    query    string  true        "The id of the webhook"
// @Param   body    body   object.Webhook  true        "The details of the webhook"
// @Success 200 {object} controllers.Response The Response object
// @router /update-webhook [post]
func (c *ApiController) UpdateWebhook() {
	id := c.Input().Get("id")

	var webhook object.Webhook
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &webhook)
	if err != nil {
		panic(err)
	}

	c.Data["json"] = wrapActionResponse(object.UpdateWebhook(id, &webhook))
	c.ServeJSON()
}

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
		panic(err)
	}

	c.Data["json"] = wrapActionResponse(object.AddWebhook(&webhook))
	c.ServeJSON()
}

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
		panic(err)
	}

	c.Data["json"] = wrapActionResponse(object.DeleteWebhook(&webhook))
	c.ServeJSON()
}
