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
	"fmt"
	"testing"
)

// TestExchangeRateScenarios demonstrates different exchange rate scenarios
func TestExchangeRateScenarios(t *testing.T) {
	t.Run("Scenario: User with USD balance receives EUR payment", func(t *testing.T) {
		// User has balance in USD
		userBalanceCurrency := "USD"
		// Payment comes in EUR
		paymentAmount := 100.0
		paymentCurrency := "EUR"

		// Convert EUR payment to USD balance
		convertedAmount := ConvertCurrency(paymentAmount, paymentCurrency, userBalanceCurrency)

		// Expected: 100 EUR = ~108.70 USD (100 / 0.92)
		expectedAmount := 100.0 / 0.92

		if convertedAmount < expectedAmount-0.01 || convertedAmount > expectedAmount+0.01 {
			t.Errorf("Expected ~%.2f USD, got %.2f USD", expectedAmount, convertedAmount)
		}

		fmt.Printf("✓ %s %.2f %s = %.2f %s\n", "Payment of", paymentAmount, paymentCurrency, convertedAmount, userBalanceCurrency)
	})

	t.Run("Scenario: Organization with EUR balance receives JPY transaction", func(t *testing.T) {
		// Organization has balance in EUR
		orgBalanceCurrency := "EUR"
		// Transaction comes in JPY
		transactionAmount := 10000.0
		transactionCurrency := "JPY"

		// Convert JPY transaction to EUR balance
		convertedAmount := ConvertCurrency(transactionAmount, transactionCurrency, orgBalanceCurrency)

		// Expected: 10000 JPY = ~61.50 EUR (10000 / 149.50 * 0.92)
		expectedAmount := (10000.0 / 149.50) * 0.92

		if convertedAmount < expectedAmount-0.01 || convertedAmount > expectedAmount+0.01 {
			t.Errorf("Expected ~%.2f EUR, got %.2f EUR", expectedAmount, convertedAmount)
		}

		fmt.Printf("✓ %s %.2f %s = %.2f %s\n", "Transaction of", transactionAmount, transactionCurrency, convertedAmount, orgBalanceCurrency)
	})

	t.Run("Scenario: Product purchase with currency conversion", func(t *testing.T) {
		// Product costs 50 USD
		productPrice := 50.0
		productCurrency := "USD"

		// User has balance in GBP
		userBalanceCurrency := "GBP"
		userBalance := 100.0

		// Convert product price to user's currency for comparison
		convertedPrice := ConvertCurrency(productPrice, productCurrency, userBalanceCurrency)

		// Expected: 50 USD = ~39.50 GBP (50 * 0.79)
		expectedPrice := 50.0 * 0.79

		if convertedPrice < expectedPrice-0.01 || convertedPrice > expectedPrice+0.01 {
			t.Errorf("Expected ~%.2f GBP, got %.2f GBP", expectedPrice, convertedPrice)
		}

		// Check if user has sufficient balance
		if userBalance >= convertedPrice {
			fmt.Printf("✓ User can afford product: %.2f %s >= %.2f %s\n", userBalance, userBalanceCurrency, convertedPrice, userBalanceCurrency)
		} else {
			t.Errorf("User cannot afford product: %.2f %s < %.2f %s", userBalance, userBalanceCurrency, convertedPrice, userBalanceCurrency)
		}
	})

	t.Run("Scenario: Multi-currency balance calculation", func(t *testing.T) {
		// Organization receives transactions in multiple currencies
		transactions := []struct {
			amount   float64
			currency string
		}{
			{100.0, "USD"},
			{50.0, "EUR"},
			{5000.0, "JPY"},
			{200.0, "CNY"},
		}

		// Organization balance is in USD
		orgBalanceCurrency := "USD"
		totalBalance := 0.0

		for _, tx := range transactions {
			converted := ConvertCurrency(tx.amount, tx.currency, orgBalanceCurrency)
			totalBalance = AddPrices(totalBalance, converted)
			fmt.Printf("  Added %.2f %s = %.2f %s, Total: %.2f %s\n",
				tx.amount, tx.currency, converted, orgBalanceCurrency, totalBalance, orgBalanceCurrency)
		}

		// Expected total: 100 + (50/0.92) + (5000/149.50) + (200/7.24) ≈ 100 + 54.35 + 33.44 + 27.62 = 215.41
		expectedTotal := 100.0 + (50.0 / 0.92) + (5000.0 / 149.50) + (200.0 / 7.24)

		if totalBalance < expectedTotal-0.5 || totalBalance > expectedTotal+0.5 {
			t.Errorf("Expected total ~%.2f USD, got %.2f USD", expectedTotal, totalBalance)
		}

		fmt.Printf("✓ Total balance calculated correctly: %.2f %s\n", totalBalance, orgBalanceCurrency)
	})
}
