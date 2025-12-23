// Copyright 2025 The Casdoor Authors. All Rights Reserved.
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
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/beego/beego/v2/server/web/context"
)

var forbiddenChars = `/?:#&%=+;`

func FieldValidationFilter(ctx *context.Context) {
	if ctx.Input.Method() != "POST" {
		return
	}

	urlPath := ctx.Request.URL.Path
	if !(strings.HasPrefix(urlPath, "/api/add-") || strings.HasPrefix(urlPath, "/api/update-")) {
		return
	}

	bodyBytes, err := io.ReadAll(ctx.Request.Body)
	if err != nil || len(bodyBytes) == 0 {
		return
	}

	ctx.Request.Body = io.NopCloser(strings.NewReader(string(bodyBytes)))

	var requestData map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &requestData); err != nil {
		return
	}

	if value, ok := requestData["name"].(string); ok {
		if strings.ContainsAny(value, forbiddenChars) {
			responseError(ctx, fmt.Sprintf("Field 'name' contains forbidden characters: %q", forbiddenChars))
			return
		}
	}
}
