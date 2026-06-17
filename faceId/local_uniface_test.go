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

package faceId

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLocalUniFaceProviderCheckCallsCompareEndpoint(t *testing.T) {
	var requestBody map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/compare" {
			t.Fatalf("expected path /v1/compare, got %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Fatalf("expected application/json content type, got %s", r.Header.Get("Content-Type"))
		}
		if r.Header.Get("Authorization") != "Bearer secret" {
			t.Fatalf("expected bearer token authorization, got %s", r.Header.Get("Authorization"))
		}

		err := json.NewDecoder(r.Body).Decode(&requestBody)
		if err != nil {
			t.Fatal(err)
		}

		_, _ = w.Write([]byte(`{"matched":true,"score":0.82,"threshold":0.6}`))
	}))
	defer server.Close()

	provider := NewLocalUniFaceProvider(server.URL+"/", "secret")
	matched, err := provider.Check("login-image", "registered-image")
	if err != nil {
		t.Fatal(err)
	}
	if !matched {
		t.Fatal("expected match")
	}
	if requestBody["imageA"] != "login-image" {
		t.Fatalf("expected imageA login-image, got %#v", requestBody["imageA"])
	}
	if requestBody["imageB"] != "registered-image" {
		t.Fatalf("expected imageB registered-image, got %#v", requestBody["imageB"])
	}
}

func TestLocalUniFaceProviderCheckReturnsFalseForNonMatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"matched":false,"score":0.3,"threshold":0.6}`))
	}))
	defer server.Close()

	provider := NewLocalUniFaceProvider(server.URL, "")
	matched, err := provider.Check("login-image", "registered-image")
	if err != nil {
		t.Fatal(err)
	}
	if matched {
		t.Fatal("expected non-match")
	}
}

func TestLocalUniFaceProviderCheckReturnsErrorForServiceError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"detail":"no face detected"}`, http.StatusBadRequest)
	}))
	defer server.Close()

	provider := NewLocalUniFaceProvider(server.URL, "")
	_, err := provider.Check("login-image", "registered-image")
	if err == nil {
		t.Fatal("expected error")
	}
}
