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
	"strings"
	"testing"
)

func TestTotpMfaUtil(t *testing.T) {
	// Test creating a new TOTP MFA util
	config := &MfaProps{
		MfaType: TotpType,
	}

	totpMfa := NewTotpMfaUtil(config)
	if totpMfa == nil {
		t.Error("NewTotpMfaUtil returned nil")
	}

	if totpMfa.MfaType != TotpType {
		t.Errorf("Expected MFA type %s, got %s", TotpType, totpMfa.MfaType)
	}
}

func TestTotpMfaInitiate_WithCustomIssuer(t *testing.T) {
	totpMfa := NewTotpMfaUtil(nil)
	
	// Test with custom issuer (application display name)
	customIssuer := "My Application"
	mfaProps, err := totpMfa.Initiate("test/user", customIssuer)
	if err != nil {
		t.Errorf("Initiate failed: %v", err)
	}

	if mfaProps == nil {
		t.Error("Initiate returned nil mfaProps")
	}

	if mfaProps.MfaType != TotpType {
		t.Errorf("Expected MFA type %s, got %s", TotpType, mfaProps.MfaType)
	}

	if mfaProps.Secret == "" {
		t.Error("Secret should not be empty")
	}

	if mfaProps.URL == "" {
		t.Error("URL should not be empty")
	}

	// Verify the URL contains the custom issuer (URL-encoded or plain)
	if !strings.Contains(mfaProps.URL, customIssuer) && !strings.Contains(mfaProps.URL, "My%20Application") {
		t.Errorf("URL should contain custom issuer '%s', got: %s", customIssuer, mfaProps.URL)
	}

	// Verify the URL contains the user ID
	if !strings.Contains(mfaProps.URL, "test/user") {
		t.Errorf("URL should contain user ID 'test/user', got: %s", mfaProps.URL)
	}
}

func TestTotpMfaInitiate_WithEmptyIssuer(t *testing.T) {
	totpMfa := NewTotpMfaUtil(nil)
	
	// Test with empty issuer (should default to "Casdoor")
	mfaProps, err := totpMfa.Initiate("test/user", "")
	if err != nil {
		t.Errorf("Initiate failed: %v", err)
	}

	if mfaProps == nil {
		t.Error("Initiate returned nil mfaProps")
	}

	// Verify the URL contains the default issuer "Casdoor"
	if !strings.Contains(mfaProps.URL, "Casdoor") {
		t.Errorf("URL should contain default issuer 'Casdoor', got: %s", mfaProps.URL)
	}
}

func TestGetMfaUtil_Totp(t *testing.T) {
	config := &MfaProps{
		MfaType: TotpType,
		Secret:  "testsecret",
	}

	mfaUtil := GetMfaUtil(TotpType, config)
	if mfaUtil == nil {
		t.Error("GetMfaUtil returned nil for TOTP type")
	}

	totpMfa, ok := mfaUtil.(*TotpMfa)
	if !ok {
		t.Error("GetMfaUtil did not return TotpMfa type")
	}

	if totpMfa.MfaType != TotpType {
		t.Errorf("Expected MFA type %s, got %s", TotpType, totpMfa.MfaType)
	}
}

func TestTotpType(t *testing.T) {
	// Test that TotpType constant is defined correctly
	if TotpType != "app" {
		t.Errorf("Expected TotpType to be 'app', got '%s'", TotpType)
	}
}
