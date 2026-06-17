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
	"encoding/json"
	"testing"

	"github.com/golang-jwt/jwt/v5"
)

// TestLogoutTokenClaimsSidIsSeparateFromJti guards bug A: the session id must be
// emitted as the `sid` claim and must not overwrite the `jti` (RegisteredClaims.ID).
func TestLogoutTokenClaimsSidIsSeparateFromJti(t *testing.T) {
	claims := LogoutTokenClaims{
		Events: map[string]interface{}{
			"http://schemas.openid.net/event/backchannel-logout": map[string]interface{}{},
		},
		Sid: "session-123",
		RegisteredClaims: jwt.RegisteredClaims{
			ID: "jti-456",
		},
	}

	raw, err := json.Marshal(claims)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if got := decoded["sid"]; got != "session-123" {
		t.Errorf("sid = %v, want session-123", got)
	}
	if got := decoded["jti"]; got != "jti-456" {
		t.Errorf("jti = %v, want jti-456 (sid must not overwrite jti)", got)
	}
	if decoded["sid"] == decoded["jti"] {
		t.Errorf("sid and jti must be distinct, both = %v", decoded["sid"])
	}
}

// TestLogoutTokenSidOmittedWhenEmpty ensures we don't emit an empty sid claim.
func TestLogoutTokenSidOmittedWhenEmpty(t *testing.T) {
	claims := LogoutTokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{ID: "jti-only"},
	}

	raw, err := json.Marshal(claims)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if _, ok := decoded["sid"]; ok {
		t.Errorf("sid should be omitted when empty, got %v", decoded["sid"])
	}
}

func TestGetClientIdFromClaims(t *testing.T) {
	tests := []struct {
		name     string
		audience jwt.ClaimStrings
		want     string
	}{
		{"single", jwt.ClaimStrings{"client-abc"}, "client-abc"},
		{"first non-empty", jwt.ClaimStrings{"", "client-xyz"}, "client-xyz"},
		{"empty", jwt.ClaimStrings{}, ""},
		{"all empty", jwt.ClaimStrings{"", ""}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims := &Claims{}
			claims.Audience = tt.audience
			if got := getClientIdFromClaims(claims); got != tt.want {
				t.Errorf("getClientIdFromClaims() = %q, want %q", got, tt.want)
			}
		})
	}
}
