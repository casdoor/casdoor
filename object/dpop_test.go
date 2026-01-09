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
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"testing"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/golang-jwt/jwt/v5"
)

func TestComputeJwkThumbprint(t *testing.T) {
	// Test RSA key thumbprint
	rsaKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate RSA key: %v", err)
	}

	joseKey := jose.JSONWebKey{Key: &rsaKey.PublicKey}
	jwkMap := make(map[string]interface{})
	jwkBytes, err := joseKey.MarshalJSON()
	if err != nil {
		t.Fatalf("Failed to marshal JWK: %v", err)
	}
	err = json.Unmarshal(jwkBytes, &jwkMap)
	if err != nil {
		t.Fatalf("Failed to unmarshal JWK: %v", err)
	}

	thumbprint, err := computeJwkThumbprint(jwkMap)
	if err != nil {
		t.Fatalf("Failed to compute JWK thumbprint: %v", err)
	}

	if thumbprint == "" {
		t.Error("Thumbprint should not be empty")
	}

	// Verify the thumbprint is base64url encoded
	_, err = base64.RawURLEncoding.DecodeString(thumbprint)
	if err != nil {
		t.Errorf("Thumbprint is not valid base64url: %v", err)
	}
}

func TestValidateHtu(t *testing.T) {
	tests := []struct {
		name       string
		htu        string
		requestUri string
		expected   bool
	}{
		{
			name:       "exact match",
			htu:        "https://example.com/token",
			requestUri: "https://example.com/token",
			expected:   true,
		},
		{
			name:       "with query params - should match",
			htu:        "https://example.com/token",
			requestUri: "https://example.com/token?code=abc",
			expected:   true,
		},
		{
			name:       "different path",
			htu:        "https://example.com/token",
			requestUri: "https://example.com/other",
			expected:   false,
		},
		{
			name:       "different domain",
			htu:        "https://example.com/token",
			requestUri: "https://other.com/token",
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validateHtu(tt.htu, tt.requestUri)
			if result != tt.expected {
				t.Errorf("validateHtu(%q, %q) = %v, want %v", tt.htu, tt.requestUri, result, tt.expected)
			}
		})
	}
}

func TestComputeAth(t *testing.T) {
	accessToken := "test-access-token"
	ath := computeAth(accessToken)

	if ath == "" {
		t.Error("ATH should not be empty")
	}

	// Verify the ath is base64url encoded
	_, err := base64.RawURLEncoding.DecodeString(ath)
	if err != nil {
		t.Errorf("ATH is not valid base64url: %v", err)
	}

	// Verify it's consistent
	ath2 := computeAth(accessToken)
	if ath != ath2 {
		t.Error("ATH should be consistent for the same token")
	}

	// Verify different tokens produce different ATH
	ath3 := computeAth("different-token")
	if ath == ath3 {
		t.Error("Different tokens should produce different ATH")
	}
}

func TestValidateJti(t *testing.T) {
	jti1 := "unique-jti-1"
	jti2 := "unique-jti-2"

	// First validation should succeed
	err := validateJti(jti1)
	if err != nil {
		t.Errorf("First validation should succeed: %v", err)
	}

	// Same JTI should fail (replay protection)
	err = validateJti(jti1)
	if err == nil {
		t.Error("Second validation with same JTI should fail")
	}

	// Different JTI should succeed
	err = validateJti(jti2)
	if err != nil {
		t.Errorf("Validation with different JTI should succeed: %v", err)
	}
}

func TestGetDPoPSigningAlgValuesSupported(t *testing.T) {
	algs := GetDPoPSigningAlgValuesSupported()

	if len(algs) == 0 {
		t.Error("Should return at least one supported algorithm")
	}

	// Check that common algorithms are present
	expectedAlgs := []string{"RS256", "ES256"}
	for _, expected := range expectedAlgs {
		found := false
		for _, alg := range algs {
			if alg == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected algorithm %s not found in supported algorithms", expected)
		}
	}
}

func createTestDPoPProof(t *testing.T, privateKey *rsa.PrivateKey, htm string, htu string, jti string, includeAth bool, accessToken string) string {
	publicKey := &privateKey.PublicKey

	// Create JWK
	jwk := jose.JSONWebKey{Key: publicKey, Algorithm: string(jose.RS256)}
	jwkMap := make(map[string]interface{})
	jwkBytes, err := jwk.MarshalJSON()
	if err != nil {
		t.Fatalf("Failed to marshal JWK: %v", err)
	}
	err = json.Unmarshal(jwkBytes, &jwkMap)
	if err != nil {
		t.Fatalf("Failed to unmarshal JWK: %v", err)
	}

	// Create claims
	claims := DPoPProofClaims{
		Jti: jti,
		Htm: htm,
		Htu: htu,
		Iat: time.Now().Unix(),
	}

	if includeAth {
		claims.Ath = computeAth(accessToken)
	}

	// Create token with custom header
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["typ"] = "dpop+jwt"
	token.Header["jwk"] = jwkMap

	// Sign the token
	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		t.Fatalf("Failed to sign token: %v", err)
	}

	return tokenString
}

func TestValidateDPoPProof_Basic(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate RSA key: %v", err)
	}

	httpMethod := "POST"
	httpUri := "https://example.com/token"
	jti := "test-jti-unique-123"

	dpopProof := createTestDPoPProof(t, privateKey, httpMethod, httpUri, jti, false, "")

	jkt, err := ValidateDPoPProof(dpopProof, httpMethod, httpUri, "")
	if err != nil {
		t.Errorf("ValidateDPoPProof failed: %v", err)
	}

	if jkt == "" {
		t.Error("JKT should not be empty")
	}
}

func TestValidateDPoPProof_WithAccessToken(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate RSA key: %v", err)
	}

	httpMethod := "GET"
	httpUri := "https://example.com/userinfo"
	jti := "test-jti-with-ath-456"
	accessToken := "test-access-token-value"

	dpopProof := createTestDPoPProof(t, privateKey, httpMethod, httpUri, jti, true, accessToken)

	jkt, err := ValidateDPoPProof(dpopProof, httpMethod, httpUri, accessToken)
	if err != nil {
		t.Errorf("ValidateDPoPProof with access token failed: %v", err)
	}

	if jkt == "" {
		t.Error("JKT should not be empty")
	}
}
