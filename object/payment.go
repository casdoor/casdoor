// Copyright 2022 The Casdoor Authors. All Rights Reserved.
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

	"github.com/casdoor/casdoor/util"
	"xorm.io/core"
)

type Payment struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`
	DisplayName string `xorm:"varchar(100)" json:"displayName"`

	Provider     string `xorm:"varchar(100)" json:"provider"`
	Type         string `xorm:"varchar(100)" json:"type"`
	Organization string `xorm:"varchar(100)" json:"organization"`
	User         string `xorm:"varchar(100)" json:"user"`
	Good         string `xorm:"varchar(100)" json:"good"`
	Amount       string `xorm:"varchar(100)" json:"amount"`
	Currency     string `xorm:"varchar(100)" json:"currency"`

	State string `xorm:"varchar(100)" json:"state"`
}

func GetPaymentCount(owner, field, value string) int {
	session := GetSession(owner, -1, -1, field, value, "", "")
	count, err := session.Count(&Payment{})
	if err != nil {
		panic(err)
	}

	return int(count)
}

func GetPayments(owner string) []*Payment {
	payments := []*Payment{}
	err := adapter.Engine.Desc("created_time").Find(&payments, &Payment{Owner: owner})
	if err != nil {
		panic(err)
	}

	return payments
}

func GetPaginationPayments(owner string, offset, limit int, field, value, sortField, sortOrder string) []*Payment {
	payments := []*Payment{}
	session := GetSession(owner, offset, limit, field, value, sortField, sortOrder)
	err := session.Find(&payments)
	if err != nil {
		panic(err)
	}

	return payments
}

func getPayment(owner string, name string) *Payment {
	if owner == "" || name == "" {
		return nil
	}

	payment := Payment{Owner: owner, Name: name}
	existed, err := adapter.Engine.Get(&payment)
	if err != nil {
		panic(err)
	}

	if existed {
		return &payment
	} else {
		return nil
	}
}

func GetPayment(id string) *Payment {
	owner, name := util.GetOwnerAndNameFromId(id)
	return getPayment(owner, name)
}

func UpdatePayment(id string, payment *Payment) bool {
	owner, name := util.GetOwnerAndNameFromId(id)
	if getPayment(owner, name) == nil {
		return false
	}

	affected, err := adapter.Engine.ID(core.PK{owner, name}).AllCols().Update(payment)
	if err != nil {
		panic(err)
	}

	return affected != 0
}

func AddPayment(payment *Payment) bool {
	affected, err := adapter.Engine.Insert(payment)
	if err != nil {
		panic(err)
	}

	return affected != 0
}

func DeletePayment(payment *Payment) bool {
	affected, err := adapter.Engine.ID(core.PK{payment.Owner, payment.Name}).Delete(&Payment{})
	if err != nil {
		panic(err)
	}

	return affected != 0
}

func NotifyPayment(id string, state string) bool {
	owner, name := util.GetOwnerAndNameFromId(id)
	payment := getPayment(owner, name)
	if payment == nil {
		return false
	}

	payment.State = state

	affected, err := adapter.Engine.ID(core.PK{owner, name}).AllCols().Update(payment)
	if err != nil {
		panic(err)
	}

	return affected != 0
}

func (payment *Payment) GetId() string {
	return fmt.Sprintf("%s/%s", payment.Owner, payment.Name)
}
