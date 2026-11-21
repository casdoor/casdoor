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
	"context"

	"github.com/casdoor/casdoor/object"
)

// GetMySessions
// @Title GetMySessions
// @Tag Session API
// @Description list the signed-in sessions of the current user, with hashed session ids
// @Success 200 {array} object.UserSessionItem The Response object
// @router /get-my-sessions [get]
func (c *ApiController) GetMySessions() {
	user := c.GetCurrentUser()
	if user == nil || object.IsAppUser(user.GetId()) {
		c.ResponseError(c.T("general:Please login first"))
		return
	}

	items, err := object.GetUserSessionItems(user.Owner, user.Name, c.Ctx.Input.CruSession.SessionID(context.Background()))
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	c.ResponseOk(items)
}

// DeleteMySession
// @Title DeleteMySession
// @Tag Session API
// @Description sign the current user out of another of its sessions
// @Param   sessionId     query    string  true        "The hashed session ID from get-my-sessions"
// @Success 200 {object} controllers.Response The Response object
// @router /delete-my-session [post]
func (c *ApiController) DeleteMySession() {
	user := c.GetCurrentUser()
	if user == nil || object.IsAppUser(user.GetId()) {
		c.ResponseError(c.T("general:Please login first"))
		return
	}

	sessionId := c.Ctx.Input.Query("sessionId")
	curSessionId := c.Ctx.Input.CruSession.SessionID(context.Background())
	if sessionId == "" || sessionId == object.GetSessionIdHash(curSessionId) {
		c.ResponseError(c.T("session:The current session cannot be ended here, sign out instead"))
		return
	}

	c.Data["json"] = WrapActionResponse(object.LogoutUserSessionByHash(user, sessionId, c.Ctx.Request.Host))
	c.ServeJSON()
}
