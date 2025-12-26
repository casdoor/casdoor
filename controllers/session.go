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
	"context"
	"encoding/json"
	"fmt"

	"github.com/beego/beego/v2/core/utils/pagination"
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
	limit := c.Ctx.Input.Query("pageSize")
	page := c.Ctx.Input.Query("p")
	field := c.Ctx.Input.Query("field")
	value := c.Ctx.Input.Query("value")
	sortField := c.Ctx.Input.Query("sortField")
	sortOrder := c.Ctx.Input.Query("sortOrder")
	owner := c.Ctx.Input.Query("owner")

	if limit == "" || page == "" {
		sessions, err := object.GetSessions(owner)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		c.ResponseOk(sessions)
	} else {
		limit := util.ParseInt(limit)
		count, err := object.GetSessionCount(owner, field, value)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}
		paginator := pagination.NewPaginator(c.Ctx.Request, limit, count)
		sessions, err := object.GetPaginationSessions(owner, paginator.Offset(), limit, field, value, sortField, sortOrder)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		c.ResponseOk(sessions, paginator.Nums())
	}
}

// GetSingleSession
// @Title GetSingleSession
// @Tag Session API
// @Description Get session for one user in one application.
// @Param   sessionPkId     query    string  true        "The session ID in format: organization/user/application (e.g., built-in/admin/app-built-in)"
// @Success 200 {array} string The Response object
// @router /get-session [get]
func (c *ApiController) GetSingleSession() {
	id := c.Ctx.Input.Query("sessionPkId")

	session, err := object.GetSingleSession(id)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(session)
}

// UpdateSession
// @Title UpdateSession
// @Tag Session API
// @Description Update session for one user in one application.
// @Param   body     body    object.Session  true        "The session object to update"
// @Success 200 {object} controllers.Response The Response object
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
// @Param   body     body    object.Session  true        "The session object to add"
// @Success 200 {object} controllers.Response The Response object
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
// @Param   body     body    object.Session  true        "The session object to delete"
// @Success 200 {object} controllers.Response The Response object
// @router /delete-session [post]
func (c *ApiController) DeleteSession() {
	var session object.Session
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &session)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	curSessionId := c.Ctx.Input.CruSession.SessionID(context.Background())

	sessionId := c.Ctx.Input.Query("sessionId")
	if curSessionId == sessionId && sessionId != "" {
		c.ResponseError(fmt.Sprintf(c.T("session:session id %s is the current session and cannot be deleted"), curSessionId))
		return
	}

	if sessionId != "" {
		c.Data["json"] = wrapActionResponse(object.DeleteSessionId(util.GetSessionId(session.Owner, session.Name, session.Application), sessionId))
		c.ServeJSON()
		return
	}

	c.Data["json"] = wrapActionResponse(object.DeleteSession(util.GetSessionId(session.Owner, session.Name, session.Application), curSessionId))
	c.ServeJSON()
}

// IsSessionDuplicated
// @Title IsSessionDuplicated
// @Tag Session API
// @Description Check if there are other different sessions for one user in one application.
// @Param   sessionPkId     query    string  true        "The session ID in format: organization/user/application (e.g., built-in/admin/app-built-in)"
// @Param   sessionId     query    string  true        "The specific session ID to check"
// @Success 200 {array} string The Response object
// @router /is-session-duplicated [get]
func (c *ApiController) IsSessionDuplicated() {
	id := c.Ctx.Input.Query("sessionPkId")
	sessionId := c.Ctx.Input.Query("sessionId")

	isUserSessionDuplicated, err := object.IsSessionDuplicated(id, sessionId)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(isUserSessionDuplicated)
}
