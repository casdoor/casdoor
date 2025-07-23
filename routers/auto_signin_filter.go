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

package routers

import (
	"fmt"
	"strings"

	"github.com/beego/beego/context"
	"github.com/casdoor/casdoor/errorx"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

func AutoSigninFilter(ctx *context.Context) {
	urlPath := ctx.Request.URL.Path
	if strings.HasPrefix(urlPath, "/api/login/oauth/access_token") {
		return
	}
	//if getSessionUser(ctx) != "" {
	//	return
	//}

	// GET parameter like "/page?access_token=123" or
	// HTTP Bearer token like "Authorization: Bearer 123"
	accessToken := ctx.Input.Query("accessToken")
	if accessToken == "" {
		accessToken = ctx.Input.Query("access_token")
	}
	if accessToken == "" {
		accessToken = parseBearerToken(ctx)
	}

	if accessToken != "" {
		token, err := object.GetTokenByAccessToken(accessToken)
		if err != nil {
			util.LogWarning(ctx, "Get access token failed, err:%s ", err)
			responseError(ctx, errorx.AuthenticationErr)
			return
		}

		if token == nil {
			responseError(ctx, errorx.UnAuthenticationErr)
			return
		}

		isExpired, _ := util.IsTokenExpired(token.CreatedTime, token.ExpiresIn)
		if isExpired {
			responseError(ctx, errorx.AuthenticationExpiredErr)
			return
		}

		userId := util.GetId(token.Organization, token.User)
		application, err := object.GetApplicationByUserId(fmt.Sprintf("app/%s", token.Application))
		if err != nil {
			util.LogWarning(ctx, "GetApplicationByUserId failed, err:%s ", err)
			responseError(ctx, errorx.AuthenticationErr)
			return
		}
		if application == nil {
			util.LogWarning(ctx, fmt.Sprintf("No application is found for userId: app/%s", token.Application))
			responseError(ctx, errorx.UnauthorizedErrFunc("用户无权限访问该应用"))
			return
		}

		setSessionUser(ctx, userId)
		setSessionOidc(ctx, token.Scope, application.ClientId)
		return
	}

	accessKey := ctx.Input.Query("accessKey")
	accessSecret := ctx.Input.Query("accessSecret")
	if accessKey != "" && accessSecret != "" {
		userId, err := getUsernameByKeys(ctx)
		if err != nil {
			responseError(ctx, err)
		}

		setSessionUser(ctx, userId)
	}

	// "/page?clientId=123&clientSecret=456"
	userId, err := getUsernameByClientIdSecret(ctx)
	if err != nil {
		responseError(ctx, err)
		return
	}
	if userId != "" {
		setSessionUser(ctx, userId)
		return
	}

	// "/page?username=built-in/admin&password=123"
	userId = ctx.Input.Query("username")
	password := ctx.Input.Query("password")
	if userId != "" && password != "" && ctx.Input.Query("grant_type") == "" {
		owner, name := util.GetOwnerAndNameFromId(userId)
		_, err = object.CheckUserPassword(owner, name, password, "en")
		if err != nil {
			responseError(ctx, err)
			return
		}

		setSessionUser(ctx, userId)
	}
}
