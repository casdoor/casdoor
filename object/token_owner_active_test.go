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

func TestTokenIsOwnerActive_WithoutUserLookup(t *testing.T) {
	t.Run("nil token", func(t *testing.T) {
		var token *Token
		ok, err := token.IsOwnerActive()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ok {
			t.Fatal("expected nil token to be inactive")
		}
	})

	t.Run("client_credentials", func(t *testing.T) {
		token := &Token{GrantType: "client_credentials", Organization: "hyperion", User: ""}
		ok, err := token.IsOwnerActive()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !ok {
			t.Fatal("expected client_credentials token to be active")
		}
	})

	t.Run("empty user", func(t *testing.T) {
		token := &Token{GrantType: "authorization_code", Organization: "hyperion", User: ""}
		ok, err := token.IsOwnerActive()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !ok {
			t.Fatal("expected empty user to skip forbidden check")
		}
	})
}
