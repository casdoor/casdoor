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
)

func TestGenerateLogoutSignature(t *testing.T) {
	// Test that the signature generation is deterministic
	clientSecret := "test-secret-key"
	owner := "test-org"
	name := "test-user"
	nonce := "test-nonce-123"
	timestamp := int64(1699900000)
	sessionIds := []string{"session-1", "session-2"}
	accessTokenHashes := []string{"hash-1", "hash-2"}

	sig1 := generateLogoutSignature(clientSecret, owner, name, nonce, timestamp, sessionIds, accessTokenHashes)
	sig2 := generateLogoutSignature(clientSecret, owner, name, nonce, timestamp, sessionIds, accessTokenHashes)

	if sig1 != sig2 {
		t.Errorf("Signature should be deterministic, got %s and %s", sig1, sig2)
	}

	// Test that different inputs produce different signatures
	sig3 := generateLogoutSignature(clientSecret, owner, "different-user", nonce, timestamp, sessionIds, accessTokenHashes)
	if sig1 == sig3 {
		t.Error("Different inputs should produce different signatures")
	}

	// Test with different client secret
	sig4 := generateLogoutSignature("different-secret", owner, name, nonce, timestamp, sessionIds, accessTokenHashes)
	if sig1 == sig4 {
		t.Error("Different client secrets should produce different signatures")
	}
}

func TestVerifySsoLogoutSignature(t *testing.T) {
	clientSecret := "test-secret-key"
	owner := "test-org"
	name := "test-user"
	nonce := "test-nonce-123"
	timestamp := int64(1699900000)
	sessionIds := []string{"session-1", "session-2"}
	accessTokenHashes := []string{"hash-1", "hash-2"}

	// Generate a valid signature
	signature := generateLogoutSignature(clientSecret, owner, name, nonce, timestamp, sessionIds, accessTokenHashes)

	// Create a notification with the valid signature
	notification := &SsoLogoutNotification{
		Owner:             owner,
		Name:              name,
		Nonce:             nonce,
		Timestamp:         timestamp,
		SessionIds:        sessionIds,
		AccessTokenHashes: accessTokenHashes,
		Signature:         signature,
	}

	// Verify with correct secret
	if !VerifySsoLogoutSignature(clientSecret, notification) {
		t.Error("Valid signature should be verified successfully")
	}

	// Verify with wrong secret
	if VerifySsoLogoutSignature("wrong-secret", notification) {
		t.Error("Invalid signature should not be verified")
	}

	// Verify with tampered data
	tamperedNotification := &SsoLogoutNotification{
		Owner:             owner,
		Name:              "tampered-user", // Changed
		Nonce:             nonce,
		Timestamp:         timestamp,
		SessionIds:        sessionIds,
		AccessTokenHashes: accessTokenHashes,
		Signature:         signature, // Same signature
	}
	if VerifySsoLogoutSignature(clientSecret, tamperedNotification) {
		t.Error("Tampered notification should not be verified")
	}
}

func TestSsoLogoutNotificationStructure(t *testing.T) {
	sessionTokenMap := map[string][]string{
		"session-1": {"hash-1"},
		"session-2": {"hash-2"},
	}

	notification := SsoLogoutNotification{
		Owner:             "test-org",
		Name:              "test-user",
		DisplayName:       "Test User",
		Email:             "test@example.com",
		Phone:             "+1234567890",
		Id:                "user-123",
		Event:             "sso-logout",
		SessionIds:        []string{"session-1", "session-2"},
		AccessTokenHashes: []string{"hash-1", "hash-2"},
		SessionTokenMap:   sessionTokenMap,
		Nonce:             "nonce-123",
		Timestamp:         1699900000,
		Signature:         "sig-123",
	}

	// Verify all fields are set correctly
	if notification.Owner != "test-org" {
		t.Errorf("Owner mismatch, got %s", notification.Owner)
	}
	if notification.Name != "test-user" {
		t.Errorf("Name mismatch, got %s", notification.Name)
	}
	if notification.Event != "sso-logout" {
		t.Errorf("Event mismatch, got %s", notification.Event)
	}
	if len(notification.SessionIds) != 2 {
		t.Errorf("SessionIds count mismatch, got %d", len(notification.SessionIds))
	}
	if len(notification.AccessTokenHashes) != 2 {
		t.Errorf("AccessTokenHashes count mismatch, got %d", len(notification.AccessTokenHashes))
	}
	if len(notification.SessionTokenMap) != 2 {
		t.Errorf("SessionTokenMap count mismatch, got %d", len(notification.SessionTokenMap))
	}
	if len(notification.SessionTokenMap["session-1"]) != 1 {
		t.Errorf("SessionTokenMap[session-1] should have 1 token, got %d", len(notification.SessionTokenMap["session-1"]))
	}
	if notification.SessionTokenMap["session-1"][0] != "hash-1" {
		t.Errorf("SessionTokenMap[session-1][0] should be hash-1, got %s", notification.SessionTokenMap["session-1"][0])
	}
}

func TestGenerateLogoutSignatureWithEmptyArrays(t *testing.T) {
	clientSecret := "test-secret-key"
	owner := "test-org"
	name := "test-user"
	nonce := "test-nonce-123"
	timestamp := int64(1699900000)

	// Test with empty session IDs and token hashes
	sig1 := generateLogoutSignature(clientSecret, owner, name, nonce, timestamp, []string{}, []string{})
	sig2 := generateLogoutSignature(clientSecret, owner, name, nonce, timestamp, nil, nil)

	// Empty slice and nil should produce the same signature
	if sig1 != sig2 {
		t.Errorf("Empty slice and nil should produce the same signature, got %s and %s", sig1, sig2)
	}

	// Should be different from non-empty arrays
	sig3 := generateLogoutSignature(clientSecret, owner, name, nonce, timestamp, []string{"session-1"}, []string{"hash-1"})
	if sig1 == sig3 {
		t.Error("Empty arrays should produce different signature from non-empty arrays")
	}
}
