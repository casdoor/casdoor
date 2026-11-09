// Copyright 2026 The Casdoor Authors. All Rights Reserved.
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
	"strings"

	"github.com/beego/beego/v2/server/web/context"
	"github.com/casdoor/casdoor/tracing"
	"github.com/casdoor/casdoor/util"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func isTracedPath(path string) bool {
	return strings.HasPrefix(path, "/api/") || strings.HasPrefix(path, "/.well-known/") || strings.HasPrefix(path, "/cas/") ||
		strings.HasPrefix(path, "/scim/") || strings.HasPrefix(path, "/login/oauth/")
}

// TracingFilter starts the server span of an API request
func TracingFilter(ctx *context.Context) {
	if !tracing.Enabled() || !isTracedPath(ctx.Request.URL.Path) {
		return
	}
	spanCtx, span := tracing.StartRequest(ctx.Request, ctx.Request.URL.Path)
	ctx.Input.SetData("otelContext", spanCtx)
	ctx.Input.SetData("otelSpan", span)
}

// TracingEndFilter ends the server span with the response status and who made the request
func TracingEndFilter(ctx *context.Context) {
	span, ok := ctx.Input.GetData("otelSpan").(trace.Span)
	if !ok {
		return
	}
	ctx.Input.SetData("otelSpan", nil)

	owner, name := getSubject(ctx)
	status := ctx.ResponseWriter.Status
	if status == 0 {
		status = 200
	}
	tracing.EndRequest(span, status,
		attribute.String("casdoor.organization", owner),
		attribute.String("casdoor.user", name),
		attribute.String("client.address", util.GetClientIpFromRequest(ctx.Request)),
	)
}
