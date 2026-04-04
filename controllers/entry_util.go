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
	"github.com/casdoor/casdoor/log"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

func responseOtlpError(ctx *context.Context, status int, body []byte, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	req := ctx.Request
	bodyInfo := "(no body)"
	if len(body) > 0 {
		bodyInfo = fmt.Sprintf("%d bytes: %q", len(body), truncate(body, 256))
	}
	fmt.Printf("responseOtlpError: [%d] %s | %s %s | remoteAddr=%s | Content-Type=%s | User-Agent=%s | body=%s\n",
		status, msg,
		req.Method, req.URL.Path,
		req.RemoteAddr,
		req.Header.Get("Content-Type"),
		req.Header.Get("User-Agent"),
		bodyInfo,
	)
	ctx.Output.SetStatus(status)
	ctx.Output.Body([]byte(msg))
}

func truncate(b []byte, max int) []byte {
	if len(b) <= max {
		return b
	}
	return b[:max]
}

func resolveOpenClawProvider(ctx *context.Context) (*log.OpenClawProvider, int, error) {
	clientIP := util.GetClientIpFromRequest(ctx.Request)
	provider, err := object.GetOpenClawProviderByIP(clientIP)
	if err != nil {
		return nil, 500, fmt.Errorf("provider lookup failed: %w", err)
	}
	if provider == nil {
		return nil, 403, fmt.Errorf("forbidden: no OpenClaw provider configured for IP %s", clientIP)
	}
	return provider, 0, nil
}

func readProtobufBody(ctx *context.Context) []byte {
	if !strings.HasPrefix(ctx.Input.Header("Content-Type"), "application/x-protobuf") {
		preview, _ := io.ReadAll(io.LimitReader(ctx.Request.Body, 256))
		responseOtlpError(ctx, 415, preview, "unsupported content type")
		return nil
	}
	body, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		responseOtlpError(ctx, 400, nil, "read body failed")
		return nil
	}
	return body
}
