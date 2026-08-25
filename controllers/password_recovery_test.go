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

func TestIsMaskedPasswordRecovery(t *testing.T) {
	t.Setenv("enableErrorMask", "true")

	if !isMaskedPasswordRecovery(ForgetVerification) {
		t.Fatal("forgot-password recovery should be masked when enableErrorMask is true")
	}
	if isMaskedPasswordRecovery(LoginVerification) {
		t.Fatal("login verification must not use forgot-password masking")
	}
}

func TestMaskedPasswordRecoveryPublicUser(t *testing.T) {
	first := getMaskedPasswordRecoveryPublicUser("known-user")
	second := getMaskedPasswordRecoveryPublicUser("unknown-user")

	if first.Email != second.Email || first.Phone != second.Phone {
		t.Fatal("masked recovery responses must expose the same contact placeholders")
	}
	if first.Name != "known-user" || second.Name != "unknown-user" {
		t.Fatal("masked recovery responses should only echo the supplied identifier")
	}
}

func TestSelectMaskedPasswordRecoveryType(t *testing.T) {
	user := &object.User{Email: "user@example.com", Phone: "1234567"}

	tests := []struct {
		name          string
		preferredType string
		canEmail      bool
		canPhone      bool
		expected      string
	}{
		{name: "preferred email", preferredType: object.VerifyTypeEmail, canEmail: true, canPhone: true, expected: object.VerifyTypeEmail},
		{name: "preferred phone", preferredType: object.VerifyTypePhone, canEmail: true, canPhone: true, expected: object.VerifyTypePhone},
		{name: "fall back to email", preferredType: object.VerifyTypePhone, canEmail: true, canPhone: false, expected: object.VerifyTypeEmail},
		{name: "no provider", preferredType: object.VerifyTypeEmail, canEmail: false, canPhone: false, expected: ""},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := selectMaskedPasswordRecoveryType(user, test.preferredType, test.canEmail, test.canPhone)
			if actual != test.expected {
				t.Fatalf("expected %q, got %q", test.expected, actual)
			}
		})
	}
}

func TestShouldEnableMaskedPasswordRecoveryCaptcha(t *testing.T) {
	application := &object.Application{
		Providers: []*object.ProviderItem{{
			Rule:     "Dynamic",
			Provider: &object.Provider{Category: "Captcha"},
		}},
	}

	if !shouldEnableMaskedPasswordRecoveryCaptcha(application, "127.0.0.1") {
		t.Fatal("dynamic CAPTCHA must not depend on whether the recovery account exists")
	}
}
