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

// TestTokenExchangeValidation tests the validation logic for token exchange parameters
func TestTokenExchangeValidation(t *testing.T) {
	// Test 1: Missing subject_token should return error
	app := &Application{
		ClientSecret: "test-secret",
		TokenFormat:  "JWT",
	}

	_, tokenError, err := GetTokenExchangeToken(app, "test-secret", "", "", "", "", "localhost")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if tokenError == nil {
		t.Error("Expected error for missing subject_token")
	}
	if tokenError != nil && tokenError.Error != InvalidRequest {
		t.Errorf("Expected InvalidRequest error, got: %s", tokenError.Error)
	}

	// Test 2: Invalid client_secret should return error
	_, tokenError, err = GetTokenExchangeToken(app, "wrong-secret", "some-token", "", "", "", "localhost")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if tokenError == nil {
		t.Error("Expected error for invalid client_secret")
	}
	if tokenError != nil && tokenError.Error != InvalidClient {
		t.Errorf("Expected InvalidClient error, got: %s", tokenError.Error)
	}

	// Test 3: Unsupported token type should return error
	_, tokenError, err = GetTokenExchangeToken(app, "test-secret", "some-token", "urn:ietf:params:oauth:token-type:unsupported", "", "", "localhost")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if tokenError == nil {
		t.Error("Expected error for unsupported token type")
	}
	if tokenError != nil && tokenError.Error != InvalidRequest {
		t.Errorf("Expected InvalidRequest error, got: %s", tokenError.Error)
	}
}

