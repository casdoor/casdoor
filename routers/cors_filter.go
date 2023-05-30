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
)

const (
	headerOrigin       = "Origin"
	headerAllowOrigin  = "Access-Control-Allow-Origin"
	headerAllowMethods = "Access-Control-Allow-Methods"
	headerAllowHeaders = "Access-Control-Allow-Headers"
)

func CorsFilter(ctx *context.Context) {
	origin := ctx.Input.Header(headerOrigin)
	originConf := conf.GetConfigString("origin")

	if origin != "" && originConf != "" && origin != originConf {
		ok, err := object.IsOriginAllowed(origin)
		if err != nil {
			panic(err)
		}

		if ok {
			ctx.Output.Header(headerAllowOrigin, origin)
			ctx.Output.Header(headerAllowMethods, "POST, GET, OPTIONS, DELETE")
			ctx.Output.Header(headerAllowHeaders, "Content-Type, Authorization")
		} else {
			ctx.ResponseWriter.WriteHeader(http.StatusForbidden)
			return
		}

		if ctx.Input.Method() == "OPTIONS" {
			ctx.ResponseWriter.WriteHeader(http.StatusOK)
			return
		}
	}

	if ctx.Input.Method() == "OPTIONS" {
		ctx.Output.Header(headerAllowOrigin, "*")
		ctx.Output.Header(headerAllowMethods, "POST, GET, OPTIONS, DELETE")
		ctx.ResponseWriter.WriteHeader(http.StatusOK)
		return
	}
}
