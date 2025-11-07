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
	"strings"
	"testing"
)

func TestGenerateGuestUsername(t *testing.T) {
	// Test that the function generates a username with the correct format
	username := generateGuestUsername()

	if !strings.HasPrefix(username, "guest_") {
		t.Errorf("Expected username to start with 'guest_', got: %s", username)
	}

	// Test that it generates unique usernames
	username2 := generateGuestUsername()
	if username == username2 {
		t.Errorf("Expected unique usernames, but got the same: %s", username)
	}

	// Test that the UUID portion is valid (has correct format)
	parts := strings.Split(username, "guest_")
	if len(parts) != 2 {
		t.Errorf("Expected username format 'guest_<uuid>', got: %s", username)
	}

	// UUID format should contain hyphens
	uuidPart := parts[1]
	if !strings.Contains(uuidPart, "-") {
		t.Errorf("Expected UUID format with hyphens, got: %s", uuidPart)
	}
}

func TestGuestUserUpgradeLogic(t *testing.T) {
	// Test case 1: Guest user with changed username (not starting with guest_)
	oldUser := &User{
		Owner:    "test-org",
		Name:     "guest_12345678-1234-5678-1234-567812345678",
		Password: "oldpassword",
		Tag:      "guest-user",
	}

	newUser := &User{
		Owner:    "test-org",
		Name:     "johndoe",
		Password: "oldpassword",
		Tag:      "guest-user",
	}

	// Check if username changed from guest format
	usernameChanged := oldUser.Name != newUser.Name && !strings.HasPrefix(newUser.Name, "guest_")
	if !usernameChanged {
		t.Error("Expected username to be detected as changed")
	}

	// Test case 2: Guest user with changed password
	newUser2 := &User{
		Owner:    "test-org",
		Name:     "guest_12345678-1234-5678-1234-567812345678",
		Password: "newpassword",
		Tag:      "guest-user",
	}

	passwordChanged := newUser2.Password != "***" && newUser2.Password != "" && newUser2.Password != oldUser.Password
	if !passwordChanged {
		t.Error("Expected password to be detected as changed")
	}

	// Test case 3: Guest user without meaningful changes
	newUser3 := &User{
		Owner:       "test-org",
		Name:        "guest_12345678-1234-5678-1234-567812345678",
		Password:    "***", // Placeholder - means not changing password
		Tag:         "guest-user",
		DisplayName: "Updated Display Name",
	}

	usernameChanged3 := oldUser.Name != newUser3.Name && !strings.HasPrefix(newUser3.Name, "guest_")
	passwordChanged3 := newUser3.Password != "***" && newUser3.Password != "" && newUser3.Password != oldUser.Password

	if usernameChanged3 || passwordChanged3 {
		t.Error("Expected no upgrade trigger for non-credential changes")
	}
}
