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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/beego/beego/v2/server/web/context"
)

func TestSetCorsHeadersWithCredentials(t *testing.T) {
	tests := []struct {
		name             string
		allowHeaders     string
		origin           string
		allowCredentials bool
		wantHeaders      string
		wantCredentials  string
	}{
		{"configured preflight", "Content-Type, Authorization, X-Requested-With", "https://example.com", true, "Content-Type, Authorization, X-Requested-With", "true"},
		{"empty configuration uses legacy headers", "", "https://example.com", true, "Content-Type, Authorization", "true"},
		{"non-credentialed origin", "X-Requested-With", "https://example.com", false, "X-Requested-With", ""},
		{"no origin", "X-Requested-With", "", false, "", ""},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv("corsAllowHeaders", test.allowHeaders)
			request := httptest.NewRequest(http.MethodOptions, "/api/login/oauth/access_token", nil)
			request.Header.Set(headerOrigin, test.origin)
			request.Header.Set("Access-Control-Request-Headers", "X-Requested-With")
			recorder := httptest.NewRecorder()
			ctx := context.NewContext()
			ctx.Reset(recorder, request)

			setCorsHeadersWithCredentials(ctx, test.origin, test.allowCredentials)

			if got := recorder.Code; got != http.StatusOK {
				t.Errorf("status = %d, want %d", got, http.StatusOK)
			}
			if got := recorder.Header().Get(headerAllowHeaders); got != test.wantHeaders {
				t.Errorf("%s = %q, want %q", headerAllowHeaders, got, test.wantHeaders)
			}
			if got := recorder.Header().Get(headerAllowCredentials); got != test.wantCredentials {
				t.Errorf("%s = %q, want %q", headerAllowCredentials, got, test.wantCredentials)
			}
		})
	}
}
