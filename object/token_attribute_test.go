// Copyright 2026 The Casdoor Authors. All Rights Reserved.
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
	"reflect"
	"testing"
)

func TestReplaceAttributeValue(t *testing.T) {
	// Create a test user with roles
	user := &User{
		Owner: "test-org",
		Name:  "test-user",
		Email: "test@example.com",
		Id:    "test-id-123",
		Phone: "+1234567890",
		Roles: []*Role{
			{Name: "admin"},
			{Name: "developer"},
			{Name: "viewer"},
		},
		Permissions: []*Permission{
			{Name: "read"},
			{Name: "write"},
			{Name: "delete"},
		},
		Groups: []string{"engineering", "product"},
	}

	tests := []struct {
		name     string
		value    string
		expected []string
	}{
		{
			name:     "Replace $user.roles",
			value:    "$user.roles",
			expected: []string{"admin", "developer", "viewer"},
		},
		{
			name:     "Replace $user.permissions",
			value:    "$user.permissions",
			expected: []string{"read", "write", "delete"},
		},
		{
			name:     "Replace $user.groups",
			value:    "$user.groups",
			expected: []string{"engineering", "product"},
		},
		{
			name:     "Replace $user.owner",
			value:    "$user.owner",
			expected: []string{"test-org"},
		},
		{
			name:     "Replace $user.name",
			value:    "$user.name",
			expected: []string{"test-user"},
		},
		{
			name:     "Replace $user.email",
			value:    "$user.email",
			expected: []string{"test@example.com"},
		},
		{
			name:     "Replace $user.id",
			value:    "$user.id",
			expected: []string{"test-id-123"},
		},
		{
			name:     "Replace $user.phone",
			value:    "$user.phone",
			expected: []string{"+1234567890"},
		},
		{
			name:     "Multiple replacements in template",
			value:    "User $user.name has email $user.email",
			expected: []string{"User test-user has email test@example.com"},
		},
		{
			name:     "Static value (no replacement)",
			value:    "static-value",
			expected: []string{"static-value"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := replaceAttributeValue(user, tt.value)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("replaceAttributeValue() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetClaimsCustomWithTokenAttributes(t *testing.T) {
	// Create a test user with roles
	user := &User{
		Owner: "test-org",
		Name:  "test-user",
		Email: "test@example.com",
		Roles: []*Role{
			{Name: "admin"},
			{Name: "developer"},
		},
		Permissions: []*Permission{
			{Name: "read"},
			{Name: "write"},
		},
		Groups: []string{"engineering"},
	}

	claims := Claims{
		User: user,
	}

	// Test the warpgate_roles use case
	tokenAttributes := []*JwtItem{
		{
			Name:  "warpgate_roles",
			Value: "$user.roles",
			Type:  "Array",
		},
		{
			Name:  "custom_email",
			Value: "$user.email",
			Type:  "String",
		},
		{
			Name:  "user_groups",
			Value: "$user.groups",
			Type:  "Array",
		},
	}

	result := getClaimsCustom(claims, []string{}, tokenAttributes)

	// Check warpgate_roles is an array
	if rolesValue, ok := result["warpgate_roles"]; ok {
		roles, ok := rolesValue.([]string)
		if !ok {
			t.Errorf("warpgate_roles should be []string, got %T", rolesValue)
		}
		expectedRoles := []string{"admin", "developer"}
		if !reflect.DeepEqual(roles, expectedRoles) {
			t.Errorf("warpgate_roles = %v, want %v", roles, expectedRoles)
		}
	} else {
		t.Error("warpgate_roles not found in claims")
	}

	// Check custom_email is a string (first element of array)
	if emailValue, ok := result["custom_email"]; ok {
		email, ok := emailValue.(string)
		if !ok {
			t.Errorf("custom_email should be string, got %T", emailValue)
		}
		if email != "test@example.com" {
			t.Errorf("custom_email = %v, want %v", email, "test@example.com")
		}
	} else {
		t.Error("custom_email not found in claims")
	}

	// Check user_groups is an array
	if groupsValue, ok := result["user_groups"]; ok {
		groups, ok := groupsValue.([]string)
		if !ok {
			t.Errorf("user_groups should be []string, got %T", groupsValue)
		}
		expectedGroups := []string{"engineering"}
		if !reflect.DeepEqual(groups, expectedGroups) {
			t.Errorf("user_groups = %v, want %v", groups, expectedGroups)
		}
	} else {
		t.Error("user_groups not found in claims")
	}
}

func TestGetClaimsCustomWithEmptyUser(t *testing.T) {
	// Test with nil user
	nilResult := replaceAttributeValue(nil, "$user.roles")
	if nilResult != nil {
		t.Errorf("replaceAttributeValue with nil user should return nil, got %v", nilResult)
	}

	// Test with user without roles - should return empty slice
	userWithoutRoles := &User{
		Name: "test-user",
	}
	result := replaceAttributeValue(userWithoutRoles, "$user.roles")
	// When user has no roles, getUserRoleNames returns an empty slice
	if len(result) != 0 {
		t.Errorf("replaceAttributeValue with empty roles should return empty array, got %v with length %d", result, len(result))
	}
}
