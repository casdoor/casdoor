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

//go:build !skipCi
// +build !skipCi

package object

import (
	"testing"
)

func TestRadiusMfaUtil(t *testing.T) {
	// Test creating a RADIUS MFA util without config
	radiusMfa := NewRadiusMfaUtil(nil)
	if radiusMfa == nil {
		t.Fatal("NewRadiusMfaUtil returned nil")
	}
	if radiusMfa.MfaType != RadiusType {
		t.Fatalf("Expected MfaType %s, got %s", RadiusType, radiusMfa.MfaType)
	}

	// Test creating a RADIUS MFA util with config
	config := &MfaProps{
		MfaType: RadiusType,
		Secret:  "test-provider-id",
	}
	radiusMfa = NewRadiusMfaUtil(config)
	if radiusMfa == nil {
		t.Fatal("NewRadiusMfaUtil with config returned nil")
	}
	if radiusMfa.MfaType != RadiusType {
		t.Fatalf("Expected MfaType %s, got %s", RadiusType, radiusMfa.MfaType)
	}
}

func TestGetMfaUtilWithRadius(t *testing.T) {
	// Test that GetMfaUtil returns RadiusMfa for RadiusType
	config := &MfaProps{
		MfaType: RadiusType,
	}
	mfaUtil := GetMfaUtil(RadiusType, config)
	if mfaUtil == nil {
		t.Fatal("GetMfaUtil returned nil for RadiusType")
	}

	// Test that it implements MfaInterface
	_, ok := mfaUtil.(*RadiusMfa)
	if !ok {
		t.Fatal("GetMfaUtil did not return RadiusMfa for RadiusType")
	}
}

func TestRadiusTypeConstant(t *testing.T) {
	// Test that RadiusType constant is defined correctly
	if RadiusType != "radius" {
		t.Fatalf("Expected RadiusType to be 'radius', got '%s'", RadiusType)
	}
}

func TestGetAllMfaPropsIncludesRadius(t *testing.T) {
	// Create a test user
	user := &User{
		Owner: "test-org",
		Name:  "test-user",
	}

	// Get all MFA props
	mfaProps := GetAllMfaProps(user, false)

	// Check that we have 4 MFA types (SMS, Email, TOTP, RADIUS)
	if len(mfaProps) != 4 {
		t.Fatalf("Expected 4 MFA props, got %d", len(mfaProps))
	}

	// Check that RADIUS is included
	hasRadius := false
	for _, prop := range mfaProps {
		if prop.MfaType == RadiusType {
			hasRadius = true
			break
		}
	}
	if !hasRadius {
		t.Fatal("RADIUS MFA type not found in GetAllMfaProps")
	}
}
