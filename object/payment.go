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
	"github.com/go-pay/gopay"
	"github.com/go-pay/gopay/alipay"
	"xorm.io/core"
)

type Payment struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`
	DisplayName string `xorm:"varchar(100)" json:"displayName"`

	Provider     string  `xorm:"varchar(100)" json:"provider"`
	Type         string  `xorm:"varchar(100)" json:"type"`
	Organization string  `xorm:"varchar(100)" json:"organization"`
	User         string  `xorm:"varchar(100)" json:"user"`
	ProductId    string  `xorm:"varchar(100)" json:"productId"`
	ProductName  string  `xorm:"varchar(100)" json:"productName"`
	Price        float64 `json:"price"`
	Currency     string  `xorm:"varchar(100)" json:"currency"`

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

func NotifyPayment(bm gopay.BodyMap) bool {
	owner := "admin"
	productName := bm.Get("subject")
	paymentId := bm.Get("out_trade_no")
	priceString := bm.Get("total_amount")
	price := util.ParseFloat(priceString)
	productId := bm.Get("productId")
	providerId := bm.Get("providerId")

	product := getProduct(owner, productId)
	if product == nil {
		panic(fmt.Errorf("the product: %s does not exist", productId))
	}

	if productName != product.DisplayName {
		panic(fmt.Errorf("the payment's product name: %s doesn't equal to the expected product name: %s", productName, product.DisplayName))
	}

	if price != product.Price {
		panic(fmt.Errorf("the payment's price: %f doesn't equal to the expected price: %f", price, product.Price))
	}

	payment := getPayment(owner, paymentId)
	if payment == nil {
		panic(fmt.Errorf("the payment: %s does not exist", paymentId))
	}

	provider, err := product.getProvider(providerId)
	if err != nil {
		panic(err)
	}

	cert := getCert(owner, provider.Cert)
	if cert == nil {
		panic(fmt.Errorf("the cert: %s does not exist", provider.Cert))
	}

	ok, err := alipay.VerifySignWithCert(cert.AuthorityPublicKey, bm)
	if err != nil {
		panic(err)
	}

	if ok {
		payment.State = "Paid"
	} else {
		if cert == nil {
			panic(fmt.Errorf("VerifySignWithCert() failed: %v", ok))
		}
		//payment.State = "Failed"
	}

	affected, err := adapter.Engine.ID(core.PK{owner, paymentId}).AllCols().Update(payment)
	if err != nil {
		panic(err)
	}

	return affected != 0
}

func (payment *Payment) GetId() string {
	return fmt.Sprintf("%s/%s", payment.Owner, payment.Name)
}
