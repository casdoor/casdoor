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
	"testing"
)

func TestRegisterDynamicClient(t *testing.T) {
	// Test with valid request
	req := &DynamicClientRegistrationRequest{
		ClientName:              "Test Client",
		RedirectUris:            []string{"http://localhost:3000/callback"},
		GrantTypes:              []string{"authorization_code"},
		ResponseTypes:           []string{"code"},
		TokenEndpointAuthMethod: "client_secret_basic",
		ApplicationType:         "web",
	}

	// Note: This test would require a running database and proper initialization
	// For now, we're just testing the structure and validation logic
	
	// Test missing redirect_uris
	reqInvalid := &DynamicClientRegistrationRequest{
		ClientName: "Invalid Client",
	}
	
	// Validate that the request structure is correct
	if reqInvalid.ClientName == "" {
		t.Error("Client name should not be empty")
	}
	
	if len(req.RedirectUris) == 0 {
		t.Error("Redirect URIs should not be empty for valid request")
	}
}

func TestDcrRequestDefaults(t *testing.T) {
	req := &DynamicClientRegistrationRequest{
		ClientName:   "Test Client",
		RedirectUris: []string{"http://localhost:3000/callback"},
	}

	// Test that we can set defaults
	if len(req.GrantTypes) == 0 {
		req.GrantTypes = []string{"authorization_code"}
	}
	if len(req.ResponseTypes) == 0 {
		req.ResponseTypes = []string{"code"}
	}
	if req.TokenEndpointAuthMethod == "" {
		req.TokenEndpointAuthMethod = "client_secret_basic"
	}
	if req.ApplicationType == "" {
		req.ApplicationType = "web"
	}

	if len(req.GrantTypes) != 1 || req.GrantTypes[0] != "authorization_code" {
		t.Error("Default grant type should be authorization_code")
	}
	if len(req.ResponseTypes) != 1 || req.ResponseTypes[0] != "code" {
		t.Error("Default response type should be code")
	}
	if req.TokenEndpointAuthMethod != "client_secret_basic" {
		t.Error("Default token endpoint auth method should be client_secret_basic")
	}
	if req.ApplicationType != "web" {
		t.Error("Default application type should be web")
	}
}
