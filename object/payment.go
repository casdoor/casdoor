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

	"github.com/casdoor/casdoor/pp"

	"github.com/casdoor/casdoor/util"
	"github.com/xorm-io/core"
)

type Payment struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`
	DisplayName string `xorm:"varchar(100)" json:"displayName"`
	// Payment Provider Info
	Provider string `xorm:"varchar(100)" json:"provider"`
	Type     string `xorm:"varchar(100)" json:"type"`
	// Product Info
	ProductName        string  `xorm:"varchar(100)" json:"productName"`
	ProductDisplayName string  `xorm:"varchar(100)" json:"productDisplayName"`
	Detail             string  `xorm:"varchar(255)" json:"detail"`
	Tag                string  `xorm:"varchar(100)" json:"tag"`
	Currency           string  `xorm:"varchar(100)" json:"currency"`
	Price              float64 `json:"price"`
	ReturnUrl          string  `xorm:"varchar(1000)" json:"returnUrl"`
	// Payer Info
	User         string `xorm:"varchar(100)" json:"user"`
	PersonName   string `xorm:"varchar(100)" json:"personName"`
	PersonIdCard string `xorm:"varchar(100)" json:"personIdCard"`
	PersonEmail  string `xorm:"varchar(100)" json:"personEmail"`
	PersonPhone  string `xorm:"varchar(100)" json:"personPhone"`
	// Invoice Info
	InvoiceType   string `xorm:"varchar(100)" json:"invoiceType"`
	InvoiceTitle  string `xorm:"varchar(100)" json:"invoiceTitle"`
	InvoiceTaxId  string `xorm:"varchar(100)" json:"invoiceTaxId"`
	InvoiceRemark string `xorm:"varchar(100)" json:"invoiceRemark"`
	InvoiceUrl    string `xorm:"varchar(255)" json:"invoiceUrl"`
	// Order Info
	OutOrderId string          `xorm:"varchar(100)" json:"outOrderId"`
	PayUrl     string          `xorm:"varchar(2000)" json:"payUrl"`
	SuccessUrl string          `xorm:"varchar(2000)" json:"successUrl"` // `successUrl` is redirected from `payUrl` after pay success
	State      pp.PaymentState `xorm:"varchar(100)" json:"state"`
	Message    string          `xorm:"varchar(2000)" json:"message"`
}

func GetPaymentCount(owner, field, value string) (int64, error) {
	session := GetSession(owner, -1, -1, field, value, "", "")
	return session.Count(&Payment{Owner: owner})
}

func GetPayments(owner string) ([]*Payment, error) {
	payments := []*Payment{}
	err := ormer.Engine.Desc("created_time").Find(&payments, &Payment{Owner: owner})
	if err != nil {
		return nil, err
	}

	return payments, nil
}

func GetUserPayments(owner, user string) ([]*Payment, error) {
	payments := []*Payment{}
	err := ormer.Engine.Desc("created_time").Find(&payments, &Payment{Owner: owner, User: user})
	if err != nil {
		return nil, err
	}

	return payments, nil
}

func GetPaginationPayments(owner string, offset, limit int, field, value, sortField, sortOrder string) ([]*Payment, error) {
	payments := []*Payment{}
	session := GetSession(owner, offset, limit, field, value, sortField, sortOrder)
	err := session.Find(&payments, &Payment{Owner: owner})
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
	existed, err := ormer.Engine.Get(&payment)
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

	affected, err := ormer.Engine.ID(core.PK{owner, name}).AllCols().Update(payment)
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func AddPayment(payment *Payment) (bool, error) {
	affected, err := ormer.Engine.Insert(payment)
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func DeletePayment(payment *Payment) (bool, error) {
	affected, err := ormer.Engine.ID(core.PK{payment.Owner, payment.Name}).Delete(&Payment{})
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func notifyPayment(body []byte, owner string, paymentName string) (*Payment, *pp.NotifyResult, error) {
	payment, err := getPayment(owner, paymentName)
	if err != nil {
		return nil, nil, err
	}
	if payment == nil {
		err = fmt.Errorf("the payment: %s does not exist", paymentName)
		return nil, nil, err
	}

	provider, err := getProvider(owner, payment.Provider)
	if err != nil {
		return nil, nil, err
	}
	pProvider, err := GetPaymentProvider(provider)
	if err != nil {
		return nil, nil, err
	}

	product, err := getProduct(owner, payment.ProductName)
	if err != nil {
		return nil, nil, err
	}
	if product == nil {
		err = fmt.Errorf("the product: %s does not exist", payment.ProductName)
		return nil, nil, err
	}

	notifyResult, err := pProvider.Notify(body, payment.OutOrderId)
	if err != nil {
		return payment, nil, err
	}
	if notifyResult.PaymentStatus != pp.PaymentStatePaid {
		return payment, notifyResult, nil
	}
	// Only check paid payment
	if notifyResult.ProductDisplayName != "" && notifyResult.ProductDisplayName != product.DisplayName {
		err = fmt.Errorf("the payment's product name: %s doesn't equal to the expected product name: %s", notifyResult.ProductDisplayName, product.DisplayName)
		return payment, nil, err
	}

	if notifyResult.Price != product.Price {
		err = fmt.Errorf("the payment's price: %f doesn't equal to the expected price: %f", notifyResult.Price, product.Price)
		return payment, nil, err
	}

	return payment, notifyResult, nil
}

func NotifyPayment(body []byte, owner string, paymentName string) (*Payment, error) {
	payment, notifyResult, err := notifyPayment(body, owner, paymentName)
	if payment != nil {
		if err != nil {
			payment.State = pp.PaymentStateError
			payment.Message = err.Error()
		} else {
			payment.State = notifyResult.PaymentStatus
			payment.Message = notifyResult.NotifyMessage
		}
		_, err = UpdatePayment(payment.GetId(), payment)
		if err != nil {
			return nil, err
		}
	}

	return payment, nil
}

func invoicePayment(payment *Payment) (string, error) {
	provider, err := getProvider(payment.Owner, payment.Provider)
	if err != nil {
		panic(err)
	}

	if provider == nil {
		return "", fmt.Errorf("the payment provider: %s does not exist", payment.Provider)
	}

	pProvider, err := GetPaymentProvider(provider)
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
	if payment.State != pp.PaymentStatePaid {
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
