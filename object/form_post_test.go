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
	"strings"
	"testing"
)

func TestGenerateFormPostResponse(t *testing.T) {
	tests := []struct {
		name          string
		redirectUri   string
		parameters    map[string]string
		expectError   bool
		expectContain []string
	}{
		{
			name:        "Valid form_post response",
			redirectUri: "https://example.com/callback",
			parameters: map[string]string{
				"code":  "test_code_123",
				"state": "test_state_456",
			},
			expectError: false,
			expectContain: []string{
				"https://example.com/callback",
				"test_code_123",
				"test_state_456",
				"document.getElementById('form').submit()",
				"method=\"post\"",
			},
		},
		{
			name:        "Empty redirect URI",
			redirectUri: "",
			parameters: map[string]string{
				"code": "test_code_123",
			},
			expectError: true,
		},
		{
			name:        "Invalid redirect URI",
			redirectUri: "not-a-valid-uri",
			parameters: map[string]string{
				"code": "test_code_123",
			},
			expectError: true,
		},
		{
			name:        "XSS prevention test",
			redirectUri: "https://example.com/callback",
			parameters: map[string]string{
				"code":  "<script>alert('xss')</script>",
				"state": "normal_state",
			},
			expectError: false,
			expectContain: []string{
				"&amp;lt;script&amp;gt;alert(&amp;#39;xss&amp;#39;)&amp;lt;/script&amp;gt;",
				"normal_state",
			},
		},
		{
			name:        "Only code parameter",
			redirectUri: "https://example.com/callback",
			parameters: map[string]string{
				"code": "test_code_only",
			},
			expectError: false,
			expectContain: []string{
				"test_code_only",
				"https://example.com/callback",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GenerateFormPostResponse(tt.redirectUri, tt.parameters)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			for _, expected := range tt.expectContain {
				if !strings.Contains(result, expected) {
					t.Errorf("Expected result to contain '%s', but it didn't. Result: %s", expected, result)
				}
			}

			// Check that result is valid HTML
			if !strings.Contains(result, "<html>") || !strings.Contains(result, "</html>") {
				t.Errorf("Result should be valid HTML")
			}

			// Check security headers are mentioned in comments/docs
			if !strings.Contains(result, "no-cache") {
				t.Errorf("Result should include cache control measures")
			}
		})
	}
}

func TestIsValidResponseMode(t *testing.T) {
	tests := []struct {
		mode   string
		expect bool
	}{
		{"query", true},
		{"fragment", true},
		{"form_post", true},
		{"invalid", false},
		{"", false},
		{"QUERY", false}, // case sensitive
	}

	for _, tt := range tests {
		t.Run(tt.mode, func(t *testing.T) {
			result := IsValidResponseMode(tt.mode)
			if result != tt.expect {
				t.Errorf("isValidResponseMode(%q) = %v, expected %v", tt.mode, result, tt.expect)
			}
		})
	}
}
