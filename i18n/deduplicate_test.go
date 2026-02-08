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

package i18n

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

// DuplicateInfo represents information about a duplicate key
type DuplicateInfo struct {
	Key          string
	OldPrefix    string
	NewPrefix    string
	OldPrefixKey string // e.g., "general:Submitter"
	NewPrefixKey string // e.g., "permission:Submitter"
}

// findDuplicateKeysInJSON finds duplicate keys across the entire JSON file
// Returns a list of duplicate information showing old and new prefix:key pairs
// The order is determined by the order keys appear in the JSON file (git history)
func findDuplicateKeysInJSON(filePath string) ([]DuplicateInfo, error) {
	// Read the JSON file
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	// Track the first occurrence of each key (prefix where it was first seen)
	keyFirstPrefix := make(map[string]string)
	var duplicates []DuplicateInfo

	// To preserve order, we need to parse the JSON with order preservation
	// We'll use a decoder to read through the top-level object
	decoder := json.NewDecoder(bytes.NewReader(fileContent))

	// Read the opening brace of the top-level object
	token, err := decoder.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to read token: %w", err)
	}
	if delim, ok := token.(json.Delim); !ok || delim != '{' {
		return nil, fmt.Errorf("expected object start, got %v", token)
	}

	// Read all namespaces in order
	for decoder.More() {
		// Read the namespace (prefix) name
		token, err := decoder.Token()
		if err != nil {
			return nil, fmt.Errorf("failed to read namespace: %w", err)
		}

		prefix, ok := token.(string)
		if !ok {
			return nil, fmt.Errorf("expected string namespace, got %v", token)
		}

		// Read the namespace object as raw message
		var namespaceData map[string]string
		if err := decoder.Decode(&namespaceData); err != nil {
			return nil, fmt.Errorf("failed to decode namespace %s: %w", prefix, err)
		}

		// Now check each key in this namespace
		for key := range namespaceData {
			// Check if this key was already seen in a different prefix
			if firstPrefix, exists := keyFirstPrefix[key]; exists {
				// This is a duplicate - the key exists in another prefix
				duplicates = append(duplicates, DuplicateInfo{
					Key:          key,
					OldPrefix:    firstPrefix,
					NewPrefix:    prefix,
					OldPrefixKey: fmt.Sprintf("%s:%s", firstPrefix, key),
					NewPrefixKey: fmt.Sprintf("%s:%s", prefix, key),
				})
			} else {
				// First time seeing this key, record the prefix
				keyFirstPrefix[key] = prefix
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
		for _, dup := range duplicates {
			t.Errorf("  i18next.t(\"%s\") duplicates with i18next.t(\"%s\")", dup.NewPrefixKey, dup.OldPrefixKey)
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
		for _, dup := range duplicates {
			t.Errorf("  i18n.Translate(\"%s\") duplicates with i18n.Translate(\"%s\")", dup.NewPrefixKey, dup.OldPrefixKey)
		}
		t.Fail()
	}
}
