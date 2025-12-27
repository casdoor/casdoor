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

func TestRadiusMfaUtil(t *testing.T) {
	// Test creating a new RADIUS MFA util without provider lookup
	config := &MfaProps{
		MfaType: RadiusType,
		Secret:  "testuser",
	}

	radiusMfa := NewRadiusMfaUtil(config)
	if radiusMfa == nil {
		t.Error("NewRadiusMfaUtil returned nil")
	}

	if radiusMfa.MfaType != RadiusType {
		t.Errorf("Expected MFA type %s, got %s", RadiusType, radiusMfa.MfaType)
	}

	// Test Initiate
	mfaProps, err := radiusMfa.Initiate("test/user", "")
	if err != nil {
		t.Errorf("Initiate failed: %v", err)
	}

	if mfaProps == nil {
		t.Error("Initiate returned nil mfaProps")
	}

	if mfaProps.MfaType != RadiusType {
		t.Errorf("Expected MFA type %s, got %s", RadiusType, mfaProps.MfaType)
	}
}

func TestGetMfaUtil_Radius(t *testing.T) {
	config := &MfaProps{
		MfaType: RadiusType,
		Secret:  "testuser",
	}

	mfaUtil := GetMfaUtil(RadiusType, config)
	if mfaUtil == nil {
		t.Error("GetMfaUtil returned nil for RADIUS type")
	}

	radiusMfa, ok := mfaUtil.(*RadiusMfa)
	if !ok {
		t.Error("GetMfaUtil did not return RadiusMfa type")
	}

	if radiusMfa.MfaType != RadiusType {
		t.Errorf("Expected MFA type %s, got %s", RadiusType, radiusMfa.MfaType)
	}
}

func TestRadiusMfaType(t *testing.T) {
	// Test that RadiusType constant is defined correctly
	if RadiusType != "radius" {
		t.Errorf("Expected RadiusType to be 'radius', got '%s'", RadiusType)
	}
}
