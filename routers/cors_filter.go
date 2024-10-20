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
	"net/http"

	"github.com/beego/beego/context"
	"github.com/casdoor/casdoor/conf"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

const (
	headerOrigin           = "Origin"
	headerAllowOrigin      = "Access-Control-Allow-Origin"
	headerAllowMethods     = "Access-Control-Allow-Methods"
	headerAllowHeaders     = "Access-Control-Allow-Headers"
	headerAllowCredentials = "Access-Control-Allow-Credentials"
)

func setCorsHeaders(ctx *context.Context, origin string) {
	ctx.Output.Header(headerAllowOrigin, origin)
	ctx.Output.Header(headerAllowMethods, "POST, GET, OPTIONS, DELETE")
	ctx.Output.Header(headerAllowHeaders, "Content-Type, Authorization")
	ctx.Output.Header(headerAllowCredentials, "true")

	if ctx.Input.Method() == "OPTIONS" {
		ctx.ResponseWriter.WriteHeader(http.StatusOK)
	}
}

func CorsFilter(ctx *context.Context) {
	origin := ctx.Input.Header(headerOrigin)
	originConf := conf.GetConfigString("origin")
	originHostname := getHostname(origin)
	host := removePort(ctx.Request.Host)

	if origin == "null" {
		origin = ""
	}

	isValid, err := util.IsValidOrigin(origin)
	if err != nil {
		ctx.ResponseWriter.WriteHeader(http.StatusForbidden)
		responseError(ctx, err.Error())
		return
	}
	if isValid {
		setCorsHeaders(ctx, origin)
		return
	}

	if originHostname == "appleid.apple.com" {
		setCorsHeaders(ctx, origin)
		return
	}

	if ctx.Request.Method == "POST" && ctx.Request.RequestURI == "/api/login/oauth/access_token" {
		setCorsHeaders(ctx, origin)
		return
	}

	if ctx.Request.RequestURI == "/api/userinfo" {
		setCorsHeaders(ctx, origin)
		return
	}

	if origin != "" {
		if origin == originConf {
			setCorsHeaders(ctx, origin)
		} else if originHostname == host {
			setCorsHeaders(ctx, origin)
		} else if isHostIntranet(host) {
			setCorsHeaders(ctx, origin)
		} else {
			ok, err := object.IsOriginAllowed(origin)
			if err != nil {
				panic(err)
			}

			if ok {
				setCorsHeaders(ctx, origin)
			} else {
				ctx.ResponseWriter.WriteHeader(http.StatusForbidden)
				return
			}
		}
	}

	if ctx.Input.Method() == "OPTIONS" {
		ctx.Output.Header(headerAllowOrigin, "*")
		ctx.Output.Header(headerAllowMethods, "POST, GET, OPTIONS, DELETE")
		ctx.ResponseWriter.WriteHeader(http.StatusOK)
		return
	}
}
