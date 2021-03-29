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

	"github.com/astaxie/beego/context"
	"github.com/casdoor/casdoor/object"
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

func AutoLoginFilter(ctx *context.Context) {
	query := ctx.Request.URL.RawQuery
	// query == "?access_token=123"
	accessToken := parseQuery(query, "accessToken")
	if accessToken == "" {
		return
	}

	if getSessionUser(ctx) != "" {
		return
	}

	claims, err := object.ParseJwtToken(accessToken)
	if err != nil {
		panic(err)
	}

	userId := fmt.Sprintf("%s/%s", claims.Organization, claims.Username)
	setSessionUser(ctx, userId)
}
