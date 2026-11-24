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
	"fmt"
	"net/http"
	"strings"

	"github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/context"
)

// NotFoundApiFilter answers an API that doesn't exist with a JSON error instead of the HTML 404 page
func NotFoundApiFilter(ctx *context.Context) {
	urlPath := ctx.Request.URL.Path
	if !strings.HasPrefix(urlPath, "/api/") || ctx.Input.RunController != nil {
		return
	}

	if _, ok := web.BeeApp.Handlers.FindRouter(ctx); ok {
		return
	}

	ctx.Output.SetStatus(http.StatusNotFound)
	responseError(ctx, fmt.Sprintf("the API: %s %s does not exist, see /swagger for the list of APIs", ctx.Request.Method, urlPath))
}
