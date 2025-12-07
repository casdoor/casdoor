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
)

func TestGetMapFromOriginalUser_EmptyTimestampFields(t *testing.T) {
	// Create a syncer with table columns
	syncer := &Syncer{
		TableColumns: []*TableColumn{
			{Name: "name", CasdoorName: "Name", Type: "string"},
			{Name: "created_time", CasdoorName: "CreatedTime", Type: "timestamp"},
			{Name: "updated_time", CasdoorName: "UpdatedTime", Type: "timestamp"},
			{Name: "deleted_time", CasdoorName: "DeletedTime", Type: "timestamp"},
			{Name: "email", CasdoorName: "Email", Type: "string"},
			{Name: "score", CasdoorName: "Score", Type: "int"},
		},
	}

	// Create a user with empty timestamp fields
	user := &OriginalUser{
		Name:        "testuser",
		CreatedTime: "2025-12-07T10:00:00+08:00",
		UpdatedTime: "", // Empty - should be skipped
		DeletedTime: "", // Empty - should be skipped
		Email:       "test@example.com",
		Score:       100,
	}

	// Get the map
	m := syncer.getMapFromOriginalUser(user)

	// Verify that Name is included (string type, even if empty would be included)
	if _, ok := m["name"]; !ok {
		t.Error("Expected 'name' to be in the map")
	}

	// Verify that CreatedTime is included (non-empty timestamp)
	if _, ok := m["created_time"]; !ok {
		t.Error("Expected 'created_time' to be in the map")
	}

	// Verify that UpdatedTime is NOT included (empty timestamp)
	if _, ok := m["updated_time"]; ok {
		t.Error("Expected 'updated_time' to be excluded from the map (empty timestamp)")
	}

	// Verify that DeletedTime is NOT included (empty timestamp)
	if _, ok := m["deleted_time"]; ok {
		t.Error("Expected 'deleted_time' to be excluded from the map (empty timestamp)")
	}

	// Verify that Email is included (string type)
	if _, ok := m["email"]; !ok {
		t.Error("Expected 'email' to be in the map")
	}

	// Verify that Score is included (int type, non-empty)
	if _, ok := m["score"]; !ok {
		t.Error("Expected 'score' to be in the map")
	}
}

func TestGetMapFromOriginalUser_EmptyStringFields(t *testing.T) {
	// Create a syncer with string type columns
	syncer := &Syncer{
		TableColumns: []*TableColumn{
			{Name: "name", CasdoorName: "Name", Type: "string"},
			{Name: "email", CasdoorName: "Email", Type: "string"},
			{Name: "phone", CasdoorName: "Phone", Type: "string"},
		},
	}

	// Create a user with some empty string fields
	user := &OriginalUser{
		Name:  "testuser",
		Email: "",    // Empty string - should still be included for string types
		Phone: "123", // Non-empty
	}

	// Get the map
	m := syncer.getMapFromOriginalUser(user)

	// Verify that all string fields are included, even if empty
	if _, ok := m["name"]; !ok {
		t.Error("Expected 'name' to be in the map")
	}

	if _, ok := m["email"]; !ok {
		t.Error("Expected 'email' to be in the map (empty string fields should be included)")
	}

	if _, ok := m["phone"]; !ok {
		t.Error("Expected 'phone' to be in the map")
	}
}
