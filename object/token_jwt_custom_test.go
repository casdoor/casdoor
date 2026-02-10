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
	"reflect"
	"testing"
)

func TestGetUserFieldValue(t *testing.T) {
	// Create a test user
	user := &User{
		Owner:       "test-org",
		Name:        "test-user",
		Id:          "test-id",
		DisplayName: "Test User",
		Email:       "test@example.com",
		Phone:       "+1234567890",
		Tag:         "test-tag",
		Groups:      []string{"group1", "group2"},
		Properties:  map[string]string{"custom_field": "custom_value"},
		Roles: []*Role{
			{Name: "admin"},
			{Name: "user"},
		},
		Permissions: []*Permission{
			{Name: "read"},
			{Name: "write"},
		},
	}

	tests := []struct {
		name      string
		fieldName string
		wantValue interface{}
		wantFound bool
	}{
		{
			name:      "Get Owner field",
			fieldName: "Owner",
			wantValue: "test-org",
			wantFound: true,
		},
		{
			name:      "Get Name field",
			fieldName: "Name",
			wantValue: "test-user",
			wantFound: true,
		},
		{
			name:      "Get DisplayName field",
			fieldName: "DisplayName",
			wantValue: "Test User",
			wantFound: true,
		},
		{
			name:      "Get Email field",
			fieldName: "Email",
			wantValue: "test@example.com",
			wantFound: true,
		},
		{
			name:      "Get Roles field",
			fieldName: "Roles",
			wantValue: []string{"admin", "user"},
			wantFound: true,
		},
		{
			name:      "Get Permissions field",
			fieldName: "Permissions",
			wantValue: []string{"read", "write"},
			wantFound: true,
		},
		{
			name:      "Get permissionNames field",
			fieldName: "permissionNames",
			wantValue: []string{"read", "write"},
			wantFound: true,
		},
		{
			name:      "Get Properties field",
			fieldName: "Properties.custom_field",
			wantValue: "custom_value",
			wantFound: true,
		},
		{
			name:      "Get non-existent field",
			fieldName: "NonExistentField",
			wantValue: nil,
			wantFound: false,
		},
		{
			name:      "Get non-existent property",
			fieldName: "Properties.non_existent",
			wantValue: nil,
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotValue, gotFound := getUserFieldValue(user, tt.fieldName)
			if gotFound != tt.wantFound {
				t.Errorf("getUserFieldValue() gotFound = %v, want %v", gotFound, tt.wantFound)
				return
			}
			if tt.wantFound && !reflect.DeepEqual(gotValue, tt.wantValue) {
				t.Errorf("getUserFieldValue() gotValue = %v, want %v", gotValue, tt.wantValue)
			}
		})
	}
}

func TestGetClaimsCustomWithExistingField(t *testing.T) {
	// Create test user
	user := &User{
		Owner:       "test-org",
		Name:        "test-user",
		Id:          "test-id",
		DisplayName: "Test User",
		Email:       "test@example.com",
		Roles: []*Role{
			{Name: "admin"},
			{Name: "user"},
		},
	}

	// Create test claims
	claims := Claims{
		User: user,
	}

	// Test with "Existing Field" category
	tokenAttributes := []*JwtItem{
		{
			Name:     "warpgate_roles",
			Value:    "Roles",
			Type:     "Array",
			Category: "Existing Field",
		},
		{
			Name:     "user_email",
			Value:    "Email",
			Type:     "String",
			Category: "Existing Field",
		},
		{
			Name:     "static_value",
			Value:    "test-value",
			Type:     "String",
			Category: "Static Value",
		},
	}

	result := getClaimsCustom(claims, []string{}, tokenAttributes)

	// Check warpgate_roles
	if roles, ok := result["warpgate_roles"]; ok {
		rolesSlice, ok := roles.([]string)
		if !ok {
			t.Errorf("warpgate_roles should be []string, got %T", roles)
		} else if len(rolesSlice) != 2 || rolesSlice[0] != "admin" || rolesSlice[1] != "user" {
			t.Errorf("warpgate_roles = %v, want [admin user]", rolesSlice)
		}
	} else {
		t.Error("warpgate_roles not found in claims")
	}

	// Check user_email
	if email, ok := result["user_email"]; ok {
		if email != "test@example.com" {
			t.Errorf("user_email = %v, want test@example.com", email)
		}
	} else {
		t.Error("user_email not found in claims")
	}

	// Check static_value
	if staticValue, ok := result["static_value"]; ok {
		if staticValue != "test-value" {
			t.Errorf("static_value = %v, want test-value", staticValue)
		}
	} else {
		t.Error("static_value not found in claims")
	}
}
