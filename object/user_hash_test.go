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

	"github.com/casdoor/casdoor/util"
)

func TestUpdateUserHashWithColumnsSpecified(t *testing.T) {
	InitConfig()

	// Create a test user
	user := &User{
		Owner:        "built-in",
		Name:         "test_hash_user_" + util.GenerateId()[:8],
		CreatedTime:  util.GetCurrentTime(),
		UpdatedTime:  util.GetCurrentTime(),
		Id:           util.GenerateId(),
		Password:     "test123",
		PasswordType: "plain",
		DisplayName:  "Test Hash User",
	}

	// Add the user
	_, err := AddUser(user, "en")
	if err != nil {
		t.Fatalf("Failed to add user: %v", err)
	}

	// Clean up after test
	defer func() {
		_, _ = deleteUser(user)
	}()

	// Get the initial hash
	initialUser, err := getUser(user.Owner, user.Name)
	if err != nil {
		t.Fatalf("Failed to get user: %v", err)
	}
	initialHash := initialUser.Hash

	// Update user password with specific columns (simulating /api/set-password)
	initialUser.Password = "newpassword123"
	organization, err := GetOrganizationByUser(initialUser)
	if err != nil {
		t.Fatalf("Failed to get organization: %v", err)
	}
	initialUser.UpdateUserPassword(organization)

	// Update with specific columns, not including "hash" explicitly
	columns := []string{"password", "password_salt", "password_type"}
	_, err = UpdateUser(initialUser.GetId(), initialUser, columns, false)
	if err != nil {
		t.Fatalf("Failed to update user: %v", err)
	}

	// Retrieve the user again to verify hash was updated in database
	updatedUser, err := getUser(user.Owner, user.Name)
	if err != nil {
		t.Fatalf("Failed to get updated user: %v", err)
	}

	// Verify the hash was updated in the database
	if updatedUser.Hash == initialHash {
		t.Errorf("Hash was not updated in database. Initial hash: %s, Updated hash: %s", initialHash, updatedUser.Hash)
	}

	// Verify the hash is not empty
	if updatedUser.Hash == "" {
		t.Errorf("Hash should not be empty after update")
	}

	// Verify password was actually changed
	if updatedUser.Password == user.Password {
		t.Errorf("Password was not updated")
	}
}

func TestUpdateUserHashDeduplication(t *testing.T) {
	InitConfig()

	// Create a test user
	user := &User{
		Owner:        "built-in",
		Name:         "test_hash_dedup_" + util.GenerateId()[:8],
		CreatedTime:  util.GetCurrentTime(),
		UpdatedTime:  util.GetCurrentTime(),
		Id:           util.GenerateId(),
		Password:     "test456",
		PasswordType: "plain",
		DisplayName:  "Test Hash Dedup User",
	}

	// Add the user
	_, err := AddUser(user, "en")
	if err != nil {
		t.Fatalf("Failed to add user: %v", err)
	}

	// Clean up after test
	defer func() {
		_, _ = deleteUser(user)
	}()

	// Get the user
	testUser, err := getUser(user.Owner, user.Name)
	if err != nil {
		t.Fatalf("Failed to get user: %v", err)
	}

	// Update user with "hash" already in columns list
	testUser.Password = "newpassword456"
	organization, err := GetOrganizationByUser(testUser)
	if err != nil {
		t.Fatalf("Failed to get organization: %v", err)
	}
	testUser.UpdateUserPassword(organization)

	// Update with "hash" explicitly included
	columns := []string{"password", "password_salt", "password_type", "hash"}
	_, err = UpdateUser(testUser.GetId(), testUser, columns, false)
	if err != nil {
		t.Fatalf("Failed to update user: %v", err)
	}

	// Retrieve the user again to verify it worked
	updatedUser, err := getUser(user.Owner, user.Name)
	if err != nil {
		t.Fatalf("Failed to get updated user: %v", err)
	}

	// Verify the hash is not empty
	if updatedUser.Hash == "" {
		t.Errorf("Hash should not be empty after update")
	}
}

func TestUpdateUserHashEmptyColumns(t *testing.T) {
	InitConfig()

	// Create a test user
	user := &User{
		Owner:        "built-in",
		Name:         "test_hash_empty_" + util.GenerateId()[:8],
		CreatedTime:  util.GetCurrentTime(),
		UpdatedTime:  util.GetCurrentTime(),
		Id:           util.GenerateId(),
		Password:     "test789",
		PasswordType: "plain",
		DisplayName:  "Test Hash Empty Cols User",
	}

	// Add the user
	_, err := AddUser(user, "en")
	if err != nil {
		t.Fatalf("Failed to add user: %v", err)
	}

	// Clean up after test
	defer func() {
		_, _ = deleteUser(user)
	}()

	// Get the initial hash
	initialUser, err := getUser(user.Owner, user.Name)
	if err != nil {
		t.Fatalf("Failed to get user: %v", err)
	}
	initialHash := initialUser.Hash

	// Update user with empty columns (should use default columns including hash)
	initialUser.Password = "newpassword789"
	organization, err := GetOrganizationByUser(initialUser)
	if err != nil {
		t.Fatalf("Failed to get organization: %v", err)
	}
	initialUser.UpdateUserPassword(organization)

	// Update with empty columns
	_, err = UpdateUser(initialUser.GetId(), initialUser, []string{}, false)
	if err != nil {
		t.Fatalf("Failed to update user: %v", err)
	}

	// Retrieve the user again to verify hash was updated
	updatedUser, err := getUser(user.Owner, user.Name)
	if err != nil {
		t.Fatalf("Failed to get updated user: %v", err)
	}

	// Verify the hash was updated
	if updatedUser.Hash == initialHash {
		t.Errorf("Hash was not updated. Initial hash: %s, Updated hash: %s", initialHash, updatedUser.Hash)
	}

	// Verify the hash is not empty
	if updatedUser.Hash == "" {
		t.Errorf("Hash should not be empty after update")
	}
}
