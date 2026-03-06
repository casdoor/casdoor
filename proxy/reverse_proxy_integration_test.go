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
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// TestReverseProxyIntegration tests the reverse proxy with a real backend server
func TestReverseProxyIntegration(t *testing.T) {
	// Create a test backend server that echoes the request path
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify headers
		headers := []string{
			"X-Forwarded-For",
			"X-Forwarded-Proto",
			"X-Real-IP",
			"X-Forwarded-Host",
		}

		for _, header := range headers {
			if r.Header.Get(header) == "" {
				t.Errorf("Expected header %s to be set", header)
			}
		}

		// Echo the path and query
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Path: " + r.URL.Path + "\n"))
		w.Write([]byte("Query: " + r.URL.RawQuery + "\n"))
		w.Write([]byte("Host: " + r.Host + "\n"))
	}))
	defer backend.Close()

	// Set up the application lookup
	SetApplicationLookup(func(domain string) *Application {
		if domain == "myapp.example.com" {
			return &Application{
				Owner:        "test-owner",
				Name:         "my-app",
				UpstreamHost: backend.URL,
			}
		}
		return nil
	})

	// Test various request paths
	tests := []struct {
		name     string
		path     string
		query    string
		expected string
	}{
		{"Simple path", "/", "", "Path: /\n"},
		{"Path with segments", "/api/v1/users", "", "Path: /api/v1/users\n"},
		{"Path with query", "/search", "q=test&limit=10", "Query: q=test&limit=10\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "http://myapp.example.com" + tt.path
			if tt.query != "" {
				url += "?" + tt.query
			}

			req := httptest.NewRequest("GET", url, nil)
			req.Host = "myapp.example.com"
			w := httptest.NewRecorder()

			HandleReverseProxy(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Expected status 200, got %d", w.Code)
			}

			body, _ := io.ReadAll(w.Body)
			bodyStr := string(body)

			if !strings.Contains(bodyStr, tt.expected) {
				t.Errorf("Expected response to contain %q, got %q", tt.expected, bodyStr)
			}
		})
	}
}

// TestReverseProxyWebSocket tests that WebSocket upgrade headers are preserved
func TestReverseProxyWebSocket(t *testing.T) {
	// Note: WebSocket upgrade through httptest.ResponseRecorder has limitations
	// This test verifies that WebSocket headers are passed through, but
	// full WebSocket functionality would need integration testing with real servers

	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify WebSocket headers are present
		if r.Header.Get("Upgrade") == "websocket" &&
			r.Header.Get("Connection") != "" &&
			r.Header.Get("Sec-WebSocket-Version") != "" &&
			r.Header.Get("Sec-WebSocket-Key") != "" {
			// Headers are present - this is what we're testing
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("WebSocket headers received"))
		} else {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Missing WebSocket headers"))
		}
	}))
	defer backend.Close()

	SetApplicationLookup(func(domain string) *Application {
		if domain == "ws.example.com" {
			return &Application{
				Owner:        "test-owner",
				Name:         "ws-app",
				UpstreamHost: backend.URL,
			}
		}
		return nil
	})

	req := httptest.NewRequest("GET", "http://ws.example.com/ws", nil)
	req.Host = "ws.example.com"
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Sec-WebSocket-Version", "13")
	req.Header.Set("Sec-WebSocket-Key", "dGhlIHNhbXBsZSBub25jZQ==")

	w := httptest.NewRecorder()
	HandleReverseProxy(w, req)

	body, _ := io.ReadAll(w.Body)
	bodyStr := string(body)

	// We expect the headers to be passed through to the backend
	if !strings.Contains(bodyStr, "WebSocket headers received") {
		t.Errorf("WebSocket headers were not properly forwarded. Got: %s", bodyStr)
	}
}

// TestReverseProxyUpstreamHostVariations tests different UpstreamHost formats
func TestReverseProxyUpstreamHostVariations(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer backend.Close()

	// Parse backend URL to get host
	backendURL, err := url.Parse(backend.URL)
	if err != nil {
		t.Fatalf("Failed to parse backend URL: %v", err)
	}

	tests := []struct {
		name         string
		upstreamHost string
		shouldWork   bool
	}{
		{"Full URL", backend.URL, true},
		{"Host only", backendURL.Host, true},
		{"Empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetApplicationLookup(func(domain string) *Application {
				if domain == "test.example.com" {
					return &Application{
						Owner:        "test-owner",
						Name:         "test-app",
						UpstreamHost: tt.upstreamHost,
					}
				}
				return nil
			})

			req := httptest.NewRequest("GET", "http://test.example.com/", nil)
			req.Host = "test.example.com"
			w := httptest.NewRecorder()

			HandleReverseProxy(w, req)

			if tt.shouldWork {
				if w.Code != http.StatusOK {
					t.Errorf("Expected status 200, got %d", w.Code)
				}
			} else {
				if w.Code == http.StatusOK {
					t.Errorf("Expected failure, but got status 200")
				}
			}
		})
	}
}
