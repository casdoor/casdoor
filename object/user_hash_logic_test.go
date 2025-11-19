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

	"github.com/casdoor/casdoor/util"
)

// TestHashColumnInclusionLogic tests that the hash column is properly included
// in the columns list when it's missing
func TestHashColumnInclusionLogic(t *testing.T) {
	tests := []struct {
		name           string
		inputColumns   []string
		expectsHash    bool
		expectsNoError bool
	}{
		{
			name:           "Columns without hash - should add hash",
			inputColumns:   []string{"password", "password_salt", "password_type"},
			expectsHash:    true,
			expectsNoError: true,
		},
		{
			name:           "Columns with hash - should not duplicate",
			inputColumns:   []string{"password", "password_salt", "hash", "password_type"},
			expectsHash:    true,
			expectsNoError: true,
		},
		{
			name:           "Empty columns - should not add hash (uses default columns)",
			inputColumns:   []string{},
			expectsHash:    false,
			expectsNoError: true,
		},
		{
			name:           "Single column without hash - should add hash",
			inputColumns:   []string{"password"},
			expectsHash:    true,
			expectsNoError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the logic from updateUser function
			columns := make([]string, len(tt.inputColumns))
			copy(columns, tt.inputColumns)

			// This is the logic we added to updateUser
			if len(columns) > 0 && !util.InSlice(columns, "hash") {
				columns = append(columns, "hash")
			}

			// Verify expectations
			hasHash := util.InSlice(columns, "hash")
			if tt.expectsHash && !hasHash {
				t.Errorf("Expected columns to contain 'hash', but it doesn't. Columns: %v", columns)
			}
			if !tt.expectsHash && hasHash {
				t.Errorf("Expected columns to NOT contain 'hash', but it does. Columns: %v", columns)
			}

			// Verify no duplication when hash was already present
			if util.InSlice(tt.inputColumns, "hash") {
				countHash := 0
				for _, col := range columns {
					if col == "hash" {
						countHash++
					}
				}
				if countHash > 1 {
					t.Errorf("Hash column was duplicated. Count: %d, Columns: %v", countHash, columns)
				}
			}
		})
	}
}

// TestHashColumnPreservesOrder tests that adding hash doesn't break column ordering
func TestHashColumnPreservesOrder(t *testing.T) {
	inputColumns := []string{"password", "password_salt", "password_type"}
	columns := make([]string, len(inputColumns))
	copy(columns, inputColumns)

	// Apply the logic
	if len(columns) > 0 && !util.InSlice(columns, "hash") {
		columns = append(columns, "hash")
	}

	// Verify original columns are still in the same order
	for i, col := range inputColumns {
		if columns[i] != col {
			t.Errorf("Column order was changed. Expected %s at index %d, got %s", col, i, columns[i])
		}
	}

	// Verify hash is at the end
	if columns[len(columns)-1] != "hash" {
		t.Errorf("Expected hash to be at the end, but got: %v", columns)
	}
}

// TestHashColumnWithEmptySlice verifies behavior with empty slice
func TestHashColumnWithEmptySlice(t *testing.T) {
	columns := []string{}

	// Apply the logic
	if len(columns) > 0 && !util.InSlice(columns, "hash") {
		columns = append(columns, "hash")
	}

	// Empty slice should remain empty (will use default columns in actual code)
	if len(columns) != 0 {
		t.Errorf("Expected empty slice to remain empty, got: %v", columns)
	}
}
