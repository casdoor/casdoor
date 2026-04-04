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

package controllers

import (
	"fmt"
	"io"
	"strings"

	"github.com/beego/beego/v2/server/web/context"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

func responseOtlpError(ctx *context.Context, status int, format string, args ...interface{}) {
	ctx.Output.SetStatus(status)
	ctx.Output.Body([]byte(fmt.Sprintf(format, args...)))
}

func resolveOpenClawProvider(ctx *context.Context) string {
	clientIP := util.GetClientIpFromRequest(ctx.Request)
	provider, err := object.GetOpenClawProviderByIP(clientIP)
	if err != nil {
		responseOtlpError(ctx, 500, "provider lookup failed: %v", err)
		return ""
	}
	if provider == nil {
		responseOtlpError(ctx, 403, "forbidden: no OpenClaw provider configured for IP %s", clientIP)
		return ""
	}
	return provider.Name
}

func readProtobufBody(ctx *context.Context) []byte {
	if !strings.HasPrefix(ctx.Input.Header("Content-Type"), "application/x-protobuf") {
		responseOtlpError(ctx, 415, "unsupported content type")
		return nil
	}
	body, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		responseOtlpError(ctx, 400, "read body failed")
		return nil
	}
	return body
}
