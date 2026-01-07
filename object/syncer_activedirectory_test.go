// Copyright 2025 The Casdoor Authors. All Rights Reserved.
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
	"unicode/utf8"
)

func TestConvertGUIDToString(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected string
	}{
		{
			name:     "Valid 16-byte GUID",
			input:    []byte{0xe7, 0xe8, 0x17, 0x4a, 0x1b, 0x2c, 0x3d, 0x4e, 0x5f, 0x60, 0x71, 0x82, 0x93, 0xa4, 0xb5, 0xc6},
			expected: "4a17e8e7-2c1b-4e3d-5f60-718293a4b5c6",
		},
		{
			name:     "All zeros GUID",
			input:    []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			expected: "00000000-0000-0000-0000-000000000000",
		},
		{
			name:     "All 0xFF GUID",
			input:    []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			expected: "ffffffff-ffff-ffff-ffff-ffffffffffff",
		},
		{
			name:     "Invalid length - too short",
			input:    []byte{0xe7, 0xe8, 0x17, 0x4a},
			expected: "",
		},
		{
			name:     "Invalid length - too long",
			input:    []byte{0xe7, 0xe8, 0x17, 0x4a, 0x1b, 0x2c, 0x3d, 0x4e, 0x5f, 0x60, 0x71, 0x82, 0x93, 0xa4, 0xb5, 0xc6, 0xd7},
			expected: "",
		},
		{
			name:     "Empty array",
			input:    []byte{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertGUIDToString(tt.input)
			if result != tt.expected {
				t.Errorf("convertGUIDToString() = %v, want %v", result, tt.expected)
			}
			// Ensure result is valid UTF-8
			if result != "" && !utf8.ValidString(result) {
				t.Errorf("convertGUIDToString() returned invalid UTF-8: %v", result)
			}
		})
	}
}

func TestSanitizeUTF8(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		isValid  bool
	}{
		{
			name:     "Valid UTF-8 string",
			input:    "Hello, World!",
			expected: "Hello, World!",
			isValid:  true,
		},
		{
			name:     "UTF-8 with special characters",
			input:    "你好世界 🌍",
			expected: "你好世界 🌍",
			isValid:  true,
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
			isValid:  true,
		},
		{
			name:     "String with invalid UTF-8 bytes",
			input:    "Hello\xc0\x80World",
			expected: "HelloWorld",
			isValid:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeUTF8(tt.input)
			if !utf8.ValidString(result) {
				t.Errorf("sanitizeUTF8() returned invalid UTF-8: %v", result)
			}
			if tt.isValid && result != tt.expected {
				t.Errorf("sanitizeUTF8() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestConvertGUIDToStringIsUTF8Safe(t *testing.T) {
	// Test with the exact problematic bytes mentioned in the issue
	problematicBytes := []byte{0xe7, 0xe8, 0x17, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	
	result := convertGUIDToString(problematicBytes)
	
	// Check that result is not empty
	if result == "" {
		t.Error("convertGUIDToString() returned empty string for valid 16-byte input")
	}
	
	// Check that result is valid UTF-8
	if !utf8.ValidString(result) {
		t.Errorf("convertGUIDToString() returned invalid UTF-8: %v (bytes: %v)", result, []byte(result))
	}
	
	// Check that result has the expected UUID format length
	if len(result) != 36 {
		t.Errorf("convertGUIDToString() returned UUID with unexpected length: %d, want 36", len(result))
	}
	
	// Check UUID format with dashes in correct positions
	if result[8] != '-' || result[13] != '-' || result[18] != '-' || result[23] != '-' {
		t.Errorf("convertGUIDToString() returned UUID with incorrect format: %v", result)
	}
}
