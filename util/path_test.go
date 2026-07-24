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

package util

import (
	"strings"
	"testing"
)

func TestFilterQueryMasksSensitiveParams(t *testing.T) {
	uri := "/api/logout?id_token_hint=eyJhbGciOiJSUzI1NiJ9.abc&post_logout_redirect_uri=http://127.0.0.1:8000/&accessToken=secret"
	filtered := FilterQuery(uri, []string{"accessToken", "id_token_hint"})

	if strings.Contains(filtered, "id_token_hint") {
		t.Errorf("id_token_hint should be filtered, got %q", filtered)
	}
	if strings.Contains(filtered, "accessToken") {
		t.Errorf("accessToken should be filtered, got %q", filtered)
	}
	if !strings.Contains(filtered, "post_logout_redirect_uri=") {
		t.Errorf("non-sensitive query params should be kept, got %q", filtered)
	}
	if !strings.HasPrefix(filtered, "/api/logout?") {
		t.Errorf("path should be preserved, got %q", filtered)
	}
}
