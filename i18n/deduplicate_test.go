// Copyright 2026 The Casdoor Authors. All Rights Reserved.
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

//go:build !skipCi

package i18n

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

// findDuplicateKeysInJSON finds duplicate keys across the entire JSON file
// Returns a list of duplicate keys found anywhere in the file
func findDuplicateKeysInJSON(filePath string) ([]string, error) {
	// Read the JSON file
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	// Parse JSON into a map to check structure
	var data map[string]map[string]string
	if err := json.Unmarshal(fileContent, &data); err != nil {
		return nil, fmt.Errorf("failed to parse JSON from %s: %w", filePath, err)
	}

	// Track all keys seen across the entire file
	keySeen := make(map[string]bool)
	var duplicates []string

	// Decode the top-level object
	var rawData map[string]json.RawMessage
	if err := json.Unmarshal(fileContent, &rawData); err != nil {
		return nil, fmt.Errorf("failed to decode JSON: %w", err)
	}

	// For each namespace, check all keys across the whole file
	for _, rawNamespace := range rawData {
		// Use a custom decoder to read all keys
		decoder := json.NewDecoder(bytes.NewReader(rawNamespace))
		
		// Read the opening brace
		token, err := decoder.Token()
		if err != nil {
			return nil, fmt.Errorf("failed to read token: %w", err)
		}
		if delim, ok := token.(json.Delim); !ok || delim != '{' {
			return nil, fmt.Errorf("expected object start, got %v", token)
		}

		// Read all key-value pairs
		for decoder.More() {
			// Read the key
			token, err := decoder.Token()
			if err != nil {
				return nil, fmt.Errorf("failed to read key: %w", err)
			}
			
			key, ok := token.(string)
			if !ok {
				return nil, fmt.Errorf("expected string key, got %v", token)
			}

			// Check if this key was already seen anywhere in the file
			if keySeen[key] {
				// Only add to duplicates list if not already there
				alreadyRecorded := false
				for _, dup := range duplicates {
					if dup == key {
						alreadyRecorded = true
						break
					}
				}
				if !alreadyRecorded {
					duplicates = append(duplicates, key)
				}
			}
			keySeen[key] = true

			// Skip the value
			var value interface{}
			if err := decoder.Decode(&value); err != nil {
				return nil, fmt.Errorf("failed to decode value: %w", err)
			}
		}
	}

	return duplicates, nil
}

// TestDeduplicateFrontendI18n checks for duplicate i18n keys in the frontend en.json file
func TestDeduplicateFrontendI18n(t *testing.T) {
	filePath := "../web/src/locales/en/data.json"

	// Find duplicate keys
	duplicates, err := findDuplicateKeysInJSON(filePath)
	if err != nil {
		t.Fatalf("Failed to check for duplicates in frontend i18n file: %v", err)
	}

	// Print all duplicates and fail the test if any are found
	if len(duplicates) > 0 {
		t.Errorf("Found duplicate i18n keys in frontend file (%s):", filePath)
		for _, key := range duplicates {
			t.Errorf("  Duplicate key: '%s'", key)
		}
		t.Fail()
	}
}

// TestDeduplicateBackendI18n checks for duplicate i18n keys in the backend en.json file
func TestDeduplicateBackendI18n(t *testing.T) {
	filePath := "../i18n/locales/en/data.json"

	// Find duplicate keys
	duplicates, err := findDuplicateKeysInJSON(filePath)
	if err != nil {
		t.Fatalf("Failed to check for duplicates in backend i18n file: %v", err)
	}

	// Print all duplicates and fail the test if any are found
	if len(duplicates) > 0 {
		t.Errorf("Found duplicate i18n keys in backend file (%s):", filePath)
		for _, key := range duplicates {
			t.Errorf("  Duplicate key: '%s'", key)
		}
		t.Fail()
	}
}
