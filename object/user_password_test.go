// Copyright 2024 The Casdoor Authors. All Rights Reserved.
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

	"github.com/casdoor/casdoor/cred"
)

// TestBackwardCompatibilityPasswordVerification tests that password verification
// works with both organization salt (legacy) and user salt (new)
func TestBackwardCompatibilityPasswordVerification(t *testing.T) {
	organization := &Organization{
		PasswordType: "sha512-salt",
		PasswordSalt: "orgFixedSalt123",
	}

	plainPassword := "testPassword123"

	// Scenario 1: Old user with organization salt
	oldUser := &User{
		Name:         "old-user",
		Password:     plainPassword,
		PasswordSalt: organization.PasswordSalt, // Old behavior: uses org salt
	}

	credManager := cred.GetCredManager(organization.PasswordType)
	oldUser.Password = credManager.GetHashedPassword(plainPassword, organization.PasswordSalt)

	// Verify old user password works with org salt
	if !credManager.IsPasswordCorrect(plainPassword, oldUser.Password, organization.PasswordSalt) {
		t.Error("Old user password verification with org salt should succeed")
	}

	// Scenario 2: New user with random salt
	newUser := &User{
		Name:     "new-user",
		Password: plainPassword,
	}

	newUser.UpdateUserPassword(organization)

	// Verify new user has a different salt
	if newUser.PasswordSalt == organization.PasswordSalt {
		t.Error("New user should have a random salt, not org salt")
	}

	// Verify new user password works with user salt
	if !credManager.IsPasswordCorrect(plainPassword, newUser.Password, newUser.PasswordSalt) {
		t.Error("New user password verification with user salt should succeed")
	}

	// Verify new user password doesn't work with org salt
	if credManager.IsPasswordCorrect(plainPassword, newUser.Password, organization.PasswordSalt) {
		t.Error("New user password verification with org salt should fail")
	}
}

// TestPasswordResetUsesRandomSalt tests that resetting a password generates a new random salt
func TestPasswordResetUsesRandomSalt(t *testing.T) {
	organization := &Organization{
		PasswordType: "sha512-salt",
		PasswordSalt: "orgSalt456",
	}

	user := &User{
		Name:     "test-user",
		Password: "oldPassword123",
	}

	// First password set
	user.UpdateUserPassword(organization)
	firstSalt := user.PasswordSalt

	// Password reset
	user.Password = "newPassword456"
	user.UpdateUserPassword(organization)
	secondSalt := user.PasswordSalt

	// Verify that password reset generates a new salt
	if firstSalt == secondSalt {
		t.Error("Password reset should generate a new random salt")
	}

	// Verify both salts are not the organization salt
	if firstSalt == organization.PasswordSalt {
		t.Error("First salt should not be the organization salt")
	}
	if secondSalt == organization.PasswordSalt {
		t.Error("Second salt should not be the organization salt")
	}
}
