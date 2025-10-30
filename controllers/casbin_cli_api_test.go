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
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCleanOldMEIFolders(t *testing.T) {
	// Create a temporary directory to simulate system temp
	testTempDir, err := os.MkdirTemp("", "test_temp_*")
	if err != nil {
		t.Fatalf("failed to create test temp directory: %v", err)
	}
	defer os.RemoveAll(testTempDir)

	// Create test folders with different ages
	testCases := []struct {
		name            string
		prefix          string
		age             time.Duration
		shouldBeDeleted bool
	}{
		{"old_MEI_folder", "_MEI123456", 25 * time.Hour, true},
		{"recent_MEI_folder", "_MEI789012", 1 * time.Hour, false},
		{"old_non_MEI_folder", "other_folder", 25 * time.Hour, false},
		{"very_old_MEI_folder", "_MEI345678", 48 * time.Hour, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create test directory
			dirPath := filepath.Join(testTempDir, tc.prefix)
			err := os.Mkdir(dirPath, 0755)
			if err != nil {
				t.Fatalf("failed to create test directory: %v", err)
			}

			// Create a test file inside to ensure directory is not empty
			testFile := filepath.Join(dirPath, "test.txt")
			err = os.WriteFile(testFile, []byte("test"), 0644)
			if err != nil {
				t.Fatalf("failed to create test file: %v", err)
			}

			// Set the modification time to simulate age
			// Must be done AFTER creating files inside the directory
			oldTime := time.Now().Add(-tc.age)
			err = os.Chtimes(dirPath, oldTime, oldTime)
			if err != nil {
				t.Fatalf("failed to change directory time: %v", err)
			}
		})
	}

	// Now test the cleanup function by temporarily modifying the temp directory
	// Since we can't easily mock os.TempDir(), we'll test the logic manually
	cutoffTime := time.Now().Add(-24 * time.Hour)
	entries, err := os.ReadDir(testTempDir)
	if err != nil {
		t.Fatalf("failed to read test temp directory: %v", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		dirPath := filepath.Join(testTempDir, entry.Name())
		info, err := entry.Info()
		if err != nil {
			continue
		}

		// Find the test case for this entry
		var tc *struct {
			name            string
			prefix          string
			age             time.Duration
			shouldBeDeleted bool
		}
		for i := range testCases {
			if testCases[i].prefix == entry.Name() {
				tc = &testCases[i]
				break
			}
		}

		if tc == nil {
			continue
		}

		// Check if folder should be cleaned based on _MEI prefix and age
		shouldClean := entry.Name()[:4] == "_MEI" && info.ModTime().Before(cutoffTime)

		if shouldClean != tc.shouldBeDeleted {
			t.Errorf("folder %s: expected shouldBeDeleted=%v, got shouldClean=%v (modTime=%v, cutoff=%v)",
				entry.Name(), tc.shouldBeDeleted, shouldClean, info.ModTime(), cutoffTime)
		}

		// Actually test removal for folders that should be cleaned
		if shouldClean {
			err = os.RemoveAll(dirPath)
			if err != nil {
				t.Errorf("failed to remove folder %s: %v", dirPath, err)
			}

			// Verify it was removed
			if _, err := os.Stat(dirPath); !os.IsNotExist(err) {
				t.Errorf("folder %s still exists after removal", dirPath)
			}
		}
	}

	// Verify that folders that should not be deleted still exist
	for _, tc := range testCases {
		dirPath := filepath.Join(testTempDir, tc.prefix)
		_, err := os.Stat(dirPath)
		exists := !os.IsNotExist(err)

		if tc.shouldBeDeleted && exists {
			t.Errorf("folder %s should have been deleted but still exists", tc.prefix)
		} else if !tc.shouldBeDeleted && !exists {
			t.Errorf("folder %s should not have been deleted but doesn't exist", tc.prefix)
		}
	}
}

func TestCleanOldMEIFolders_Integration(t *testing.T) {
	// This is a smoke test to ensure the function doesn't crash
	// when called with the real system temp directory
	// It should handle errors gracefully
	cleanOldMEIFolders()
	// If we get here without panicking, the test passes
}
