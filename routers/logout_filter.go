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
	"fmt"
	"sync"
	"time"

	"github.com/beego/beego/context"
	"github.com/beego/beego/logs"
	"github.com/casdoor/casdoor/conf"
)

var (
	logoutMinutes   = time.Minute * 30
	cookie2LastTime sync.Map
)

func init() {
	logoutMinutesInt, err := conf.GetConfigInt64("logoutMinutes")
	if err != nil {
		logs.Info(fmt.Sprintf("get logoutMinutes failed, err:%v. use default time duration: 30 minutes", err))
	} else {
		logoutMinutes = time.Minute * time.Duration(logoutMinutesInt)
	}
}

func LogoutFilter(ctx *context.Context) {
	owner, name := getSubject(ctx)
	if owner != "anonymous" && name != "anonymous" {
		sessionId := ctx.Input.CruSession.SessionID()
		currentTime := time.Now()
		if cookieTime, exist := cookie2LastTime.Load(sessionId); exist && cookieTime.(time.Time).Add(logoutMinutes).Before(currentTime) {
			cookie2LastTime.Delete(sessionId)
			ctx.Input.CruSession.Set("username", "")
			ctx.Input.CruSession.Set("accessToken", "")
			ctx.Input.CruSession.Delete("SessionData")
			responseError(ctx, T(ctx, "auth:Long time of no operation"))
			return
		}
		cookie2LastTime.Store(sessionId, currentTime)
	}
}
