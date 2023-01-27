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

// DeleteSession
// @Title DeleteSession
// @Tag Session API
// @Description Delete session by userId
// @Param   ID     query    string  true        "The ID(owner/name) of user."
// @Success 200 {array} string The Response object
// @router /delete-session [post]
func (c *ApiController) DeleteSession() {
	var session object.Session
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &session)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = wrapActionResponse(object.DeleteSession(util.GetId(session.Owner, session.Name)))
	c.ServeJSON()
}

// GetSessions
// @Title GetSessions
// @Tag Session API
// @Description Get organization user sessions
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

// AddUserSession
// @Title AddUserSession
// @Tag Session API
// @Description Add application user sessions
// @Param   ID     				string  "The ID(owner/application/name) of user."
// @Param	sessionId			string	"sessionId to be added"
// @Param	sessionCreateTime	string	"unixTimeStamp"
// @Success 200 {array} string The Response object
// @router /add-user-session [post]
func (c *ApiController) AddUserSession() {
	var session object.Session
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &session)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if len(session.SessionId) != 1 {
		c.Data["json"] = &Response{Status: "error", Msg: "only one session can be added at a time"}
	} else {
		c.Data["json"] = wrapActionResponse(object.AddUserSession(session.Owner, session.Application, session.Name, session.SessionId[0], session.CreatedTime))
	}

	c.ServeJSON()
}

// DeleteUserSession
// @Title DeleteUserSession
// @Tag Session API
// @Description Delete application user sessions
// @Param   ID  string  "The ID(owner/application/name) of user."
// @Success 200 {array} string The Response object
// @router /delete-user-session [post]
func (c *ApiController) DeleteUserSession() {
	var session object.Session
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &session)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = wrapActionResponse(object.DeleteUserSession(session.Owner, session.Application, session.Name))
	c.ServeJSON()
}

// IsUserSessionDuplicated
// @Title IsUserSessionDuplicated
// @Tag Session API
// @Description Judge Whether this application user session is repeated
// @Param   ID     				string  "The ID(owner/application/name) of user."
// @Param	sessionId			string	"sessionId to be checked"
// @Param	sessionCreateTime	string	"unixTimeStamp"
// @Success 200 {array} string The Response object
// @router /is-user-session-duplicated [post]
func (c *ApiController) IsUserSessionDuplicated() {
	var session object.Session
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &session)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if len(session.SessionId) != 1 {
		c.Data["json"] = &Response{Status: "error", Msg: "only one session can be checked at a time"}
	} else {
		isUserSessionDuplicated := object.IsUserSessionDuplicated(session.Owner, session.Application, session.Name, session.SessionId[0], session.CreatedTime)
		c.Data["json"] = &Response{Status: "ok", Msg: "", Data: isUserSessionDuplicated}
	}
	c.ServeJSON()
}
