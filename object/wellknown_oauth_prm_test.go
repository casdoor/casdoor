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
	"strings"
	"testing"
)

func TestGetOauthProtectedResourceMetadata(t *testing.T) {
	// Test global discovery
	host := "door.casdoor.com"
	metadata := GetOauthProtectedResourceMetadata(host)

	// Verify required fields are present
	if metadata.Resource == "" {
		t.Error("Resource field should not be empty")
	}

	if len(metadata.AuthorizationServers) == 0 {
		t.Error("AuthorizationServers should not be empty")
	}

	// Verify resource and auth server match for global discovery
	if metadata.Resource != metadata.AuthorizationServers[0] {
		t.Errorf("For global discovery, Resource (%s) should match AuthorizationServers[0] (%s)",
			metadata.Resource, metadata.AuthorizationServers[0])
	}

	// Verify it starts with https for proper domain
	if len(metadata.Resource) < 8 || metadata.Resource[:8] != "https://" {
		t.Errorf("Resource should start with https:// for domain, got: %s", metadata.Resource)
	}

	// Verify bearer methods supported
	if len(metadata.BearerMethodsSupported) == 0 {
		t.Error("BearerMethodsSupported should not be empty")
	}

	// Verify scopes supported
	if len(metadata.ScopesSupported) == 0 {
		t.Error("ScopesSupported should not be empty")
	}
}

func TestGetOauthProtectedResourceMetadataByApplication(t *testing.T) {
	// Test application-specific discovery
	host := "door.casdoor.com"
	appName := "my-app"
	metadata := GetOauthProtectedResourceMetadataByApplication(host, appName)

	// Verify required fields are present
	if metadata.Resource == "" {
		t.Error("Resource field should not be empty")
	}

	if len(metadata.AuthorizationServers) == 0 {
		t.Error("AuthorizationServers should not be empty")
	}

	// Verify resource includes application name
	expectedSuffix := "/.well-known/" + appName
	if !strings.HasSuffix(metadata.Resource, expectedSuffix) {
		t.Errorf("Resource should end with %s, got: %s", expectedSuffix, metadata.Resource)
	}

	// Verify auth server includes application name
	if !strings.HasSuffix(metadata.AuthorizationServers[0], expectedSuffix) {
		t.Errorf("AuthorizationServers[0] should end with %s, got: %s", expectedSuffix, metadata.AuthorizationServers[0])
	}

	// Verify resource and auth server match for application-specific discovery
	if metadata.Resource != metadata.AuthorizationServers[0] {
		t.Errorf("For application-specific discovery, Resource (%s) should match AuthorizationServers[0] (%s)",
			metadata.Resource, metadata.AuthorizationServers[0])
	}
}

func TestOauthProtectedResourceMetadataLocalhost(t *testing.T) {
	// Test localhost (should use http://)
	host := "localhost:8000"
	metadata := GetOauthProtectedResourceMetadata(host)

	// Verify it starts with http for localhost
	if len(metadata.Resource) < 7 || metadata.Resource[:7] != "http://" {
		t.Errorf("Resource should start with http:// for localhost, got: %s", metadata.Resource)
	}

	// Verify the host is included
	if !strings.HasSuffix(metadata.Resource, host) {
		t.Errorf("Resource should end with %s, got: %s", host, metadata.Resource)
	}
}
