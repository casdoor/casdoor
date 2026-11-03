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

	"github.com/golang-jwt/jwt/v5"
)

func TestGetTokenHost(t *testing.T) {
	sign := func(issuer string) string {
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{Issuer: issuer})
		signed, err := token.SignedString([]byte("test-key"))
		if err != nil {
			t.Fatal(err)
		}
		return signed
	}

	tests := []struct {
		name        string
		accessToken string
		want        string
	}{
		{"https issuer", sign("https://door.casdoor.com"), "door.casdoor.com"},
		{"host with port", sign("http://casdoor:8000"), "casdoor:8000"},
		{"issuer with a path", sign("https://door.casdoor.com/.well-known/org"), "door.casdoor.com"},
		{"no issuer", sign(""), ""},
		{"not a JWT", "opaque-token", ""},
		{"no token", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getTokenHost(&Token{AccessToken: tt.accessToken})
			if got != tt.want {
				t.Errorf("getTokenHost() = %q, want %q", got, tt.want)
			}
		})
	}
}
