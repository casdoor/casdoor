// Copyright 2024 The Casdoor Authors. All Rights Reserved.
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
	"net/http/httptest"
	"testing"

	"github.com/beego/beego/context"
)

func TestCorsFilterAllowsAcsEndpoint(t *testing.T) {
	tests := []struct {
		name          string
		method        string
		requestURI    string
		origin        string
		expectForbid  bool
		expectHeaders bool
	}{
		{
			name:          "POST /api/acs with Azure AD origin should be allowed",
			method:        "POST",
			requestURI:    "/api/acs",
			origin:        "https://login.microsoftonline.com",
			expectForbid:  false,
			expectHeaders: true,
		},
		{
			name:          "POST /api/acs with empty origin should pass through",
			method:        "POST",
			requestURI:    "/api/acs",
			origin:        "",
			expectForbid:  false,
			expectHeaders: false,
		},
		{
			name:          "POST /api/login/oauth/access_token should be allowed",
			method:        "POST",
			requestURI:    "/api/login/oauth/access_token",
			origin:        "https://external-service.com",
			expectForbid:  false,
			expectHeaders: true,
		},
		{
			name:          "GET /api/userinfo should be allowed",
			method:        "GET",
			requestURI:    "/api/userinfo",
			origin:        "https://external-service.com",
			expectForbid:  false,
			expectHeaders: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.requestURI, nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			w := httptest.NewRecorder()

			ctx := context.NewContext()
			ctx.Reset(w, req)
			ctx.Request = req
			ctx.ResponseWriter = &context.Response{
				ResponseWriter: w,
			}

			CorsFilter(ctx)

			// Check if 403 was returned
			if tt.expectForbid && w.Code != http.StatusForbidden {
				t.Errorf("Expected status 403, got %d", w.Code)
			}
			if !tt.expectForbid && w.Code == http.StatusForbidden {
				t.Errorf("Expected status not 403, got %d", w.Code)
			}

			// Check CORS headers
			if tt.expectHeaders {
				allowOrigin := w.Header().Get(headerAllowOrigin)
				if allowOrigin != tt.origin {
					t.Errorf("Expected Access-Control-Allow-Origin: %s, got: %s", tt.origin, allowOrigin)
				}
			}
		})
	}
}
