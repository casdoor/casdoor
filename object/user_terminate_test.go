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

import "testing"

func TestShouldTerminateUserAccess(t *testing.T) {
	active := &User{IsForbidden: false, IsDeleted: false}
	forbidden := &User{IsForbidden: true, IsDeleted: false}
	deleted := &User{IsForbidden: false, IsDeleted: true}

	tests := []struct {
		name    string
		oldUser *User
		newUser *User
		columns []string
		want    bool
	}{
		{
			name:    "forbid with column",
			oldUser: active,
			newUser: forbidden,
			columns: []string{"is_forbidden"},
			want:    true,
		},
		{
			name:    "forbid without column",
			oldUser: active,
			newUser: forbidden,
			columns: []string{"display_name"},
			want:    false,
		},
		{
			name:    "forbid full update",
			oldUser: active,
			newUser: forbidden,
			columns: nil,
			want:    true,
		},
		{
			name:    "already forbidden",
			oldUser: forbidden,
			newUser: forbidden,
			columns: []string{"is_forbidden"},
			want:    false,
		},
		{
			name:    "unforbid",
			oldUser: forbidden,
			newUser: active,
			columns: []string{"is_forbidden"},
			want:    false,
		},
		{
			name:    "soft delete with column",
			oldUser: active,
			newUser: deleted,
			columns: []string{"is_deleted"},
			want:    true,
		},
		{
			name:    "nil users",
			oldUser: nil,
			newUser: forbidden,
			columns: []string{"is_forbidden"},
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldTerminateUserAccess(tt.oldUser, tt.newUser, tt.columns)
			if got != tt.want {
				t.Fatalf("shouldTerminateUserAccess() = %v, want %v", got, tt.want)
			}
		})
	}
}
