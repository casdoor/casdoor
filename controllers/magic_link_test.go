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

package controllers

import (
	"testing"

	"github.com/casdoor/casdoor/object"
)

func TestGetMagicLinkAuthAction(t *testing.T) {
	if got := getMagicLinkAuthAction(true); got != object.MagicLinkAuthActionSignupNewUser {
		t.Errorf("getMagicLinkAuthAction(true) = %q, want %q", got, object.MagicLinkAuthActionSignupNewUser)
	}
	if got := getMagicLinkAuthAction(false); got != object.MagicLinkAuthActionSigninExistingUser {
		t.Errorf("getMagicLinkAuthAction(false) = %q, want %q", got, object.MagicLinkAuthActionSigninExistingUser)
	}
}

func TestCheckMagicLinkOAuthPayload(t *testing.T) {
	expected := map[string]string{
		"clientId":              "client-id",
		"responseType":          "code",
		"redirectUri":           "https://example.com/callback",
		"scope":                 "openid",
		"state":                 "the-state",
		"nonce":                 "",
		"code_challenge_method": "S256",
		"code_challenge":        "the-challenge",
	}

	copyOf := func(source map[string]string) map[string]string {
		target := map[string]string{}
		for key, value := range source {
			target[key] = value
		}
		return target
	}

	if err := checkMagicLinkOAuthPayload(expected, copyOf(expected)); err != nil {
		t.Errorf("checkMagicLinkOAuthPayload() error = %v, want nil for an identical request", err)
	}

	// a link issued for one redirect URI must not finish another one
	tampered := copyOf(expected)
	tampered["redirectUri"] = "https://attacker.example.com/callback"
	if err := checkMagicLinkOAuthPayload(expected, tampered); err == nil {
		t.Errorf("checkMagicLinkOAuthPayload() error = nil, want an error for a changed redirectUri")
	}

	// a parameter that is dropped entirely is a change too
	missing := copyOf(expected)
	delete(missing, "state")
	if err := checkMagicLinkOAuthPayload(expected, missing); err == nil {
		t.Errorf("checkMagicLinkOAuthPayload() error = nil, want an error for a missing state")
	}

	// a parameter the link does not carry is expected to be absent
	blank := copyOf(expected)
	blank["nonce"] = "injected"
	if err := checkMagicLinkOAuthPayload(expected, blank); err == nil {
		t.Errorf("checkMagicLinkOAuthPayload() error = nil, want an error for an added nonce")
	}
}
