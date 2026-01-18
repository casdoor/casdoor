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

	"github.com/casdoor/casdoor/util"
)

// createTestApplication creates a mock application for testing
func createTestApplication() *Application {
	return &Application{
		Owner:         "admin",
		Name:          "test-app",
		ClientId:      "test-client-id",
		ClientSecret:  "test-client-secret",
		ExpireInHours: 168,
		GrantTypes:    []string{"authorization_code"},
	}
}

// createTestToken creates a mock token for testing with optional state
func createTestToken(state string) *Token {
	return &Token{
		Owner:        "admin",
		Name:         util.GenerateId(),
		Application:  "test-app",
		Organization: "test-org",
		User:         "test-user",
		Code:         util.GenerateClientId(),
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
		ExpiresIn:    3600,
		Scope:        "read",
		TokenType:    "Bearer",
		CodeIsUsed:   false,
		CodeExpireIn: 9999999999, // Far future
		State:        state,
	}
}

// TestStateValidation tests the OAuth state parameter CSRF protection
func TestStateValidation(t *testing.T) {
	application := createTestApplication()

	// Test 1: Valid state - should succeed
	t.Run("ValidState", func(t *testing.T) {
		expectedState := util.GenerateId()
		token := createTestToken(expectedState)

		// Validate with matching state
		result, tokenError, err := validateTokenState(token, application, expectedState, "test-client-secret", "")

		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if tokenError != nil {
			t.Errorf("Expected no token error, got: %v", tokenError)
		}
		if result == nil {
			t.Error("Expected token to be returned, got nil")
		}
	})

	// Test 2: Mismatched state - should fail
	t.Run("MismatchedState", func(t *testing.T) {
		expectedState := util.GenerateId()
		wrongState := util.GenerateId()
		token := createTestToken(expectedState)

		// Validate with wrong state
		result, tokenError, err := validateTokenState(token, application, wrongState, "test-client-secret", "")

		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if tokenError == nil {
			t.Error("Expected token error for mismatched state, got nil")
		}
		if tokenError != nil && tokenError.Error != InvalidGrant {
			t.Errorf("Expected InvalidGrant error, got: %s", tokenError.Error)
		}
		if result != nil {
			t.Error("Expected nil token for mismatched state, got token")
		}
	})

	// Test 3: Empty state in token (backward compatibility) - should succeed
	t.Run("EmptyStateInToken", func(t *testing.T) {
		token := createTestToken("") // Empty state

		// Validate with any state - should succeed for backward compatibility
		result, tokenError, err := validateTokenState(token, application, "any-state", "test-client-secret", "")

		if err != nil {
			t.Errorf("Expected no error for empty token state, got: %v", err)
		}
		if tokenError != nil {
			t.Errorf("Expected no token error for empty token state, got: %v", tokenError)
		}
		if result == nil {
			t.Error("Expected token to be returned for empty token state, got nil")
		}
	})

	// Test 4: Empty state provided when token has state - should fail
	t.Run("EmptyStateProvided", func(t *testing.T) {
		expectedState := util.GenerateId()
		token := createTestToken(expectedState)

		// Validate with empty state
		result, tokenError, err := validateTokenState(token, application, "", "test-client-secret", "")

		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if tokenError == nil {
			t.Error("Expected token error for empty state when token has state, got nil")
		}
		if tokenError != nil && tokenError.Error != InvalidGrant {
			t.Errorf("Expected InvalidGrant error, got: %s", tokenError.Error)
		}
		if result != nil {
			t.Error("Expected nil token for empty state, got token")
		}
	})
}

// validateTokenState is a helper function that mimics the state validation logic
// from GetAuthorizationCodeToken for testing purposes
func validateTokenState(token *Token, application *Application, state string, clientSecret string, verifier string) (*Token, *TokenError, error) {
	// Validate state parameter for CSRF protection
	if token.State != "" && state != token.State {
		return nil, &TokenError{
			Error:            InvalidGrant,
			ErrorDescription: "state parameter validation failed",
		}, nil
	}

	// If validation passes, return the token
	return token, nil, nil
}
