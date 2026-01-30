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

func TestStringArrayToStructWithProperties(t *testing.T) {
	// Test data that includes properties field as JSON
	testData := [][]string{
		{"owner", "name", "password", "display_name", "properties"},
		{"test-org", "test-user", "password123", "Test User", `{"key1":"value1","key2":"value2"}`},
	}

	users, err := StringArrayToStruct[User](testData)
	if err != nil {
		t.Errorf("StringArrayToStruct failed: %v", err)
		return
	}

	if len(users) != 1 {
		t.Errorf("Expected 1 user, got %d", len(users))
		return
	}

	user := users[0]
	if user.Owner != "test-org" {
		t.Errorf("Expected owner 'test-org', got '%s'", user.Owner)
	}
	if user.Name != "test-user" {
		t.Errorf("Expected name 'test-user', got '%s'", user.Name)
	}

	// Check if properties field was parsed correctly
	if user.Properties == nil {
		t.Errorf("Expected properties to be parsed, got nil")
		return
	}

	if len(user.Properties) != 2 {
		t.Errorf("Expected 2 properties, got %d", len(user.Properties))
		return
	}

	if user.Properties["key1"] != "value1" {
		t.Errorf("Expected properties['key1'] = 'value1', got '%s'", user.Properties["key1"])
	}

	if user.Properties["key2"] != "value2" {
		t.Errorf("Expected properties['key2'] = 'value2', got '%s'", user.Properties["key2"])
	}
}

func TestStringArrayToStructWithEmptyProperties(t *testing.T) {
	// Test data with empty properties
	testData := [][]string{
		{"owner", "name", "password", "display_name", "properties"},
		{"test-org", "test-user", "password123", "Test User", ""},
	}

	users, err := StringArrayToStruct[User](testData)
	if err != nil {
		t.Errorf("StringArrayToStruct failed: %v", err)
		return
	}

	if len(users) != 1 {
		t.Errorf("Expected 1 user, got %d", len(users))
		return
	}

	user := users[0]
	// Empty properties should be nil (handled by the condition v == "" || v == "{}" in StringArrayToStruct)
	if user.Properties != nil && len(user.Properties) > 0 {
		t.Errorf("Expected properties to be empty/nil, got %v", user.Properties)
	}
}

func TestStringArrayToStructWithNullProperties(t *testing.T) {
	// Test data with null properties
	testData := [][]string{
		{"owner", "name", "password", "display_name", "properties"},
		{"test-org", "test-user", "password123", "Test User", "null"},
	}

	users, err := StringArrayToStruct[User](testData)
	if err != nil {
		t.Errorf("StringArrayToStruct failed: %v", err)
		return
	}

	if len(users) != 1 {
		t.Errorf("Expected 1 user, got %d", len(users))
		return
	}

	user := users[0]
	// null properties should be skipped
	if user.Properties != nil && len(user.Properties) > 0 {
		t.Errorf("Expected properties to be empty/nil, got %v", user.Properties)
	}
}
