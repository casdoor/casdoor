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
	"encoding/json"
	"testing"
)

func TestLdapCustomAttributesJSON(t *testing.T) {
	// Test that CustomAttributes can be properly marshaled and unmarshaled
	ldap := &Ldap{
		Id:    "test-ldap",
		Owner: "test-org",
		CustomAttributes: map[string]string{
			"department":     "userDepartment",
			"employeeNumber": "employeeId",
			"team":           "userTeam",
		},
	}

	// Marshal to JSON
	data, err := json.Marshal(ldap)
	if err != nil {
		t.Fatalf("Failed to marshal LDAP to JSON: %v", err)
	}

	// Unmarshal from JSON
	var unmarshaled Ldap
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal LDAP from JSON: %v", err)
	}

	// Verify CustomAttributes are preserved
	if len(unmarshaled.CustomAttributes) != 3 {
		t.Errorf("Expected 3 custom attributes, got %d", len(unmarshaled.CustomAttributes))
	}

	if unmarshaled.CustomAttributes["department"] != "userDepartment" {
		t.Errorf("Expected department -> userDepartment, got %s", unmarshaled.CustomAttributes["department"])
	}

	if unmarshaled.CustomAttributes["employeeNumber"] != "employeeId" {
		t.Errorf("Expected employeeNumber -> employeeId, got %s", unmarshaled.CustomAttributes["employeeNumber"])
	}

	if unmarshaled.CustomAttributes["team"] != "userTeam" {
		t.Errorf("Expected team -> userTeam, got %s", unmarshaled.CustomAttributes["team"])
	}
}

func TestLdapEmptyCustomAttributes(t *testing.T) {
	// Test that empty/nil CustomAttributes are handled correctly
	ldap := &Ldap{
		Id:               "test-ldap",
		Owner:            "test-org",
		CustomAttributes: nil,
	}

	data, err := json.Marshal(ldap)
	if err != nil {
		t.Fatalf("Failed to marshal LDAP with nil CustomAttributes: %v", err)
	}

	var unmarshaled Ldap
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal LDAP with nil CustomAttributes: %v", err)
	}

	// nil map should be handled gracefully
	if unmarshaled.CustomAttributes == nil {
		// This is acceptable
	} else if len(unmarshaled.CustomAttributes) != 0 {
		t.Errorf("Expected empty or nil custom attributes, got %d", len(unmarshaled.CustomAttributes))
	}
}
