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

package routers

import (
	"strconv"
	"sync"
	"time"

	"github.com/beego/beego/context"
	"github.com/casdoor/casdoor/conf"
)

var (
	logoutMinutes        int
	requestTimeMap      sync.Map
)

func init() {
	logoutMinutes, err := strconv.Atoi(conf.GetConfigString("logoutMinutes"))
	if err != nil || logoutMinutes < 0 {
		logoutMinutes = 0
	}
}

func timeoutLogout(ctx *context.Context, sessionId string) {
	requestTimeMap.Delete(sessionId)
	ctx.Input.CruSession.Set("username", "")
	ctx.Input.CruSession.Set("accessToken", "")
	ctx.Input.CruSession.Delete("SessionData")
	responseError(ctx, fmt.Sprintf(T(ctx, "auth:Timeout for inactivity of %d minutes"), logoutMinutes))
}

func LogoutFilter(ctx *context.Context) {
	if logoutMinutes <= 0 {
		return
	}

	owner, name := getSubject(ctx)
	if owner == "anonymous" || name == "anonymous" {
		return
	}

	sessionId := ctx.Input.CruSession.SessionID()
	currentTime := time.Now()
	preRequestTime, has := requestTimeMap.Load(sessionId)
	requestTimeMap.Store(sessionId, currentTime)
	if has && preRequestTime.(time.Time).Add(time.Minute * time.Duration(logoutMinutes)).Before(currentTime) {
		timeoutLogout(ctx, sessionId)
	}
}
