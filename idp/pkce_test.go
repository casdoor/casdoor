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

package idp

import (
	"testing"
)

func TestTwitterIdProvider_CodeVerifier(t *testing.T) {
	// Create Twitter provider with default verifier
	provider := NewTwitterIdProvider("test_client_id", "test_client_secret", "http://localhost/callback")
	if provider.CodeVerifier != "" {
		t.Errorf("Expected empty CodeVerifier by default, got: %s", provider.CodeVerifier)
	}

	// Set custom verifier
	customVerifier := "custom-test-verifier-123"
	provider.CodeVerifier = customVerifier
	if provider.CodeVerifier != customVerifier {
		t.Errorf("Expected CodeVerifier to be %s, got: %s", customVerifier, provider.CodeVerifier)
	}
}

func TestGothIdProvider_CodeVerifier(t *testing.T) {
	// Create Goth provider for Fitbit
	provider, err := NewGothIdProvider("Fitbit", "test_client_id", "test_client_secret", "", "", "http://localhost/callback", "")
	if err != nil {
		t.Fatalf("Failed to create GothIdProvider: %v", err)
	}

	if provider.CodeVerifier != "" {
		t.Errorf("Expected empty CodeVerifier by default, got: %s", provider.CodeVerifier)
	}

	// Set custom verifier
	customVerifier := "custom-test-verifier-456"
	provider.CodeVerifier = customVerifier
	if provider.CodeVerifier != customVerifier {
		t.Errorf("Expected CodeVerifier to be %s, got: %s", customVerifier, provider.CodeVerifier)
	}
}

func TestProviderInfo_CodeVerifier(t *testing.T) {
	// Create ProviderInfo with code verifier
	providerInfo := &ProviderInfo{
		Type:         "Twitter",
		ClientId:     "test_client_id",
		ClientSecret: "test_client_secret",
		RedirectUrl:  "http://localhost/callback",
		CodeVerifier: "test-verifier-789",
	}

	if providerInfo.CodeVerifier != "test-verifier-789" {
		t.Errorf("Expected CodeVerifier to be test-verifier-789, got: %s", providerInfo.CodeVerifier)
	}

	// Test GetIdProvider with code verifier
	idProvider, err := GetIdProvider(providerInfo, providerInfo.RedirectUrl)
	if err != nil {
		t.Fatalf("Failed to get IdProvider: %v", err)
	}

	twitterProvider, ok := idProvider.(*TwitterIdProvider)
	if !ok {
		t.Fatalf("Expected TwitterIdProvider, got: %T", idProvider)
	}

	if twitterProvider.CodeVerifier != "test-verifier-789" {
		t.Errorf("Expected TwitterIdProvider.CodeVerifier to be test-verifier-789, got: %s", twitterProvider.CodeVerifier)
	}
}
