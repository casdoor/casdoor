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
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/golang-jwt/jwt/v5"
)

// DPoPProofClaims represents the claims in a DPoP proof JWT
type DPoPProofClaims struct {
	Jti string `json:"jti"`
	Htm string `json:"htm"`
	Htu string `json:"htu"`
	Iat int64  `json:"iat"`
	Ath string `json:"ath,omitempty"`
	jwt.RegisteredClaims
}

// DPoPHeader represents the JOSE header of a DPoP proof
type DPoPHeader struct {
	Typ string                 `json:"typ"`
	Alg string                 `json:"alg"`
	Jwk map[string]interface{} `json:"jwk"`
}

// dpopNonceStore stores nonces for DPoP replay protection
var dpopNonceStore = &sync.Map{}
var cleanupStarted sync.Once

const (
	dpopNonceExpiration        = 5 * time.Minute
	dpopJtiExpiration          = 1 * time.Hour
	dpopIatToleranceSeconds    = 60
	dpopCleanupIntervalSeconds = 300 // Clean up expired JTIs every 5 minutes
)

// DPoPNonce represents a stored nonce with expiration
type DPoPNonce struct {
	Value     string
	ExpiresAt time.Time
}

// startCleanupTask starts a background goroutine to periodically clean up expired JTIs
func startCleanupTask() {
	cleanupStarted.Do(func() {
		go func() {
			ticker := time.NewTicker(dpopCleanupIntervalSeconds * time.Second)
			defer ticker.Stop()

			for range ticker.C {
				cleanupExpiredJtis()
			}
		}()
	})
}

// ValidateDPoPProof validates a DPoP proof JWT according to RFC 9449
func ValidateDPoPProof(dpopProof string, httpMethod string, httpUri string, accessToken string) (string, error) {
	if dpopProof == "" {
		return "", nil
	}

	// Parse the JWT without validation first to get the header
	parts := strings.Split(dpopProof, ".")
	if len(parts) != 3 {
		return "", fmt.Errorf("invalid DPoP proof format")
	}

	// Decode header
	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return "", fmt.Errorf("failed to decode DPoP proof header: %w", err)
	}

	var header DPoPHeader
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return "", fmt.Errorf("failed to parse DPoP proof header: %w", err)
	}

	// Validate header
	if header.Typ != "dpop+jwt" {
		return "", fmt.Errorf("invalid DPoP proof typ header, expected 'dpop+jwt', got '%s'", header.Typ)
	}

	// Validate algorithm
	supportedAlgs := []string{"RS256", "RS512", "ES256", "ES384", "ES512", "PS256", "PS384", "PS512"}
	algSupported := false
	for _, alg := range supportedAlgs {
		if header.Alg == alg {
			algSupported = true
			break
		}
	}
	if !algSupported {
		return "", fmt.Errorf("unsupported DPoP proof algorithm: %s", header.Alg)
	}

	// Validate JWK is present
	if header.Jwk == nil {
		return "", fmt.Errorf("DPoP proof missing jwk in header")
	}

	// Validate required JWK fields
	if _, ok := header.Jwk["kty"]; !ok {
		return "", fmt.Errorf("DPoP proof jwk missing kty")
	}

	// Extract and verify public key
	publicKey, err := extractPublicKeyFromJwk(header.Jwk, header.Alg)
	if err != nil {
		return "", fmt.Errorf("failed to extract public key from DPoP proof: %w", err)
	}

	// Parse and validate the JWT with the public key
	token, err := jwt.ParseWithClaims(dpopProof, &DPoPProofClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify the algorithm matches
		if token.Method.Alg() != header.Alg {
			return nil, fmt.Errorf("algorithm mismatch")
		}
		return publicKey, nil
	})

	if err != nil {
		return "", fmt.Errorf("failed to validate DPoP proof signature: %w", err)
	}

	if !token.Valid {
		return "", fmt.Errorf("invalid DPoP proof token")
	}

	claims, ok := token.Claims.(*DPoPProofClaims)
	if !ok {
		return "", fmt.Errorf("failed to parse DPoP proof claims")
	}

	// Validate required claims
	if claims.Jti == "" {
		return "", fmt.Errorf("DPoP proof missing jti claim")
	}
	if claims.Htm == "" {
		return "", fmt.Errorf("DPoP proof missing htm claim")
	}
	if claims.Htu == "" {
		return "", fmt.Errorf("DPoP proof missing htu claim")
	}

	// Validate htm matches the HTTP method
	if !strings.EqualFold(claims.Htm, httpMethod) {
		return "", fmt.Errorf("DPoP proof htm claim mismatch, expected '%s', got '%s'", httpMethod, claims.Htm)
	}

	// Validate htu matches the HTTP URI (without query and fragment)
	if !validateHtu(claims.Htu, httpUri) {
		return "", fmt.Errorf("DPoP proof htu claim mismatch")
	}

	// Validate iat is recent (within configured tolerance window)
	now := time.Now().Unix()
	if claims.Iat == 0 {
		return "", fmt.Errorf("DPoP proof missing iat claim")
	}
	if claims.Iat > now+dpopIatToleranceSeconds || claims.Iat < now-dpopIatToleranceSeconds {
		return "", fmt.Errorf("DPoP proof iat claim outside acceptable time window")
	}

	// Validate jti hasn't been used before (replay protection)
	if err := validateJti(claims.Jti); err != nil {
		return "", err
	}

	// If access token is provided, validate ath claim
	if accessToken != "" {
		expectedAth := computeAth(accessToken)
		if claims.Ath != expectedAth {
			return "", fmt.Errorf("DPoP proof ath claim mismatch")
		}
	}

	// Compute and return the JWK thumbprint (jkt)
	jkt, err := computeJwkThumbprint(header.Jwk)
	if err != nil {
		return "", fmt.Errorf("failed to compute JWK thumbprint: %w", err)
	}

	return jkt, nil
}

// extractPublicKeyFromJwk extracts the public key from a JWK using go-jose
func extractPublicKeyFromJwk(jwk map[string]interface{}, alg string) (interface{}, error) {
	jwkBytes, err := json.Marshal(jwk)
	if err != nil {
		return nil, err
	}

	// Parse using go-jose library
	var joseJwk jose.JSONWebKey
	if err := json.Unmarshal(jwkBytes, &joseJwk); err != nil {
		return nil, fmt.Errorf("failed to parse JWK: %w", err)
	}

	// Validate the key is public
	if !joseJwk.IsPublic() {
		return nil, fmt.Errorf("JWK must be a public key")
	}

	// Return the public key
	return joseJwk.Key, nil
}

// computeJwkThumbprint computes the JWK thumbprint according to RFC 7638
func computeJwkThumbprint(jwk map[string]interface{}) (string, error) {
	kty, ok := jwk["kty"].(string)
	if !ok {
		return "", fmt.Errorf("missing kty in JWK")
	}

	var requiredFields map[string]interface{}

	switch kty {
	case "RSA":
		requiredFields = map[string]interface{}{
			"e":   jwk["e"],
			"kty": jwk["kty"],
			"n":   jwk["n"],
		}
	case "EC":
		requiredFields = map[string]interface{}{
			"crv": jwk["crv"],
			"kty": jwk["kty"],
			"x":   jwk["x"],
			"y":   jwk["y"],
		}
	case "OKP":
		requiredFields = map[string]interface{}{
			"crv": jwk["crv"],
			"kty": jwk["kty"],
			"x":   jwk["x"],
		}
	default:
		return "", fmt.Errorf("unsupported key type: %s", kty)
	}

	// Marshal with sorted keys (lexicographic order)
	thumbprintInput, err := json.Marshal(requiredFields)
	if err != nil {
		return "", err
	}

	// Compute SHA-256 hash
	hash := sha256.Sum256(thumbprintInput)

	// Base64url encode
	return base64.RawURLEncoding.EncodeToString(hash[:]), nil
}

// validateHtu validates the htu claim matches the request URI
func validateHtu(htu string, requestUri string) bool {
	// Remove query and fragment from both
	htuClean := strings.Split(htu, "?")[0]
	htuClean = strings.Split(htuClean, "#")[0]

	requestUriClean := strings.Split(requestUri, "?")[0]
	requestUriClean = strings.Split(requestUriClean, "#")[0]

	return htuClean == requestUriClean
}

// validateJti validates the jti hasn't been used before
func validateJti(jti string) error {
	// Start cleanup task on first use
	startCleanupTask()

	now := time.Now()

	// Check if jti exists
	if _, exists := dpopNonceStore.Load(jti); exists {
		return fmt.Errorf("DPoP proof jti has already been used (replay attack)")
	}

	// Store the jti with expiration
	dpopNonceStore.Store(jti, now.Add(dpopJtiExpiration))

	return nil
}

// cleanupExpiredJtis removes expired jtis from the store
func cleanupExpiredJtis() {
	now := time.Now()
	dpopNonceStore.Range(func(key, value interface{}) bool {
		if expiresAt, ok := value.(time.Time); ok {
			if now.After(expiresAt) {
				dpopNonceStore.Delete(key)
			}
		}
		return true
	})
}

// computeAth computes the ath claim (access token hash)
func computeAth(accessToken string) string {
	hash := sha256.Sum256([]byte(accessToken))
	return base64.RawURLEncoding.EncodeToString(hash[:])
}

// GetDPoPSigningAlgValuesSupported returns the list of supported DPoP signing algorithms
func GetDPoPSigningAlgValuesSupported() []string {
	return []string{"RS256", "RS512", "ES256", "ES384", "ES512", "PS256", "PS384", "PS512"}
}
