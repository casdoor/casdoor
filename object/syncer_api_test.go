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

package object

import (
	"testing"
)

func TestIsApiBasedSyncer(t *testing.T) {
	tests := []struct {
		name     string
		syncType string
		expected bool
	}{
		{"DingTalk should be API-based", "DingTalk", true},
		{"WeCom should be API-based", "WeCom", true},
		{"Azure AD should be API-based", "Azure AD", true},
		{"Google Workspace should be API-based", "Google Workspace", true},
		{"Active Directory should be API-based", "Active Directory", true},
		{"Database should not be API-based", "Database", false},
		{"Keycloak should not be API-based", "Keycloak", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			syncer := &Syncer{Type: tt.syncType}
			result := syncer.isApiBasedSyncer()
			if result != tt.expected {
				t.Errorf("isApiBasedSyncer() for type %s = %v, want %v", tt.syncType, result, tt.expected)
			}
		})
	}
}

func TestGetKeyColumn_ApiSyncer(t *testing.T) {
	syncer := &Syncer{
		Type:         "DingTalk",
		TableColumns: []*TableColumn{}, // Empty for API syncer
	}

	column := syncer.getKeyColumn()
	if column == nil {
		t.Fatal("getKeyColumn() returned nil for API syncer")
	}

	if column.CasdoorName != "Name" {
		t.Errorf("getKeyColumn() CasdoorName = %s, want Name", column.CasdoorName)
	}

	if !column.IsKey {
		t.Error("getKeyColumn() IsKey should be true")
	}
}

func TestGetCasdoorColumns_ApiSyncer(t *testing.T) {
	syncer := &Syncer{
		Type:         "DingTalk",
		TableColumns: []*TableColumn{}, // Empty for API syncer
	}

	columns := syncer.getCasdoorColumns()
	if len(columns) == 0 {
		t.Error("getCasdoorColumns() returned empty for API syncer")
	}

	// Check that important fields are included
	expectedFields := map[string]bool{
		"display_name": true,
		"email":        true,
		"phone":        true,
	}

	for _, col := range columns {
		if expectedFields[col] {
			delete(expectedFields, col)
		}
	}

	if len(expectedFields) > 0 {
		t.Errorf("getCasdoorColumns() missing expected fields: %v", expectedFields)
	}
}

func TestCalculateHash_ApiSyncer(t *testing.T) {
	syncer := &Syncer{
		Type:         "DingTalk",
		Organization: "test-org",
		TableColumns: []*TableColumn{}, // Empty for API syncer
	}

	user := &OriginalUser{
		Name:        "testuser",
		DisplayName: "Test User",
		Email:       "test@example.com",
		Phone:       "1234567890",
	}

	hash1 := syncer.calculateHash(user)
	if hash1 == "" {
		t.Error("calculateHash() returned empty hash")
	}

	// Change user data and verify hash changes
	user.Email = "newemail@example.com"
	hash2 := syncer.calculateHash(user)

	if hash1 == hash2 {
		t.Error("calculateHash() should return different hash when user data changes")
	}
}

func TestGetMapFromOriginalUser_ApiSyncer(t *testing.T) {
	syncer := &Syncer{
		Type:         "DingTalk",
		TableColumns: []*TableColumn{}, // Empty for API syncer
	}

	user := &OriginalUser{
		Name:        "testuser",
		DisplayName: "Test User",
		Email:       "test@example.com",
	}

	userMap := syncer.getMapFromOriginalUser(user)

	// For API syncers, should return full map
	if userMap["Name"] != "testuser" {
		t.Errorf("getMapFromOriginalUser() Name = %s, want testuser", userMap["Name"])
	}

	if userMap["DisplayName"] != "Test User" {
		t.Errorf("getMapFromOriginalUser() DisplayName = %s, want Test User", userMap["DisplayName"])
	}

	if userMap["Email"] != "test@example.com" {
		t.Errorf("getMapFromOriginalUser() Email = %s, want test@example.com", userMap["Email"])
	}
}
