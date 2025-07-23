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
	"net"
	"net/url"
	"strings"

	"github.com/beego/beego/context"
	"github.com/casdoor/casdoor/conf"
	"github.com/casdoor/casdoor/errorx"
	"github.com/casdoor/casdoor/i18n"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
	"github.com/zdzh/errorx/errcode"
)


type MsgResponse struct {
	Code int         `json:"code"`
	Msg  string      `json:"message"`
	Data interface{} `json:"data"`
}

type Response struct {
	Status string      `json:"status"`
	Msg    string      `json:"msg"`
	Data   interface{} `json:"data"`
	Data2  interface{} `json:"data2"`
}

func responseError(ctx *context.Context, err error, data ...interface{}) {
	resp := WrapResponse(data, err)
	err2 := ctx.Output.JSON(resp, true, false)
	if err2 != nil {
		panic(err2)
	}
}


func WrapResponse(data interface{}, err error) *MsgResponse {
	resp := &MsgResponse{
		Code: 0,
		Msg:  "成功",
		Data: data,
	}

	if err != nil {
		enableErrorMask2 := conf.GetConfigBool("enableErrorMask2")
		enableErrorMask := conf.GetConfigBool("enableErrorMask")
		if enableErrorMask2 {
			err = errorx.DefaultErr
		} else if enableErrorMask {
			errStr := err.Error()
			if strings.HasPrefix(errStr, "The user: ") && strings.HasSuffix(errStr, " doesn't exist") || strings.HasPrefix(errStr, "用户: ") && strings.HasSuffix(errStr, "不存在") {
				err = errorx.LoginErr
			}
		}
		resp.Code = errcode.Code(err)
		resp.Msg = err.Error()

	}
	return resp
}


func getAcceptLanguage(ctx *context.Context) string {
	language := ctx.Request.Header.Get("Accept-Language")
	return conf.GetLanguage(language)
}

func T(ctx *context.Context, error string) string {
	return i18n.Translate(getAcceptLanguage(ctx), error)
}

func denyRequest(ctx *context.Context) {
	responseError(ctx, errorx.UnauthorizedErr)

}

func getUsernameByClientIdSecret(ctx *context.Context) (string, error) {
	clientId, clientSecret, ok := ctx.Request.BasicAuth()
	if !ok {
		clientId = ctx.Input.Query("clientId")
		clientSecret = ctx.Input.Query("clientSecret")
	}

	if clientId == "" || clientSecret == "" {
		return "", nil
	}

	application, err := object.GetApplicationByClientId(clientId)
	if err != nil {
		return "", err
	}
	if application == nil {
		return "", fmt.Errorf("Application not found for client ID: %s", clientId)
	}

	if application.ClientSecret != clientSecret {
		return "", fmt.Errorf("Incorrect client secret for application: %s", application.Name)
	}

	return fmt.Sprintf("app/%s", application.Name), nil
}

func getUsernameByKeys(ctx *context.Context) (string, error) {
	accessKey, accessSecret := getKeys(ctx)
	user, err := object.GetUserByAccessKey(accessKey)
	if err != nil {
		return "", err
	}

	if user == nil {
		return "", fmt.Errorf("user not found for access key: %s", accessKey)
	}

	if accessSecret != user.AccessSecret {
		return "", fmt.Errorf("incorrect access secret for user: %s", user.Name)
	}

	return user.GetId(), nil
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

func getHostname(s string) string {
	if s == "" {
		return ""
	}

	l, err := url.Parse(s)
	if err != nil {
		panic(err)
	}

	res := l.Hostname()
	return res
}

func removePort(s string) string {
	ipStr, _, err := net.SplitHostPort(s)
	if err != nil {
		ipStr = s
	}
	return ipStr
}
