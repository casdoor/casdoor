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

	"github.com/casdoor/casdoor/util"
)

func TestUserOrganizationStruct(t *testing.T) {
	// Test UserOrganization struct creation
	userOrg := &UserOrganization{
		Owner:        "built-in",
		Name:         "admin",
		Organization: "test-org",
		CreatedTime:  util.GetCurrentTime(),
		IsDefault:    false,
	}

	if userOrg.Owner != "built-in" {
		t.Errorf("Expected owner 'built-in', got '%s'", userOrg.Owner)
	}
	if userOrg.Name != "admin" {
		t.Errorf("Expected name 'admin', got '%s'", userOrg.Name)
	}
	if userOrg.Organization != "test-org" {
		t.Errorf("Expected organization 'test-org', got '%s'", userOrg.Organization)
	}
	if userOrg.IsDefault != false {
		t.Error("Expected IsDefault to be false")
	}
}

func TestUserOrganizationGetId(t *testing.T) {
	userOrg := &UserOrganization{
		Owner:        "built-in",
		Name:         "admin",
		Organization: "test-org",
		CreatedTime:  util.GetCurrentTime(),
		IsDefault:    false,
	}

	expectedId := "built-in/admin/test-org"
	actualId := userOrg.GetId()

	if actualId != expectedId {
		t.Errorf("Expected ID '%s', got '%s'", expectedId, actualId)
	}
}

// Integration tests below require database connection
// To run these tests, ensure database is configured and running

// func TestUserOrganizationIntegration(t *testing.T) {
// 	InitConfig()
//
// 	// Test creating a user organization relationship
// 	userOrg := &UserOrganization{
// 		Owner:        "built-in",
// 		Name:         "admin",
// 		Organization: "test-org",
// 		CreatedTime:  util.GetCurrentTime(),
// 		IsDefault:    false,
// 	}
//
// 	// Add the relationship
// 	added, err := AddUserOrganization(userOrg)
// 	if err != nil {
// 		t.Errorf("Failed to add user organization: %v", err)
// 	}
// 	if !added {
// 		t.Error("User organization was not added")
// 	}
//
// 	// Get the relationship
// 	retrieved, err := GetUserOrganization("built-in", "admin", "test-org")
// 	if err != nil {
// 		t.Errorf("Failed to get user organization: %v", err)
// 	}
// 	if retrieved == nil {
// 		t.Error("User organization not found")
// 	}
// 	if retrieved.Organization != "test-org" {
// 		t.Errorf("Expected organization 'test-org', got '%s'", retrieved.Organization)
// 	}
//
// 	// Clean up
// 	deleted, err := DeleteUserOrganization("built-in", "admin", "test-org")
// 	if err != nil {
// 		t.Errorf("Failed to delete user organization: %v", err)
// 	}
// 	if !deleted {
// 		t.Error("User organization was not deleted")
// 	}
// }

