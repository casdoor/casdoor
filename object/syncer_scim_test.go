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

func TestSCIMUserToOriginalUser(t *testing.T) {
	provider := &SCIMSyncerProvider{
		Syncer: &Syncer{
			Host:     "https://example.com/scim/v2",
			User:     "testuser",
			Password: "testtoken",
		},
	}

	// Test case 1: Full SCIM user with all fields
	scimUser := &SCIMUser{
		ID:          "user-123",
		ExternalID:  "ext-123",
		UserName:    "john.doe",
		DisplayName: "John Doe",
		Name: SCIMName{
			GivenName:  "John",
			FamilyName: "Doe",
			Formatted:  "John Doe",
		},
		Title:        "Software Engineer",
		PreferredLan: "en-US",
		Active:       true,
		Emails: []SCIMEmail{
			{Value: "john.doe@example.com", Primary: true, Type: "work"},
			{Value: "john@personal.com", Primary: false, Type: "home"},
		},
		PhoneNumbers: []SCIMPhoneNumber{
			{Value: "+1-555-1234", Primary: true, Type: "work"},
			{Value: "+1-555-5678", Primary: false, Type: "mobile"},
		},
		Addresses: []SCIMAddress{
			{
				StreetAddress: "123 Main St",
				Locality:      "San Francisco",
				Region:        "CA",
				PostalCode:    "94102",
				Country:       "USA",
				Formatted:     "123 Main St, San Francisco, CA 94102, USA",
				Primary:       true,
				Type:          "work",
			},
		},
	}

	originalUser := provider.scimUserToOriginalUser(scimUser)

	// Verify basic fields
	if originalUser.Id != "user-123" {
		t.Errorf("Expected Id to be 'user-123', got '%s'", originalUser.Id)
	}
	if originalUser.ExternalId != "ext-123" {
		t.Errorf("Expected ExternalId to be 'ext-123', got '%s'", originalUser.ExternalId)
	}
	if originalUser.Name != "john.doe" {
		t.Errorf("Expected Name to be 'john.doe', got '%s'", originalUser.Name)
	}
	if originalUser.DisplayName != "John Doe" {
		t.Errorf("Expected DisplayName to be 'John Doe', got '%s'", originalUser.DisplayName)
	}
	if originalUser.FirstName != "John" {
		t.Errorf("Expected FirstName to be 'John', got '%s'", originalUser.FirstName)
	}
	if originalUser.LastName != "Doe" {
		t.Errorf("Expected LastName to be 'Doe', got '%s'", originalUser.LastName)
	}
	if originalUser.Title != "Software Engineer" {
		t.Errorf("Expected Title to be 'Software Engineer', got '%s'", originalUser.Title)
	}
	if originalUser.Language != "en-US" {
		t.Errorf("Expected Language to be 'en-US', got '%s'", originalUser.Language)
	}

	// Verify primary email is selected
	if originalUser.Email != "john.doe@example.com" {
		t.Errorf("Expected Email to be 'john.doe@example.com', got '%s'", originalUser.Email)
	}

	// Verify primary phone is selected
	if originalUser.Phone != "+1-555-1234" {
		t.Errorf("Expected Phone to be '+1-555-1234', got '%s'", originalUser.Phone)
	}

	// Verify address fields
	if originalUser.Location != "San Francisco" {
		t.Errorf("Expected Location to be 'San Francisco', got '%s'", originalUser.Location)
	}
	if originalUser.Region != "CA" {
		t.Errorf("Expected Region to be 'CA', got '%s'", originalUser.Region)
	}

	// Verify active status is inverted to IsForbidden
	if originalUser.IsForbidden != false {
		t.Errorf("Expected IsForbidden to be false for active user, got %v", originalUser.IsForbidden)
	}

	// Test case 2: Inactive SCIM user
	inactiveUser := &SCIMUser{
		ID:       "user-456",
		UserName: "jane.doe",
		Active:   false,
	}

	inactiveOriginalUser := provider.scimUserToOriginalUser(inactiveUser)
	if inactiveOriginalUser.IsForbidden != true {
		t.Errorf("Expected IsForbidden to be true for inactive user, got %v", inactiveOriginalUser.IsForbidden)
	}

	// Test case 3: SCIM user with no primary email/phone (should use first)
	noPrimaryUser := &SCIMUser{
		ID:       "user-789",
		UserName: "bob.smith",
		Emails: []SCIMEmail{
			{Value: "bob@example.com", Primary: false, Type: "work"},
			{Value: "bob@personal.com", Primary: false, Type: "home"},
		},
		PhoneNumbers: []SCIMPhoneNumber{
			{Value: "+1-555-9999", Primary: false, Type: "work"},
		},
	}

	noPrimaryOriginalUser := provider.scimUserToOriginalUser(noPrimaryUser)
	if noPrimaryOriginalUser.Email != "bob@example.com" {
		t.Errorf("Expected first email when no primary, got '%s'", noPrimaryOriginalUser.Email)
	}
	if noPrimaryOriginalUser.Phone != "+1-555-9999" {
		t.Errorf("Expected first phone when no primary, got '%s'", noPrimaryOriginalUser.Phone)
	}

	// Test case 4: Display name construction from first/last name when empty
	noDisplayNameUser := &SCIMUser{
		ID:       "user-101",
		UserName: "alice.jones",
		Name: SCIMName{
			GivenName:  "Alice",
			FamilyName: "Jones",
		},
	}

	noDisplayNameOriginalUser := provider.scimUserToOriginalUser(noDisplayNameUser)
	if noDisplayNameOriginalUser.DisplayName != "Alice Jones" {
		t.Errorf("Expected DisplayName to be constructed as 'Alice Jones', got '%s'", noDisplayNameOriginalUser.DisplayName)
	}
}

func TestSCIMBuildEndpoint(t *testing.T) {
	tests := []struct {
		host     string
		expected string
	}{
		{"https://example.com/scim/v2", "https://example.com/scim/v2/Users"},
		{"https://example.com/scim/v2/", "https://example.com/scim/v2/Users"},
		{"http://localhost:8080/scim", "http://localhost:8080/scim/Users"},
	}

	for _, test := range tests {
		provider := &SCIMSyncerProvider{
			Syncer: &Syncer{Host: test.host},
		}
		endpoint := provider.buildSCIMEndpoint()
		if endpoint != test.expected {
			t.Errorf("For host '%s', expected endpoint '%s', got '%s'", test.host, test.expected, endpoint)
		}
	}
}

func TestGetSyncerProviderSCIM(t *testing.T) {
	syncer := &Syncer{
		Type: "SCIM",
		Host: "https://example.com/scim/v2",
	}

	provider := GetSyncerProvider(syncer)

	if _, ok := provider.(*SCIMSyncerProvider); !ok {
		t.Errorf("Expected SCIMSyncerProvider for type 'SCIM', got %T", provider)
	}
}
