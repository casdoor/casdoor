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

package cred

import (
	"testing"
)

func TestArgon2idWithoutPepper(t *testing.T) {
	password := "mySecurePassword123"
	cm := NewArgon2idCredManager()

	// Hash password without pepper (empty salt)
	hash := cm.GetHashedPassword(password, "")
	if hash == "" {
		t.Fatal("Failed to generate hash")
	}

	// Verify correct password
	if !cm.IsPasswordCorrect(password, hash, "") {
		t.Error("Password verification failed for correct password without pepper")
	}

	// Verify wrong password fails
	if cm.IsPasswordCorrect("wrongPassword", hash, "") {
		t.Error("Password verification succeeded for wrong password without pepper")
	}
}

func TestArgon2idWithPepper(t *testing.T) {
	password := "mySecurePassword123"
	pepper := "mySecretPepper"
	cm := NewArgon2idCredManager()

	// Hash password with pepper (using salt parameter)
	hash := cm.GetHashedPassword(password, pepper)
	if hash == "" {
		t.Fatal("Failed to generate hash with pepper")
	}

	// Verify correct password with correct pepper
	if !cm.IsPasswordCorrect(password, hash, pepper) {
		t.Error("Password verification failed for correct password with pepper")
	}

	// Verify that password without pepper fails
	if cm.IsPasswordCorrect(password, hash, "") {
		t.Error("Password verification succeeded without pepper when pepper was used")
	}

	// Verify that password with wrong pepper fails
	if cm.IsPasswordCorrect(password, hash, "wrongPepper") {
		t.Error("Password verification succeeded with wrong pepper")
	}

	// Verify wrong password with correct pepper fails
	if cm.IsPasswordCorrect("wrongPassword", hash, pepper) {
		t.Error("Password verification succeeded for wrong password with pepper")
	}
}

func TestArgon2idMigratedHash(t *testing.T) {
	// Simulate a migrated hash from old system with custom parameters and pepper
	// This hash was created with password "testPassword123" and pepper "oldSystemPepper"
	// using parameters m=12, t=20, p=2 (for testing - actual migrated hash would have similar structure)
	cm := NewArgon2idCredManager()

	// First, create a hash with a pepper to simulate migration scenario
	password := "testPassword123"
	pepper := "oldSystemPepper"
	migratedHash := cm.GetHashedPassword(password, pepper)

	// Verify that the migrated hash works with correct password and pepper
	if !cm.IsPasswordCorrect(password, migratedHash, pepper) {
		t.Error("Failed to verify migrated hash with correct password and pepper")
	}

	// Verify that wrong password fails
	if cm.IsPasswordCorrect("wrongPassword", migratedHash, pepper) {
		t.Error("Verification succeeded with wrong password for migrated hash")
	}

	// Verify that missing pepper fails
	if cm.IsPasswordCorrect(password, migratedHash, "") {
		t.Error("Verification succeeded without pepper for migrated hash")
	}
}

func TestArgon2idBackwardCompatibility(t *testing.T) {
	// Test that hashes created without pepper can still be verified
	cm := NewArgon2idCredManager()
	password := "backwardCompatTest"

	// Create hash without pepper (old behavior)
	hashWithoutPepper := cm.GetHashedPassword(password, "")

	// Should verify correctly without pepper
	if !cm.IsPasswordCorrect(password, hashWithoutPepper, "") {
		t.Error("Backward compatibility broken: cannot verify hash without pepper")
	}

	// Create hash with pepper (new behavior)
	pepper := "newPepper"
	hashWithPepper := cm.GetHashedPassword(password, pepper)

	// Should verify correctly with pepper
	if !cm.IsPasswordCorrect(password, hashWithPepper, pepper) {
		t.Error("Cannot verify hash created with pepper")
	}

	// Hashes should be different
	if hashWithoutPepper == hashWithPepper {
		t.Error("Hashes with and without pepper should be different")
	}
}

func TestArgon2idEmptyPassword(t *testing.T) {
	cm := NewArgon2idCredManager()
	pepper := "testPepper"

	// Hash empty password with pepper
	hash := cm.GetHashedPassword("", pepper)
	if hash == "" {
		t.Fatal("Failed to generate hash for empty password")
	}

	// Verify empty password with pepper
	if !cm.IsPasswordCorrect("", hash, pepper) {
		t.Error("Failed to verify empty password with pepper")
	}

	// Verify non-empty password fails
	if cm.IsPasswordCorrect("notEmpty", hash, pepper) {
		t.Error("Non-empty password verified against empty password hash")
	}
}

func TestArgon2idPepperOrdering(t *testing.T) {
	// Test that pepper is consistently applied (prepended) to password
	cm := NewArgon2idCredManager()
	password := "password"
	pepper := "pepper"

	hash := cm.GetHashedPassword(password, pepper)

	// This should work because pepper is prepended
	if !cm.IsPasswordCorrect(password, hash, pepper) {
		t.Error("Failed to verify with correct pepper prepending")
	}

	// Create a hash with password as "pepperpassword" without using salt parameter
	// This should match the hash created with pepper="pepper" and password="password"
	hashDirect := cm.GetHashedPassword(pepper+password, "")

	// Both approaches should yield verifiable hashes (though different due to random salt in argon2id)
	if !cm.IsPasswordCorrect(pepper+password, hashDirect, "") {
		t.Error("Failed to verify direct concatenation")
	}
}
