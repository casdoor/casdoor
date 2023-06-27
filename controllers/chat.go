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

// GetChats
// @Title GetChats
// @Tag Chat API
// @Description get chats
// @Param   owner     query    string  true        "The owner of chats"
// @Success 200 {array} object.Chat The Response object
// @router /get-chats [get]
func (c *ApiController) GetChats() {
	owner := "admin"
	limit := c.Input().Get("pageSize")
	page := c.Input().Get("p")
	field := c.Input().Get("field")
	value := c.Input().Get("value")
	sortField := c.Input().Get("sortField")
	sortOrder := c.Input().Get("sortOrder")

	if limit == "" || page == "" {
		maskedChats, err := object.GetMaskedChats(object.GetChats(owner))
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		c.Data["json"] = maskedChats
		c.ServeJSON()
	} else {
		limit := util.ParseInt(limit)
		count, err := object.GetChatCount(owner, field, value)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		paginator := pagination.SetPaginator(c.Ctx, limit, count)
		chats, err := object.GetMaskedChats(object.GetPaginationChats(owner, paginator.Offset(), limit, field, value, sortField, sortOrder))
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		c.ResponseOk(chats, paginator.Nums())
	}
}

// GetChat
// @Title GetChat
// @Tag Chat API
// @Description get chat
// @Param   id     query    string  true        "The id ( owner/name ) of the chat"
// @Success 200 {object} object.Chat The Response object
// @router /get-chat [get]
func (c *ApiController) GetChat() {
	id := c.Input().Get("id")

	maskedChat, err := object.GetMaskedChat(object.GetChat(id))
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = maskedChat
	c.ServeJSON()
}

// UpdateChat
// @Title UpdateChat
// @Tag Chat API
// @Description update chat
// @Param   id     query    string  true        "The id ( owner/name ) of the chat"
// @Param   body    body   object.Chat  true        "The details of the chat"
// @Success 200 {object} controllers.Response The Response object
// @router /update-chat [post]
func (c *ApiController) UpdateChat() {
	id := c.Input().Get("id")

	var chat object.Chat
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &chat)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = wrapActionResponse(object.UpdateChat(id, &chat))
	c.ServeJSON()
}

// AddChat
// @Title AddChat
// @Tag Chat API
// @Description add chat
// @Param   body    body   object.Chat  true        "The details of the chat"
// @Success 200 {object} controllers.Response The Response object
// @router /add-chat [post]
func (c *ApiController) AddChat() {
	var chat object.Chat
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &chat)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = wrapActionResponse(object.AddChat(&chat))
	c.ServeJSON()
}

// DeleteChat
// @Title DeleteChat
// @Tag Chat API
// @Description delete chat
// @Param   body    body   object.Chat  true        "The details of the chat"
// @Success 200 {object} controllers.Response The Response object
// @router /delete-chat [post]
func (c *ApiController) DeleteChat() {
	var chat object.Chat
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &chat)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = wrapActionResponse(object.DeleteChat(&chat))
	c.ServeJSON()
}
