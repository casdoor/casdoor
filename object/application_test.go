// Copyright 2023 The Casdoor Authors. All Rights Reserved.
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

func TestApplication_IsRedirectUriValid(t *testing.T) {
	tests := []struct {
		name         string
		redirectUris []string
		testUri      string
		expected     bool
	}{
		{
			name:         "Empty redirect URIs with empty string should not match any URI",
			redirectUris: []string{""},
			testUri:      "https://malicious.com/callback",
			expected:     false,
		},
		{
			name:         "Empty redirect URIs array with empty string should still allow localhost (valid origin)",
			redirectUris: []string{""},
			testUri:      "http://localhost:8080/callback",
			expected:     true, // localhost is always allowed via IsValidOrigin
		},
		{
			name:         "Mixed valid and empty redirect URIs should match valid URI",
			redirectUris: []string{"", "https://example.com/callback"},
			testUri:      "https://example.com/callback",
			expected:     true,
		},
		{
			name:         "Mixed valid and empty redirect URIs should not match invalid URI",
			redirectUris: []string{"", "https://example.com/callback"},
			testUri:      "https://malicious.com/callback",
			expected:     false,
		},
		{
			name:         "Valid redirect URI with regex pattern",
			redirectUris: []string{"https://.*\\.example\\.com/callback"},
			testUri:      "https://sub.example.com/callback",
			expected:     true,
		},
		{
			name:         "Valid redirect URI should match exactly",
			redirectUris: []string{"https://example.com/callback"},
			testUri:      "https://example.com/callback",
			expected:     true,
		},
		{
			name:         "Valid redirect URI should not match different URI",
			redirectUris: []string{"https://example.com/callback"},
			testUri:      "https://other.com/callback",
			expected:     false,
		},
		{
			name:         "Multiple empty strings should not match any URI",
			redirectUris: []string{"", "", ""},
			testUri:      "https://any.com/callback",
			expected:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := &Application{
				RedirectUris: tt.redirectUris,
			}
			result := app.IsRedirectUriValid(tt.testUri)
			if result != tt.expected {
				t.Errorf("IsRedirectUriValid() = %v, want %v", result, tt.expected)
			}
		})
	}
}
