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

package service

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

func TestLocalizationSiteFailureStatusAndOpaqueReason(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "https://example.invalid/", nil)
	request.Header.Set("Accept-Language", "zh-TW")
	reason := `customer:opaque <reason> 100%`
	value := util.MessageErrorf("site:Blocked by Casdoor: %s", reason)
	for _, fixture := range []struct {
		status int
		write  func(http.ResponseWriter, *http.Request, error)
	}{
		{http.StatusInternalServerError, responseErrorFrom}, {http.StatusForbidden, responseForbiddenFrom},
	} {
		recorder := httptest.NewRecorder()
		fixture.write(recorder, request, value)
		if recorder.Code != fixture.status || recorder.Body.String() != "Casdoor 已封鎖此請求："+reason {
			t.Fatalf("status or custom reason changed: %d %q", recorder.Code, recorder.Body.String())
		}
		if value.Error() != "Blocked by Casdoor: "+reason {
			t.Fatal("diagnostic original changed")
		}
	}
	recorder := httptest.NewRecorder()
	recorder.WriteHeader(http.StatusBadRequest)
	responseErrorWithoutCodeFrom(recorder, request, value)
	if recorder.Code != http.StatusBadRequest {
		t.Fatal("preselected status was overwritten")
	}
}

func TestLocalizationForwardAuthAndCallbackEarlyFailures(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "https://example.invalid/", nil)
	request.Header.Set("Accept-Language", "zh-TW")
	recorder := httptest.NewRecorder()
	HandleForwardAuth(recorder, request)
	if recorder.Code != http.StatusBadRequest || !strings.Contains(recorder.Body.String(), "缺少 X-Forwarded-Host 或 X-Original-URL 標頭") || recorder.Header().Get("Location") != "" || len(recorder.Result().Cookies()) != 0 {
		t.Fatalf("missing-header outcome changed: %d %q", recorder.Code, recorder.Body.String())
	}
	for _, query := range []url.Values{{"state": {"synthetic"}}, {"code": {"synthetic"}}} {
		recorder := httptest.NewRecorder()
		handleSiteCallback(recorder, request, &object.Site{}, "https", "example.invalid", query)
		if recorder.Code != http.StatusInternalServerError || !strings.HasPrefix(recorder.Body.String(), "Casdoor 網站錯誤：") || recorder.Header().Get("Location") != "" || len(recorder.Result().Cookies()) != 0 {
			t.Fatalf("invalid-callback outcome changed: %d %q", recorder.Code, recorder.Body.String())
		}
	}
	allowed, status, err := evaluateSiteRules(&object.Site{}, request)
	if !allowed || status != 0 || err != nil {
		t.Fatalf("empty rules no longer allow: %t %d %v", allowed, status, err)
	}
}
