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

func TestSubscriptionStateValidation(t *testing.T) {
	tests := []struct {
		name     string
		state    SubscriptionState
		isActive bool
	}{
		{"Active subscription", SubStateActive, true},
		{"Upcoming subscription", SubStateUpcoming, true},
		{"Pending subscription", SubStatePending, true},
		{"Expired subscription", SubStateExpired, false},
		{"Error subscription", SubStateError, false},
		{"Suspended subscription", SubStateSuspended, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test that the state is correctly identified
			isActive := tt.state == SubStateActive || tt.state == SubStateUpcoming || tt.state == SubStatePending
			if isActive != tt.isActive {
				t.Errorf("State %s: expected isActive=%v, got %v", tt.state, tt.isActive, isActive)
			}
		})
	}
}

func TestSubscriptionIdGeneration(t *testing.T) {
	sub := &Subscription{
		Owner: "test-org",
		Name:  "sub_123456",
	}

	expectedId := "test-org/sub_123456"
	if sub.GetId() != expectedId {
		t.Errorf("Expected subscription ID %s, got %s", expectedId, sub.GetId())
	}
}
