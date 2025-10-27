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
	"time"

	"github.com/casdoor/casdoor/util"
)

func TestTransactionBalanceIntegration(t *testing.T) {
	InitConfig()

	// Create test organization
	testOrg := &Organization{
		Owner:          "admin",
		Name:           "test-org-" + util.GenerateId(),
		CreatedTime:    util.GetCurrentTime(),
		DisplayName:    "Test Organization",
		InitialBalance: 100.0,
		Balance:        100.0,
		Currency:       "USD",
	}

	// Add organization
	success, err := AddOrganization(testOrg)
	if err != nil {
		t.Fatalf("Failed to add organization: %v", err)
	}
	if !success {
		t.Fatal("Failed to add organization")
	}

	// Create test user
	testUser := &User{
		Owner:          testOrg.Name,
		Name:           "test-user-" + util.GenerateId(),
		CreatedTime:    util.GetCurrentTime(),
		DisplayName:    "Test User",
		InitialBalance: 50.0,
		Balance:        50.0,
		Currency:       "USD",
	}

	// Add user
	success, err = AddUser(testUser, "en")
	if err != nil {
		t.Fatalf("Failed to add user: %v", err)
	}
	if !success {
		t.Fatal("Failed to add user")
	}

	// Create test transaction
	transaction := &Transaction{
		Owner:              testOrg.Name,
		Name:               "test-transaction-" + util.GenerateId(),
		CreatedTime:        util.GetCurrentTime(),
		DisplayName:        "Test Transaction",
		User:               testUser.Name,
		Amount:             10.0,
		Currency:           "USD",
		Provider:           "test-provider",
		Category:           "Payment",
		Type:               "Test",
		ProductName:        "test-product",
		ProductDisplayName: "Test Product",
		Detail:             "Test transaction for balance integration",
		State:              "Paid",
	}

	// Test 1: Add transaction and verify balance updates
	success, err = AddTransaction(transaction)
	if err != nil {
		t.Fatalf("Failed to add transaction: %v", err)
	}
	if !success {
		t.Fatal("Failed to add transaction")
	}

	// Wait a moment for balance updates
	time.Sleep(100 * time.Millisecond)

	// Verify user balance increased
	updatedUser, err := getUser(testUser.Owner, testUser.Name)
	if err != nil {
		t.Fatalf("Failed to get updated user: %v", err)
	}
	expectedUserBalance := 60.0 // 50 (initial) + 10 (transaction)
	if updatedUser.Balance != expectedUserBalance {
		t.Errorf("User balance mismatch. Expected: %.2f, Got: %.2f", expectedUserBalance, updatedUser.Balance)
	}

	// Verify organization balance increased
	updatedOrg, err := getOrganization("admin", testOrg.Name)
	if err != nil {
		t.Fatalf("Failed to get updated organization: %v", err)
	}
	expectedOrgBalance := 110.0 // 100 (initial) + 10 (transaction)
	if updatedOrg.Balance != expectedOrgBalance {
		t.Errorf("Organization balance mismatch. Expected: %.2f, Got: %.2f", expectedOrgBalance, updatedOrg.Balance)
	}

	// Test 2: Update transaction amount and verify balance changes
	transaction.Amount = 15.0
	success, err = UpdateTransaction(transaction.GetId(), transaction)
	if err != nil {
		t.Fatalf("Failed to update transaction: %v", err)
	}
	if !success {
		t.Fatal("Failed to update transaction")
	}

	time.Sleep(100 * time.Millisecond)

	// Verify user balance reflects the change (increase of 5)
	updatedUser, err = getUser(testUser.Owner, testUser.Name)
	if err != nil {
		t.Fatalf("Failed to get updated user after transaction update: %v", err)
	}
	expectedUserBalance = 65.0 // 60 + 5 (difference in amount)
	if updatedUser.Balance != expectedUserBalance {
		t.Errorf("User balance after update mismatch. Expected: %.2f, Got: %.2f", expectedUserBalance, updatedUser.Balance)
	}

	// Verify organization balance reflects the change
	updatedOrg, err = getOrganization("admin", testOrg.Name)
	if err != nil {
		t.Fatalf("Failed to get updated organization after transaction update: %v", err)
	}
	expectedOrgBalance = 115.0 // 110 + 5 (difference in amount)
	if updatedOrg.Balance != expectedOrgBalance {
		t.Errorf("Organization balance after update mismatch. Expected: %.2f, Got: %.2f", expectedOrgBalance, updatedOrg.Balance)
	}

	// Test 3: Delete transaction and verify balance reversal
	success, err = DeleteTransaction(transaction)
	if err != nil {
		t.Fatalf("Failed to delete transaction: %v", err)
	}
	if !success {
		t.Fatal("Failed to delete transaction")
	}

	time.Sleep(100 * time.Millisecond)

	// Verify user balance reverted
	updatedUser, err = getUser(testUser.Owner, testUser.Name)
	if err != nil {
		t.Fatalf("Failed to get updated user after transaction deletion: %v", err)
	}
	expectedUserBalance = 50.0 // Back to initial balance
	if updatedUser.Balance != expectedUserBalance {
		t.Errorf("User balance after deletion mismatch. Expected: %.2f, Got: %.2f", expectedUserBalance, updatedUser.Balance)
	}

	// Verify organization balance reverted
	updatedOrg, err = getOrganization("admin", testOrg.Name)
	if err != nil {
		t.Fatalf("Failed to get updated organization after transaction deletion: %v", err)
	}
	expectedOrgBalance = 100.0 // Back to initial balance
	if updatedOrg.Balance != expectedOrgBalance {
		t.Errorf("Organization balance after deletion mismatch. Expected: %.2f, Got: %.2f", expectedOrgBalance, updatedOrg.Balance)
	}

	// Cleanup
	deleteUser(testUser)
	deleteOrganization(testOrg)
}
