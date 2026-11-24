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
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/beego/beego/v2/server/web"
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
		checkRequestForm(ctx, nil)
		return
	}

	limit := getRequestBodyLimit(ctx)
	if ctx.Request.ContentLength > limit {
		responseRequestBodyTooLarge(ctx, limit)
		return
	}

	body, err := io.ReadAll(http.MaxBytesReader(ctx.ResponseWriter, ctx.Request.Body, limit))
	var maxBytesErr *http.MaxBytesError
	if errors.As(err, &maxBytesErr) {
		responseRequestBodyTooLarge(ctx, limit)
		return
	}
	if err != nil {
		return
	}
	if !checkRequestForm(ctx, body) || len(body) == 0 {
		return
	}

	// Restore Request.Body so Beego's subsequent CopyBody and ParseForm can read it.
	ctx.Request.Body = io.NopCloser(bytes.NewBuffer(body))
	// Cache the raw bytes directly so controllers always have access to them.
	ctx.Input.RequestBody = body
}

// checkRequestForm parses the form the way Beego does before routing, a malformed one is answered with
// 400 here, Beego would answer it with 500
func checkRequestForm(ctx *context.Context, body []byte) bool {
	request := ctx.Request.Clone(ctx.Request.Context())
	request.Body = io.NopCloser(bytes.NewReader(body))

	var err error
	if ctx.Input.IsUpload() {
		err = request.ParseMultipartForm(web.BConfig.MaxMemory)
		if request.MultipartForm != nil {
			_ = request.MultipartForm.RemoveAll()
		}
	} else {
		err = request.ParseForm()
	}
	if err == nil {
		return true
	}

	ctx.Output.SetStatus(http.StatusBadRequest)
	responseError(ctx, fmt.Sprintf("failed to parse the request: %s", err.Error()))
	return false
}

// getRequestBodyLimit applies the same limits Beego enforces after the filters, the body
// is read into memory here before Beego gets to check it
func getRequestBodyLimit(ctx *context.Context) int64 {
	if ctx.Input.IsUpload() {
		return web.BConfig.MaxUploadSize
	}
	return web.BConfig.MaxMemory
}

func responseRequestBodyTooLarge(ctx *context.Context, limit int64) {
	ctx.Output.SetStatus(http.StatusRequestEntityTooLarge)
	responseError(ctx, fmt.Sprintf("the request body is larger than the limit: %d bytes", limit))
}
