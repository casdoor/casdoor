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
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestGetClaimsCustom(t *testing.T) {
	// Create a test user
	testUser := &User{
		Owner: "test-org",
		Name:  "test-user",
		Id:    "test-id-123",
		Tag:   "test-tag",
	}

	// Create test claims
	testClaims := Claims{
		User:         testUser,
		TokenType:    "access-token",
		Nonce:        "",
		Tag:          "",
		Scope:        "",
		Azp:          "test-client-id",
		Provider:     "",
		SigninMethod: "",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "http://localhost:8000",
			Subject:   "test-id-123",
			Audience:  []string{"test-client-id"},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			NotBefore: jwt.NewNumericDate(time.Now()),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        "test-jti",
		},
	}

	tests := []struct {
		name             string
		tokenFields      []string
		expectedFields   []string
		unexpectedFields []string
	}{
		{
			name:             "Only selected fields (Name, Id)",
			tokenFields:      []string{"Name", "Id"},
			expectedFields:   []string{"iss", "sub", "aud", "exp", "nbf", "iat", "jti", "tokenType", "azp", "name", "id"},
			unexpectedFields: []string{"nonce", "tag", "scope", "signinMethod", "provider"},
		},
		{
			name:             "Empty token fields",
			tokenFields:      []string{},
			expectedFields:   []string{"iss", "sub", "aud", "exp", "nbf", "iat", "jti", "tokenType", "azp"},
			unexpectedFields: []string{"nonce", "tag", "scope", "signinMethod", "provider"},
		},
		{
			name:             "Include provider when selected",
			tokenFields:      []string{"Name", "provider"},
			expectedFields:   []string{"iss", "sub", "aud", "exp", "nbf", "iat", "jti", "tokenType", "azp", "name", "provider"},
			unexpectedFields: []string{"nonce", "tag", "scope", "signinMethod"},
		},
		{
			name:             "Include scope when selected",
			tokenFields:      []string{"Id", "scope"},
			expectedFields:   []string{"iss", "sub", "aud", "exp", "nbf", "iat", "jti", "tokenType", "azp", "id", "scope"},
			unexpectedFields: []string{"nonce", "tag", "signinMethod", "provider"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getClaimsCustom(testClaims, tt.tokenFields, nil)

			// Check expected fields are present
			for _, field := range tt.expectedFields {
				if _, ok := result[field]; !ok {
					t.Errorf("Expected field %s to be present in claims", field)
				}
			}

			// Check unexpected fields are NOT present
			for _, field := range tt.unexpectedFields {
				if _, ok := result[field]; ok {
					t.Errorf("Expected field %s to NOT be present in claims, but it was found", field)
				}
			}
		})
	}
}

func TestGetClaimsCustomWithNonEmptyValuesNotSelected(t *testing.T) {
	// Create a test user
	testUser := &User{
		Owner: "test-org",
		Name:  "test-user",
		Id:    "test-id-123",
		Tag:   "test-tag",
	}

	// Create test claims with non-empty optional fields
	testClaims := Claims{
		User:         testUser,
		TokenType:    "access-token",
		Nonce:        "test-nonce",
		Tag:          "test-tag",
		Scope:        "openid profile",
		Azp:          "test-client-id",
		Provider:     "github",
		SigninMethod: "oauth",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "http://localhost:8000",
			Subject:   "test-id-123",
			Audience:  []string{"test-client-id"},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			NotBefore: jwt.NewNumericDate(time.Now()),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        "test-jti",
		},
	}

	// Test with only Name and Id selected, even though other fields have non-empty values
	// they should NOT be included since they are not selected
	tokenFields := []string{"Name", "Id"}
	result := getClaimsCustom(testClaims, tokenFields, nil)

	// These fields should be present
	expectedPresentFields := []string{
		"iss", "sub", "aud", "exp", "nbf", "iat", "jti",
		"tokenType", "azp", "name", "id",
	}

	for _, field := range expectedPresentFields {
		if _, ok := result[field]; !ok {
			t.Errorf("Expected field %s to be present in claims", field)
		}
	}

	// These fields should NOT be present even though they have non-empty values
	// because they were not selected in tokenFields
	unexpectedFields := []string{"nonce", "scope", "provider", "signinMethod", "tag"}
	for _, field := range unexpectedFields {
		if _, ok := result[field]; ok {
			t.Errorf("Expected field %s to NOT be present in claims (not selected in tokenFields)", field)
		}
	}
}

func TestGetClaimsCustomWithSelectedFieldsIncludingEmpty(t *testing.T) {
	// Create a test user
	testUser := &User{
		Owner: "test-org",
		Name:  "test-user",
		Id:    "test-id-123",
	}

	// Create test claims with empty optional fields
	testClaims := Claims{
		User:         testUser,
		TokenType:    "access-token",
		Nonce:        "", // empty but will be selected
		Tag:          "",
		Scope:        "", // empty but will be selected
		Azp:          "test-client-id",
		Provider:     "github", // not selected but has value
		SigninMethod: "",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "http://localhost:8000",
			Subject:   "test-id-123",
			Audience:  []string{"test-client-id"},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			NotBefore: jwt.NewNumericDate(time.Now()),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        "test-jti",
		},
	}

	// Test with Name, Id, nonce, and scope selected
	tokenFields := []string{"Name", "Id", "nonce", "scope"}
	result := getClaimsCustom(testClaims, tokenFields, nil)

	// These fields should be present (including selected empty fields)
	expectedPresentFields := []string{
		"iss", "sub", "aud", "exp", "nbf", "iat", "jti",
		"tokenType", "azp", "name", "id", "nonce", "scope",
	}

	for _, field := range expectedPresentFields {
		if _, ok := result[field]; !ok {
			t.Errorf("Expected field %s to be present in claims", field)
		}
	}

	// Provider should NOT be present even though it has a non-empty value
	// because it was not selected in tokenFields
	unexpectedFields := []string{"provider", "signinMethod", "tag"}
	for _, field := range unexpectedFields {
		if _, ok := result[field]; ok {
			t.Errorf("Expected field %s to NOT be present in claims (not selected)", field)
		}
	}

	// Verify empty selected fields have empty values
	if result["nonce"] != "" {
		t.Errorf("Expected nonce to be empty string, got %v", result["nonce"])
	}
	if result["scope"] != "" {
		t.Errorf("Expected scope to be empty string, got %v", result["scope"])
	}
}
