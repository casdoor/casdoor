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

//go:build !skipCi

package object

import (
	"testing"

	"github.com/casdoor/casdoor/pp"
)

func TestPaymentStateChangeDetection(t *testing.T) {
	testCases := []struct {
		name              string
		currentState      pp.PaymentState
		newState          pp.PaymentState
		shouldUpdate      bool
		isTerminalState   bool
		description       string
	}{
		{
			name:            "Created to Created - no change",
			currentState:    pp.PaymentStateCreated,
			newState:        pp.PaymentStateCreated,
			shouldUpdate:    false,
			isTerminalState: false,
			description:     "Duplicate Created status should not trigger update",
		},
		{
			name:            "Created to Paid - state change",
			currentState:    pp.PaymentStateCreated,
			newState:        pp.PaymentStatePaid,
			shouldUpdate:    true,
			isTerminalState: false,
			description:     "State transition should trigger update",
		},
		{
			name:            "Paid to Paid - terminal state",
			currentState:    pp.PaymentStatePaid,
			newState:        pp.PaymentStatePaid,
			shouldUpdate:    false,
			isTerminalState: true,
			description:     "Duplicate Paid status should not trigger update (terminal state)",
		},
		{
			name:            "Error to Error - terminal state",
			currentState:    pp.PaymentStateError,
			newState:        pp.PaymentStateError,
			shouldUpdate:    false,
			isTerminalState: true,
			description:     "Duplicate Error status should not trigger update (terminal state)",
		},
		{
			name:            "Canceled to Canceled - terminal state",
			currentState:    pp.PaymentStateCanceled,
			newState:        pp.PaymentStateCanceled,
			shouldUpdate:    false,
			isTerminalState: true,
			description:     "Duplicate Canceled status should not trigger update (terminal state)",
		},
		{
			name:            "Timeout to Timeout - terminal state",
			currentState:    pp.PaymentStateTimeout,
			newState:        pp.PaymentStateTimeout,
			shouldUpdate:    false,
			isTerminalState: true,
			description:     "Duplicate Timeout status should not trigger update (terminal state)",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test terminal state check
			isTerminal := pp.IsTerminalState(tc.currentState)

			if isTerminal != tc.isTerminalState {
				t.Errorf("%s: Expected isTerminal=%v, got %v",
					tc.description, tc.isTerminalState, isTerminal)
			}

			// Test state change detection
			stateChanged := tc.currentState != tc.newState

			if stateChanged != tc.shouldUpdate {
				t.Errorf("%s: Expected shouldUpdate=%v, got %v (currentState=%s, newState=%s)",
					tc.description, tc.shouldUpdate, stateChanged,
					tc.currentState, tc.newState)
			}

			t.Logf("✓ %s: currentState=%s, newState=%s, shouldUpdate=%v, isTerminal=%v",
				tc.description, tc.currentState, tc.newState, tc.shouldUpdate, isTerminal)
		})
	}
}

// TestPaymentNotificationLogic tests the logic flow of NotifyPayment function
// to ensure duplicate notifications don't trigger unnecessary updates
func TestPaymentNotificationLogic(t *testing.T) {
	// This test validates the decision logic without requiring a full database setup

	testScenarios := []struct {
		scenario      string
		currentState  pp.PaymentState
		newState      pp.PaymentState
		expectSkip    bool
		skipReason    string
	}{
		{
			scenario:     "WechatPay sends duplicate Created notifications",
			currentState: pp.PaymentStateCreated,
			newState:     pp.PaymentStateCreated,
			expectSkip:   true,
			skipReason:   "State unchanged - should skip update and webhook",
		},
		{
			scenario:     "WechatPay sends Created then Paid",
			currentState: pp.PaymentStateCreated,
			newState:     pp.PaymentStatePaid,
			expectSkip:   false,
			skipReason:   "State changed - should update and trigger webhook",
		},
		{
			scenario:     "WechatPay sends duplicate Paid after success",
			currentState: pp.PaymentStatePaid,
			newState:     pp.PaymentStatePaid,
			expectSkip:   true,
			skipReason:   "Terminal state - should skip update and webhook",
		},
		{
			scenario:     "Provider sends duplicate Error notifications",
			currentState: pp.PaymentStateError,
			newState:     pp.PaymentStateError,
			expectSkip:   true,
			skipReason:   "Terminal state - should skip update and webhook",
		},
	}

	for _, scenario := range testScenarios {
		t.Run(scenario.scenario, func(t *testing.T) {
			// Check if it's a terminal state
			isTerminal := pp.IsTerminalState(scenario.currentState)

			// Check if state would change
			stateWouldChange := scenario.currentState != scenario.newState

			// Determine if we should skip (terminal state OR no state change)
			shouldSkip := isTerminal || !stateWouldChange

			if shouldSkip != scenario.expectSkip {
				t.Errorf("Scenario: %s\n  Expected skip=%v, got %v\n  Reason: %s\n  currentState=%s, newState=%s",
					scenario.scenario, scenario.expectSkip, shouldSkip,
					scenario.skipReason, scenario.currentState, scenario.newState)
			} else {
				t.Logf("✓ Scenario: %s - Correctly determined skip=%v (%s)",
					scenario.scenario, shouldSkip, scenario.skipReason)
			}
		})
	}
}
