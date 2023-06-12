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
	"strings"

	"github.com/beego/beego/utils/pagination"
	"github.com/casdoor/casdoor/ai"
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
	organization := c.Input().Get("organization")
	limit := c.Input().Get("pageSize")
	page := c.Input().Get("p")
	field := c.Input().Get("field")
	value := c.Input().Get("value")
	sortField := c.Input().Get("sortField")
	sortOrder := c.Input().Get("sortOrder")
	chat := c.Input().Get("chat")

	if limit == "" || page == "" {
		var messages []*object.Message
		var err error
		if chat == "" {
			messages, err = object.GetMessages(owner)
		} else {
			messages, err = object.GetChatMessages(chat)
		}

		if err != nil {
			panic(err)
		}

		c.Data["json"] = object.GetMaskedMessages(messages)
		c.ServeJSON()
	} else {
		limit := util.ParseInt(limit)
		count, err := object.GetMessageCount(owner, organization, field, value)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		paginator := pagination.SetPaginator(c.Ctx, limit, count)
		paginationMessages, err := object.GetPaginationMessages(owner, organization, paginator.Offset(), limit, field, value, sortField, sortOrder)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		messages := object.GetMaskedMessages(paginationMessages)
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
	message, err := object.GetMessage(id)
	if err != nil {
		panic(err)
	}

	c.Data["json"] = object.GetMaskedMessage(message)
	c.ServeJSON()
}

func (c *ApiController) ResponseErrorStream(errorText string) {
	event := fmt.Sprintf("event: myerror\ndata: %s\n\n", errorText)
	_, err := c.Ctx.ResponseWriter.Write([]byte(event))
	if err != nil {
		panic(err)
	}
}

// GetMessageAnswer
// @Title GetMessageAnswer
// @Tag Message API
// @Description get message answer
// @Param   id     query    string  true        "The id ( owner/name ) of the message"
// @Success 200 {object} object.Message The Response object
// @router /get-message-answer [get]
func (c *ApiController) GetMessageAnswer() {
	id := c.Input().Get("id")

	c.Ctx.ResponseWriter.Header().Set("Content-Type", "text/event-stream")
	c.Ctx.ResponseWriter.Header().Set("Cache-Control", "no-cache")
	c.Ctx.ResponseWriter.Header().Set("Connection", "keep-alive")

	message, err := object.GetMessage(id)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if message == nil {
		c.ResponseErrorStream(fmt.Sprintf(c.T("chat:The message: %s is not found"), id))
		return
	}

	if message.Author != "AI" || message.ReplyTo == "" || message.Text != "" {
		c.ResponseErrorStream(c.T("chat:The message is invalid"))
		return
	}

	chatId := util.GetId("admin", message.Chat)
	chat, err := object.GetChat(chatId)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if chat == nil || chat.Organization != message.Organization {
		c.ResponseErrorStream(fmt.Sprintf(c.T("chat:The chat: %s is not found"), chatId))
		return
	}

	if chat.Type != "AI" {
		c.ResponseErrorStream(c.T("chat:The chat type must be \"AI\""))
		return
	}

	questionMessage, err := object.GetMessage(message.ReplyTo)
	if questionMessage == nil {
		c.ResponseErrorStream(fmt.Sprintf(c.T("chat:The message: %s is not found"), id))
		return
	}

	providerId := util.GetId(chat.Owner, chat.User2)
	provider, err := object.GetProvider(providerId)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if provider == nil {
		c.ResponseErrorStream(fmt.Sprintf(c.T("chat:The provider: %s is not found"), providerId))
		return
	}

	if provider.Category != "AI" || provider.ClientSecret == "" {
		c.ResponseErrorStream(fmt.Sprintf(c.T("chat:The provider: %s is invalid"), providerId))
		return
	}

	c.Ctx.ResponseWriter.Header().Set("Content-Type", "text/event-stream")
	c.Ctx.ResponseWriter.Header().Set("Cache-Control", "no-cache")
	c.Ctx.ResponseWriter.Header().Set("Connection", "keep-alive")

	authToken := provider.ClientSecret
	question := questionMessage.Text
	var stringBuilder strings.Builder

	fmt.Printf("Question: [%s]\n", questionMessage.Text)
	fmt.Printf("Answer: [")

	err = ai.QueryAnswerStream(authToken, question, c.Ctx.ResponseWriter, &stringBuilder)
	if err != nil {
		c.ResponseErrorStream(err.Error())
		return
	}

	fmt.Printf("]\n")

	event := fmt.Sprintf("event: end\ndata: %s\n\n", "end")
	_, err = c.Ctx.ResponseWriter.Write([]byte(event))
	if err != nil {
		panic(err)
	}

	answer := stringBuilder.String()

	message.Text = answer
	_, err = object.UpdateMessage(message.GetId(), message)
	if err != nil {
		panic(err)
	}
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

	var chat *object.Chat
	if message.Chat != "" {
		chatId := util.GetId("admin", message.Chat)
		chat, err = object.GetChat(chatId)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		if chat == nil || chat.Organization != message.Organization {
			c.ResponseError(fmt.Sprintf(c.T("chat:The chat: %s is not found"), chatId))
			return
		}
	}

	affected, err := object.AddMessage(&message)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if affected {
		if chat != nil && chat.Type == "AI" {
			answerMessage := &object.Message{
				Owner:        message.Owner,
				Name:         fmt.Sprintf("message_%s", util.GetRandomName()),
				CreatedTime:  util.GetCurrentTimeEx(message.CreatedTime),
				Organization: message.Organization,
				Chat:         message.Chat,
				ReplyTo:      message.GetId(),
				Author:       "AI",
				Text:         "",
			}
			_, err = object.AddMessage(answerMessage)
			if err != nil {
				c.ResponseError(err.Error())
				return
			}
		}
	}

	c.Data["json"] = wrapActionResponse(affected)
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
