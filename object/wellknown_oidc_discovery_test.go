// Copyright 2025 The Casdoor Authors. All Rights Reserved.
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
	host := "localhost:8000"

	// Test without application name
	discovery := GetOidcDiscovery(host, "")

	// Verify that all required fields are populated
	if discovery.Issuer == "" {
		t.Error("Issuer should not be empty")
	}

	if discovery.AuthorizationEndpoint == "" {
		t.Error("AuthorizationEndpoint should not be empty")
	}

	if discovery.TokenEndpoint == "" {
		t.Error("TokenEndpoint should not be empty")
	}

	if discovery.JwksUri == "" {
		t.Error("JwksUri should not be empty")
	}

	// Verify that code_challenge_methods_supported includes S256
	if len(discovery.CodeChallengeMethodsSupported) == 0 {
		t.Error("CodeChallengeMethodsSupported should not be empty")
	}

	found := false
	for _, method := range discovery.CodeChallengeMethodsSupported {
		if method == "S256" {
			found = true
			break
		}
	}

	if !found {
		t.Error("CodeChallengeMethodsSupported should include S256")
	}

	// Test with application name
	applicationName := "test-app"
	discoveryWithApp := GetOidcDiscovery(host, applicationName)

	// Verify that the issuer includes the application name
	expectedIssuerSuffix := "/.well-known/" + applicationName
	if len(discoveryWithApp.Issuer) < len(expectedIssuerSuffix) ||
		discoveryWithApp.Issuer[len(discoveryWithApp.Issuer)-len(expectedIssuerSuffix):] != expectedIssuerSuffix {
		t.Errorf("Issuer should end with %s when application name is provided, got: %s", expectedIssuerSuffix, discoveryWithApp.Issuer)
	}

	// Verify that code_challenge_methods_supported is also present for application-specific discovery
	if len(discoveryWithApp.CodeChallengeMethodsSupported) == 0 {
		t.Error("CodeChallengeMethodsSupported should not be empty for application-specific discovery")
	}
}
