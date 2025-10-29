// Copyright 2023 The Casdoor Authors. All Rights Reserved.
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
	"github.com/beego/beego/context"
	"github.com/casdoor/casdoor/object"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MetricsHandler wraps the Prometheus metrics handler with admin authentication
func MetricsHandler(ctx *context.Context) {
	// Get the session user
	username := getSessionUser(ctx)
	if username == "" {
		responseError(ctx, T(ctx, "general:Please login first"))
		return
	}

	// Check if user is app user (they have admin access)
	if object.IsAppUser(username) {
		promhttp.Handler().ServeHTTP(ctx.ResponseWriter, ctx.Request)
		return
	}

	// Get the user object
	user, err := object.GetUser(username)
	if err != nil {
		responseError(ctx, err.Error())
		return
	}

	if user == nil {
		responseError(ctx, T(ctx, "general:The user doesn't exist"))
		return
	}

	// Check if user is admin
	if user.Owner == "built-in" || user.IsAdmin {
		promhttp.Handler().ServeHTTP(ctx.ResponseWriter, ctx.Request)
		return
	}

	// User is not admin, deny access
	responseError(ctx, T(ctx, "general:this operation requires administrator to perform"))
}
