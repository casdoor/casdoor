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
	"testing"

	"github.com/golang-jwt/jwt/v5"
)

func TestGetClaimsCustom_AlwaysIncludesNonceAndScope(t *testing.T) {
	// Create test claims
	claims := Claims{
		User:         &User{Owner: "admin", Name: "testuser"},
		TokenType:    "access-token",
		Nonce:        "test-nonce",
		Scope:        "openid profile",
		SigninMethod: "Password",
		Provider:     "GitHub",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:  "https://example.com",
			Subject: "user123",
		},
	}

	// Test 1: Empty token fields - nonce and scope should still be present
	tokenFields := []string{}
	result := getClaimsCustom(claims, tokenFields, nil)

	if _, ok := result["nonce"]; !ok {
		t.Error("nonce should always be present in JWT-Custom tokens")
	}
	if result["nonce"] != "test-nonce" {
		t.Errorf("Expected nonce to be 'test-nonce', got %v", result["nonce"])
	}

	if _, ok := result["scope"]; !ok {
		t.Error("scope should always be present in JWT-Custom tokens")
	}
	if result["scope"] != "openid profile" {
		t.Errorf("Expected scope to be 'openid profile', got %v", result["scope"])
	}

	// Test 2: signinMethod and provider should NOT be present if not selected
	if _, ok := result["signinMethod"]; ok {
		t.Error("signinMethod should not be present when not selected in tokenFields")
	}
	if _, ok := result["provider"]; ok {
		t.Error("provider should not be present when not selected in tokenFields")
	}

	// Test 3: signinMethod and provider SHOULD be present when selected
	tokenFieldsWithOptional := []string{"signinMethod", "provider"}
	resultWithOptional := getClaimsCustom(claims, tokenFieldsWithOptional, nil)

	if _, ok := resultWithOptional["signinMethod"]; !ok {
		t.Error("signinMethod should be present when selected in tokenFields")
	}
	if resultWithOptional["signinMethod"] != "Password" {
		t.Errorf("Expected signinMethod to be 'Password', got %v", resultWithOptional["signinMethod"])
	}

	if _, ok := resultWithOptional["provider"]; !ok {
		t.Error("provider should be present when selected in tokenFields")
	}
	if resultWithOptional["provider"] != "GitHub" {
		t.Errorf("Expected provider to be 'GitHub', got %v", resultWithOptional["provider"])
	}
}

func TestGetClaimsCustom_EmptyNonceAndScope(t *testing.T) {
	// Create test claims with empty nonce and scope
	claims := Claims{
		User:      &User{Owner: "admin", Name: "testuser"},
		TokenType: "access-token",
		Nonce:     "",
		Scope:     "",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:  "https://example.com",
			Subject: "user123",
		},
	}

	tokenFields := []string{}
	result := getClaimsCustom(claims, tokenFields, nil)

	// nonce and scope should be present even if empty
	if _, ok := result["nonce"]; !ok {
		t.Error("nonce should be present in JWT-Custom tokens even if empty")
	}
	if result["nonce"] != "" {
		t.Errorf("Expected nonce to be empty string, got %v", result["nonce"])
	}

	if _, ok := result["scope"]; !ok {
		t.Error("scope should be present in JWT-Custom tokens even if empty")
	}
	if result["scope"] != "" {
		t.Errorf("Expected scope to be empty string, got %v", result["scope"])
	}
}
