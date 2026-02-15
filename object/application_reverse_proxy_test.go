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

package object

import (
	"encoding/json"
	"testing"
)

func TestApplicationReverseProxyFields(t *testing.T) {
	// Test that reverse proxy fields can be marshaled and unmarshaled
	app := &Application{
		Owner:        "admin",
		Name:         "test-app",
		Domain:       "blog.example.com",
		OtherDomains: []string{"www.blog.example.com", "blog2.example.com"},
		UpstreamHost: "localhost:8080",
		SslMode:      "HTTPS Only",
		SslCert:      "cert-test",
	}

	// Marshal to JSON
	data, err := json.Marshal(app)
	if err != nil {
		t.Fatalf("Failed to marshal application: %v", err)
	}

	// Unmarshal from JSON
	var app2 Application
	err = json.Unmarshal(data, &app2)
	if err != nil {
		t.Fatalf("Failed to unmarshal application: %v", err)
	}

	// Verify fields
	if app2.Domain != app.Domain {
		t.Errorf("Domain mismatch: expected %s, got %s", app.Domain, app2.Domain)
	}
	if len(app2.OtherDomains) != len(app.OtherDomains) {
		t.Errorf("OtherDomains length mismatch: expected %d, got %d", len(app.OtherDomains), len(app2.OtherDomains))
	}
	if app2.UpstreamHost != app.UpstreamHost {
		t.Errorf("UpstreamHost mismatch: expected %s, got %s", app.UpstreamHost, app2.UpstreamHost)
	}
	if app2.SslMode != app.SslMode {
		t.Errorf("SslMode mismatch: expected %s, got %s", app.SslMode, app2.SslMode)
	}
	if app2.SslCert != app.SslCert {
		t.Errorf("SslCert mismatch: expected %s, got %s", app.SslCert, app2.SslCert)
	}
}
