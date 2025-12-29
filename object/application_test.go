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
)

func TestExtendApplicationWithSignupItems(t *testing.T) {
	// Test that signup items are initialized when empty
	app := &Application{}
	err := extendApplicationWithSignupItems(app)
	if err != nil {
		t.Errorf("extendApplicationWithSignupItems() error = %v", err)
		return
	}

	if len(app.SignupItems) == 0 {
		t.Error("SignupItems should be initialized but is empty")
		return
	}

	// Check that all default items are present
	expectedItems := []string{"ID", "Username", "Display name", "Password", "Confirm password", "Email", "Phone", "Agreement"}
	if len(app.SignupItems) != len(expectedItems) {
		t.Errorf("Expected %d signup items, got %d", len(expectedItems), len(app.SignupItems))
	}

	for i, item := range app.SignupItems {
		if item.Name != expectedItems[i] {
			t.Errorf("Expected item name %s at position %d, got %s", expectedItems[i], i, item.Name)
		}
	}
}

func TestExtendApplicationWithSignupItems_NotEmpty(t *testing.T) {
	// Test that existing signup items are not overwritten
	existingItem := &SignupItem{Name: "Custom", Visible: true, Required: false}
	app := &Application{
		SignupItems: []*SignupItem{existingItem},
	}

	err := extendApplicationWithSignupItems(app)
	if err != nil {
		t.Errorf("extendApplicationWithSignupItems() error = %v", err)
		return
	}

	if len(app.SignupItems) != 1 {
		t.Errorf("Expected 1 signup item (should not be overwritten), got %d", len(app.SignupItems))
	}

	if app.SignupItems[0].Name != "Custom" {
		t.Errorf("Expected custom item to be preserved, got %s", app.SignupItems[0].Name)
	}
}

func TestExtendApplicationWithSigninItems(t *testing.T) {
	// Test that signin items are initialized when empty
	app := &Application{}
	err := extendApplicationWithSigninItems(app)
	if err != nil {
		t.Errorf("extendApplicationWithSigninItems() error = %v", err)
		return
	}

	if len(app.SigninItems) == 0 {
		t.Error("SigninItems should be initialized but is empty")
		return
	}

	// Check that default items are present
	expectedItems := []string{"Back button", "Languages", "Logo", "Signin methods", "Username", "Password", "Verification code", "Agreement", "Forgot password?", "Login button", "Signup link", "Providers"}
	if len(app.SigninItems) != len(expectedItems) {
		t.Errorf("Expected %d signin items, got %d", len(expectedItems), len(app.SigninItems))
	}
}

func TestExtendApplicationWithSigninMethods(t *testing.T) {
	// Test that signin methods are initialized when empty
	app := &Application{
		EnablePassword: true, // Enable password signin method
	}
	err := extendApplicationWithSigninMethods(app)
	if err != nil {
		t.Errorf("extendApplicationWithSigninMethods() error = %v", err)
		return
	}

	if len(app.SigninMethods) == 0 {
		t.Error("SigninMethods should be initialized but is empty")
		return
	}

	// Should have at least Password method when EnablePassword is true
	hasPassword := false
	for _, method := range app.SigninMethods {
		if method.Name == "Password" {
			hasPassword = true
			break
		}
	}

	if !hasPassword {
		t.Error("Expected default signin methods to include Password when EnablePassword is true")
	}
}

func TestExtendApplicationWithSigninMethods_EmptyDefaults(t *testing.T) {
	// Test that signin methods get default Password method even when no flags are enabled
	app := &Application{}
	err := extendApplicationWithSigninMethods(app)
	if err != nil {
		t.Errorf("extendApplicationWithSigninMethods() error = %v", err)
		return
	}

	// The function adds a default Password method even if SigninMethods is still empty
	if len(app.SigninMethods) == 0 {
		t.Error("SigninMethods should have at least a default Password method")
		return
	}
}
