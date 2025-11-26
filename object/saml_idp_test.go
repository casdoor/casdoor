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
	// SAML 2.0 xs:dateTime pattern: YYYY-MM-DDTHH:MM:SSZ
	samlTimePattern := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$`)

	testCases := []struct {
		name     string
		time     time.Time
		expected string
	}{
		{
			name:     "regular time",
			time:     time.Date(2025, 10, 17, 9, 24, 35, 0, time.UTC),
			expected: "2025-10-17T09:24:35Z",
		},
		{
			name:     "midnight",
			time:     time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			expected: "2025-01-01T00:00:00Z",
		},
		{
			name:     "end of year",
			time:     time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC),
			expected: "2025-12-31T23:59:59Z",
		},
		{
			name:     "single digit month and day",
			time:     time.Date(2025, 1, 5, 8, 5, 3, 0, time.UTC),
			expected: "2025-01-05T08:05:03Z",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			formatted := tc.time.Format(SAMLTimeFormat)
			if formatted != tc.expected {
				t.Errorf("SAMLTimeFormat produced %q, expected %q", formatted, tc.expected)
			}
			if !samlTimePattern.MatchString(formatted) {
				t.Errorf("SAMLTimeFormat produced invalid xs:dateTime format: %q", formatted)
			}
		})
	}
}

func TestSAMLTimeFormatCompliance(t *testing.T) {
	// Test that SAMLTimeFormat produces timestamps that:
	// 1. Include the full date (YYYY-MM-DD)
	// 2. Include the time (HH:MM:SS)
	// 3. End with Z for UTC timezone

	now := time.Now().UTC()
	formatted := now.Format(SAMLTimeFormat)

	// Must start with date in YYYY-MM-DD format
	if len(formatted) < 10 {
		t.Fatalf("Formatted time %q is too short, must include full date", formatted)
	}

	datePrefix := formatted[:10]
	if datePrefix[4] != '-' || datePrefix[7] != '-' {
		t.Errorf("Date portion %q does not follow YYYY-MM-DD format", datePrefix)
	}

	// Must have 'T' separator
	if formatted[10] != 'T' {
		t.Errorf("Missing 'T' separator in timestamp %q", formatted)
	}

	// Must end with 'Z' for UTC
	if formatted[len(formatted)-1] != 'Z' {
		t.Errorf("Timestamp %q must end with 'Z' for UTC timezone", formatted)
	}

	// Total length should be exactly 20 characters: YYYY-MM-DDTHH:MM:SSZ
	if len(formatted) != 20 {
		t.Errorf("Timestamp %q has incorrect length %d, expected 20", formatted, len(formatted))
	}
}
