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

	"github.com/stretchr/testify/assert"
)

func TestParseGroupNameFromDN(t *testing.T) {
	testCases := []struct {
		name     string
		dn       string
		expected string
	}{
		{
			name:     "Standard Active Directory DN",
			dn:       "CN=Domain Admins,OU=Groups,DC=example,DC=com",
			expected: "Domain Admins",
		},
		{
			name:     "OpenLDAP DN",
			dn:       "cn=developers,ou=groups,dc=example,dc=org",
			expected: "developers",
		},
		{
			name:     "DN with spaces",
			dn:       "CN=Project Managers,OU=Marketing,DC=company,DC=net",
			expected: "Project Managers",
		},
		{
			name:     "DN with special characters",
			dn:       "CN=IT-Support-Team,OU=IT,DC=domain,DC=local",
			expected: "IT-Support-Team",
		},
		{
			name:     "Empty DN",
			dn:       "",
			expected: "",
		},
		{
			name:     "DN without CN",
			dn:       "OU=Groups,DC=example,DC=com",
			expected: "",
		},
		{
			name:     "DN with extra spaces",
			dn:       "CN=Engineers,OU=Tech,DC=example,DC=com",
			expected: "Engineers",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := parseGroupNameFromDN(tc.dn)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestExtractGroupNamesFromMemberOf(t *testing.T) {
	testCases := []struct {
		name      string
		memberOfs []string
		expected  []string
	}{
		{
			name: "Multiple groups from Active Directory",
			memberOfs: []string{
				"CN=Domain Admins,OU=Groups,DC=example,DC=com",
				"CN=IT Support,OU=Groups,DC=example,DC=com",
				"CN=Developers,OU=Engineering,DC=example,DC=com",
			},
			expected: []string{"Domain Admins", "IT Support", "Developers"},
		},
		{
			name: "Single group from OpenLDAP",
			memberOfs: []string{
				"cn=developers,ou=groups,dc=example,dc=org",
			},
			expected: []string{"developers"},
		},
		{
			name:      "Empty memberOf list",
			memberOfs: []string{},
			expected:  nil,
		},
		{
			name: "Mixed valid and invalid DNs",
			memberOfs: []string{
				"CN=Valid Group,OU=Groups,DC=example,DC=com",
				"OU=InvalidDN,DC=example,DC=com",
				"CN=Another Valid,OU=Teams,DC=example,DC=com",
			},
			expected: []string{"Valid Group", "Another Valid"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := extractGroupNamesFromMemberOf(tc.memberOfs)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestAutoAdjustLdapUser_WithMemberOf(t *testing.T) {
	// Test that AutoAdjustLdapUser properly preserves MemberOfs
	users := []LdapUser{
		{
			Uid:         "testuser",
			Cn:          "Test User",
			DisplayName: "Test User",
			Email:       "test@example.com",
			MemberOf:    "CN=Admins,OU=Groups,DC=example,DC=com",
			MemberOfs: []string{
				"CN=Admins,OU=Groups,DC=example,DC=com",
				"CN=Developers,OU=Groups,DC=example,DC=com",
			},
		},
	}

	result := AutoAdjustLdapUser(users)

	assert.Len(t, result, 1)
	assert.Equal(t, "CN=Admins,OU=Groups,DC=example,DC=com", result[0].MemberOf)
	assert.Len(t, result[0].MemberOfs, 2)
	assert.Equal(t, "CN=Admins,OU=Groups,DC=example,DC=com", result[0].MemberOfs[0])
	assert.Equal(t, "CN=Developers,OU=Groups,DC=example,DC=com", result[0].MemberOfs[1])
}
