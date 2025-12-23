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

package controllers

import (
	"testing"
)

func TestHashString(t *testing.T) {
	// Test that hash function generates consistent values
	testCases := []struct {
		input    string
		expected bool // whether it should be in valid range
	}{
		{"testuser", true},
		{"admin", true},
		{"user123", true},
		{"", true},
	}

	for _, tc := range testCases {
		result := hashString(tc.input)
		// Should be in range 1000-60000
		if result < 1000 || result >= 60000 {
			if tc.expected {
				t.Errorf("hashString(%q) = %d, want value between 1000 and 60000", tc.input, result)
			}
		}

		// Test consistency
		result2 := hashString(tc.input)
		if result != result2 {
			t.Errorf("hashString(%q) not consistent: %d != %d", tc.input, result, result2)
		}
	}
}

func TestFreeIPAUserResult(t *testing.T) {
	// Test that FreeIPAUserResult structure is properly formed
	result := FreeIPAUserResult{
		UID:          []string{"testuser"},
		UIDNumber:    []string{"1000"},
		GIDNumber:    []string{"1000"},
		CN:           []string{"Test User"},
		DisplayName:  []string{"Test User"},
		Mail:         []string{"test@example.com"},
		HomeDirectory: []string{"/home/testuser"},
		LoginShell:   []string{"/bin/bash"},
		MemberOf:     []string{"group1", "group2"},
	}

	if len(result.UID) != 1 || result.UID[0] != "testuser" {
		t.Errorf("Expected UID to be [testuser], got %v", result.UID)
	}

	if len(result.MemberOf) != 2 {
		t.Errorf("Expected 2 groups in MemberOf, got %d", len(result.MemberOf))
	}
}
