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

package routers

import (
	stdcontext "context"
	"net/http"
	"net/url"
	"time"

	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/context"
	"github.com/casdoor/casdoor/controllers"
	"github.com/casdoor/casdoor/util"
)

// extendSlidingSession moves the end of a sliding session to SlidingHours from now once less than half
// of it is left, and sends the session cookie again so that the browser keeps it as long
func extendSlidingSession(ctx *context.Context, sessionData *controllers.SessionData) {
	window := sessionData.SlidingHours * 3600
	now := time.Now().Unix()
	if sessionData.ExpireTime-now > window/2 {
		return
	}

	sessionData.ExpireTime = now + window
	err := ctx.Input.CruSession.Set(stdcontext.Background(), "SessionData", util.StructToJson(sessionData))
	if err != nil {
		logs.Error("Failed to extend the sliding session, error: %s", err)
		return
	}

	sessionConfig := web.BConfig.WebConfig.Session
	maxAge := sessionConfig.SessionCookieLifeTime
	if int64(maxAge) < window {
		maxAge = int(window)
	}

	cookie := &http.Cookie{
		Name:     sessionConfig.SessionName,
		Value:    url.QueryEscape(ctx.Input.CruSession.SessionID(stdcontext.Background())),
		Path:     "/",
		Domain:   sessionConfig.SessionDomain,
		MaxAge:   maxAge,
		Expires:  time.Now().Add(time.Duration(maxAge) * time.Second),
		HttpOnly: !sessionConfig.SessionEnableSidInHTTPHeader,
		Secure:   web.BConfig.Listen.EnableHTTPS && ctx.Request.TLS != nil,
		SameSite: sessionConfig.SessionCookieSameSite,
	}
	http.SetCookie(ctx.ResponseWriter, cookie)
}
