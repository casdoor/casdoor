// Copyright 2024 The Casdoor Authors. All Rights Reserved.
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

// GetTickets
// @Title GetTickets
// @Tag Ticket API
// @Description get tickets
// @Param   owner     query    string  true        "The owner of tickets"
// @Success 200 {array} object.Ticket The Response object
// @router /get-tickets [get]
func (c *ApiController) GetTickets() {
	owner := c.Input().Get("owner")
	limit := c.Input().Get("pageSize")
	page := c.Input().Get("p")
	field := c.Input().Get("field")
	value := c.Input().Get("value")
	sortField := c.Input().Get("sortField")
	sortOrder := c.Input().Get("sortOrder")

	user := c.getCurrentUser()
	isAdmin := c.IsAdmin()

	var tickets []*object.Ticket
	var err error

	if limit == "" || page == "" {
		if isAdmin {
			tickets, err = object.GetTickets(owner)
		} else {
			tickets, err = object.GetUserTickets(owner, user.GetId())
		}
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		c.ResponseOk(tickets)
	} else {
		limit := util.ParseInt(limit)
		var count int64

		if isAdmin {
			count, err = object.GetTicketCount(owner, field, value)
		} else {
			// For non-admin users, only show their own tickets
			tickets, err = object.GetUserTickets(owner, user.GetId())
			if err != nil {
				c.ResponseError(err.Error())
				return
			}
			count = int64(len(tickets))
		}

		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		paginator := pagination.SetPaginator(c.Ctx, limit, count)

		if isAdmin {
			tickets, err = object.GetPaginationTickets(owner, paginator.Offset(), limit, field, value, sortField, sortOrder)
		}

		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		c.ResponseOk(tickets, paginator.Nums())
	}
}

// GetTicket
// @Title GetTicket
// @Tag Ticket API
// @Description get ticket
// @Param   id     query    string  true        "The id ( owner/name ) of the ticket"
// @Success 200 {object} object.Ticket The Response object
// @router /get-ticket [get]
func (c *ApiController) GetTicket() {
	id := c.Input().Get("id")

	ticket, err := object.GetTicket(id)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	// Check permission: user can only view their own tickets unless they are admin
	user := c.getCurrentUser()
	isAdmin := c.IsAdmin()

	if ticket != nil && !isAdmin && ticket.User != user.GetId() {
		c.ResponseError(c.T("auth:Unauthorized operation"))
		return
	}

	c.ResponseOk(ticket)
}

// UpdateTicket
// @Title UpdateTicket
// @Tag Ticket API
// @Description update ticket
// @Param   id     query    string  true        "The id ( owner/name ) of the ticket"
// @Param   body    body   object.Ticket  true        "The details of the ticket"
// @Success 200 {object} controllers.Response The Response object
// @router /update-ticket [post]
func (c *ApiController) UpdateTicket() {
	id := c.Input().Get("id")

	var ticket object.Ticket
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &ticket)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	// Check permission
	user := c.getCurrentUser()
	isAdmin := c.IsAdmin()

	existingTicket, err := object.GetTicket(id)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if existingTicket == nil {
		c.ResponseError(c.T("ticket:Ticket not found"))
		return
	}

	// Normal users can only close their own tickets
	if !isAdmin {
		if existingTicket.User != user.GetId() {
			c.ResponseError(c.T("auth:Unauthorized operation"))
			return
		}
		// Normal users can only change state to "Closed"
		if ticket.State != "Closed" && ticket.State != existingTicket.State {
			c.ResponseError(c.T("auth:Unauthorized operation"))
			return
		}
		// Preserve original fields that users shouldn't modify
		ticket.Owner = existingTicket.Owner
		ticket.Name = existingTicket.Name
		ticket.User = existingTicket.User
		ticket.CreatedTime = existingTicket.CreatedTime
	}

	c.Data["json"] = wrapActionResponse(object.UpdateTicket(id, &ticket))
	c.ServeJSON()
}

// AddTicket
// @Title AddTicket
// @Tag Ticket API
// @Description add ticket
// @Param   body    body   object.Ticket  true        "The details of the ticket"
// @Success 200 {object} controllers.Response The Response object
// @router /add-ticket [post]
func (c *ApiController) AddTicket() {
	var ticket object.Ticket
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &ticket)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	// Set the user field to the current user
	user := c.getCurrentUser()
	ticket.User = user.GetId()

	c.Data["json"] = wrapActionResponse(object.AddTicket(&ticket))
	c.ServeJSON()
}

// DeleteTicket
// @Title DeleteTicket
// @Tag Ticket API
// @Description delete ticket
// @Param   body    body   object.Ticket  true        "The details of the ticket"
// @Success 200 {object} controllers.Response The Response object
// @router /delete-ticket [post]
func (c *ApiController) DeleteTicket() {
	var ticket object.Ticket
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &ticket)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	// Only admins can delete tickets
	if !c.IsAdmin() {
		c.ResponseError(c.T("auth:Unauthorized operation"))
		return
	}

	c.Data["json"] = wrapActionResponse(object.DeleteTicket(&ticket))
	c.ServeJSON()
}

// AddTicketMessage
// @Title AddTicketMessage
// @Tag Ticket API
// @Description add a message to a ticket
// @Param   id     query    string  true        "The id ( owner/name ) of the ticket"
// @Param   body    body   object.TicketMessage  true        "The message to add"
// @Success 200 {object} controllers.Response The Response object
// @router /add-ticket-message [post]
func (c *ApiController) AddTicketMessage() {
	id := c.Input().Get("id")

	var message object.TicketMessage
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &message)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	// Check permission
	user := c.getCurrentUser()
	isAdmin := c.IsAdmin()

	ticket, err := object.GetTicket(id)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if ticket == nil {
		c.ResponseError(c.T("ticket:Ticket not found"))
		return
	}

	// Users can only add messages to their own tickets, admins can add to any ticket
	if !isAdmin && ticket.User != user.GetId() {
		c.ResponseError(c.T("auth:Unauthorized operation"))
		return
	}

	// Set the author and admin flag
	message.Author = user.GetId()
	message.IsAdmin = isAdmin

	c.Data["json"] = wrapActionResponse(object.AddTicketMessage(id, &message))
	c.ServeJSON()
}
