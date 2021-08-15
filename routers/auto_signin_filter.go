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
	"net/url"

	"github.com/astaxie/beego/context"
	"github.com/casbin/casdoor/controllers"
	"github.com/casbin/casdoor/object"
	"github.com/casbin/casdoor/util"
)

func getSessionUser(ctx *context.Context) string {
	user := ctx.Input.CruSession.Get("username")
	if user == nil {
		return ""
	}

	return user.(string)
}

func setSessionUser(ctx *context.Context, user string) {
	err := ctx.Input.CruSession.Set("username", user)
	if err != nil {
		panic(err)
	}

	// https://github.com/beego/beego/issues/3445#issuecomment-455411915
	ctx.Input.CruSession.SessionRelease(ctx.ResponseWriter)
}

func returnRequest(ctx *context.Context, msg string) {
	w := ctx.ResponseWriter
	w.WriteHeader(200)
	resp := &controllers.Response{Status: "error", Msg: msg}
	_, err := w.Write([]byte(util.StructToJson(resp)))
	if err != nil {
		panic(err)
	}
}

func AutoSigninFilter(ctx *context.Context) {
	//if getSessionUser(ctx) != "" {
	//	return
	//}

	query := ctx.Request.URL.RawQuery
	queryMap, err := url.ParseQuery(query)
	if err != nil {
		panic(err)
	}

	// "/page?access_token=123"
	accessToken := queryMap.Get("accessToken")
	if accessToken != "" {
		claims, err := object.ParseJwtToken(accessToken)
		if err != nil {
			returnRequest(ctx, "Invalid JWT token")
			return
		}

		userId := fmt.Sprintf("%s/%s", claims.User.Owner, claims.User.Name)
		setSessionUser(ctx, userId)
		return
	}

	// "/page?username=abc&password=123"
	userId := queryMap.Get("username")
	password := queryMap.Get("password")
	if userId != "" && password != "" {
		owner, name := util.GetOwnerAndNameFromId(userId)
		_, msg := object.CheckUserPassword(owner, name, password)
		if msg != "" {
			returnRequest(ctx, msg)
			return
		}

		setSessionUser(ctx, userId)
		return
	}
}
