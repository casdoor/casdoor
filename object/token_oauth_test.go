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

func TestHasOpenIDScope(t *testing.T) {
	tests := []struct {
		name     string
		scope    string
		expected bool
	}{
		{
			name:     "empty scope",
			scope:    "",
			expected: false,
		},
		{
			name:     "openid only",
			scope:    "openid",
			expected: true,
		},
		{
			name:     "openid with other scopes",
			scope:    "openid profile email",
			expected: true,
		},
		{
			name:     "no openid",
			scope:    "profile email",
			expected: false,
		},
		{
			name:     "openid at end",
			scope:    "profile email openid",
			expected: true,
		},
		{
			name:     "openid in middle",
			scope:    "profile openid email",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hasOpenIDScope(tt.scope)
			if result != tt.expected {
				t.Errorf("hasOpenIDScope(%q) = %v, want %v", tt.scope, result, tt.expected)
			}
		})
	}
}
