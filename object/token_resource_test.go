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

func TestValidateResourceURI(t *testing.T) {
	tests := []struct {
		name     string
		resource string
		wantErr  bool
	}{
		{
			name:     "empty resource is valid",
			resource: "",
			wantErr:  false,
		},
		{
			name:     "valid https URI",
			resource: "https://mcp-server.example.com",
			wantErr:  false,
		},
		{
			name:     "valid http URI",
			resource: "http://localhost:8080/api",
			wantErr:  false,
		},
		{
			name:     "valid URI with path",
			resource: "https://api.example.com/v1/resources",
			wantErr:  false,
		},
		{
			name:     "invalid relative URI",
			resource: "/api/v1",
			wantErr:  true,
		},
		{
			name:     "invalid URI missing scheme",
			resource: "example.com",
			wantErr:  true,
		},
		{
			name:     "invalid URI with only path",
			resource: "api/v1/resources",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateResourceURI(tt.resource)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateResourceURI() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
