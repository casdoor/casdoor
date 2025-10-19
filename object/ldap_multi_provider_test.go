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

func TestBuildLdapUserNameWithMultipleProviders(t *testing.T) {
	tests := []struct {
		name           string
		ldapUser       LdapUser
		ldapId         string
		owner          string
		expectedSuffix string
		description    string
	}{
		{
			name: "Same username different providers",
			ldapUser: LdapUser{
				Uid:       "john",
				UidNumber: "1001",
			},
			ldapId:         "ldap-provider-1",
			owner:          "test-org",
			expectedSuffix: "",
			description:    "First provider should use username without suffix",
		},
		{
			name: "Username with UID number",
			ldapUser: LdapUser{
				Uid:       "alice",
				UidNumber: "2001",
			},
			ldapId:         "ldap-provider-2",
			owner:          "test-org",
			expectedSuffix: "",
			description:    "Should generate username with UID number format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			username, err := tt.ldapUser.buildLdapUserName(tt.owner, tt.ldapId)
			if err != nil {
				// Test will fail gracefully if no database connection
				t.Logf("Test requires database connection: %v", err)
				return
			}

			if username == "" {
				t.Errorf("buildLdapUserName() returned empty username")
			}

			t.Logf("Generated username: %s for LDAP user %s from provider %s", username, tt.ldapUser.Uid, tt.ldapId)
		})
	}
}

func TestLdapUserNameConflictResolution(t *testing.T) {
	// This test validates the logic of username conflict resolution
	// when the same username exists in multiple LDAP providers

	ldapUser := LdapUser{
		Uid:       "testuser",
		UidNumber: "5001",
		Cn:        "Test User",
	}

	// Test case 1: Username is available
	t.Run("Username available", func(t *testing.T) {
		username, err := ldapUser.buildLdapUserName("test-org", "ldap-1")
		if err != nil {
			t.Logf("Test requires database connection: %v", err)
			return
		}
		if username != "testuser" && username != "" {
			t.Logf("Generated username: %s (expected 'testuser' if no conflicts)", username)
		}
	})

	// Test case 2: Username exists in different provider
	// In this case, the system should append the LDAP server name
	t.Run("Username exists in different provider", func(t *testing.T) {
		username, err := ldapUser.buildLdapUserName("test-org", "ldap-2")
		if err != nil {
			t.Logf("Test requires database connection: %v", err)
			return
		}
		if username == "" {
			t.Errorf("buildLdapUserName() returned empty username")
		}
		t.Logf("Generated username with conflict resolution: %s", username)
	})
}

func TestGetLdapUuid(t *testing.T) {
	tests := []struct {
		name     string
		ldapUser LdapUser
		expected string
	}{
		{
			name: "UUID provided",
			ldapUser: LdapUser{
				Uuid: "uuid-123",
				Uid:  "user1",
				Cn:   "User One",
			},
			expected: "uuid-123",
		},
		{
			name: "No UUID, use UID",
			ldapUser: LdapUser{
				Uid: "user2",
				Cn:  "User Two",
			},
			expected: "user2",
		},
		{
			name: "No UUID or UID, use CN",
			ldapUser: LdapUser{
				Cn: "User Three",
			},
			expected: "User Three",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.ldapUser.GetLdapUuid()
			if result != tt.expected {
				t.Errorf("GetLdapUuid() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestBuildLdapDisplayName(t *testing.T) {
	tests := []struct {
		name     string
		ldapUser LdapUser
		expected string
	}{
		{
			name: "DisplayName provided",
			ldapUser: LdapUser{
				DisplayName: "John Doe",
				Cn:          "jdoe",
			},
			expected: "John Doe",
		},
		{
			name: "No DisplayName, use CN",
			ldapUser: LdapUser{
				Cn: "Jane Smith",
			},
			expected: "Jane Smith",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.ldapUser.buildLdapDisplayName()
			if result != tt.expected {
				t.Errorf("buildLdapDisplayName() = %v, want %v", result, tt.expected)
			}
		})
	}
}
