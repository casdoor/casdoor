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
)

func TestUpdateUserPasswordRandomSalt(t *testing.T) {
	organization := &Organization{
		PasswordType: "sha512-salt",
		PasswordSalt: "546435464ad566", // Fixed organization salt
	}

	// Create two users with the same password
	user1 := &User{
		Name:     "test-user-1",
		Password: "testPassword123",
	}

	user2 := &User{
		Name:     "test-user-2",
		Password: "testPassword123",
	}

	// Update passwords for both users
	user1.UpdateUserPassword(organization)
	user2.UpdateUserPassword(organization)

	// Verify that each user has a unique salt
	if user1.PasswordSalt == user2.PasswordSalt {
		t.Errorf("Expected different salts for different users, but both have: %s", user1.PasswordSalt)
	}

	// Verify that salts are not empty
	if user1.PasswordSalt == "" {
		t.Error("Expected user1 to have a non-empty salt")
	}
	if user2.PasswordSalt == "" {
		t.Error("Expected user2 to have a non-empty salt")
	}

	// Verify that salts are not the organization salt
	if user1.PasswordSalt == organization.PasswordSalt {
		t.Error("Expected user1 salt to be different from organization salt")
	}
	if user2.PasswordSalt == organization.PasswordSalt {
		t.Error("Expected user2 salt to be different from organization salt")
	}

	// Verify that hashed passwords are different (since salts are different)
	if user1.Password == user2.Password {
		t.Error("Expected different password hashes due to different salts")
	}

	// Verify that password type is set
	if user1.PasswordType != organization.PasswordType {
		t.Errorf("Expected password type to be %s, got %s", organization.PasswordType, user1.PasswordType)
	}
	if user2.PasswordType != organization.PasswordType {
		t.Errorf("Expected password type to be %s, got %s", organization.PasswordType, user2.PasswordType)
	}
}

func TestUpdateUserPasswordWithDifferentTypes(t *testing.T) {
	testCases := []string{
		"plain",
		"md5-salt",
		"salt", // This is sha256-salt
		"sha512-salt",
		"bcrypt",
		"pbkdf2-salt",
	}

	for _, passwordType := range testCases {
		t.Run(passwordType, func(t *testing.T) {
			organization := &Organization{
				PasswordType: passwordType,
				PasswordSalt: "orgSalt123",
			}

			user := &User{
				Name:     "test-user",
				Password: "testPassword123",
			}

			user.UpdateUserPassword(organization)

			// Verify salt is generated and different from org salt
			if user.PasswordSalt == "" {
				t.Errorf("Expected non-empty salt for password type %s", passwordType)
			}

			// For password types that use salt, verify it's not the org salt
			if passwordType != "plain" && user.PasswordSalt == organization.PasswordSalt {
				t.Errorf("Expected user salt to be different from org salt for password type %s", passwordType)
			}

			// Verify password type is set
			if user.PasswordType != organization.PasswordType {
				t.Errorf("Expected password type to be %s, got %s", organization.PasswordType, user.PasswordType)
			}
		})
	}
}
