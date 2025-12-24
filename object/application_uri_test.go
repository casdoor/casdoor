// Copyright 2021 The Casdoor Authors. All Rights Reserved.
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

func TestApplication_IsUriProtected(t *testing.T) {
	tests := []struct {
		name          string
		protectedUris []string
		publicUris    []string
		testUri       string
		expected      bool
	}{
		{
			name:          "No configuration - all URIs protected by default",
			protectedUris: []string{},
			publicUris:    []string{},
			testUri:       "https://app.example.com/api",
			expected:      true,
		},
		{
			name:          "ProtectedUris configured - matching URI is protected",
			protectedUris: []string{"https://app\\.example\\.com/api$"},
			publicUris:    []string{},
			testUri:       "https://app.example.com/api",
			expected:      true,
		},
		{
			name:          "ProtectedUris configured - non-matching URI is not protected",
			protectedUris: []string{"https://app\\.example\\.com/api$"},
			publicUris:    []string{},
			testUri:       "https://app.example.com/api-another",
			expected:      false,
		},
		{
			name:          "PublicUris configured - matching URI is not protected",
			protectedUris: []string{},
			publicUris:    []string{"https://app.example.com/api-another"},
			testUri:       "https://app.example.com/api-another",
			expected:      false,
		},
		{
			name:          "PublicUris configured - non-matching URI is protected",
			protectedUris: []string{},
			publicUris:    []string{"https://app.example.com/api-another"},
			testUri:       "https://app.example.com/api",
			expected:      true,
		},
		{
			name:          "Both configured - public takes precedence",
			protectedUris: []string{"https://app.example.com/.*"},
			publicUris:    []string{"https://app.example.com/api-another"},
			testUri:       "https://app.example.com/api-another",
			expected:      false,
		},
		{
			name:          "Regex pattern in ProtectedUris",
			protectedUris: []string{"https://app.example.com/api.*"},
			publicUris:    []string{},
			testUri:       "https://app.example.com/api/users",
			expected:      true,
		},
		{
			name:          "Regex pattern not matching",
			protectedUris: []string{"https://app.example.com/api$"},
			publicUris:    []string{},
			testUri:       "https://app.example.com/api/users",
			expected:      false,
		},
		{
			name:          "Regex pattern in PublicUris",
			protectedUris: []string{},
			publicUris:    []string{".*/public/.*"},
			testUri:       "https://app.example.com/public/health",
			expected:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := &Application{
				ProtectedUris: tt.protectedUris,
				PublicUris:    tt.publicUris,
			}
			result := app.IsUriProtected(tt.testUri)
			if result != tt.expected {
				t.Errorf("IsUriProtected() = %v, expected %v for URI %s with ProtectedUris=%v, PublicUris=%v",
					result, tt.expected, tt.testUri, tt.protectedUris, tt.publicUris)
			}
		})
	}
}
