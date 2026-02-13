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

func TestDingtalkUserToOriginalUser_WithTableColumns(t *testing.T) {
	// Create a test syncer with TableColumns configuration
	syncer := &Syncer{
		Type: "DingTalk",
		TableColumns: []*TableColumn{
			{Name: "userid", CasdoorName: "Id", IsHashed: true},
			{Name: "unionid", CasdoorName: "Name", IsHashed: true},
			{Name: "name", CasdoorName: "DisplayName", IsHashed: true},
			{Name: "email", CasdoorName: "Email", IsHashed: true},
			{Name: "mobile", CasdoorName: "Phone", IsHashed: true},
			{Name: "avatar", CasdoorName: "Avatar", IsHashed: true},
			{Name: "title", CasdoorName: "Title", IsHashed: true},
			{Name: "active", CasdoorName: "IsForbidden", IsHashed: true},
		},
	}

	provider := &DingtalkSyncerProvider{Syncer: syncer}

	// Create a test DingTalk user
	dingtalkUser := &DingtalkUser{
		UserId:     "test_userid",
		UnionId:    "test_unionid",
		Name:       "Test User",
		Email:      "test@example.com",
		Mobile:     "1234567890",
		Avatar:     "http://example.com/avatar.jpg",
		Position:   "Engineer",
		JobNumber:  "EMP001",
		Active:     true,
		Department: []int{1, 2},
	}

	// Convert to OriginalUser
	originalUser := provider.dingtalkUserToOriginalUser(dingtalkUser)

	// Verify field mappings based on TableColumns
	if originalUser.Id != "test_userid" {
		t.Errorf("Expected Id to be 'test_userid', got '%s'", originalUser.Id)
	}

	if originalUser.Name != "test_unionid" {
		t.Errorf("Expected Name to be 'test_unionid' (mapped from unionid), got '%s'", originalUser.Name)
	}

	if originalUser.DisplayName != "Test User" {
		t.Errorf("Expected DisplayName to be 'Test User', got '%s'", originalUser.DisplayName)
	}

	if originalUser.Email != "test@example.com" {
		t.Errorf("Expected Email to be 'test@example.com', got '%s'", originalUser.Email)
	}

	if originalUser.Phone != "1234567890" {
		t.Errorf("Expected Phone to be '1234567890', got '%s'", originalUser.Phone)
	}

	if originalUser.Avatar != "http://example.com/avatar.jpg" {
		t.Errorf("Expected Avatar to be 'http://example.com/avatar.jpg', got '%s'", originalUser.Avatar)
	}

	if originalUser.Title != "Engineer" {
		t.Errorf("Expected Title to be 'Engineer', got '%s'", originalUser.Title)
	}

	if originalUser.IsForbidden != false {
		t.Errorf("Expected IsForbidden to be false (active=true), got %v", originalUser.IsForbidden)
	}

	if originalUser.DingTalk != "test_userid" {
		t.Errorf("Expected DingTalk to be 'test_userid', got '%s'", originalUser.DingTalk)
	}

	if len(originalUser.Groups) != 2 {
		t.Errorf("Expected 2 groups, got %d", len(originalUser.Groups))
	}
}

func TestDingtalkUserToOriginalUser_WithUserIdAsName(t *testing.T) {
	// Test with userid mapped to Name field
	syncer := &Syncer{
		Type: "DingTalk",
		TableColumns: []*TableColumn{
			{Name: "userid", CasdoorName: "Id", IsHashed: true},
			{Name: "userid", CasdoorName: "Name", IsHashed: true},
			{Name: "name", CasdoorName: "DisplayName", IsHashed: true},
			{Name: "email", CasdoorName: "Email", IsHashed: true},
		},
	}

	provider := &DingtalkSyncerProvider{Syncer: syncer}

	dingtalkUser := &DingtalkUser{
		UserId:     "test_userid",
		UnionId:    "test_unionid",
		Name:       "Test User",
		Email:      "test@example.com",
		Department: []int{},
	}

	originalUser := provider.dingtalkUserToOriginalUser(dingtalkUser)

	if originalUser.Name != "test_userid" {
		t.Errorf("Expected Name to be 'test_userid' (mapped from userid), got '%s'", originalUser.Name)
	}
}

func TestDingtalkUserToOriginalUser_WithEmailAsName(t *testing.T) {
	// Test with email mapped to Name field
	syncer := &Syncer{
		Type: "DingTalk",
		TableColumns: []*TableColumn{
			{Name: "userid", CasdoorName: "Id", IsHashed: true},
			{Name: "email", CasdoorName: "Name", IsHashed: true},
			{Name: "name", CasdoorName: "DisplayName", IsHashed: true},
		},
	}

	provider := &DingtalkSyncerProvider{Syncer: syncer}

	dingtalkUser := &DingtalkUser{
		UserId:     "test_userid",
		UnionId:    "test_unionid",
		Name:       "Test User",
		Email:      "test@example.com",
		Department: []int{},
	}

	originalUser := provider.dingtalkUserToOriginalUser(dingtalkUser)

	if originalUser.Name != "test@example.com" {
		t.Errorf("Expected Name to be 'test@example.com' (mapped from email), got '%s'", originalUser.Name)
	}
}

func TestDingtalkUserToOriginalUser_WithMobileAsName(t *testing.T) {
	// Test with mobile mapped to Name field
	syncer := &Syncer{
		Type: "DingTalk",
		TableColumns: []*TableColumn{
			{Name: "userid", CasdoorName: "Id", IsHashed: true},
			{Name: "mobile", CasdoorName: "Name", IsHashed: true},
			{Name: "name", CasdoorName: "DisplayName", IsHashed: true},
		},
	}

	provider := &DingtalkSyncerProvider{Syncer: syncer}

	dingtalkUser := &DingtalkUser{
		UserId:     "test_userid",
		UnionId:    "test_unionid",
		Name:       "Test User",
		Mobile:     "1234567890",
		Department: []int{},
	}

	originalUser := provider.dingtalkUserToOriginalUser(dingtalkUser)

	if originalUser.Name != "1234567890" {
		t.Errorf("Expected Name to be '1234567890' (mapped from mobile), got '%s'", originalUser.Name)
	}
}

func TestDingtalkUserToOriginalUser_WithoutTableColumns(t *testing.T) {
	// Test backward compatibility without TableColumns
	syncer := &Syncer{
		Type:         "DingTalk",
		TableColumns: nil,
	}

	provider := &DingtalkSyncerProvider{Syncer: syncer}

	dingtalkUser := &DingtalkUser{
		UserId:     "test_userid",
		UnionId:    "test_unionid",
		Name:       "Test User",
		Email:      "test@example.com",
		Mobile:     "1234567890",
		Avatar:     "http://example.com/avatar.jpg",
		Position:   "Engineer",
		Active:     true,
		Department: []int{1},
	}

	originalUser := provider.dingtalkUserToOriginalUser(dingtalkUser)

	// Should use default mapping with unionid as Name
	if originalUser.Id != "test_userid" {
		t.Errorf("Expected Id to be 'test_userid', got '%s'", originalUser.Id)
	}

	if originalUser.Name != "test_unionid" {
		t.Errorf("Expected Name to be 'test_unionid' (default behavior), got '%s'", originalUser.Name)
	}

	if originalUser.DisplayName != "Test User" {
		t.Errorf("Expected DisplayName to be 'Test User', got '%s'", originalUser.DisplayName)
	}

	if originalUser.Email != "test@example.com" {
		t.Errorf("Expected Email to be 'test@example.com', got '%s'", originalUser.Email)
	}
}

func TestDingtalkUserToOriginalUser_WithEmptyUnionId(t *testing.T) {
	// Test backward compatibility without TableColumns and empty UnionId
	syncer := &Syncer{
		Type:         "DingTalk",
		TableColumns: nil,
	}

	provider := &DingtalkSyncerProvider{Syncer: syncer}

	dingtalkUser := &DingtalkUser{
		UserId:     "test_userid",
		UnionId:    "", // Empty UnionId
		Name:       "Test User",
		Email:      "test@example.com",
		Department: []int{},
	}

	originalUser := provider.dingtalkUserToOriginalUser(dingtalkUser)

	// Should fallback to userid as Name when unionid is empty
	if originalUser.Name != "test_userid" {
		t.Errorf("Expected Name to be 'test_userid' (fallback when unionid is empty), got '%s'", originalUser.Name)
	}
}

func TestGetDingtalkUserFieldValue(t *testing.T) {
	provider := &DingtalkSyncerProvider{Syncer: &Syncer{}}

	dingtalkUser := &DingtalkUser{
		UserId:    "test_userid",
		UnionId:   "test_unionid",
		Name:      "Test User",
		Email:     "test@example.com",
		Mobile:    "1234567890",
		Avatar:    "http://example.com/avatar.jpg",
		Position:  "Engineer",
		JobNumber: "EMP001",
		Active:    true,
	}

	tests := []struct {
		fieldName     string
		expectedValue string
	}{
		{"userid", "test_userid"},
		{"unionid", "test_unionid"},
		{"name", "Test User"},
		{"email", "test@example.com"},
		{"mobile", "1234567890"},
		{"avatar", "http://example.com/avatar.jpg"},
		{"title", "Engineer"},
		{"job_number", "EMP001"},
		{"active", "false"}, // active=true means IsForbidden=false
		{"unknown_field", ""},
	}

	for _, test := range tests {
		value := provider.getDingtalkUserFieldValue(dingtalkUser, test.fieldName)
		if value != test.expectedValue {
			t.Errorf("getDingtalkUserFieldValue(%s) = '%s', expected '%s'", test.fieldName, value, test.expectedValue)
		}
	}
}
