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

package routers

import (
	"fmt"
	"time"

	"github.com/astaxie/beego/context"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

func AutoSigninFilter(ctx *context.Context) {
	//if getSessionUser(ctx) != "" {
	//	return
	//}

	// "/page?access_token=123"
	accessToken := ctx.Input.Query("accessToken")
	if accessToken != "" {
		cert := object.GetDefaultCert()
		claims, err := object.ParseJwtToken(accessToken, cert)
		if err != nil {
			responseError(ctx, "invalid JWT token")
			return
		}
		if time.Now().Unix() > claims.ExpiresAt.Unix() {
			responseError(ctx, "expired JWT token")
		}

		userId := fmt.Sprintf("%s/%s", claims.User.Owner, claims.User.Name)
		setSessionUser(ctx, userId)
		setSessionOidc(ctx, claims.Scope, claims.Audience[0])
		return
	}

	// "/page?clientId=123&clientSecret=456"
	userId := getUsernameByClientIdSecret(ctx)
	if userId != "" {
		setSessionUser(ctx, userId)
		return
	}

	// "/page?username=abc&password=123"
	userId = ctx.Input.Query("username")
	password := ctx.Input.Query("password")
	if userId != "" && password != "" {
		owner, name := util.GetOwnerAndNameFromId(userId)
		_, msg := object.CheckUserPassword(owner, name, password)
		if msg != "" {
			responseError(ctx, msg)
			return
		}

		setSessionUser(ctx, userId)
		return
	}

	// HTTP Bearer token
	// Authorization: Bearer bearerToken
	bearerToken := parseBearerToken(ctx)
	if bearerToken != "" {
		cert := object.GetDefaultCert()
		claims, err := object.ParseJwtToken(bearerToken, cert)
		if err != nil {
			responseError(ctx, err.Error())
			return
		}

		setSessionUser(ctx, fmt.Sprintf("%s/%s", claims.Owner, claims.Name))
		setSessionExpire(ctx, claims.ExpiresAt.Unix())
		setSessionOidc(ctx, claims.Scope, claims.Audience[0])
	}
}
