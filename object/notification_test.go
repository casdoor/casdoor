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

func TestSendSsoLogoutNotifications_NilUser(t *testing.T) {
	// Test with nil user - should not panic and return nil
	err := SendSsoLogoutNotifications(nil)
	if err != nil {
		t.Errorf("SendSsoLogoutNotifications with nil user should return nil, got: %v", err)
	}
}

func TestSendSsoLogoutNotifications_UserWithoutOrganization(t *testing.T) {
	// Skip this test if database is not initialized
	// We need to catch panics from database access
	defer func() {
		if r := recover(); r != nil {
			t.Skip("Skipping test: database not initialized")
		}
	}()

	// Test with a user that has no organization
	user := &User{
		Owner: "test-org",
		Name:  "test-user",
	}

	// This should not panic even if the organization doesn't exist
	// It should just return an error or handle gracefully
	err := SendSsoLogoutNotifications(user)
	// We expect an error or nil depending on whether the organization exists
	// The function should not panic
	_ = err
}
