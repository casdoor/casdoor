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
	"github.com/casdoor/casdoor/conf"
	"github.com/casdoor/casdoor/i18n"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

type Response struct {
	Status string      `json:"status"`
	Msg    string      `json:"msg"`
	Data   interface{} `json:"data"`
	Data2  interface{} `json:"data2"`
}

func responseError(ctx *context.Context, error string, data ...interface{}) {
	resp := Response{Status: "error", Msg: error}
	switch len(data) {
	case 2:
		resp.Data2 = data[1]
		fallthrough
	case 1:
		resp.Data = data[0]
	}

	err := ctx.Output.JSON(resp, true, false)
	if err != nil {
		panic(err)
	}
}

func getAcceptLanguage(ctx *context.Context) string {
	language := ctx.Request.Header.Get("Accept-Language")
	return conf.GetLanguage(language)
}

func T(ctx *context.Context, error string) string {
	return i18n.Translate(getAcceptLanguage(ctx), error)
}

func denyRequest(ctx *context.Context) {
	responseError(ctx, T(ctx, "auth:Unauthorized operation"))
}

func getUsernameByClientIdSecret(ctx *context.Context) string {
	clientId, clientSecret, ok := ctx.Request.BasicAuth()
	if !ok {
		clientId = ctx.Input.Query("clientId")
		clientSecret = ctx.Input.Query("clientSecret")
	}

	if clientId == "" || clientSecret == "" {
		return ""
	}

	application := object.GetApplicationByClientId(clientId)
	if application == nil || application.ClientSecret != clientSecret {
		return ""
	}

	return fmt.Sprintf("app/%s", application.Name)
}

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

func setSessionExpire(ctx *context.Context, ExpireTime int64) {
	SessionData := struct{ ExpireTime int64 }{ExpireTime: ExpireTime}
	err := ctx.Input.CruSession.Set("SessionData", util.StructToJson(SessionData))
	if err != nil {
		panic(err)
	}
	ctx.Input.CruSession.SessionRelease(ctx.ResponseWriter)
}

func setSessionOidc(ctx *context.Context, scope string, aud string) {
	err := ctx.Input.CruSession.Set("scope", scope)
	if err != nil {
		panic(err)
	}
	err = ctx.Input.CruSession.Set("aud", aud)
	if err != nil {
		panic(err)
	}
	ctx.Input.CruSession.SessionRelease(ctx.ResponseWriter)
}

func parseBearerToken(ctx *context.Context) string {
	header := ctx.Request.Header.Get("Authorization")
	tokens := strings.Split(header, " ")
	if len(tokens) != 2 {
		return ""
	}

	prefix := tokens[0]
	if prefix != "Bearer" {
		return ""
	}

	return tokens[1]
}
