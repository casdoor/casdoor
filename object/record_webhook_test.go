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
	"encoding/json"
	"testing"
)

func TestFilterRecordObject(t *testing.T) {
	// Test with a valid object
	object := `{"name":"test","email":"test@example.com","phone":"123456","age":25}`
	fields := []string{"name", "email"}

	result := filterRecordObject(object, fields)

	var filtered map[string]interface{}
	err := json.Unmarshal([]byte(result), &filtered)
	if err != nil {
		t.Errorf("Failed to unmarshal filtered object: %v", err)
	}

	if len(filtered) != 2 {
		t.Errorf("Expected 2 fields, got %d", len(filtered))
	}

	if filtered["name"] != "test" {
		t.Errorf("Expected name to be 'test', got %v", filtered["name"])
	}

	if filtered["email"] != "test@example.com" {
		t.Errorf("Expected email to be 'test@example.com', got %v", filtered["email"])
	}

	if _, exists := filtered["phone"]; exists {
		t.Error("Field 'phone' should not exist in filtered object")
	}

	if _, exists := filtered["age"]; exists {
		t.Error("Field 'age' should not exist in filtered object")
	}
}

func TestFilterRecordObjectWithEmptyFields(t *testing.T) {
	object := `{"name":"test","email":"test@example.com"}`
	fields := []string{}

	result := filterRecordObject(object, fields)

	var filtered map[string]interface{}
	err := json.Unmarshal([]byte(result), &filtered)
	if err != nil {
		t.Errorf("Failed to unmarshal filtered object: %v", err)
	}

	if len(filtered) != 0 {
		t.Errorf("Expected 0 fields, got %d", len(filtered))
	}
}

func TestFilterRecordObjectWithInvalidJSON(t *testing.T) {
	object := "invalid json"
	fields := []string{"name"}

	result := filterRecordObject(object, fields)

	if result != object {
		t.Errorf("Expected invalid JSON to be returned as-is, got %s", result)
	}
}

func TestFilterRecordObjectWithEmptyString(t *testing.T) {
	object := ""
	fields := []string{"name"}

	result := filterRecordObject(object, fields)

	if result != object {
		t.Errorf("Expected empty string to be returned as-is, got %s", result)
	}
}

func TestFilterRecordObjectIsolation(t *testing.T) {
	// This test verifies that filtering doesn't affect the original object
	// and that multiple filter operations work independently
	originalObject := `{"name":"test","email":"test@example.com","phone":"123456","address":"home"}`

	// First filter
	fields1 := []string{"name", "email"}
	result1 := filterRecordObject(originalObject, fields1)

	var filtered1 map[string]interface{}
	err := json.Unmarshal([]byte(result1), &filtered1)
	if err != nil {
		t.Errorf("Failed to unmarshal first filtered object: %v", err)
	}

	if len(filtered1) != 2 {
		t.Errorf("Expected 2 fields in first filter, got %d", len(filtered1))
	}

	// Second filter with different fields
	fields2 := []string{"phone", "address"}
	result2 := filterRecordObject(originalObject, fields2)

	var filtered2 map[string]interface{}
	err = json.Unmarshal([]byte(result2), &filtered2)
	if err != nil {
		t.Errorf("Failed to unmarshal second filtered object: %v", err)
	}

	if len(filtered2) != 2 {
		t.Errorf("Expected 2 fields in second filter, got %d", len(filtered2))
	}

	// Verify each filter has the correct fields
	if _, exists := filtered1["name"]; !exists {
		t.Error("First filter should contain 'name'")
	}
	if _, exists := filtered1["email"]; !exists {
		t.Error("First filter should contain 'email'")
	}
	if _, exists := filtered1["phone"]; exists {
		t.Error("First filter should not contain 'phone'")
	}

	if _, exists := filtered2["phone"]; !exists {
		t.Error("Second filter should contain 'phone'")
	}
	if _, exists := filtered2["address"]; !exists {
		t.Error("Second filter should contain 'address'")
	}
	if _, exists := filtered2["name"]; exists {
		t.Error("Second filter should not contain 'name'")
	}
}
