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

func TestGetPoliciesIncludesGroups(t *testing.T) {
	permission := &Permission{
		Owner:     "org",
		Name:      "perm",
		Users:     []string{"org/alice"},
		Groups:    []string{"org/dev"},
		Roles:     []string{"org/admin"},
		Resources: []string{"data1"},
		Actions:   []string{"read"},
		Effect:    "Allow",
	}

	got := getPolicies(permission)
	want := [][]string{
		{"org/alice", "data1", "read", "allow", "", "org/perm"},
		{"group:org/dev", "data1", "read", "allow", "", "org/perm"},
		{"org/admin", "data1", "read", "allow", "", "org/perm"},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("getPolicies() = %#v, want %#v", got, want)
	}
}

func TestGetPoliciesIncludesGroupsWithDomains(t *testing.T) {
	permission := &Permission{
		Owner:     "org",
		Name:      "perm",
		Groups:    []string{"org/dev"},
		Domains:   []string{"domain1"},
		Resources: []string{"data1"},
		Actions:   []string{"read"},
		Effect:    "Deny",
	}

	got := getPolicies(permission)
	want := [][]string{
		{"group:org/dev", "domain1", "data1", "read", "deny", "org/perm"},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("getPolicies() = %#v, want %#v", got, want)
	}
}
