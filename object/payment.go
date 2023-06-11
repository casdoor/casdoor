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
	"github.com/xorm-io/core"
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

	Detail   string  `xorm:"varchar(255)" json:"detail"`
	Tag      string  `xorm:"varchar(100)" json:"tag"`
	Currency string  `xorm:"varchar(100)" json:"currency"`
	Price    float64 `json:"price"`

	PayUrl    string `xorm:"varchar(2000)" json:"payUrl"`
	ReturnUrl string `xorm:"varchar(1000)" json:"returnUrl"`
	State     string `xorm:"varchar(100)" json:"state"`
	Message   string `xorm:"varchar(2000)" json:"message"`

	PersonName    string `xorm:"varchar(100)" json:"personName"`
	PersonIdCard  string `xorm:"varchar(100)" json:"personIdCard"`
	PersonEmail   string `xorm:"varchar(100)" json:"personEmail"`
	PersonPhone   string `xorm:"varchar(100)" json:"personPhone"`
	InvoiceType   string `xorm:"varchar(100)" json:"invoiceType"`
	InvoiceTitle  string `xorm:"varchar(100)" json:"invoiceTitle"`
	InvoiceTaxId  string `xorm:"varchar(100)" json:"invoiceTaxId"`
	InvoiceRemark string `xorm:"varchar(100)" json:"invoiceRemark"`
	InvoiceUrl    string `xorm:"varchar(255)" json:"invoiceUrl"`
}

func GetPaymentCount(owner, organization, field, value string) (int64, error) {
	session := GetSession(owner, -1, -1, field, value, "", "")
	return session.Count(&Payment{Organization: organization})
}

func GetPayments(owner string) ([]*Payment, error) {
	payments := []*Payment{}
	err := adapter.Engine.Desc("created_time").Find(&payments, &Payment{Owner: owner})
	if err != nil {
		return nil, err
	}

	return payments, nil
}

func GetUserPayments(owner string, organization string, user string) ([]*Payment, error) {
	payments := []*Payment{}
	err := adapter.Engine.Desc("created_time").Find(&payments, &Payment{Owner: owner, Organization: organization, User: user})
	if err != nil {
		return nil, err
	}

	return payments, nil
}

func GetPaginationPayments(owner, organization string, offset, limit int, field, value, sortField, sortOrder string) ([]*Payment, error) {
	payments := []*Payment{}
	session := GetSession(owner, offset, limit, field, value, sortField, sortOrder)
	err := session.Find(&payments, &Payment{Organization: organization})
	if err != nil {
		return nil, err
	}

	return payments, nil
}

func getPayment(owner string, name string) (*Payment, error) {
	if owner == "" || name == "" {
		return nil, nil
	}

	payment := Payment{Owner: owner, Name: name}
	existed, err := adapter.Engine.Get(&payment)
	if err != nil {
		return nil, err
	}

	if existed {
		return &payment, nil
	} else {
		return nil, nil
	}
}

func GetPayment(id string) (*Payment, error) {
	owner, name := util.GetOwnerAndNameFromId(id)
	return getPayment(owner, name)
}

func UpdatePayment(id string, payment *Payment) (bool, error) {
	owner, name := util.GetOwnerAndNameFromId(id)
	if p, err := getPayment(owner, name); err != nil {
		return false, err
	} else if p == nil {
		return false, nil
	}

	affected, err := adapter.Engine.ID(core.PK{owner, name}).AllCols().Update(payment)
	if err != nil {
		panic(err)
	}

	return affected != 0, nil
}

func AddPayment(payment *Payment) (bool, error) {
	affected, err := adapter.Engine.Insert(payment)
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func DeletePayment(payment *Payment) (bool, error) {
	affected, err := adapter.Engine.ID(core.PK{payment.Owner, payment.Name}).Delete(&Payment{})
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func notifyPayment(request *http.Request, body []byte, owner string, providerName string, productName string, paymentName string, orderId string) (*Payment, error, string) {
	provider, err := getProvider(owner, providerName)
	if err != nil {
		panic(err)
	}

	pProvider, cert, err := provider.getPaymentProvider()
	if err != nil {
		panic(err)
	}

	payment, err := getPayment(owner, paymentName)
	if err != nil {
		panic(err)
	}

	if payment == nil {
		err = fmt.Errorf("the payment: %s does not exist", paymentName)
		return nil, err, pProvider.GetResponseError(err)
	}

	product, err := getProduct(owner, productName)
	if err != nil {
		panic(err)
	}

	if product == nil {
		err = fmt.Errorf("the product: %s does not exist", productName)
		return payment, err, pProvider.GetResponseError(err)
	}

	productDisplayName, paymentName, price, productName, providerName, err := pProvider.Notify(request, body, cert.AuthorityPublicKey, orderId)
	if err != nil {
		return payment, err, pProvider.GetResponseError(err)
	}

	if productDisplayName != "" && productDisplayName != product.DisplayName {
		err = fmt.Errorf("the payment's product name: %s doesn't equal to the expected product name: %s", productDisplayName, product.DisplayName)
		return payment, err, pProvider.GetResponseError(err)
	}

	if price != product.Price {
		err = fmt.Errorf("the payment's price: %f doesn't equal to the expected price: %f", price, product.Price)
		return payment, err, pProvider.GetResponseError(err)
	}

	err = nil
	return payment, err, pProvider.GetResponseError(err)
}

func NotifyPayment(request *http.Request, body []byte, owner string, providerName string, productName string, paymentName string, orderId string) (error, string) {
	payment, err, errorResponse := notifyPayment(request, body, owner, providerName, productName, paymentName, orderId)
	if payment != nil {
		if err != nil {
			payment.State = "Error"
			payment.Message = err.Error()
		} else {
			payment.State = "Paid"
		}

		_, err = UpdatePayment(payment.GetId(), payment)
		if err != nil {
			panic(err)
		}
	}

	return err, errorResponse
}

func invoicePayment(payment *Payment) (string, error) {
	provider, err := getProvider(payment.Owner, payment.Provider)
	if err != nil {
		panic(err)
	}

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

func InvoicePayment(payment *Payment) (string, error) {
	if payment.State != "Paid" {
		return "", fmt.Errorf("the payment state is supposed to be: \"%s\", got: \"%s\"", "Paid", payment.State)
	}

	invoiceUrl, err := invoicePayment(payment)
	if err != nil {
		return "", err
	}

	payment.InvoiceUrl = invoiceUrl
	affected, err := UpdatePayment(payment.GetId(), payment)
	if err != nil {
		return "", err
	}

	if !affected {
		return "", fmt.Errorf("failed to update the payment: %s", payment.Name)
	}

	return invoiceUrl, nil
}

func (payment *Payment) GetId() string {
	return fmt.Sprintf("%s/%s", payment.Owner, payment.Name)
}
