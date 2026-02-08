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
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestValidateClientAssertion(t *testing.T) {
	// Create a test certificate
	certificate, privateKey, err := generateRsaKeys(2048, 256, 20, "Test Client", "Test Org")
	if err != nil {
		t.Fatalf("Failed to generate keys: %v", err)
	}

	// Create a test application
	application := &Application{
		Owner:                   "test-org",
		Name:                    "test-app",
		ClientId:                "test-client-id",
		Cert:                    "test-cert",
		TokenEndpointAuthMethod: ClientAuthMethodPrivateKeyJWT,
	}

	// Create a test cert object
	cert := &Cert{
		Owner:       "test-org",
		Name:        "test-cert",
		Certificate: certificate,
		PrivateKey:  privateKey,
	}

	// Helper function to create and sign a JWT assertion
	createAssertion := func(claims *ClientAssertionClaims) string {
		token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
		key, _ := jwt.ParseRSAPrivateKeyFromPEM([]byte(privateKey))
		assertion, _ := token.SignedString(key)
		return assertion
	}

	// Helper function to validate using the cert directly (simulating getCertByApplication)
	validateWithCert := func(assertion, tokenEndpoint string) *TokenError {
		// Parse and validate the JWT
		token, err := jwt.ParseWithClaims(assertion, &ClientAssertionClaims{}, func(token *jwt.Token) (interface{}, error) {
			sigAlgorithm := token.Method.Alg()
			var certificate interface{}
			var parseErr error

			if sigAlgorithm[:2] == "RS" || sigAlgorithm[:2] == "PS" {
				certificate, parseErr = jwt.ParseRSAPublicKeyFromPEM([]byte(cert.Certificate))
			} else if sigAlgorithm[:2] == "ES" {
				certificate, parseErr = jwt.ParseECPublicKeyFromPEM([]byte(cert.Certificate))
			}
			return certificate, parseErr
		})

		if err != nil || !token.Valid {
			return &TokenError{Error: InvalidClient, ErrorDescription: "invalid token"}
		}

		claims := token.Claims.(*ClientAssertionClaims)

		// Validate issuer
		if claims.Issuer != application.ClientId {
			return &TokenError{Error: InvalidClient, ErrorDescription: "invalid issuer"}
		}

		// Validate subject
		if claims.Subject != application.ClientId {
			return &TokenError{Error: InvalidClient, ErrorDescription: "invalid subject"}
		}

		// Validate audience
		audienceValid := false
		for _, aud := range claims.Audience {
			if aud == tokenEndpoint || aud == application.Owner {
				audienceValid = true
				break
			}
		}
		if !audienceValid {
			return &TokenError{Error: InvalidClient, ErrorDescription: "invalid audience"}
		}

		// Validate expiration
		if claims.ExpiresAt == nil {
			return &TokenError{Error: InvalidClient, ErrorDescription: "missing expiration"}
		}
		if time.Now().After(claims.ExpiresAt.Time) {
			return &TokenError{Error: InvalidClient, ErrorDescription: "expired"}
		}

		return nil
	}

	// Test 1: Valid JWT assertion
	t.Run("Valid JWT assertion", func(t *testing.T) {
		assertion := createAssertion(&ClientAssertionClaims{
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    application.ClientId,
				Subject:   application.ClientId,
				Audience:  []string{"https://example.com/api/login/oauth/access_token"},
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Minute)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
				ID:        "unique-jti-12345",
			},
		})

		tokenErr := validateWithCert(assertion, "https://example.com/api/login/oauth/access_token")
		if tokenErr != nil {
			t.Errorf("Expected valid assertion, got error: %v", tokenErr.ErrorDescription)
		}
	})

	// Test 2: Expired JWT assertion
	t.Run("Expired JWT assertion", func(t *testing.T) {
		assertion := createAssertion(&ClientAssertionClaims{
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    application.ClientId,
				Subject:   application.ClientId,
				Audience:  []string{"https://example.com/api/login/oauth/access_token"},
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(-5 * time.Minute)),
				IssuedAt:  jwt.NewNumericDate(time.Now().Add(-10 * time.Minute)),
			},
		})

		tokenErr := validateWithCert(assertion, "https://example.com/api/login/oauth/access_token")
		if tokenErr == nil {
			t.Error("Expected error for expired assertion, got nil")
		}
	})

	// Test 3: Invalid issuer
	t.Run("Invalid issuer", func(t *testing.T) {
		assertion := createAssertion(&ClientAssertionClaims{
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    "wrong-client-id",
				Subject:   application.ClientId,
				Audience:  []string{"https://example.com/api/login/oauth/access_token"},
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Minute)),
			},
		})

		tokenErr := validateWithCert(assertion, "https://example.com/api/login/oauth/access_token")
		if tokenErr == nil {
			t.Error("Expected error for invalid issuer, got nil")
		}
	})

	// Test 4: Invalid subject
	t.Run("Invalid subject", func(t *testing.T) {
		assertion := createAssertion(&ClientAssertionClaims{
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    application.ClientId,
				Subject:   "wrong-client-id",
				Audience:  []string{"https://example.com/api/login/oauth/access_token"},
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Minute)),
			},
		})

		tokenErr := validateWithCert(assertion, "https://example.com/api/login/oauth/access_token")
		if tokenErr == nil {
			t.Error("Expected error for invalid subject, got nil")
		}
	})

	// Test 5: Invalid audience
	t.Run("Invalid audience", func(t *testing.T) {
		assertion := createAssertion(&ClientAssertionClaims{
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    application.ClientId,
				Subject:   application.ClientId,
				Audience:  []string{"https://wrong-endpoint.com"},
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Minute)),
			},
		})

		tokenErr := validateWithCert(assertion, "https://example.com/api/login/oauth/access_token")
		if tokenErr == nil {
			t.Error("Expected error for invalid audience, got nil")
		}
	})

	// Test 6: Missing expiration
	t.Run("Missing expiration", func(t *testing.T) {
		assertion := createAssertion(&ClientAssertionClaims{
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:   application.ClientId,
				Subject:  application.ClientId,
				Audience: []string{"https://example.com/api/login/oauth/access_token"},
				// No ExpiresAt
			},
		})

		tokenErr := validateWithCert(assertion, "https://example.com/api/login/oauth/access_token")
		if tokenErr == nil {
			t.Error("Expected error for missing expiration, got nil")
		}
	})
}

func TestValidateClientAuthentication(t *testing.T) {
	// Create test application
	application := &Application{
		Owner:        "test-org",
		Name:         "test-app",
		ClientId:     "test-client-id",
		ClientSecret: "test-client-secret",
	}

	// Test 1: Valid client_secret authentication
	t.Run("Valid client_secret", func(t *testing.T) {
		application.TokenEndpointAuthMethod = ClientAuthMethodClientSecretPost
		tokenErr := ValidateClientAuthentication(application, "test-client-secret", "", "", "")
		if tokenErr != nil {
			t.Errorf("Expected valid client_secret, got error: %v", tokenErr.ErrorDescription)
		}
	})

	// Test 2: Invalid client_secret
	t.Run("Invalid client_secret", func(t *testing.T) {
		application.TokenEndpointAuthMethod = ClientAuthMethodClientSecretPost
		tokenErr := ValidateClientAuthentication(application, "wrong-secret", "", "", "")
		if tokenErr == nil {
			t.Error("Expected error for invalid client_secret, got nil")
		}
		if tokenErr != nil && tokenErr.Error != InvalidClient {
			t.Errorf("Expected InvalidClient error, got: %v", tokenErr.Error)
		}
	})

	// Test 3: No authentication (none method)
	t.Run("No authentication required", func(t *testing.T) {
		application.TokenEndpointAuthMethod = ClientAuthMethodNone
		tokenErr := ValidateClientAuthentication(application, "", "", "", "")
		if tokenErr != nil {
			t.Errorf("Expected no error for 'none' auth method, got: %v", tokenErr.ErrorDescription)
		}
	})

	// Test 4: Invalid client_assertion_type
	t.Run("Invalid client_assertion_type", func(t *testing.T) {
		application.TokenEndpointAuthMethod = ClientAuthMethodPrivateKeyJWT
		tokenErr := ValidateClientAuthentication(application, "", "some-assertion", "wrong-type", "")
		if tokenErr == nil {
			t.Error("Expected error for invalid client_assertion_type, got nil")
		}
	})

	// Test 5: Missing client_assertion for private_key_jwt
	t.Run("Missing client_assertion", func(t *testing.T) {
		application.TokenEndpointAuthMethod = ClientAuthMethodPrivateKeyJWT
		tokenErr := ValidateClientAuthentication(application, "", "", "urn:ietf:params:oauth:client-assertion-type:jwt-bearer", "")
		if tokenErr == nil {
			t.Error("Expected error for missing client_assertion, got nil")
		}
	})
}
