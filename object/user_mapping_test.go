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

func TestApplyUserMapping(t *testing.T) {
	// Create a test user
	user := &User{
		Owner: "test-org",
		Name:  "test-user",
	}

	// Create mock extra claims from IDP
	extraClaims := map[string]string{
		"sub":              "user-123",
		"given_name":       "John",
		"family_name":      "Doe",
		"phone_number":     "+1234567890",
		"country_code":     "US",
		"locality":         "San Francisco",
		"region":           "California",
		"organization":     "Acme Corp",
		"job_title":        "Software Engineer",
		"website":          "https://example.com",
		"bio":              "Test bio",
		"preferred_locale": "en-US",
	}

	// Create user mapping configuration
	userMapping := map[string]string{
		"firstName":   "given_name",
		"lastName":    "family_name",
		"phone":       "phone_number",
		"countryCode": "country_code",
		"location":    "locality",
		"region":      "region",
		"affiliation": "organization",
		"title":       "job_title",
		"homepage":    "website",
		"bio":         "bio",
		"language":    "preferred_locale",
	}

	// Apply user mapping
	applyUserMapping(user, extraClaims, userMapping)

	// Verify mappings
	if user.FirstName != "John" {
		t.Errorf("Expected FirstName to be 'John', got '%s'", user.FirstName)
	}
	if user.LastName != "Doe" {
		t.Errorf("Expected LastName to be 'Doe', got '%s'", user.LastName)
	}
	if user.Phone != "+1234567890" {
		t.Errorf("Expected Phone to be '+1234567890', got '%s'", user.Phone)
	}
	if user.CountryCode != "US" {
		t.Errorf("Expected CountryCode to be 'US', got '%s'", user.CountryCode)
	}
	if user.Location != "San Francisco" {
		t.Errorf("Expected Location to be 'San Francisco', got '%s'", user.Location)
	}
	if user.Region != "California" {
		t.Errorf("Expected Region to be 'California', got '%s'", user.Region)
	}
	if user.Affiliation != "Acme Corp" {
		t.Errorf("Expected Affiliation to be 'Acme Corp', got '%s'", user.Affiliation)
	}
	if user.Title != "Software Engineer" {
		t.Errorf("Expected Title to be 'Software Engineer', got '%s'", user.Title)
	}
	if user.Homepage != "https://example.com" {
		t.Errorf("Expected Homepage to be 'https://example.com', got '%s'", user.Homepage)
	}
	if user.Bio != "Test bio" {
		t.Errorf("Expected Bio to be 'Test bio', got '%s'", user.Bio)
	}
	if user.Language != "en-US" {
		t.Errorf("Expected Language to be 'en-US', got '%s'", user.Language)
	}
}

func TestApplyUserMappingDoesNotOverwriteExistingValues(t *testing.T) {
	// Create a test user with existing values
	user := &User{
		Owner:     "test-org",
		Name:      "test-user",
		FirstName: "Existing",
		LastName:  "User",
		Phone:     "+9999999999",
	}

	// Create mock extra claims from IDP
	extraClaims := map[string]string{
		"given_name":   "John",
		"family_name":  "Doe",
		"phone_number": "+1234567890",
	}

	// Create user mapping configuration
	userMapping := map[string]string{
		"firstName": "given_name",
		"lastName":  "family_name",
		"phone":     "phone_number",
	}

	// Apply user mapping
	applyUserMapping(user, extraClaims, userMapping)

	// Verify existing values are not overwritten
	if user.FirstName != "Existing" {
		t.Errorf("Expected FirstName to remain 'Existing', got '%s'", user.FirstName)
	}
	if user.LastName != "User" {
		t.Errorf("Expected LastName to remain 'User', got '%s'", user.LastName)
	}
	if user.Phone != "+9999999999" {
		t.Errorf("Expected Phone to remain '+9999999999', got '%s'", user.Phone)
	}
}

func TestApplyUserMappingWithMissingClaims(t *testing.T) {
	// Create a test user
	user := &User{
		Owner: "test-org",
		Name:  "test-user",
	}

	// Create mock extra claims with some missing fields
	extraClaims := map[string]string{
		"given_name": "John",
		// family_name is missing
	}

	// Create user mapping configuration
	userMapping := map[string]string{
		"firstName": "given_name",
		"lastName":  "family_name", // This claim doesn't exist
		"phone":     "phone_number", // This claim doesn't exist
	}

	// Apply user mapping
	applyUserMapping(user, extraClaims, userMapping)

	// Verify only available claims are mapped
	if user.FirstName != "John" {
		t.Errorf("Expected FirstName to be 'John', got '%s'", user.FirstName)
	}
	if user.LastName != "" {
		t.Errorf("Expected LastName to remain empty, got '%s'", user.LastName)
	}
	if user.Phone != "" {
		t.Errorf("Expected Phone to remain empty, got '%s'", user.Phone)
	}
}

func TestApplyUserMappingSkipsStandardFields(t *testing.T) {
	// Create a test user
	user := &User{
		Owner:       "test-org",
		Name:        "test-user",
		DisplayName: "Existing Display",
		Email:       "existing@example.com",
	}

	// Create mock extra claims
	extraClaims := map[string]string{
		"displayName": "New Display",
		"email":       "new@example.com",
		"id":          "new-id",
		"username":    "newusername",
	}

	// Try to map standard fields (should be skipped)
	userMapping := map[string]string{
		"displayName": "displayName",
		"email":       "email",
		"id":          "id",
		"username":    "username",
	}

	// Apply user mapping
	applyUserMapping(user, extraClaims, userMapping)

	// Verify standard fields are not modified by applyUserMapping
	// (they are handled elsewhere in the code)
	if user.DisplayName != "Existing Display" {
		t.Errorf("Expected DisplayName to remain 'Existing Display', got '%s'", user.DisplayName)
	}
	if user.Email != "existing@example.com" {
		t.Errorf("Expected Email to remain 'existing@example.com', got '%s'", user.Email)
	}
}
