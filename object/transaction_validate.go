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

	"github.com/casdoor/casdoor/i18n"
)

func validateBalanceForTransaction(transaction *Transaction, amount float64, lang string) error {
	currency := transaction.Currency
	if currency == "" {
		currency = "USD"
	}

	if transaction.Tag == "Organization" {
		// Validate organization balance change
		return validateOrganizationBalance("admin", transaction.Owner, amount, currency, true, lang)
	} else if transaction.Tag == "User" {
		// Validate user balance change
		if transaction.User == "" {
			return fmt.Errorf(i18n.Translate(lang, "general:User is required for User category transaction"))
		}
		if err := validateUserBalance(transaction.Owner, transaction.User, amount, currency, lang); err != nil {
			return err
		}
		// Validate organization's user balance sum change
		return validateOrganizationBalance("admin", transaction.Owner, amount, currency, false, lang)
	}
	return nil
}

func validateOrganizationBalance(owner string, name string, balance float64, currency string, isOrgBalance bool, lang string) error {
	organization, err := getOrganization(owner, name)
	if err != nil {
		return err
	}
	if organization == nil {
		return fmt.Errorf(i18n.Translate(lang, "auth:the organization: %s is not found"), fmt.Sprintf("%s/%s", owner, name))
	}

	// Convert the balance amount from transaction currency to organization's balance currency
	balanceCurrency := organization.BalanceCurrency
	if balanceCurrency == "" {
		balanceCurrency = "USD"
	}
	convertedBalance := ConvertCurrency(balance, currency, balanceCurrency)

	var newBalance float64
	if isOrgBalance {
		newBalance = AddPrices(organization.OrgBalance, convertedBalance)
		// Check organization balance credit limit
		if newBalance < organization.BalanceCredit {
			return fmt.Errorf(i18n.Translate(lang, "general:Insufficient balance: new organization balance %v would be below credit limit %v"), newBalance, organization.BalanceCredit)
		}
	} else {
		// User balance is just a sum of all users' balances, no credit limit check here
		// Individual user credit limits are checked in validateUserBalance
		newBalance = AddPrices(organization.UserBalance, convertedBalance)
	}

	// In validation mode, we don't actually update the balance
	return nil
}

func validateUserBalance(owner string, name string, balance float64, currency string, lang string) error {
	user, err := getUser(owner, name)
	if err != nil {
		return err
	}
	if user == nil {
		return fmt.Errorf(i18n.Translate(lang, "general:The user: %s is not found"), fmt.Sprintf("%s/%s", owner, name))
	}

	// Convert the balance amount from transaction currency to user's balance currency
	balanceCurrency := user.BalanceCurrency
	var org *Organization
	if balanceCurrency == "" {
		// Get organization's balance currency as fallback
		org, err = getOrganization("admin", owner)
		if err == nil && org != nil && org.BalanceCurrency != "" {
			balanceCurrency = org.BalanceCurrency
		} else {
			balanceCurrency = "USD"
		}
	}
	convertedBalance := ConvertCurrency(balance, currency, balanceCurrency)

	// Calculate new balance
	newBalance := AddPrices(user.Balance, convertedBalance)

	// Check balance credit limit
	// User.BalanceCredit takes precedence over Organization.BalanceCredit
	var balanceCredit float64
	if user.BalanceCredit != 0 {
		balanceCredit = user.BalanceCredit
	} else {
		// Get organization's balance credit as fallback
		if org == nil {
			org, err = getOrganization("admin", owner)
			if err != nil {
				return err
			}
		}
		if org != nil {
			balanceCredit = org.BalanceCredit
		}
	}

	// Validate new balance against credit limit
	if newBalance < balanceCredit {
		return fmt.Errorf(i18n.Translate(lang, "general:Insufficient balance: new balance %v would be below credit limit %v"), newBalance, balanceCredit)
	}

	// In validation mode, we don't actually update the balance
	return nil
}
