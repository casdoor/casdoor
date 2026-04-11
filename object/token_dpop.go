// Copyright 2026 The Casdoor Authors. All Rights Reserved.
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
	"crypto"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/golang-jwt/jwt/v5"
)

const dpopMaxAgeSeconds = 300

// DPoPProofClaims represents the payload claims of a DPoP proof JWT (RFC 9449).
type DPoPProofClaims struct {
	Jti string `json:"jti"`
	Htm string `json:"htm"`
	Htu string `json:"htu"`
	Ath string `json:"ath,omitempty"`
	jwt.RegisteredClaims
}

// ValidateDPoPProof validates a DPoP proof JWT as specified in RFC 9449.
//
//   - proofToken: the compact-serialized DPoP proof JWT from the DPoP HTTP header
//   - method:     the HTTP request method (e.g., "POST", "GET")
//   - htu:        the HTTP request URL without query string or fragment
//   - accessToken: the access token string; empty at the token endpoint,
//     non-empty at protected resource endpoints (enables ath claim validation)
//
// On success it returns the base64url-encoded SHA-256 JWK thumbprint (jkt) of
// the DPoP public key embedded in the proof header.
func ValidateDPoPProof(proofToken, method, htu, accessToken string) (string, error) {
	parts := strings.Split(proofToken, ".")
	if len(parts) != 3 {
		return "", fmt.Errorf("invalid DPoP proof JWT format")
	}

	// Decode and inspect the JOSE header before signature verification.
	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return "", fmt.Errorf("failed to decode DPoP proof header: %w", err)
	}

	var header struct {
		Typ string          `json:"typ"`
		Alg string          `json:"alg"`
		JWK json.RawMessage `json:"jwk"`
	}
	if err = json.Unmarshal(headerBytes, &header); err != nil {
		return "", fmt.Errorf("failed to parse DPoP proof header: %w", err)
	}

	// typ MUST be exactly "dpop+jwt" (RFC 9449 §4.2).
	if header.Typ != "dpop+jwt" {
		return "", fmt.Errorf("DPoP proof typ must be \"dpop+jwt\", got %q", header.Typ)
	}

	// alg MUST identify an asymmetric digital signature algorithm;
	// symmetric algorithms (HS*) are explicitly forbidden (RFC 9449 §4.2).
	if header.Alg == "" || strings.HasPrefix(header.Alg, "HS") {
		return "", fmt.Errorf("DPoP proof must use an asymmetric algorithm, got %q", header.Alg)
	}

	// jwk MUST be present (RFC 9449 §4.2).
	if len(header.JWK) == 0 {
		return "", fmt.Errorf("DPoP proof header must contain the jwk claim")
	}

	var jwkKey jose.JSONWebKey
	if err = jwkKey.UnmarshalJSON(header.JWK); err != nil {
		return "", fmt.Errorf("failed to parse DPoP JWK: %w", err)
	}

	// Compute the JWK SHA-256 thumbprint per RFC 7638.
	thumbprintBytes, err := jwkKey.Thumbprint(crypto.SHA256)
	if err != nil {
		return "", fmt.Errorf("failed to compute DPoP JWK thumbprint: %w", err)
	}
	jkt := base64.RawURLEncoding.EncodeToString(thumbprintBytes)

	// Verify the proof's signature using the public key embedded in the header.
	// WithoutClaimsValidation is used so that we can perform all claim checks
	// ourselves (jwt library exp/nbf validation is not appropriate here).
	t, err := jwt.ParseWithClaims(proofToken, &DPoPProofClaims{}, func(token *jwt.Token) (interface{}, error) {
		return jwkKey.Key, nil
	}, jwt.WithoutClaimsValidation())
	if err != nil || !t.Valid {
		return "", fmt.Errorf("DPoP proof signature verification failed: %w", err)
	}

	claims, ok := t.Claims.(*DPoPProofClaims)
	if !ok {
		return "", fmt.Errorf("failed to parse DPoP proof claims")
	}

	// htm MUST match the HTTP request method (RFC 9449 §4.2).
	if !strings.EqualFold(claims.Htm, method) {
		return "", fmt.Errorf("DPoP proof htm %q does not match request method %q", claims.Htm, method)
	}

	// htu MUST match the request URL without query/fragment (RFC 9449 §4.2).
	if !strings.EqualFold(claims.Htu, htu) {
		return "", fmt.Errorf("DPoP proof htu %q does not match request URL %q", claims.Htu, htu)
	}

	// iat MUST be present and within the acceptable time window (RFC 9449 §4.2).
	if claims.IssuedAt == nil {
		return "", fmt.Errorf("DPoP proof missing iat claim")
	}
	age := time.Since(claims.IssuedAt.Time).Abs()
	if age > time.Duration(dpopMaxAgeSeconds)*time.Second {
		return "", fmt.Errorf("DPoP proof iat is outside the acceptable time window (%d seconds)", dpopMaxAgeSeconds)
	}

	// jti MUST be present to support replay detection (RFC 9449 §4.2).
	if claims.Jti == "" {
		return "", fmt.Errorf("DPoP proof missing jti claim")
	}

	// ath MUST be validated at protected resource endpoints (RFC 9449 §4.2).
	// It is the base64url-encoded SHA-256 hash of the ASCII access token string.
	if accessToken != "" {
		hash := sha256.Sum256([]byte(accessToken))
		expectedAth := base64.RawURLEncoding.EncodeToString(hash[:])
		if claims.Ath != expectedAth {
			return "", fmt.Errorf("DPoP proof ath claim does not match access token hash")
		}
	}

	return jkt, nil
}

// GetDPoPHtu constructs the full DPoP htu URL for a given host and path.
// It uses the same origin-detection logic as the rest of the backend.
func GetDPoPHtu(host, path string) string {
	_, originBackend := getOriginFromHost(host)
	return originBackend + path
}
