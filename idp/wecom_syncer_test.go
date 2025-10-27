// Copyright 2021 The Casdoor Authors. All Rights Reserved.
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

package idp

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWeComSyncer_NewWeComSyncer(t *testing.T) {
	syncer := NewWeComSyncer("test-corp-id", "test-corp-secret", "1")

	if syncer.CorpId != "test-corp-id" {
		t.Errorf("Expected CorpId to be 'test-corp-id', got '%s'", syncer.CorpId)
	}

	if syncer.CorpSecret != "test-corp-secret" {
		t.Errorf("Expected CorpSecret to be 'test-corp-secret', got '%s'", syncer.CorpSecret)
	}

	if syncer.DepartmentId != "1" {
		t.Errorf("Expected DepartmentId to be '1', got '%s'", syncer.DepartmentId)
	}
}

func TestWeComSyncer_GetAccessToken(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/cgi-bin/gettoken" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"errcode":0,"errmsg":"ok","access_token":"test-token","expires_in":7200}`))
		}
	}))
	defer server.Close()

	syncer := NewWeComSyncer("test-corp-id", "test-corp-secret", "1")
	syncer.SetHttpClient(server.Client())

	// This test won't work as expected because the actual implementation calls the real WeCom API
	// It's here as a placeholder for future mock-based testing
}
