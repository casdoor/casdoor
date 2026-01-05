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

func TestDingtalkSyncerProvider(t *testing.T) {
	syncer := &Syncer{
		Type:     "DingTalk",
		User:     "test_app_key",
		Password: "test_app_secret",
	}

	provider := GetSyncerProvider(syncer)

	// Verify the provider is of the correct type
	if _, ok := provider.(*DingtalkSyncerProvider); !ok {
		t.Errorf("Expected DingtalkSyncerProvider, got %T", provider)
	}

	// Test InitAdapter (should not return error for DingTalk)
	err := provider.InitAdapter()
	if err != nil {
		t.Errorf("InitAdapter should not return error for DingTalk syncer: %v", err)
	}

	// Test Close (should not return error)
	err = provider.Close()
	if err != nil {
		t.Errorf("Close should not return error for DingTalk syncer: %v", err)
	}
}

func TestDingtalkUserToOriginalUser(t *testing.T) {
	syncer := &Syncer{
		Type: "DingTalk",
	}
	provider := &DingtalkSyncerProvider{Syncer: syncer}

	// Test with job number
	dingtalkUser := &DingtalkUser{
		UserId:    "user123",
		Name:      "Test User",
		Email:     "test@example.com",
		Mobile:    "13800138000",
		Avatar:    "http://example.com/avatar.jpg",
		Position:  "Developer",
		JobNumber: "EMP001",
		Active:    true,
	}

	originalUser := provider.dingtalkUserToOriginalUser(dingtalkUser)

	if originalUser.Id != "user123" {
		t.Errorf("Expected Id to be 'user123', got '%s'", originalUser.Id)
	}

	if originalUser.Name != "EMP001" {
		t.Errorf("Expected Name to be 'EMP001', got '%s'", originalUser.Name)
	}

	if originalUser.DisplayName != "Test User" {
		t.Errorf("Expected DisplayName to be 'Test User', got '%s'", originalUser.DisplayName)
	}

	if originalUser.Email != "test@example.com" {
		t.Errorf("Expected Email to be 'test@example.com', got '%s'", originalUser.Email)
	}

	if originalUser.IsForbidden {
		t.Errorf("Expected IsForbidden to be false for active user")
	}

	// Test without job number
	dingtalkUser2 := &DingtalkUser{
		UserId: "user456",
		Name:   "Another User",
		Active: false,
	}

	originalUser2 := provider.dingtalkUserToOriginalUser(dingtalkUser2)

	if originalUser2.Name != "user456" {
		t.Errorf("Expected Name to be 'user456', got '%s'", originalUser2.Name)
	}

	if !originalUser2.IsForbidden {
		t.Errorf("Expected IsForbidden to be true for inactive user")
	}
}

func TestDingtalkAddUpdateUser(t *testing.T) {
	syncer := &Syncer{
		Type: "DingTalk",
	}
	provider := &DingtalkSyncerProvider{Syncer: syncer}

	user := &OriginalUser{
		Id:   "test123",
		Name: "Test User",
	}

	// AddUser should not be supported
	_, err := provider.AddUser(user)
	if err == nil {
		t.Error("AddUser should return an error for DingTalk syncer")
	}

	// UpdateUser should not be supported
	_, err = provider.UpdateUser(user)
	if err == nil {
		t.Error("UpdateUser should return an error for DingTalk syncer")
	}
}
