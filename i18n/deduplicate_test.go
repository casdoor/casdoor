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

package i18n

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"
)

// findDuplicateKeysInJSON finds duplicate keys within each namespace in an i18n JSON file
// Returns a map of namespace -> list of duplicate keys
func findDuplicateKeysInJSON(filePath string) (map[string][]string, error) {
	// Read the JSON file
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	duplicates := make(map[string][]string)
	scanner := bufio.NewScanner(bytes.NewReader(fileContent))

	// Regular expression to match JSON key-value pairs
	// Matches: "key": "value" or "key": value
	keyRegex := regexp.MustCompile(`^\s*"([^"]+)"\s*:`)

	var currentNamespace string
	namespaceKeys := make(map[string]int)
	bracketDepth := 0
	inNamespaceBlock := false

	for scanner.Scan() {
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)

		// Match keys in the line BEFORE updating bracket depth
		if matches := keyRegex.FindStringSubmatch(line); len(matches) > 1 {
			key := matches[1]

			// Determine if this is a namespace (depth = 1) or a regular key (depth = 2)
			if bracketDepth == 1 && !inNamespaceBlock {
				// This is a namespace key - start tracking a new namespace
				currentNamespace = key
				namespaceKeys = make(map[string]int)
				inNamespaceBlock = true
			} else if bracketDepth == 2 && inNamespaceBlock {
				// This is a key within the current namespace
				namespaceKeys[key]++
			}
		}

		// Count opening and closing braces to track depth AFTER processing keys
		bracketDepth += strings.Count(line, "{") - strings.Count(line, "}")

		// Reset when we exit a namespace block
		if (trimmedLine == "}," || trimmedLine == "}") && bracketDepth == 1 && inNamespaceBlock {
			// End of namespace block - check for duplicates
			for k, count := range namespaceKeys {
				if count > 1 {
					if duplicates[currentNamespace] == nil {
						duplicates[currentNamespace] = []string{}
					}
					duplicates[currentNamespace] = append(duplicates[currentNamespace], k)
				}
			}
			inNamespaceBlock = false
		}
	}

	// Check the last namespace if we're still in one
	if inNamespaceBlock && currentNamespace != "" {
		for k, count := range namespaceKeys {
			if count > 1 {
				if duplicates[currentNamespace] == nil {
					duplicates[currentNamespace] = []string{}
				}
				duplicates[currentNamespace] = append(duplicates[currentNamespace], k)
			}
		}
	}

	return duplicates, scanner.Err()
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
		for namespace, keys := range duplicates {
			for _, key := range keys {
				t.Errorf("  Namespace '%s': duplicate key '%s'", namespace, key)
			}
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
		for namespace, keys := range duplicates {
			for _, key := range keys {
				t.Errorf("  Namespace '%s': duplicate key '%s'", namespace, key)
			}
		}
		t.Fail()
	}
}
