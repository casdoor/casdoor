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

package proxy

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetDomainWithoutPort(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"example.com", "example.com"},
		{"example.com:8080", "example.com"},
		{"localhost:3000", "localhost"},
		{"subdomain.example.com:443", "subdomain.example.com"},
	}

	for _, test := range tests {
		result := getDomainWithoutPort(test.input)
		if result != test.expected {
			t.Errorf("getDomainWithoutPort(%s) = %s; want %s", test.input, result, test.expected)
		}
	}
}

func TestHandleReverseProxy(t *testing.T) {
	// Create a test backend server
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check that headers are set correctly
		if r.Header.Get("X-Forwarded-For") == "" {
			t.Error("X-Forwarded-For header not set")
		}
		if r.Header.Get("X-Forwarded-Proto") == "" {
			t.Error("X-Forwarded-Proto header not set")
		}
		if r.Header.Get("X-Real-IP") == "" {
			t.Error("X-Real-IP header not set")
		}
		if r.Header.Get("X-Forwarded-Host") == "" {
			t.Error("X-Forwarded-Host header not set")
		}
		
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Backend response")
	}))
	defer backend.Close()

	// Set up a mock application lookup function
	SetApplicationLookup(func(domain string) *Application {
		if domain == "test.example.com" {
			return &Application{
				Owner:        "test-owner",
				Name:         "test-app",
				UpstreamHost: backend.URL,
			}
		}
		return nil
	})

	// Test successful proxy
	req := httptest.NewRequest("GET", "http://test.example.com/path", nil)
	req.Host = "test.example.com"
	w := httptest.NewRecorder()

	HandleReverseProxy(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Test domain not found
	req = httptest.NewRequest("GET", "http://unknown.example.com/path", nil)
	req.Host = "unknown.example.com"
	w = httptest.NewRecorder()

	HandleReverseProxy(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404 for unknown domain, got %d", w.Code)
	}

	// Test application without upstream host
	SetApplicationLookup(func(domain string) *Application {
		if domain == "no-upstream.example.com" {
			return &Application{
				Owner:        "test-owner",
				Name:         "test-app-no-upstream",
				UpstreamHost: "",
			}
		}
		return nil
	})

	req = httptest.NewRequest("GET", "http://no-upstream.example.com/path", nil)
	req.Host = "no-upstream.example.com"
	w = httptest.NewRecorder()

	HandleReverseProxy(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404 for app without upstream, got %d", w.Code)
	}
}

func TestApplicationLookup(t *testing.T) {
	// Test setting and using the application lookup function
	called := false
	SetApplicationLookup(func(domain string) *Application {
		called = true
		return &Application{
			Owner:        "test",
			Name:         "app",
			UpstreamHost: "http://localhost:8080",
		}
	})

	if applicationLookup == nil {
		t.Error("applicationLookup should not be nil after SetApplicationLookup")
	}

	app := applicationLookup("test.com")
	if !called {
		t.Error("applicationLookup function was not called")
	}
	if app == nil {
		t.Error("applicationLookup should return non-nil application")
	}
	if app.Owner != "test" {
		t.Errorf("Expected owner 'test', got '%s'", app.Owner)
	}
}
