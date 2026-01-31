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
	"fmt"
	"strings"

	"github.com/casdoor/casdoor/i18n"
	"github.com/casdoor/casdoor/util"
	"github.com/xorm-io/core"
)

type TransactionCategory string

const (
	TransactionCategoryPurchase TransactionCategory = "Purchase"
	TransactionCategoryRecharge TransactionCategory = "Recharge"
)

type Transaction struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`
	DisplayName string `xorm:"varchar(100)" json:"displayName"`

	Application string              `xorm:"varchar(100)" json:"application"`
	Domain      string              `xorm:"varchar(1000)" json:"domain"`
	Category    TransactionCategory `xorm:"varchar(100)" json:"category"`
	Type        string              `xorm:"varchar(100)" json:"type"`
	Subtype     string              `xorm:"varchar(100)" json:"subtype"`
	Provider    string              `xorm:"varchar(100)" json:"provider"`
	User        string              `xorm:"varchar(100)" json:"user"`
	Tag         string              `xorm:"varchar(100)" json:"tag"`

	Amount        float64 `json:"amount"`
	PaidAmount    float64 `json:"paidAmount"`
	GrantedAmount float64 `json:"grantedAmount"`
	Currency      string  `xorm:"varchar(100)" json:"currency"`

	Payment string `xorm:"varchar(100)" json:"payment"`

	State string `xorm:"varchar(100)" json:"state"`
}

func GetTransactionCount(owner, field, value string) (int64, error) {
	session := GetSession(owner, -1, -1, field, value, "", "")
	return session.Count(&Transaction{Owner: owner})
}

func GetTransactions(owner string) ([]*Transaction, error) {
	transactions := []*Transaction{}
	err := ormer.Engine.Desc("created_time").Find(&transactions, &Transaction{Owner: owner})
	if err != nil {
		return nil, err
	}

	return transactions, nil
}

func GetUserTransactions(owner, user string) ([]*Transaction, error) {
	transactions := []*Transaction{}
	err := ormer.Engine.Desc("created_time").Find(&transactions, &Transaction{Owner: owner, User: user})
	if err != nil {
		return nil, err
	}

	return transactions, nil
}

func GetPaginationTransactions(owner string, offset, limit int, field, value, sortField, sortOrder string) ([]*Transaction, error) {
	transactions := []*Transaction{}
	session := GetSession(owner, offset, limit, field, value, sortField, sortOrder)
	err := session.Find(&transactions, &Transaction{Owner: owner})
	if err != nil {
		return nil, err
	}

	return transactions, nil
}

func getTransaction(owner string, name string) (*Transaction, error) {
	if owner == "" || name == "" {
		return nil, nil
	}

	transaction := Transaction{Owner: owner, Name: name}
	existed, err := ormer.Engine.Get(&transaction)
	if err != nil {
		return nil, err
	}

	if existed {
		return &transaction, nil
	} else {
		return nil, nil
	}
}

func GetTransaction(id string) (*Transaction, error) {
	owner, name, err := util.GetOwnerAndNameFromIdWithError(id)
	if err != nil {
		return nil, err
	}
	return getTransaction(owner, name)
}

func UpdateTransaction(id string, transaction *Transaction, lang string) (bool, error) {
	owner, name, err := util.GetOwnerAndNameFromIdWithError(id)
	if err != nil {
		return false, err
	}
	oldTransaction, err := getTransaction(owner, name)
	if err != nil {
		return false, err
	} else if oldTransaction == nil {
		return false, nil
	}

	// Revert old balance changes
	if err := updateBalanceForTransaction(oldTransaction, -oldTransaction.Amount, lang); err != nil {
		return false, err
	}

	affected, err := ormer.Engine.ID(core.PK{owner, name}).AllCols().Update(transaction)
	if err != nil {
		return false, err
	}

	// Apply new balance changes
	if affected != 0 {
		if err := updateBalanceForTransaction(transaction, transaction.Amount, lang); err != nil {
			return false, err
		}
	}

	return affected != 0, nil
}

func AddTransaction(transaction *Transaction, lang string, dryRun bool) (bool, string, error) {
	transactionId := strings.ReplaceAll(util.GenerateId(), "-", "")
	transaction.Name = transactionId
	transaction.DisplayName = transactionId

	// In dry run mode, only validate without making changes
	if dryRun {
		err := validateBalanceForTransaction(transaction, transaction.Amount, lang)
		if err != nil {
			return false, "", err
		}

		return true, "", nil
	}

	affected, err := ormer.Engine.Insert(transaction)
	if err != nil {
		return false, "", err
	}

	if affected != 0 {
		if err := updateBalanceForTransaction(transaction, transaction.Amount, lang); err != nil {
			return false, transactionId, err
		}
	}

	return affected != 0, transactionId, nil
}

func AddInternalPaymentTransaction(transaction *Transaction, lang string) (bool, error) {
	transactionId := strings.ReplaceAll(util.GenerateId(), "-", "")
	transaction.Name = transactionId
	transaction.DisplayName = transactionId

	// Validate balance impact first
	if err := validateBalanceForTransaction(transaction, transaction.Amount, lang); err != nil {
		return false, err
	}

	affected, err := ormer.Engine.Insert(transaction)
	if err != nil {
		return false, err
	}

	if affected != 0 {
		if err := updateBalanceForTransaction(transaction, transaction.Amount, lang); err != nil {
			return false, err
		}
	}

	return affected != 0, nil
}

func AddExternalPaymentTransaction(transaction *Transaction, lang string) (bool, error) {
	transactionId := strings.ReplaceAll(util.GenerateId(), "-", "")
	transaction.Name = transactionId
	transaction.DisplayName = transactionId

	affected, err := ormer.Engine.Insert(transaction)
	if err != nil {
		return false, err
	}

	if affected != 0 && transaction.Category == TransactionCategoryRecharge {
		if err := updateBalanceForTransaction(transaction, transaction.Amount, lang); err != nil {
			return false, err
		}
	}

	return affected != 0, nil
}

func DeleteTransaction(transaction *Transaction, lang string) (bool, error) {
	// Revert balance changes before deleting
	if err := updateBalanceForTransaction(transaction, -transaction.Amount, lang); err != nil {
		return false, err
	}

	affected, err := ormer.Engine.ID(core.PK{transaction.Owner, transaction.Name}).Delete(&Transaction{})
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func (transaction *Transaction) GetId() string {
	return fmt.Sprintf("%s/%s", transaction.Owner, transaction.Name)
}

func updateBalanceForTransaction(transaction *Transaction, amount float64, lang string) error {
	currency := transaction.Currency
	if currency == "" {
		currency = "USD"
	}

	if transaction.Tag == "Organization" {
		// Update organization's own balance
		return UpdateOrganizationBalance("admin", transaction.Owner, amount, currency, true, lang)
	} else if transaction.Tag == "User" {
		// Update user's balance
		if transaction.User == "" {
			return fmt.Errorf(i18n.Translate(lang, "general:User is required for User category transaction"))
		}
		if err := UpdateUserBalance(transaction.Owner, transaction.User, amount, currency, lang); err != nil {
			return err
		}
		// Update organization's user balance sum
		return UpdateOrganizationBalance("admin", transaction.Owner, amount, currency, false, lang)
	}
	return nil
}
