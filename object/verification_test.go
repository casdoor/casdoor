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

func TestGenerateSecureToken(t *testing.T) {
	// Test that token generation works
	token, err := generateSecureToken(32)
	if err != nil {
		t.Errorf("generateSecureToken failed: %v", err)
	}

	if len(token) == 0 {
		t.Errorf("generateSecureToken returned empty token")
	}

	// Test that multiple calls generate different tokens
	token2, err := generateSecureToken(32)
	if err != nil {
		t.Errorf("generateSecureToken failed on second call: %v", err)
	}

	if token == token2 {
		t.Errorf("generateSecureToken returned same token twice (very unlikely if random)")
	}

	// Test different lengths
	shortToken, err := generateSecureToken(16)
	if err != nil {
		t.Errorf("generateSecureToken failed with length 16: %v", err)
	}

	if len(shortToken) == 0 {
		t.Errorf("generateSecureToken returned empty token for length 16")
	}
}

func TestGetRandomCode(t *testing.T) {
	// Test that code generation works
	code := getRandomCode(6)
	if len(code) != 6 {
		t.Errorf("getRandomCode returned code of length %d, expected 6", len(code))
	}

	// Test that code only contains digits
	for _, c := range code {
		if c < '0' || c > '9' {
			t.Errorf("getRandomCode returned non-digit character: %c", c)
		}
	}

	// Test different lengths
	code10 := getRandomCode(10)
	if len(code10) != 10 {
		t.Errorf("getRandomCode returned code of length %d, expected 10", len(code10))
	}
}
