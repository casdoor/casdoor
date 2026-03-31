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

const defaultWebhookEventListLimit = 100

func (c *ApiController) getScopedWebhookEventQuery() (string, string, bool) {
	organization, ok := c.RequireAdmin()
	if !ok {
		return "", "", false
	}

	owner := ""
	if c.IsGlobalAdmin() {
		owner = c.Ctx.Input.Query("owner")

		requestedOrganization := c.Ctx.Input.Query("organization")
		if requestedOrganization != "" {
			organization = requestedOrganization
		}
	}

	return owner, organization, true
}

func (c *ApiController) checkWebhookEventAccess(event *object.WebhookEvent, organization string) bool {
	if event == nil || c.IsGlobalAdmin() {
		return true
	}

	if event.Organization != organization {
		c.ResponseError(c.T("auth:Unauthorized operation"))
		return false
	}

	return true
}

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
	owner, organization, ok := c.getScopedWebhookEventQuery()
	if !ok {
		return
	}
	webhookName := c.Ctx.Input.Query("webhookName")
	status := c.Ctx.Input.Query("status")
	limit := c.Ctx.Input.Query("pageSize")
	page := c.Ctx.Input.Query("p")
	sortField := c.Ctx.Input.Query("sortField")
	sortOrder := c.Ctx.Input.Query("sortOrder")

	if limit != "" && page != "" {
		limit := util.ParseInt(limit)
		count, err := object.GetWebhookEventCount(owner, organization, webhookName, object.WebhookEventStatus(status))
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		paginator := pagination.NewPaginator(c.Ctx.Request, limit, count)
		events, err := object.GetWebhookEvents(owner, organization, webhookName, object.WebhookEventStatus(status), paginator.Offset(), limit, sortField, sortOrder)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		c.ResponseOk(events, paginator.Nums())
	} else {
		events, err := object.GetWebhookEvents(owner, organization, webhookName, object.WebhookEventStatus(status), 0, defaultWebhookEventListLimit, sortField, sortOrder)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		c.ResponseOk(events)
	}
}

// GetWebhookEvent
// @Title GetWebhookEvent
// @Tag Webhook Event API
// @Description get webhook event
// @Param   id     query    string  true        "The id ( owner/name ) of the webhook event"
// @Success 200 {object} object.WebhookEvent The Response object
// @router /get-webhook-event-detail [get]
func (c *ApiController) GetWebhookEvent() {
	organization, ok := c.RequireAdmin()
	if !ok {
		return
	}

	id := c.Ctx.Input.Query("id")

	event, err := object.GetWebhookEvent(id)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if !c.checkWebhookEventAccess(event, organization) {
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
	organization, ok := c.RequireAdmin()
	if !ok {
		return
	}

	id := c.Ctx.Input.Query("id")

	event, err := object.GetWebhookEvent(id)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if !c.checkWebhookEventAccess(event, organization) {
		return
	}

	err = object.ReplayWebhookEvent(id)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk("Webhook event replay triggered")
}

// DeleteWebhookEvent
// @Title DeleteWebhookEvent
// @Tag Webhook Event API
// @Description delete webhook event
// @Param   body    body   object.WebhookEvent  true        "The details of the webhook event"
// @Success 200 {object} controllers.Response The Response object
// @router /delete-webhook-event [post]
func (c *ApiController) DeleteWebhookEvent() {
	organization, ok := c.RequireAdmin()
	if !ok {
		return
	}

	var event object.WebhookEvent
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &event)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	storedEvent, err := object.GetWebhookEvent(event.GetId())
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if !c.checkWebhookEventAccess(storedEvent, organization) {
		return
	}

	c.Data["json"] = wrapActionResponse(object.DeleteWebhookEvent(&event))
	c.ServeJSON()
}
