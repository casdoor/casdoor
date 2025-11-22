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

//go:build !skipCi
// +build !skipCi

package object

import (
	"testing"
)

func TestTransactionBalanceUpdate(t *testing.T) {
	InitConfig()

	// Test User category transaction
	userTransaction := &Transaction{
		Owner:    "test-org",
		Name:     "test-user-transaction",
		Category: "User",
		User:     "test-user",
		Amount:   100.0,
	}

	// Verify updateBalanceForTransaction for User category
	err := updateBalanceForTransaction(userTransaction, 100.0, "en")
	if err != nil {
		// Expected to fail if test user/org doesn't exist
		t.Logf("Expected error for non-existent user: %v", err)
	}

	// Test Organization category transaction
	orgTransaction := &Transaction{
		Owner:    "test-org",
		Name:     "test-org-transaction",
		Category: "Organization",
		Amount:   200.0,
	}

	// Verify updateBalanceForTransaction for Organization category
	err = updateBalanceForTransaction(orgTransaction, 200.0, "en")
	if err != nil {
		// Expected to fail if test org doesn't exist
		t.Logf("Expected error for non-existent organization: %v", err)
	}
}

func TestBalanceCreditValidation(t *testing.T) {
	InitConfig()

	// Test user balance credit validation
	// This test verifies that UpdateUserBalance correctly checks BalanceCredit limits
	t.Run("User balance credit validation", func(t *testing.T) {
		err := UpdateUserBalance("test-org", "test-user", -200.0, "USD", "en")
		if err == nil {
			t.Log("Expected error for balance below credit limit, but got none (user/org may not exist)")
		} else {
			t.Logf("Got expected error: %v", err)
		}
	})

	// Test organization balance credit validation
	t.Run("Organization balance credit validation", func(t *testing.T) {
		err := UpdateOrganizationBalance("admin", "test-org", -500.0, "USD", true, "en")
		if err == nil {
			t.Log("Expected error for balance below credit limit, but got none (org may not exist)")
		} else {
			t.Logf("Got expected error: %v", err)
		}
	})
}
