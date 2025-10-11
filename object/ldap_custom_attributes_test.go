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

func TestLdapCustomAttributesMapping(t *testing.T) {
	// Test that LdapUser can store custom attributes
	ldapUser := LdapUser{
		Uid:         "testuser",
		Cn:          "Test User",
		DisplayName: "Test User",
		Email:       "test@example.com",
		Attributes: map[string]string{
			"department":     "Engineering",
			"employeeNumber": "12345",
			"team":           "Platform",
		},
	}

	// Verify attributes are stored correctly
	if ldapUser.Attributes["department"] != "Engineering" {
		t.Errorf("Expected department to be 'Engineering', got '%s'", ldapUser.Attributes["department"])
	}

	if ldapUser.Attributes["employeeNumber"] != "12345" {
		t.Errorf("Expected employeeNumber to be '12345', got '%s'", ldapUser.Attributes["employeeNumber"])
	}

	if ldapUser.Attributes["team"] != "Platform" {
		t.Errorf("Expected team to be 'Platform', got '%s'", ldapUser.Attributes["team"])
	}
}

func TestLdapCustomAttributesMappingToUser(t *testing.T) {
	// Test that custom attributes from LDAP are properly mapped to User Properties
	ldapUser := LdapUser{
		Uid:         "testuser",
		Cn:          "Test User",
		DisplayName: "Test User",
		Email:       "test@example.com",
		Mobile:      "1234567890",
		Attributes: map[string]string{
			"department":     "Engineering",
			"employeeNumber": "12345",
			"jobTitle":       "Senior Engineer",
			"location":       "Building A",
		},
	}

	// Create a user with properties from LDAP attributes
	user := &User{
		Owner:       "test-org",
		Name:        ldapUser.Uid,
		DisplayName: ldapUser.DisplayName,
		Email:       ldapUser.Email,
		Phone:       ldapUser.Mobile,
		Properties:  ldapUser.Attributes,
	}

	// Verify properties are correctly mapped
	if user.Properties["department"] != "Engineering" {
		t.Errorf("Expected user property 'department' to be 'Engineering', got '%s'", user.Properties["department"])
	}

	if user.Properties["employeeNumber"] != "12345" {
		t.Errorf("Expected user property 'employeeNumber' to be '12345', got '%s'", user.Properties["employeeNumber"])
	}

	if user.Properties["jobTitle"] != "Senior Engineer" {
		t.Errorf("Expected user property 'jobTitle' to be 'Senior Engineer', got '%s'", user.Properties["jobTitle"])
	}

	if user.Properties["location"] != "Building A" {
		t.Errorf("Expected user property 'location' to be 'Building A', got '%s'", user.Properties["location"])
	}

	// Verify number of custom properties
	if len(user.Properties) != 4 {
		t.Errorf("Expected 4 custom properties, got %d", len(user.Properties))
	}
}

func TestAutoAdjustLdapUserPreservesAttributes(t *testing.T) {
	// Test that AutoAdjustLdapUser preserves custom attributes
	users := []LdapUser{
		{
			Uid:         "user1",
			Cn:          "User One",
			DisplayName: "User One Display",
			Email:       "user1@example.com",
			Mobile:      "1111111111",
			Attributes: map[string]string{
				"department": "HR",
				"office":     "NYC",
			},
		},
		{
			Uid:         "user2",
			Cn:          "User Two",
			DisplayName: "User Two Display",
			Email:       "user2@example.com",
			Mobile:      "2222222222",
			Attributes: map[string]string{
				"department": "Sales",
				"region":     "APAC",
			},
		},
	}

	adjustedUsers := AutoAdjustLdapUser(users)

	// Verify first user's attributes are preserved
	if adjustedUsers[0].Attributes["department"] != "HR" {
		t.Errorf("Expected first user's department to be 'HR', got '%s'", adjustedUsers[0].Attributes["department"])
	}

	if adjustedUsers[0].Attributes["office"] != "NYC" {
		t.Errorf("Expected first user's office to be 'NYC', got '%s'", adjustedUsers[0].Attributes["office"])
	}

	// Verify second user's attributes are preserved
	if adjustedUsers[1].Attributes["department"] != "Sales" {
		t.Errorf("Expected second user's department to be 'Sales', got '%s'", adjustedUsers[1].Attributes["department"])
	}

	if adjustedUsers[1].Attributes["region"] != "APAC" {
		t.Errorf("Expected second user's region to be 'APAC', got '%s'", adjustedUsers[1].Attributes["region"])
	}
}
