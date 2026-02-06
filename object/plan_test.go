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

func TestPlanIdGeneration(t *testing.T) {
	plan := &Plan{
		Owner: "test-org",
		Name:  "premium-plan",
	}

	expectedId := "test-org/premium-plan"
	if plan.GetId() != expectedId {
		t.Errorf("Expected plan ID %s, got %s", expectedId, plan.GetId())
	}
}

func TestPlanIsOneTimeSubscription(t *testing.T) {
	tests := []struct {
		name                  string
		isOneTimeSubscription bool
		description           string
	}{
		{
			name:                  "Unrestricted plan",
			isOneTimeSubscription: false,
			description:           "Users can have multiple subscriptions",
		},
		{
			name:                  "One-time plan",
			isOneTimeSubscription: true,
			description:           "Users can only have one subscription",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plan := &Plan{
				Owner:                 "test-org",
				Name:                  "test-plan",
				IsOneTimeSubscription: tt.isOneTimeSubscription,
			}

			if plan.IsOneTimeSubscription != tt.isOneTimeSubscription {
				t.Errorf("Expected IsOneTimeSubscription=%v, got %v", tt.isOneTimeSubscription, plan.IsOneTimeSubscription)
			}
		})
	}
}

func TestGetDuration(t *testing.T) {
	tests := []struct {
		name        string
		period      string
		expectError bool
	}{
		{"Monthly period", PeriodMonthly, false},
		{"Yearly period", PeriodYearly, false},
		{"Invalid period", "Weekly", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			startTime, endTime, err := getDuration(tt.period)
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for period %s, but got none", tt.period)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for period %s: %v", tt.period, err)
				}
				if startTime == "" || endTime == "" {
					t.Errorf("Expected non-empty time strings for period %s", tt.period)
				}
			}
		})
	}
}
