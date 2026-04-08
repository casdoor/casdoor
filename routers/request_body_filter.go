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
	"bytes"
	"io"
	"net/http"

	"github.com/beego/beego/v2/server/web/context"
)

// RequestBodyFilter reads the raw request body early (before Beego's CopyBody
// and ParseForm) and caches it in Input.RequestBody. This prevents silent data
// corruption when clients send requests without a Content-Type: application/json
// header: without this filter, Beego's ParseForm may consume the body before
// controllers can read it, causing json.Unmarshal to receive empty bytes and
// produce zero-value structs that overwrite real data on AllCols().Update().
func RequestBodyFilter(ctx *context.Context) {
	if ctx.Request.Method == http.MethodGet || ctx.Request.Method == http.MethodHead {
		return
	}
	if ctx.Request.Body == nil || ctx.Request.Body == http.NoBody {
		return
	}

	body, err := io.ReadAll(ctx.Request.Body)
	if err != nil || len(body) == 0 {
		return
	}

	// Restore Request.Body so Beego's subsequent CopyBody and ParseForm can read it.
	ctx.Request.Body = io.NopCloser(bytes.NewBuffer(body))
	// Cache the raw bytes directly so controllers always have access to them.
	ctx.Input.RequestBody = body
}
