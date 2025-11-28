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
	"regexp"
	"testing"
	"time"
)

func TestSAMLTimeFormat(t *testing.T) {
	// Test that SAMLTimeFormat produces the correct format
	now := time.Now().UTC().Format(SAMLTimeFormat)

	// The format should be YYYY-MM-DDTHH:MM:SSZ
	// Example: 2025-11-28T16:14:40Z
	pattern := `^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$`
	matched, err := regexp.MatchString(pattern, now)
	if err != nil {
		t.Fatalf("Failed to compile regex: %v", err)
	}
	if !matched {
		t.Errorf("SAMLTimeFormat produced invalid format: %s, expected pattern: %s", now, pattern)
	}

	// Verify the format starts with a date (not just T)
	if now[0] == 'T' {
		t.Errorf("SAMLTimeFormat should not start with 'T', got: %s", now)
	}

	// Verify the format includes the date portion
	if len(now) < 20 {
		t.Errorf("SAMLTimeFormat should produce at least 20 characters (YYYY-MM-DDTHH:MM:SSZ), got %d: %s", len(now), now)
	}
}

func TestSAMLTimeFormatConstant(t *testing.T) {
	// Verify the constant is correct
	expectedFormat := "2006-01-02T15:04:05Z"
	if SAMLTimeFormat != expectedFormat {
		t.Errorf("SAMLTimeFormat constant is incorrect. Expected: %s, Got: %s", expectedFormat, SAMLTimeFormat)
	}
}
