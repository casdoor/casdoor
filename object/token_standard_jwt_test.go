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
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TestParseStandardJwtTokenWithRSA tests ParseStandardJwtToken with RSA signing method
func TestParseStandardJwtTokenWithRSA(t *testing.T) {
	// Generate RSA keys
	certificate, privateKey, err := generateRsaKeys(2048, 256, 1, "Test Cert", "Test Org")
	if err != nil {
		t.Fatalf("Failed to generate RSA keys: %v", err)
	}

	cert := &Cert{
		Owner:       "test-owner",
		Name:        "test-cert",
		Certificate: certificate,
		PrivateKey:  privateKey,
	}

	// Create a standard JWT token with RSA
	claims := &ClaimsStandard{
		UserStandard: &UserStandard{
			Owner: "test-owner",
			Name:  "test-user",
		},
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	
	// Parse private key for signing
	key, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(privateKey))
	if err != nil {
		t.Fatalf("Failed to parse private key: %v", err)
	}

	// Sign the token
	tokenString, err := token.SignedString(key)
	if err != nil {
		t.Fatalf("Failed to sign token: %v", err)
	}

	// Parse the token
	parsedClaims, err := ParseStandardJwtToken(tokenString, cert)
	if err != nil {
		t.Fatalf("Failed to parse standard JWT token with RSA: %v", err)
	}

	if parsedClaims.UserStandard.Name != "test-user" {
		t.Errorf("Expected user name 'test-user', got '%s'", parsedClaims.UserStandard.Name)
	}
}

// TestParseStandardJwtTokenWithES256 tests ParseStandardJwtToken with ES256 signing method
func TestParseStandardJwtTokenWithES256(t *testing.T) {
	// Generate ES256 keys
	certificate, privateKey, err := generateEsKeys(256, 1, "Test Cert", "Test Org")
	if err != nil {
		t.Fatalf("Failed to generate ES256 keys: %v", err)
	}

	cert := &Cert{
		Owner:       "test-owner",
		Name:        "test-cert",
		Certificate: certificate,
		PrivateKey:  privateKey,
	}

	// Create a standard JWT token with ES256
	claims := &ClaimsStandard{
		UserStandard: &UserStandard{
			Owner: "test-owner",
			Name:  "test-user-es256",
		},
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	
	// Parse private key for signing
	key, err := jwt.ParseECPrivateKeyFromPEM([]byte(privateKey))
	if err != nil {
		t.Fatalf("Failed to parse EC private key: %v", err)
	}

	// Sign the token
	tokenString, err := token.SignedString(key)
	if err != nil {
		t.Fatalf("Failed to sign token: %v", err)
	}

	// Parse the token
	parsedClaims, err := ParseStandardJwtToken(tokenString, cert)
	if err != nil {
		t.Fatalf("Failed to parse standard JWT token with ES256: %v", err)
	}

	if parsedClaims.UserStandard.Name != "test-user-es256" {
		t.Errorf("Expected user name 'test-user-es256', got '%s'", parsedClaims.UserStandard.Name)
	}
}

// TestParseStandardJwtTokenWithES384 tests ParseStandardJwtToken with ES384 signing method
func TestParseStandardJwtTokenWithES384(t *testing.T) {
	// Generate ES384 keys
	certificate, privateKey, err := generateEsKeys(384, 1, "Test Cert", "Test Org")
	if err != nil {
		t.Fatalf("Failed to generate ES384 keys: %v", err)
	}

	cert := &Cert{
		Owner:       "test-owner",
		Name:        "test-cert",
		Certificate: certificate,
		PrivateKey:  privateKey,
	}

	// Create a standard JWT token with ES384
	claims := &ClaimsStandard{
		UserStandard: &UserStandard{
			Owner: "test-owner",
			Name:  "test-user-es384",
		},
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES384, claims)
	
	// Parse private key for signing
	key, err := jwt.ParseECPrivateKeyFromPEM([]byte(privateKey))
	if err != nil {
		t.Fatalf("Failed to parse EC private key: %v", err)
	}

	// Sign the token
	tokenString, err := token.SignedString(key)
	if err != nil {
		t.Fatalf("Failed to sign token: %v", err)
	}

	// Parse the token
	parsedClaims, err := ParseStandardJwtToken(tokenString, cert)
	if err != nil {
		t.Fatalf("Failed to parse standard JWT token with ES384: %v", err)
	}

	if parsedClaims.UserStandard.Name != "test-user-es384" {
		t.Errorf("Expected user name 'test-user-es384', got '%s'", parsedClaims.UserStandard.Name)
	}
}

// TestParseStandardJwtTokenWithES512 tests ParseStandardJwtToken with ES512 signing method
func TestParseStandardJwtTokenWithES512(t *testing.T) {
	// Generate ES512 keys
	certificate, privateKey, err := generateEsKeys(512, 1, "Test Cert", "Test Org")
	if err != nil {
		t.Fatalf("Failed to generate ES512 keys: %v", err)
	}

	cert := &Cert{
		Owner:       "test-owner",
		Name:        "test-cert",
		Certificate: certificate,
		PrivateKey:  privateKey,
	}

	// Create a standard JWT token with ES512
	claims := &ClaimsStandard{
		UserStandard: &UserStandard{
			Owner: "test-owner",
			Name:  "test-user-es512",
		},
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES512, claims)
	
	// Parse private key for signing
	key, err := jwt.ParseECPrivateKeyFromPEM([]byte(privateKey))
	if err != nil {
		t.Fatalf("Failed to parse EC private key: %v", err)
	}

	// Sign the token
	tokenString, err := token.SignedString(key)
	if err != nil {
		t.Fatalf("Failed to sign token: %v", err)
	}

	// Parse the token
	parsedClaims, err := ParseStandardJwtToken(tokenString, cert)
	if err != nil {
		t.Fatalf("Failed to parse standard JWT token with ES512: %v", err)
	}

	if parsedClaims.UserStandard.Name != "test-user-es512" {
		t.Errorf("Expected user name 'test-user-es512', got '%s'", parsedClaims.UserStandard.Name)
	}
}
