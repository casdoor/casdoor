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
	"strings"
	"time"

	"github.com/beego/beego"
	"github.com/beego/beego/logs"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

// ApiController
// controller for handlers under /api uri
type ApiController struct {
	beego.Controller
}

// RootController
// controller for handlers directly under / (root)
type RootController struct {
	ApiController
}

type SessionData struct {
	ExpireTime int64
}

func (c *ApiController) IsGlobalAdmin() bool {
	isGlobalAdmin, _ := c.isGlobalAdmin()

	return isGlobalAdmin
}

func (c *ApiController) IsAdmin() bool {
	isGlobalAdmin, user := c.isGlobalAdmin()
	if !isGlobalAdmin && user == nil {
		return false
	}

	return isGlobalAdmin || user.IsAdmin
}

func (c *ApiController) IsAdminOrSelf(user2 *object.User) bool {
	isGlobalAdmin, user := c.isGlobalAdmin()
	if isGlobalAdmin || (user != nil && user.IsAdmin) {
		return true
	}

	if user == nil || user2 == nil {
		return false
	}

	if user.Owner == user2.Owner && user.Name == user2.Name {
		return true
	}
	return false
}

func (c *ApiController) isGlobalAdmin() (bool, *object.User) {
	username := c.GetSessionUsername()
	if object.IsAppUser(username) {
		// e.g., "app/app-casnode"
		return true, nil
	}

	user := c.getCurrentUser()
	if user == nil {
		return false, nil
	}

	return user.IsGlobalAdmin(), user
}

func (c *ApiController) getCurrentUser() *object.User {
	var user *object.User
	var err error
	userId := c.GetSessionUsername()
	if userId == "" {
		user = nil
	} else {
		user, err = object.GetUser(userId)
		if err != nil {
			c.ResponseErr(err)
			return nil
		}
	}
	return user
}

// GetSessionUsername ...
func (c *ApiController) GetSessionUsername() string {
	// check if user session expired
	sessionData := c.GetSessionData()

	if sessionData != nil &&
		sessionData.ExpireTime != 0 &&
		sessionData.ExpireTime < time.Now().Unix() {
		c.ClearUserSession()
		return ""
	}

	user := c.GetSession("username")
	if user == nil {
		return ""
	}

	return user.(string)
}

func (c *ApiController) GetSessionToken() string {
	accessToken := c.GetSession("accessToken")
	if accessToken == nil {
		return ""
	}

	return accessToken.(string)
}

func (c *ApiController) GetSessionApplication() *object.Application {
	clientId := c.GetSession("aud")
	if clientId == nil {
		return nil
	}
	application, err := object.GetApplicationByClientId(clientId.(string))
	if err != nil {
		c.ResponseErr(err)
		return nil
	}

	return application
}

func (c *ApiController) ClearUserSession() {
	c.SetSessionUsername("")
	c.SetSessionData(nil)
}

func (c *ApiController) ClearTokenSession() {
	c.SetSessionToken("")
}

func (c *ApiController) GetSessionOidc() (string, string) {
	sessionData := c.GetSessionData()
	if sessionData != nil &&
		sessionData.ExpireTime != 0 &&
		sessionData.ExpireTime < time.Now().Unix() {
		c.ClearUserSession()
		return "", ""
	}
	scopeValue := c.GetSession("scope")
	audValue := c.GetSession("aud")
	var scope, aud string
	var ok bool
	if scope, ok = scopeValue.(string); !ok {
		scope = ""
	}
	if aud, ok = audValue.(string); !ok {
		aud = ""
	}
	return scope, aud
}

// SetSessionUsername ...
func (c *ApiController) SetSessionUsername(user string) {
	c.SetSession("username", user)
}

func (c *ApiController) SetSessionToken(accessToken string) {
	c.SetSession("accessToken", accessToken)
}

// GetSessionData ...
func (c *ApiController) GetSessionData() *SessionData {
	session := c.GetSession("SessionData")
	if session == nil {
		return nil
	}

	sessionData := &SessionData{}
	err := util.JsonToStruct(session.(string), sessionData)
	if err != nil {
		logs.Error("GetSessionData failed, error: %s", err)
		return nil
	}

	return sessionData
}

// SetSessionData ...
func (c *ApiController) SetSessionData(s *SessionData) {
	if s == nil {
		c.DelSession("SessionData")
		return
	}

	c.SetSession("SessionData", util.StructToJson(s))
}

func (c *ApiController) setMfaUserSession(userId string) {
	c.SetSession(object.MfaSessionUserId, userId)
}

func (c *ApiController) getMfaUserSession() string {
	userId := c.Ctx.Input.CruSession.Get(object.MfaSessionUserId)
	if userId == nil {
		return ""
	}
	return userId.(string)
}

func (c *ApiController) setExpireForSession() {
	timestamp := time.Now().Unix()
	timestamp += 3600 * 24
	c.SetSessionData(&SessionData{
		ExpireTime: timestamp,
	})
}

func wrapActionResponse(affected bool, e ...error) *Response {
	if len(e) != 0 && e[0] != nil {
		return &Response{Status: "error", Msg: e[0].Error()}
	} else if affected {
		return &Response{Status: "ok", Msg: "", Data: "Affected"}
	} else {
		return &Response{Status: "ok", Msg: "", Data: "Unaffected"}
	}
}

func wrapErrorResponse(err error) *Response {
	if err == nil {
		return &Response{Status: "ok", Msg: ""}
	} else {
		return &Response{Status: "error", Msg: err.Error()}
	}
}

func (c *ApiController) Finish() {
	if strings.HasPrefix(c.Ctx.Input.URL(), "/api") {
		startTime := c.Ctx.Input.GetData("startTime")
		if startTime != nil {
			latency := time.Since(startTime.(time.Time)).Milliseconds()
			object.ApiLatency.WithLabelValues(c.Ctx.Input.URL(), c.Ctx.Input.Method()).Observe(float64(latency))
		}
	}
	c.Controller.Finish()
}
