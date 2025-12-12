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

package object

import (
	"testing"
)

func TestGetOidcDiscovery(t *testing.T) {
	// Test the OIDC discovery endpoint to ensure it returns all supported grant types
	host := "localhost:8000"
	applicationName := ""

	discovery := GetOidcDiscovery(host, applicationName)

	// Verify that all expected grant types are present
	expectedGrantTypes := []string{
		"authorization_code",
		"password",
		"client_credentials",
		"refresh_token",
		"urn:ietf:params:oauth:grant-type:device_code",
	}

	grantTypesMap := make(map[string]bool)
	for _, gt := range discovery.GrantTypesSupported {
		grantTypesMap[gt] = true
	}

	for _, expectedGT := range expectedGrantTypes {
		if !grantTypesMap[expectedGT] {
			t.Errorf("Expected grant type %s not found in GrantTypesSupported", expectedGT)
		}
	}

	// Verify that refresh_token is specifically included (the main fix)
	hasRefreshToken := false
	for _, gt := range discovery.GrantTypesSupported {
		if gt == "refresh_token" {
			hasRefreshToken = true
			break
		}
	}

	if !hasRefreshToken {
		t.Error("refresh_token grant type is missing from GrantTypesSupported")
	}
}

func TestGetOidcDiscoveryByApplication(t *testing.T) {
	// Test the OIDC discovery endpoint for application-specific configuration
	host := "localhost:8000"
	applicationName := "app-test"

	discovery := GetOidcDiscovery(host, applicationName)

	// Verify that refresh_token is included
	hasRefreshToken := false
	for _, gt := range discovery.GrantTypesSupported {
		if gt == "refresh_token" {
			hasRefreshToken = true
			break
		}
	}

	if !hasRefreshToken {
		t.Error("refresh_token grant type is missing from application-specific GrantTypesSupported")
	}
}
