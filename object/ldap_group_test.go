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
	"testing"
)

func TestGetParentDn(t *testing.T) {
	tests := []struct {
		name     string
		dn       string
		expected string
	}{
		{
			name:     "Simple OU hierarchy",
			dn:       "OU=Sales,OU=Departments,DC=example,DC=com",
			expected: "OU=Departments,DC=example,DC=com",
		},
		{
			name:     "Single component",
			dn:       "DC=example",
			expected: "",
		},
		{
			name:     "Group DN",
			dn:       "CN=Admins,OU=Groups,DC=example,DC=com",
			expected: "OU=Groups,DC=example,DC=com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getParentDn(tt.dn)
			if result != tt.expected {
				t.Errorf("getParentDn(%s) = %s; want %s", tt.dn, result, tt.expected)
			}
		})
	}
}

func TestDnToGroupName(t *testing.T) {
	tests := []struct {
		name     string
		owner    string
		dn       string
		expected string
	}{
		{
			name:     "Simple OU",
			owner:    "org1",
			dn:       "OU=Sales,DC=example,DC=com",
			expected: "Sales",
		},
		{
			name:     "Nested OUs",
			owner:    "org1",
			dn:       "OU=Team1,OU=Sales,OU=Departments,DC=example,DC=com",
			expected: "Departments_Sales_Team1",
		},
		{
			name:     "Group CN",
			owner:    "org1",
			dn:       "CN=Admins,OU=Groups,DC=example,DC=com",
			expected: "Groups_Admins",
		},
		{
			name:     "Empty DN",
			owner:    "org1",
			dn:       "",
			expected: "",
		},
		{
			name:     "Only DC components",
			owner:    "org1",
			dn:       "DC=example,DC=com",
			expected: "",
		},
		{
			name:     "Name with spaces",
			owner:    "org1",
			dn:       "OU=Sales Team,DC=example,DC=com",
			expected: "Sales_Team",
		},
		{
			name:     "Name with special characters",
			owner:    "org1",
			dn:       "OU=Sales & Marketing!,DC=example,DC=com",
			expected: "Sales_Marketing",
		},
		{
			name:     "Name with consecutive spaces/special chars",
			owner:    "org1",
			dn:       "OU=Sales   &   Marketing,DC=example,DC=com",
			expected: "Sales_Marketing",
		},
		{
			name:     "Name with forward slash",
			owner:    "org1",
			dn:       "OU=IT/Support,DC=example,DC=com",
			expected: "IT_Support",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := dnToGroupName(tt.owner, tt.dn)
			if result != tt.expected {
				t.Errorf("dnToGroupName(%s, %s) = %s; want %s", tt.owner, tt.dn, result, tt.expected)
			}
		})
	}
}

func TestParseDnToGroupName(t *testing.T) {
	tests := []struct {
		name     string
		dn       string
		expected string
	}{
		{
			name:     "CN entry",
			dn:       "CN=Admins,OU=Groups,DC=example,DC=com",
			expected: "Admins",
		},
		{
			name:     "OU entry",
			dn:       "OU=Sales,DC=example,DC=com",
			expected: "Sales",
		},
		{
			name:     "Empty DN",
			dn:       "",
			expected: "",
		},
		{
			name:     "No equals sign",
			dn:       "Invalid",
			expected: "Invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseDnToGroupName(tt.dn)
			if result != tt.expected {
				t.Errorf("parseDnToGroupName(%s) = %s; want %s", tt.dn, result, tt.expected)
			}
		})
	}
}
