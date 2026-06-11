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
	"mime"
	"net/http"

	"github.com/beego/beego/v2/server/web/context"
)

// ContentTypeFilter rejects requests whose Content-Type header cannot be parsed
// by mime.ParseMediaType. An unparseable Content-Type causes Beego's form parser
// to return an error that propagates as HTTP 500. We catch it here globally and
// return 400 so every endpoint is protected without per-handler boilerplate.
func ContentTypeFilter(ctx *context.Context) {
	if ctx.Request.Method == http.MethodGet || ctx.Request.Method == http.MethodHead {
		return
	}
	ct := ctx.Request.Header.Get("Content-Type")
	if ct == "" {
		return
	}
	if _, _, err := mime.ParseMediaType(ct); err != nil {
		ctx.ResponseWriter.WriteHeader(http.StatusBadRequest)
		ctx.ResponseWriter.Header().Set("Content-Type", "application/json")
		_, _ = ctx.ResponseWriter.Write([]byte(`{"status":"error","msg":"invalid Content-Type header"}`))
		ctx.ResponseWriter.Started = true
	}
}
