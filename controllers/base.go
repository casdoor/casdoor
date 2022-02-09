// Copyright 2021 The casbin Authors. All Rights Reserved.
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

	"github.com/astaxie/beego"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

// controller for handlers under /api uri
type ApiController struct {
	beego.Controller
}

// controller for handlers directly under / (root)
type RootController struct {
	ApiController
}

type SessionData struct {
	ExpireTime int64
}

func (c *ApiController) IsGlobalAdmin() bool {
	username := c.GetSessionUsername()
	if strings.HasPrefix(username, "app/") {
		// e.g., "app/app-casnode"
		return true
	}

	user := object.GetUser(username)
	if user == nil {
		return false
	}

	return user.Owner == "built-in" || user.IsGlobalAdmin
}

// GetSessionUsername ...
func (c *ApiController) GetSessionUsername() string {
	// check if user session expired
	sessionData := c.GetSessionData()
	if sessionData != nil &&
		sessionData.ExpireTime != 0 &&
		sessionData.ExpireTime < time.Now().Unix() {
		c.SetSessionUsername("")
		c.SetSessionData(nil)
		return ""
	}

	user := c.GetSession("username")
	if user == nil {
		return ""
	}

	return user.(string)
}

func (c *ApiController) GetSessionOidc() (string, string) {
	sessionData := c.GetSessionData()
	if sessionData != nil &&
		sessionData.ExpireTime != 0 &&
		sessionData.ExpireTime < time.Now().Unix() {
		c.SetSessionUsername("")
		c.SetSessionData(nil)
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

// GetSessionData ...
func (c *ApiController) GetSessionData() *SessionData {
	session := c.GetSession("SessionData")
	if session == nil {
		return nil
	}

	sessionData := &SessionData{}
	err := util.JsonToStruct(session.(string), sessionData)
	if err != nil {
		panic(err)
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

func wrapActionResponse(affected bool) *Response {
	if affected {
		return &Response{Status: "ok", Msg: "", Data: "Affected"}
	} else {
		return &Response{Status: "ok", Msg: "", Data: "Unaffected"}
	}
}
