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

// GetSessions
// @Title GetSessions
// @Tag Session API
// @Description Get organization user sessions.
// @Param   owner     query    string  true        "The organization name"
// @Success 200 {array} string The Response object
// @router /get-sessions [get]
func (c *ApiController) GetSessions() {
	limit := c.Input().Get("pageSize")
	page := c.Input().Get("p")
	field := c.Input().Get("field")
	value := c.Input().Get("value")
	sortField := c.Input().Get("sortField")
	sortOrder := c.Input().Get("sortOrder")
	owner := c.Input().Get("owner")
	if limit == "" || page == "" {
		c.Data["json"] = object.GetSessions(owner)
		c.ServeJSON()
	} else {
		limit := util.ParseInt(limit)
		paginator := pagination.SetPaginator(c.Ctx, limit, int64(object.GetSessionCount(owner, field, value)))
		sessions := object.GetPaginationSessions(owner, paginator.Offset(), limit, field, value, sortField, sortOrder)
		c.ResponseOk(sessions, paginator.Nums())
	}
}

// GetSingleSession
// @Title GetSingleSession
// @Tag Session API
// @Description Get session for one user in one application.
// @Param   ID     query    string  true        "The ID(organization/application/user) of session"
// @Success 200 {array} string The Response object
// @router /get-session [get]
func (c *ApiController) GetSingleSession() {
	id := c.Input().Get("sessionPkId")

	c.Data["json"] = object.GetSingleSession(id)
	c.ServeJSON()
}

// UpdateSession
// @Title UpdateSession
// @Tag Session API
// @Description Update session for one user in one application.
// @Param   ID     query    string  true        "The ID(organization/application/user) of session"
// @Success 200 {array} string The Response object
// @router /update-session [post]
func (c *ApiController) UpdateSession() {
	var session object.Session
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &session)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = wrapActionResponse(object.UpdateSession(util.GetSessionId(session.Owner, session.Name, session.Application), &session))
	c.ServeJSON()
}

// AddSession
// @Title AddSession
// @Tag Session API
// @Description Add session for one user in one application. If there are other existing sessions, join the session into the list.
// @Param   ID     query    string  true        "The ID(organization/application/user) of session"
// @Param   sessionId     query    string  true        "sessionId to be added"
// @Success 200 {array} string The Response object
// @router /add-session [post]
func (c *ApiController) AddSession() {
	var session object.Session
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &session)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = wrapActionResponse(object.AddSession(&session))
	c.ServeJSON()
}

// DeleteSession
// @Title DeleteSession
// @Tag Session API
// @Description Delete session for one user in one application.
// @Param   ID     query    string  true        "The ID(organization/application/user) of session"
// @Success 200 {array} string The Response object
// @router /delete-session [post]
func (c *ApiController) DeleteSession() {
	var session object.Session
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &session)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = wrapActionResponse(object.DeleteSession(util.GetSessionId(session.Owner, session.Name, session.Application)))
	c.ServeJSON()
}

// IsSessionDuplicated
// @Title IsSessionDuplicated
// @Tag Session API
// @Description Check if there are other different sessions for one user in one application.
// @Param   ID     query    string  true        "The ID(organization/application/user) of session"
// @Param   sessionId     query    string  true        "sessionId to be checked"
// @Success 200 {array} string The Response object
// @router /is-user-session-duplicated [get]
func (c *ApiController) IsSessionDuplicated() {
	id := c.Input().Get("sessionPkId")
	sessionId := c.Input().Get("sessionId")

	isUserSessionDuplicated := object.IsSessionDuplicated(id, sessionId)
	c.Data["json"] = &Response{Status: "ok", Msg: "", Data: isUserSessionDuplicated}

	c.ServeJSON()
}
