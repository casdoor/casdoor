// Copyright 2024 The Casdoor Authors. All Rights Reserved.
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

func TestGetTokensBySessionIds(t *testing.T) {
	// Test with empty session IDs
	tokens, err := GetTokensBySessionIds([]string{})
	if err != nil {
		t.Errorf("GetTokensBySessionIds with empty array should not error: %v", err)
	}
	if len(tokens) != 0 {
		t.Errorf("GetTokensBySessionIds with empty array should return empty slice, got %d tokens", len(tokens))
	}

	// Test with nil session IDs
	tokens, err = GetTokensBySessionIds(nil)
	if err != nil {
		t.Errorf("GetTokensBySessionIds with nil should not error: %v", err)
	}
	if len(tokens) != 0 {
		t.Errorf("GetTokensBySessionIds with nil should return empty slice, got %d tokens", len(tokens))
	}
}

func TestGetTokensBySessionId(t *testing.T) {
	// Test with empty session ID
	tokens, err := GetTokensBySessionId("")
	if err != nil {
		t.Errorf("GetTokensBySessionId with empty string should not error: %v", err)
	}
	if len(tokens) != 0 {
		t.Errorf("GetTokensBySessionId with empty string should return empty slice, got %d tokens", len(tokens))
	}
}
