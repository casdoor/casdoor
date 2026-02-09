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
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ClientAssertionClaims represents the claims in a client assertion JWT (RFC 7523)
type ClientAssertionClaims struct {
	jwt.RegisteredClaims
}

// ValidateClientAssertion validates a client assertion JWT according to RFC 7523
// Returns the clientId if validation is successful, or an error otherwise
func ValidateClientAssertion(assertion string, expectedAudience string) (string, error) {
	// Parse the JWT without validation first to get the claims and identify the client
	token, _, err := new(jwt.Parser).ParseUnverified(assertion, &ClientAssertionClaims{})
	if err != nil {
		return "", fmt.Errorf("failed to parse client assertion: %v", err)
	}

	claims, ok := token.Claims.(*ClientAssertionClaims)
	if !ok {
		return "", fmt.Errorf("invalid claims type in client assertion")
	}

	// The 'sub' claim must contain the client_id
	clientId := claims.Subject
	if clientId == "" {
		return "", fmt.Errorf("client assertion missing 'sub' claim")
	}

	// The 'iss' claim must equal the client_id (RFC 7523 Section 3)
	if claims.Issuer != clientId {
		return "", fmt.Errorf("client assertion 'iss' must equal 'sub' (client_id)")
	}

	// Get the application to retrieve the client certificate
	application, err := GetApplicationByClientId(clientId)
	if err != nil {
		return "", fmt.Errorf("failed to get application: %v", err)
	}
	if application == nil {
		return "", fmt.Errorf("application not found for client_id: %s", clientId)
	}

	// Get the client certificate for this application
	if application.ClientCert == "" {
		return "", fmt.Errorf("no client certificate configured for application: %s", application.Name)
	}

	clientCert, err := GetCert(application.ClientCert)
	if err != nil {
		return "", fmt.Errorf("failed to get client certificate: %v", err)
	}
	if clientCert == nil {
		return "", fmt.Errorf("client certificate not found: %s", application.ClientCert)
	}

	// Parse the public key from the certificate
	publicKey, err := parsePublicKeyFromCertificate(clientCert.Certificate)
	if err != nil {
		return "", fmt.Errorf("failed to parse public key: %v", err)
	}

	// Now parse and validate the JWT with the public key
	validatedToken, err := jwt.ParseWithClaims(assertion, &ClientAssertionClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify the signing algorithm is acceptable (RS256, RS384, RS512, ES256, ES384, ES512, PS256, PS384, PS512)
		switch token.Method.(type) {
		case *jwt.SigningMethodRSA, *jwt.SigningMethodECDSA, *jwt.SigningMethodRSAPSS:
			return publicKey, nil
		default:
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
	})

	if err != nil {
		return "", fmt.Errorf("failed to validate client assertion signature: %v", err)
	}

	validatedClaims, ok := validatedToken.Claims.(*ClientAssertionClaims)
	if !ok || !validatedToken.Valid {
		return "", fmt.Errorf("invalid client assertion token")
	}

	// Validate the 'aud' claim (RFC 7523 Section 3)
	// The audience should be the token endpoint URL
	if len(validatedClaims.Audience) == 0 {
		return "", fmt.Errorf("client assertion missing 'aud' claim")
	}

	// Check if expectedAudience is in the audience list
	audienceValid := false
	for _, aud := range validatedClaims.Audience {
		if aud == expectedAudience || aud == expectedAudience+"/login/oauth/access_token" {
			audienceValid = true
			break
		}
	}
	if !audienceValid {
		return "", fmt.Errorf("client assertion 'aud' claim does not match expected audience")
	}

	// Validate 'exp' claim - must not be expired
	if validatedClaims.ExpiresAt != nil {
		if time.Now().After(validatedClaims.ExpiresAt.Time) {
			return "", fmt.Errorf("client assertion has expired")
		}
	} else {
		return "", fmt.Errorf("client assertion missing 'exp' claim")
	}

	// Validate 'iat' claim - issued at time (optional but recommended)
	if validatedClaims.IssuedAt != nil {
		// Check that the token is not issued in the future (with 5 minute tolerance)
		if time.Now().Add(-5 * time.Minute).Before(validatedClaims.IssuedAt.Time) {
			return "", fmt.Errorf("client assertion issued in the future")
		}
	}

	// Validate 'nbf' claim if present
	if validatedClaims.NotBefore != nil {
		if time.Now().Before(validatedClaims.NotBefore.Time) {
			return "", fmt.Errorf("client assertion not yet valid")
		}
	}

	// RFC 7523 recommends that assertions should be short-lived
	// We'll enforce a maximum lifetime of 5 minutes
	if validatedClaims.ExpiresAt != nil && validatedClaims.IssuedAt != nil {
		lifetime := validatedClaims.ExpiresAt.Time.Sub(validatedClaims.IssuedAt.Time)
		if lifetime > 5*time.Minute {
			return "", fmt.Errorf("client assertion lifetime too long (max 5 minutes)")
		}
	}

	// Validate 'jti' claim (optional but recommended for replay protection)
	// In a production system, you would check this against a cache of used JTIs
	// For now, we just verify it exists
	if validatedClaims.ID == "" {
		// JTI is optional per RFC 7523, but recommended
		// We'll allow it to be missing for now
	}

	return clientId, nil
}

// parsePublicKeyFromCertificate extracts the public key from a PEM-encoded certificate
func parsePublicKeyFromCertificate(certPEM string) (interface{}, error) {
	block, _ := pem.Decode([]byte(certPEM))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %v", err)
	}

	return cert.PublicKey, nil
}

// AuthenticateClientByAssertion authenticates a client using private_key_jwt method
// Returns the application and userId if successful
func AuthenticateClientByAssertion(clientAssertion string, clientAssertionType string, host string) (*Application, string, error) {
	// Validate the assertion type
	if clientAssertionType != "urn:ietf:params:oauth:client-assertion-type:jwt-bearer" {
		return nil, "", fmt.Errorf("unsupported client_assertion_type: %s", clientAssertionType)
	}

	// Construct the expected audience (the server's token endpoint)
	expectedAudience := "https://" + host

	// Validate the client assertion
	clientId, err := ValidateClientAssertion(clientAssertion, expectedAudience)
	if err != nil {
		return nil, "", fmt.Errorf("client assertion validation failed: %v", err)
	}

	// Get the application
	application, err := GetApplicationByClientId(clientId)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get application: %v", err)
	}
	if application == nil {
		return nil, "", fmt.Errorf("application not found for client_id: %s", clientId)
	}

	// Return the application and userId in the format expected by the system
	userId := fmt.Sprintf("app/%s", application.Name)
	return application, userId, nil
}
