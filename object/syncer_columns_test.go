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

func TestGetDefaultTableColumns(t *testing.T) {
	tests := []struct {
		name         string
		syncerType   string
		wantNonEmpty bool
		wantKeyField bool
	}{
		{"DingTalk", "DingTalk", true, true},
		{"WeCom", "WeCom", true, true},
		{"Azure AD", "Azure AD", true, true},
		{"Google Workspace", "Google Workspace", true, true},
		{"Active Directory", "Active Directory", true, true},
		{"Database", "Database", false, false},
		{"Keycloak", "Keycloak", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			columns := getDefaultTableColumns(tt.syncerType)

			if tt.wantNonEmpty {
				if len(columns) == 0 {
					t.Errorf("getDefaultTableColumns(%s) returned empty columns, want non-empty", tt.syncerType)
				}

				if tt.wantKeyField {
					hasKey := false
					for _, col := range columns {
						if col.IsKey {
							hasKey = true
							break
						}
					}
					if !hasKey {
						t.Errorf("getDefaultTableColumns(%s) has no key column", tt.syncerType)
					}
				}
			} else {
				if len(columns) != 0 {
					t.Errorf("getDefaultTableColumns(%s) returned %d columns, want 0", tt.syncerType, len(columns))
				}
			}
		})
	}
}

func TestGetDefaultTableColumns_DingTalkStructure(t *testing.T) {
	columns := getDefaultTableColumns("DingTalk")

	// Verify we have expected columns
	expectedColumns := map[string]bool{
		"userid": false,
		"name":   false,
		"email":  false,
		"mobile": false,
	}

	for _, col := range columns {
		if _, ok := expectedColumns[col.Name]; ok {
			expectedColumns[col.Name] = true
		}
	}

	for name, found := range expectedColumns {
		if !found {
			t.Errorf("Expected column '%s' not found in DingTalk default columns", name)
		}
	}

	// Verify userid is the key column
	for _, col := range columns {
		if col.Name == "userid" && !col.IsKey {
			t.Errorf("userid should be marked as IsKey=true")
		}
		if col.Name == "userid" && col.CasdoorName != "Id" {
			t.Errorf("userid should map to CasdoorName 'Id', got '%s'", col.CasdoorName)
		}
	}
}
