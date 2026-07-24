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

package util

import (
	"reflect"
	"testing"
)

// DeleteSessionId uses DeleteVal to drop a Beego session id from Session.SessionId.
// If the id passed after SessionRegenerateID() does not match the stored one, the
// array stays non-empty and the Session row is incorrectly kept in the admin UI.
func TestDeleteValSessionIdLogout(t *testing.T) {
	stored := []string{"beego-session-before-logout"}

	wrongIdAfterRegenerate := DeleteVal(stored, "beego-session-after-regenerate")
	if len(wrongIdAfterRegenerate) != 1 {
		t.Fatalf("deleting regenerated id should leave stored session, got %#v", wrongIdAfterRegenerate)
	}

	correctId := DeleteVal(stored, "beego-session-before-logout")
	if len(correctId) != 0 {
		t.Fatalf("deleting the stored session id should empty the list, got %#v", correctId)
	}
}

func TestDeleteVal(t *testing.T) {
	tests := []struct {
		name   string
		values []string
		val    string
		want   []string
	}{
		{"empty", nil, "a", []string{}},
		{"remove one", []string{"a", "b", "a"}, "a", []string{"b"}},
		{"missing", []string{"a", "b"}, "c", []string{"a", "b"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DeleteVal(tt.values, tt.val)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("DeleteVal(%v, %q) = %v, want %v", tt.values, tt.val, got, tt.want)
			}
		})
	}
}
