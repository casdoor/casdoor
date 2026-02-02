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

	admin "google.golang.org/api/admin/directory/v1"
)

func TestGoogleWorkspaceUserToOriginalUser(t *testing.T) {
	provider := &GoogleWorkspaceSyncerProvider{
		Syncer: &Syncer{},
	}

	// Test case 1: Full Google Workspace user with all fields
	gwUser := &admin.User{
		Id:           "user-123",
		PrimaryEmail: "john.doe@example.com",
		Name: &admin.UserName{
			FullName:   "John Doe",
			GivenName:  "John",
			FamilyName: "Doe",
		},
		ThumbnailPhotoUrl: "https://example.com/avatar.jpg",
		Suspended:         false,
		IsAdmin:           true,
		CreationTime:      "2024-01-01T00:00:00Z",
	}

	originalUser := provider.googleWorkspaceUserToOriginalUser(gwUser)

	// Verify basic fields
	if originalUser.Id != "user-123" {
		t.Errorf("Expected Id to be 'user-123', got '%s'", originalUser.Id)
	}
	if originalUser.Name != "john.doe@example.com" {
		t.Errorf("Expected Name to be 'john.doe@example.com', got '%s'", originalUser.Name)
	}
	if originalUser.Email != "john.doe@example.com" {
		t.Errorf("Expected Email to be 'john.doe@example.com', got '%s'", originalUser.Email)
	}
	if originalUser.DisplayName != "John Doe" {
		t.Errorf("Expected DisplayName to be 'John Doe', got '%s'", originalUser.DisplayName)
	}
	if originalUser.FirstName != "John" {
		t.Errorf("Expected FirstName to be 'John', got '%s'", originalUser.FirstName)
	}
	if originalUser.LastName != "Doe" {
		t.Errorf("Expected LastName to be 'Doe', got '%s'", originalUser.LastName)
	}
	if originalUser.Avatar != "https://example.com/avatar.jpg" {
		t.Errorf("Expected Avatar to be 'https://example.com/avatar.jpg', got '%s'", originalUser.Avatar)
	}
	if originalUser.IsForbidden != false {
		t.Errorf("Expected IsForbidden to be false for non-suspended user, got %v", originalUser.IsForbidden)
	}
	if originalUser.IsAdmin != true {
		t.Errorf("Expected IsAdmin to be true, got %v", originalUser.IsAdmin)
	}

	// Test case 2: Suspended Google Workspace user
	suspendedUser := &admin.User{
		Id:           "user-456",
		PrimaryEmail: "jane.doe@example.com",
		Name: &admin.UserName{
			FullName: "Jane Doe",
		},
		Suspended: true,
	}

	suspendedOriginalUser := provider.googleWorkspaceUserToOriginalUser(suspendedUser)
	if suspendedOriginalUser.IsForbidden != true {
		t.Errorf("Expected IsForbidden to be true for suspended user, got %v", suspendedOriginalUser.IsForbidden)
	}

	// Test case 3: User with no Name object (should not panic)
	minimalUser := &admin.User{
		Id:           "user-789",
		PrimaryEmail: "bob@example.com",
	}

	minimalOriginalUser := provider.googleWorkspaceUserToOriginalUser(minimalUser)
	if minimalOriginalUser.DisplayName != "" {
		t.Errorf("Expected DisplayName to be empty for minimal user, got '%s'", minimalOriginalUser.DisplayName)
	}

	// Test case 4: Display name construction from first/last name when FullName is empty
	noFullNameUser := &admin.User{
		Id:           "user-101",
		PrimaryEmail: "alice@example.com",
		Name: &admin.UserName{
			GivenName:  "Alice",
			FamilyName: "Jones",
		},
	}

	noFullNameOriginalUser := provider.googleWorkspaceUserToOriginalUser(noFullNameUser)
	if noFullNameOriginalUser.DisplayName != "Alice Jones" {
		t.Errorf("Expected DisplayName to be constructed as 'Alice Jones', got '%s'", noFullNameOriginalUser.DisplayName)
	}
}

func TestGoogleWorkspaceGroupToOriginalGroup(t *testing.T) {
	provider := &GoogleWorkspaceSyncerProvider{
		Syncer: &Syncer{},
	}

	// Test case 1: Full Google Workspace group with all fields
	gwGroup := &admin.Group{
		Id:          "group-123",
		Email:       "team@example.com",
		Name:        "Engineering Team",
		Description: "All engineering staff",
	}

	originalGroup := provider.googleWorkspaceGroupToOriginalGroup(gwGroup)

	// Verify all fields
	if originalGroup.Id != "group-123" {
		t.Errorf("Expected Id to be 'group-123', got '%s'", originalGroup.Id)
	}
	if originalGroup.Name != "team@example.com" {
		t.Errorf("Expected Name to be 'team@example.com', got '%s'", originalGroup.Name)
	}
	if originalGroup.DisplayName != "Engineering Team" {
		t.Errorf("Expected DisplayName to be 'Engineering Team', got '%s'", originalGroup.DisplayName)
	}
	if originalGroup.Description != "All engineering staff" {
		t.Errorf("Expected Description to be 'All engineering staff', got '%s'", originalGroup.Description)
	}
	if originalGroup.Email != "team@example.com" {
		t.Errorf("Expected Email to be 'team@example.com', got '%s'", originalGroup.Email)
	}

	// Test case 2: Minimal group
	minimalGroup := &admin.Group{
		Id:    "group-456",
		Email: "minimal@example.com",
	}

	minimalOriginalGroup := provider.googleWorkspaceGroupToOriginalGroup(minimalGroup)
	if minimalOriginalGroup.DisplayName != "" {
		t.Errorf("Expected DisplayName to be empty for minimal group, got '%s'", minimalOriginalGroup.DisplayName)
	}
	if minimalOriginalGroup.Description != "" {
		t.Errorf("Expected Description to be empty for minimal group, got '%s'", minimalOriginalGroup.Description)
	}
}

func TestGetSyncerProviderGoogleWorkspace(t *testing.T) {
	syncer := &Syncer{
		Type: "Google Workspace",
		Host: "admin@example.com",
	}

	provider := GetSyncerProvider(syncer)

	if _, ok := provider.(*GoogleWorkspaceSyncerProvider); !ok {
		t.Errorf("Expected GoogleWorkspaceSyncerProvider for type 'Google Workspace', got %T", provider)
	}
}

func TestGoogleWorkspaceSyncerProviderEmptyMethods(t *testing.T) {
	provider := &GoogleWorkspaceSyncerProvider{
		Syncer: &Syncer{},
	}

	// Test AddUser returns error
	_, err := provider.AddUser(&OriginalUser{})
	if err == nil {
		t.Error("Expected AddUser to return error for read-only syncer")
	}

	// Test UpdateUser returns error
	_, err = provider.UpdateUser(&OriginalUser{})
	if err == nil {
		t.Error("Expected UpdateUser to return error for read-only syncer")
	}

	// Test Close returns no error
	err = provider.Close()
	if err != nil {
		t.Errorf("Expected Close to return nil, got error: %v", err)
	}

	// Test InitAdapter returns no error
	err = provider.InitAdapter()
	if err != nil {
		t.Errorf("Expected InitAdapter to return nil, got error: %v", err)
	}
}
