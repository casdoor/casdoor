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
	stdcontext "context"
	"fmt"
	"sync"
	"time"

	"github.com/beego/beego/v2/server/web/context"
	"github.com/casdoor/casdoor/conf"
	"github.com/casdoor/casdoor/object"
)

var (
	inactiveTimeoutMinutes int64
	requestTimeMap         sync.Map
)

func init() {
	var err error
	inactiveTimeoutMinutes, err = conf.GetConfigInt64("inactiveTimeoutMinutes")
	if err != nil {
		inactiveTimeoutMinutes = 0
	}

	go func() {
		for range time.Tick(time.Hour) {
			requestTimeMap.Range(func(key, value interface{}) bool {
				if time.Since(value.(time.Time)) > 7*24*time.Hour {
					requestTimeMap.Delete(key)
				}
				return true
			})
		}
	}()
}

func timeoutLogout(ctx *context.Context, sessionId string, minutes int64) {
	requestTimeMap.Delete(sessionId)
	ctx.Input.CruSession.Set(stdcontext.Background(), "username", "")
	ctx.Input.CruSession.Set(stdcontext.Background(), "accessToken", "")
	ctx.Input.CruSession.Delete(stdcontext.Background(), "SessionData")
	responseError(ctx, fmt.Sprintf(T(ctx, "auth:Timeout for inactivity of %d minutes"), minutes))
}

func TimeoutFilter(ctx *context.Context) {
	owner, name := getSubject(ctx)
	if owner == "anonymous" || name == "anonymous" {
		return
	}

	// the organization's security policy overrides the global setting
	minutes := inactiveTimeoutMinutes
	if policy := object.GetCachedSecurityPolicy(owner); policy != nil && policy.InactiveTimeoutMinutes > 0 {
		minutes = int64(policy.InactiveTimeoutMinutes)
	}

	sessionId := ctx.Input.CruSession.SessionID(stdcontext.Background())
	currentTime := time.Now()
	preRequestTime, has := requestTimeMap.Load(sessionId)
	requestTimeMap.Store(sessionId, currentTime)
	if has {
		// for the per-application timeout checked when the session signs in to an application
		ctx.Input.SetData("previousRequestTime", preRequestTime)
	}
	if minutes > 0 && has && preRequestTime.(time.Time).Add(time.Minute*time.Duration(minutes)).Before(currentTime) {
		timeoutLogout(ctx, sessionId, minutes)
	}
}
