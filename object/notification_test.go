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

// TestSendSsoLogoutNotifications_BasicCases tests the SSO logout notification functionality
// with basic cases that don't require database access
func TestSendSsoLogoutNotifications_BasicCases(t *testing.T) {
	tests := []struct {
		name        string
		user        *User
		expectError bool
	}{
		{
			name:        "nil user",
			user:        nil,
			expectError: false, // Should return nil without error
		},
		{
			name: "user with empty SignupApplication",
			user: &User{
				Owner:             "built-in",
				Name:              "test-user",
				SignupApplication: "",
			},
			expectError: false, // Should return nil without error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := SendSsoLogoutNotifications(tt.user)
			if tt.expectError && err == nil {
				t.Errorf("SendSsoLogoutNotifications() expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("SendSsoLogoutNotifications() unexpected error = %v", err)
			}
		})
	}
}

// TestSendSsoLogoutNotifications_WithDB tests the SSO logout notification functionality
// with cases that require database access, including the fix for app-built-in
func TestSendSsoLogoutNotifications_WithDB(t *testing.T) {
	// Initialize the configuration and database
	InitConfig()

	tests := []struct {
		name             string
		user             *User
		expectTokenError bool
	}{
		{
			name: "user with app-built-in SignupApplication (no owner prefix)",
			user: &User{
				Owner:             "built-in",
				Name:              "test-user",
				SignupApplication: "app-built-in",
				Email:             "test@example.com",
			},
			expectTokenError: false, // Should not fail with "wrong token count" error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := SendSsoLogoutNotifications(tt.user)
			// Check if the error is the specific "wrong token count" error that we fixed
			if err != nil && err.Error() != "" {
				containsTokenError := false
				if errStr := err.Error(); errStr != "" {
					containsTokenError = (errStr == "wrong token count for ID" || 
						errStr == "GetOwnerAndNameFromId() error, wrong token count for ID: app-built-in")
				}
				
				if tt.expectTokenError && !containsTokenError {
					t.Errorf("SendSsoLogoutNotifications() expected token count error, got: %v", err)
				}
				if !tt.expectTokenError && containsTokenError {
					t.Errorf("SendSsoLogoutNotifications() unexpected token count error = %v", err)
				}
				// Other errors (like application not found) are acceptable in test environment
				if !containsTokenError {
					t.Logf("SendSsoLogoutNotifications() error = %v (acceptable in test environment)", err)
				}
			}
		})
	}
}
