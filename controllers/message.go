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

// GetMessages
// @Title GetMessages
// @Tag Message API
// @Description get messages
// @Param   owner     query    string  true        "The owner of messages"
// @Success 200 {array} object.Message The Response object
// @router /get-messages [get]
func (c *ApiController) GetMessages() {
	owner := c.Input().Get("owner")
	limit := c.Input().Get("pageSize")
	page := c.Input().Get("p")
	field := c.Input().Get("field")
	value := c.Input().Get("value")
	sortField := c.Input().Get("sortField")
	sortOrder := c.Input().Get("sortOrder")
	if limit == "" || page == "" {
		c.Data["json"] = object.GetMaskedMessages(object.GetMessages(owner))
		c.ServeJSON()
	} else {
		limit := util.ParseInt(limit)
		paginator := pagination.SetPaginator(c.Ctx, limit, int64(object.GetMessageCount(owner, field, value)))
		messages := object.GetMaskedMessages(object.GetPaginationMessages(owner, paginator.Offset(), limit, field, value, sortField, sortOrder))
		c.ResponseOk(messages, paginator.Nums())
	}
}

// GetMessage
// @Title GetMessage
// @Tag Message API
// @Description get message
// @Param   id     query    string  true        "The id ( owner/name ) of the message"
// @Success 200 {object} object.Message The Response object
// @router /get-message [get]
func (c *ApiController) GetMessage() {
	id := c.Input().Get("id")

	c.Data["json"] = object.GetMaskedMessage(object.GetMessage(id))
	c.ServeJSON()
}

// UpdateMessage
// @Title UpdateMessage
// @Tag Message API
// @Description update message
// @Param   id     query    string  true        "The id ( owner/name ) of the message"
// @Param   body    body   object.Message  true        "The details of the message"
// @Success 200 {object} controllers.Response The Response object
// @router /update-message [post]
func (c *ApiController) UpdateMessage() {
	id := c.Input().Get("id")

	var message object.Message
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &message)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = wrapActionResponse(object.UpdateMessage(id, &message))
	c.ServeJSON()
}

// AddMessage
// @Title AddMessage
// @Tag Message API
// @Description add message
// @Param   body    body   object.Message  true        "The details of the message"
// @Success 200 {object} controllers.Response The Response object
// @router /add-message [post]
func (c *ApiController) AddMessage() {
	var message object.Message
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &message)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = wrapActionResponse(object.AddMessage(&message))
	c.ServeJSON()
}

// DeleteMessage
// @Title DeleteMessage
// @Tag Message API
// @Description delete message
// @Param   body    body   object.Message  true        "The details of the message"
// @Success 200 {object} controllers.Response The Response object
// @router /delete-message [post]
func (c *ApiController) DeleteMessage() {
	var message object.Message
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &message)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = wrapActionResponse(object.DeleteMessage(&message))
	c.ServeJSON()
}
