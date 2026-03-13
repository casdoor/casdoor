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

	"github.com/beego/beego/v2/core/utils/pagination"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

// GetWebhookEvents
// @Title GetWebhookEvents
// @Tag Webhook Event API
// @Description get webhook events with filtering
// @Param   owner     query    string  false       "The owner of webhook events"
// @Param   organization     query    string  false       "The organization"
// @Param   webhookName     query    string  false       "The webhook name"
// @Param   status     query    string  false       "Event status (pending, success, failed, retrying)"
// @Success 200 {array} object.WebhookEvent The Response object
// @router /get-webhook-events [get]
func (c *ApiController) GetWebhookEvents() {
	owner := c.Ctx.Input.Query("owner")
	organization := c.Ctx.Input.Query("organization")
	webhookName := c.Ctx.Input.Query("webhookName")
	status := c.Ctx.Input.Query("status")
	limit := c.Ctx.Input.Query("pageSize")
	page := c.Ctx.Input.Query("p")

	var offset int
	var limitInt int

	if limit != "" && page != "" {
		limitInt = util.ParseInt(limit)
		pageInt := util.ParseInt(page)
		offset = (pageInt - 1) * limitInt
	}

	events, err := object.GetWebhookEvents(owner, organization, webhookName, object.WebhookEventStatus(status), offset, limitInt)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if limit != "" && page != "" {
		// For pagination, we'd need to add a count function
		// For now, just return the events
		c.ResponseOk(events)
	} else {
		c.ResponseOk(events)
	}
}

// GetWebhookEvent
// @Title GetWebhookEvent
// @Tag Webhook Event API
// @Description get webhook event
// @Param   id     query    string  true        "The id ( owner/name ) of the webhook event"
// @Success 200 {object} object.WebhookEvent The Response object
// @router /get-webhook-event [get]
func (c *ApiController) GetWebhookEvent() {
	id := c.Ctx.Input.Query("id")

	event, err := object.GetWebhookEvent(id)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(event)
}

// ReplayWebhookEvent
// @Title ReplayWebhookEvent
// @Tag Webhook Event API
// @Description replay a webhook event
// @Param   id     query    string  true        "The id ( owner/name ) of the webhook event"
// @Success 200 {object} controllers.Response The Response object
// @router /replay-webhook-event [post]
func (c *ApiController) ReplayWebhookEvent() {
	id := c.Ctx.Input.Query("id")

	err := object.ReplayWebhookEvent(id)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk("Webhook event replayed successfully")
}

// ReplayWebhookEvents
// @Title ReplayWebhookEvents
// @Tag Webhook Event API
// @Description replay multiple webhook events
// @Param   owner     query    string  false       "The owner of webhook events"
// @Param   organization     query    string  false       "The organization"
// @Param   webhookName     query    string  false       "The webhook name"
// @Param   status     query    string  false       "Event status to replay (e.g., failed)"
// @Success 200 {object} controllers.Response The Response object
// @router /replay-webhook-events [post]
func (c *ApiController) ReplayWebhookEvents() {
	owner := c.Ctx.Input.Query("owner")
	organization := c.Ctx.Input.Query("organization")
	webhookName := c.Ctx.Input.Query("webhookName")
	status := c.Ctx.Input.Query("status")

	count, err := object.ReplayWebhookEvents(owner, organization, webhookName, object.WebhookEventStatus(status))
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(map[string]interface{}{
		"count":   count,
		"message": "webhook events replayed successfully",
	})
}

// DeleteWebhookEvent
// @Title DeleteWebhookEvent
// @Tag Webhook Event API
// @Description delete webhook event
// @Param   body    body   object.WebhookEvent  true        "The details of the webhook event"
// @Success 200 {object} controllers.Response The Response object
// @router /delete-webhook-event [post]
func (c *ApiController) DeleteWebhookEvent() {
	var event object.WebhookEvent
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &event)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = wrapActionResponse(object.DeleteWebhookEvent(&event))
	c.ServeJSON()
}
