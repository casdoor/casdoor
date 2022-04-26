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
	"net/http"

	"github.com/casdoor/casdoor/util"
	"xorm.io/core"
)

type Payment struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`
	DisplayName string `xorm:"varchar(100)" json:"displayName"`

	Provider           string `xorm:"varchar(100)" json:"provider"`
	Type               string `xorm:"varchar(100)" json:"type"`
	Organization       string `xorm:"varchar(100)" json:"organization"`
	User               string `xorm:"varchar(100)" json:"user"`
	ProductName        string `xorm:"varchar(100)" json:"productName"`
	ProductDisplayName string `xorm:"varchar(100)" json:"productDisplayName"`

	Detail   string  `xorm:"varchar(100)" json:"detail"`
	Tag      string  `xorm:"varchar(100)" json:"tag"`
	Currency string  `xorm:"varchar(100)" json:"currency"`
	Price    float64 `json:"price"`

	PayUrl    string `xorm:"varchar(2000)" json:"payUrl"`
	ReturnUrl string `xorm:"varchar(1000)" json:"returnUrl"`
	State     string `xorm:"varchar(100)" json:"state"`
	Message   string `xorm:"varchar(1000)" json:"message"`

	PersonName    string `xorm:"varchar(100)" json:"personName"`
	PersonIdCard  string `xorm:"varchar(100)" json:"personIdCard"`
	PersonEmail   string `xorm:"varchar(100)" json:"personEmail"`
	PersonPhone   string `xorm:"varchar(100)" json:"personPhone"`
	InvoiceType   string `xorm:"varchar(100)" json:"invoiceType"`
	InvoiceTitle  string `xorm:"varchar(100)" json:"invoiceTitle"`
	InvoiceTaxId  string `xorm:"varchar(100)" json:"invoiceTaxId"`
	InvoiceRemark string `xorm:"varchar(100)" json:"invoiceRemark"`
	InvoiceUrl    string `xorm:"varchar(100)" json:"invoiceUrl"`
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

func GetUserPayments(owner string, organization string, user string) []*Payment {
	payments := []*Payment{}
	err := adapter.Engine.Desc("created_time").Find(&payments, &Payment{Owner: owner, Organization: organization, User: user})
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

func notifyPayment(request *http.Request, body []byte, owner string, providerName string, productName string, paymentName string) (*Payment, error) {
	payment := getPayment(owner, paymentName)
	if payment == nil {
		return nil, fmt.Errorf("the payment: %s does not exist", paymentName)
	}

	product := getProduct(owner, productName)
	if product == nil {
		return nil, fmt.Errorf("the product: %s does not exist", productName)
	}

	provider, err := product.getProvider(providerName)
	if err != nil {
		return payment, err
	}

	pProvider, cert, err := provider.getPaymentProvider()
	if err != nil {
		return payment, err
	}

	productDisplayName, paymentName, price, productName, providerName, err := pProvider.Notify(request, body, cert.AuthorityPublicKey)
	if err != nil {
		return payment, err
	}

	if productDisplayName != "" && productDisplayName != product.DisplayName {
		return nil, fmt.Errorf("the payment's product name: %s doesn't equal to the expected product name: %s", productDisplayName, product.DisplayName)
	}

	if price != product.Price {
		return nil, fmt.Errorf("the payment's price: %f doesn't equal to the expected price: %f", price, product.Price)
	}

	return payment, nil
}

func NotifyPayment(request *http.Request, body []byte, owner string, providerName string, productName string, paymentName string) bool {
	payment, err := notifyPayment(request, body, owner, providerName, productName, paymentName)

	if payment != nil {
		if err != nil {
			payment.State = "Error"
			payment.Message = err.Error()
		} else {
			payment.State = "Paid"
		}

		UpdatePayment(payment.GetId(), payment)
	}

	ok := err == nil
	return ok
}

func invoicePayment(payment *Payment) (string, error) {
	provider := getProvider(payment.Owner, payment.Provider)
	if provider == nil {
		return "", fmt.Errorf("the payment provider: %s does not exist", payment.Provider)
	}

	pProvider, _, err := provider.getPaymentProvider()
	if err != nil {
		return "", err
	}

	invoiceUrl, err := pProvider.GetInvoice(payment.Name, payment.PersonName, payment.PersonIdCard, payment.PersonEmail, payment.PersonPhone, payment.InvoiceType, payment.InvoiceTitle, payment.InvoiceTaxId)
	if err != nil {
		return "", err
	}

	return invoiceUrl, nil
}

func InvoicePayment(payment *Payment) error {
	if payment.State != "Paid" {
		return fmt.Errorf("the payment state is supposed to be: \"%s\", got: \"%s\"", "Paid", payment.State)
	}

	invoiceUrl, err := invoicePayment(payment)
	if err != nil {
		return err
	}

	payment.InvoiceUrl = invoiceUrl
	affected := UpdatePayment(payment.GetId(), payment)
	if !affected {
		return fmt.Errorf("failed to update the payment: %s", payment.Name)
	}

	return nil
}

func (payment *Payment) GetId() string {
	return fmt.Sprintf("%s/%s", payment.Owner, payment.Name)
}
