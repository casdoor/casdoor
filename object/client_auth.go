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
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	ClientAuthMethodClientSecretBasic = "client_secret_basic"
	ClientAuthMethodClientSecretPost  = "client_secret_post"
	ClientAuthMethodPrivateKeyJWT     = "private_key_jwt"
	ClientAuthMethodNone              = "none"
)

// ClientAssertionClaims represents the JWT claims for client authentication as per RFC 7523
type ClientAssertionClaims struct {
	jwt.RegisteredClaims
}

// ValidateClientAuthentication validates client authentication using either client_secret or private_key_jwt
// Returns TokenError if authentication fails, nil if successful
func ValidateClientAuthentication(application *Application, clientSecret, clientAssertion, clientAssertionType string, tokenEndpoint string) *TokenError {
	// Determine the authentication method
	authMethod := application.TokenEndpointAuthMethod
	if authMethod == "" {
		// Default to client_secret_basic or client_secret_post if not specified
		if clientSecret != "" {
			authMethod = ClientAuthMethodClientSecretPost
		} else if clientAssertion != "" {
			authMethod = ClientAuthMethodPrivateKeyJWT
		} else {
			// No authentication provided
			authMethod = ClientAuthMethodNone
		}
	}

	switch authMethod {
	case ClientAuthMethodClientSecretBasic, ClientAuthMethodClientSecretPost:
		// Validate client secret
		if application.ClientSecret != clientSecret {
			return &TokenError{
				Error:            InvalidClient,
				ErrorDescription: fmt.Sprintf("client_secret is invalid for application: [%s]", application.GetId()),
			}
		}
		return nil

	case ClientAuthMethodPrivateKeyJWT:
		// Validate private_key_jwt assertion
		if clientAssertionType != "urn:ietf:params:oauth:client-assertion-type:jwt-bearer" {
			return &TokenError{
				Error:            InvalidClient,
				ErrorDescription: fmt.Sprintf("invalid client_assertion_type, expected 'urn:ietf:params:oauth:client-assertion-type:jwt-bearer', got: [%s]", clientAssertionType),
			}
		}

		if clientAssertion == "" {
			return &TokenError{
				Error:            InvalidClient,
				ErrorDescription: "client_assertion is required for private_key_jwt authentication",
			}
		}

		// Validate the JWT assertion
		return validateClientAssertion(application, clientAssertion, tokenEndpoint)

	case ClientAuthMethodNone:
		// No authentication required (e.g., for PKCE public clients)
		return nil

	default:
		return &TokenError{
			Error:            InvalidClient,
			ErrorDescription: fmt.Sprintf("unsupported token_endpoint_auth_method: [%s]", authMethod),
		}
	}
}

// validateClientAssertion validates a JWT assertion for client authentication (RFC 7523)
func validateClientAssertion(application *Application, assertion string, tokenEndpoint string) *TokenError {
	// Get the certificate/public key for this application
	cert, err := getCertByApplication(application)
	if err != nil {
		return &TokenError{
			Error:            EndpointError,
			ErrorDescription: fmt.Sprintf("failed to get certificate for application: %s", err.Error()),
		}
	}

	if cert == nil || cert.Certificate == "" {
		return &TokenError{
			Error:            InvalidClient,
			ErrorDescription: "no certificate configured for application, required for private_key_jwt authentication",
		}
	}

	// Parse and validate the JWT
	token, err := jwt.ParseWithClaims(assertion, &ClientAssertionClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing algorithm
		sigAlgorithm := token.Method.Alg()
		
		// Get the public key based on the algorithm
		var certificate interface{}
		var parseErr error
		
		if sigAlgorithm[:2] == "RS" || sigAlgorithm[:2] == "PS" {
			certificate, parseErr = jwt.ParseRSAPublicKeyFromPEM([]byte(cert.Certificate))
		} else if sigAlgorithm[:2] == "ES" {
			certificate, parseErr = jwt.ParseECPublicKeyFromPEM([]byte(cert.Certificate))
		} else {
			return nil, fmt.Errorf("unsupported signing algorithm: %s", sigAlgorithm)
		}

		if parseErr != nil {
			return nil, fmt.Errorf("failed to parse certificate: %s", parseErr.Error())
		}

		return certificate, nil
	})

	if err != nil {
		return &TokenError{
			Error:            InvalidClient,
			ErrorDescription: fmt.Sprintf("failed to validate client assertion: %s", err.Error()),
		}
	}

	if !token.Valid {
		return &TokenError{
			Error:            InvalidClient,
			ErrorDescription: "client assertion token is invalid",
		}
	}

	// Validate claims according to RFC 7523
	claims, ok := token.Claims.(*ClientAssertionClaims)
	if !ok {
		return &TokenError{
			Error:            InvalidClient,
			ErrorDescription: "invalid claims in client assertion",
		}
	}

	// Validate issuer (iss) - must be the client_id
	if claims.Issuer != application.ClientId {
		return &TokenError{
			Error:            InvalidClient,
			ErrorDescription: fmt.Sprintf("invalid issuer in client assertion, expected: [%s], got: [%s]", application.ClientId, claims.Issuer),
		}
	}

	// Validate subject (sub) - must be the client_id
	if claims.Subject != application.ClientId {
		return &TokenError{
			Error:            InvalidClient,
			ErrorDescription: fmt.Sprintf("invalid subject in client assertion, expected: [%s], got: [%s]", application.ClientId, claims.Subject),
		}
	}

	// Validate audience (aud) - must be the token endpoint URL or the authorization server's issuer identifier
	audienceValid := false
	if len(claims.Audience) > 0 {
		for _, aud := range claims.Audience {
			if aud == tokenEndpoint || aud == application.Owner {
				audienceValid = true
				break
			}
		}
	}

	if !audienceValid {
		return &TokenError{
			Error:            InvalidClient,
			ErrorDescription: fmt.Sprintf("invalid audience in client assertion, expected token endpoint or issuer"),
		}
	}

	// Validate expiration time (exp) - must be present and in the future
	if claims.ExpiresAt == nil {
		return &TokenError{
			Error:            InvalidClient,
			ErrorDescription: "expiration time (exp) is required in client assertion",
		}
	}

	if time.Now().After(claims.ExpiresAt.Time) {
		return &TokenError{
			Error:            InvalidClient,
			ErrorDescription: "client assertion has expired",
		}
	}

	// Validate not before (nbf) if present
	if claims.NotBefore != nil && time.Now().Before(claims.NotBefore.Time) {
		return &TokenError{
			Error:            InvalidClient,
			ErrorDescription: "client assertion not yet valid (nbf)",
		}
	}

	// JWT ID (jti) should be present for replay protection (recommended but not required)
	// In production, you might want to store used JTIs to prevent replay attacks

	return nil
}
