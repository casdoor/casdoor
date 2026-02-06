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
	Products            []string `xorm:"varchar(1000)" json:"products"`
	ProductsDisplayName string   `xorm:"varchar(1000)" json:"productsDisplayName"`
	Detail              string   `xorm:"varchar(255)" json:"detail"`
	Currency            string   `xorm:"varchar(100)" json:"currency"`
	Price               float64  `json:"price"`

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
	Order      string          `xorm:"varchar(100)" json:"order"` // Internal order name
	OrderObj   *Order          `xorm:"-" json:"orderObj,omitempty"`
	OutOrderId string          `xorm:"varchar(100)" json:"outOrderId"` // External payment provider's order ID
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

	err = ExtendPaymentWithOrder(payments)
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

	err = ExtendPaymentWithOrder(payments)
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

	err = ExtendPaymentWithOrder(payments)
	if err != nil {
		return nil, err
	}

	return payments, nil
}

func ExtendPaymentWithOrder(payments []*Payment) error {
	ownerOrdersMap := make(map[string][]string)
	for _, payment := range payments {
		if payment.Order != "" {
			ownerOrdersMap[payment.Owner] = append(ownerOrdersMap[payment.Owner], payment.Order)
		}
	}

	ordersMap := make(map[string]*Order)
	for owner, orderNames := range ownerOrdersMap {
		if len(orderNames) == 0 {
			continue
		}
		var orders []*Order
		err := ormer.Engine.In("name", orderNames).Find(&orders, &Order{Owner: owner})
		if err != nil {
			return err
		}

		for _, order := range orders {
			ordersMap[util.GetId(order.Owner, order.Name)] = order
		}
	}

	for _, payment := range payments {
		if payment.Order != "" {
			orderId := util.GetId(payment.Owner, payment.Order)
			if order, ok := ordersMap[orderId]; ok {
				payment.OrderObj = order
			}
		}
	}
	return nil
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
	owner, name, err := util.GetOwnerAndNameFromIdWithError(id)
	if err != nil {
		return nil, err
	}
	return getPayment(owner, name)
}

func UpdatePayment(id string, payment *Payment) (bool, error) {
	owner, name, err := util.GetOwnerAndNameFromIdWithError(id)
	if err != nil {
		return false, err
	}
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

	// Check if the order products exist
	_, err = getOrderProducts(owner, payment.Products)
	if err != nil {
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
	if notifyResult.ProductDisplayName != "" && notifyResult.ProductDisplayName != payment.ProductsDisplayName {
		err = fmt.Errorf("the payment's product name: %s doesn't equal to the expected product name: %s", notifyResult.ProductDisplayName, payment.ProductsDisplayName)
		return payment, nil, err
	}

	if notifyResult.Price != payment.Price {
		err = fmt.Errorf("the payment's price: %f doesn't equal to the expected price: %f", notifyResult.Price, payment.Price)
		return payment, nil, err
	}

	return payment, notifyResult, nil
}

func NotifyPayment(body []byte, owner string, paymentName string, lang string) (*Payment, error) {
	payment, notifyResult, err := notifyPayment(body, owner, paymentName)
	if payment == nil {
		return nil, fmt.Errorf("the payment: %s does not exist", paymentName)
	}

	// Check if payment is already in a terminal state to prevent duplicate processing
	if pp.IsTerminalState(payment.State) {
		return payment, nil
	}

	// Determine the new payment state
	var newState pp.PaymentState
	var newMessage string
	if err != nil {
		newState = pp.PaymentStateError
		newMessage = err.Error()
	} else {
		newState = notifyResult.PaymentStatus
		newMessage = notifyResult.NotifyMessage
	}

	// Check if the payment state would actually change
	// This prevents duplicate webhook events when providers send redundant notifications
	if payment.State == newState {
		return payment, nil
	}

	payment.State = newState
	payment.Message = newMessage
	_, err = UpdatePayment(payment.GetId(), payment)
	if err != nil {
		return nil, err
	}

	// Update order state based on payment status
	order, err := getOrder(payment.Owner, payment.Order)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, fmt.Errorf("the order: %s does not exist", payment.Order)
	}

	if payment.State == pp.PaymentStatePaid {
		order.State = "Paid"
		order.Message = "Payment successful"
		order.UpdateTime = util.GetCurrentTime()
	} else if payment.State == pp.PaymentStateError {
		order.State = "PaymentFailed"
		order.Message = payment.Message
		order.UpdateTime = util.GetCurrentTime()
	} else if payment.State == pp.PaymentStateCanceled {
		order.State = "Canceled"
		order.Message = "Payment was cancelled"
		order.UpdateTime = util.GetCurrentTime()
	} else if payment.State == pp.PaymentStateTimeout {
		order.State = "Timeout"
		order.Message = "Payment timed out"
		order.UpdateTime = util.GetCurrentTime()
	}
	_, err = UpdateOrder(order.GetId(), order)
	if err != nil {
		return nil, err
	}

	if payment.State == pp.PaymentStatePaid {
		// Get provider, product and user for transaction creation
		provider, err := getProvider(payment.Owner, payment.Provider)
		if err != nil {
			return nil, err
		}
		if provider == nil {
			return nil, fmt.Errorf("the provider: %s does not exist", payment.Provider)
		}

		products, err := getOrderProducts(payment.Owner, order.Products)
		if err != nil {
			return nil, err
		}
		if len(products) == 0 {
			return nil, fmt.Errorf("order has no products")
		}
		for _, product := range products {
			if !product.IsRecharge && product.Quantity <= 0 {
				return nil, fmt.Errorf("the product: %s is out of stock", product.Name)
			}
		}

		user, err := getUser(payment.Owner, payment.User)
		if err != nil {
			return nil, err
		}
		if user == nil {
			return nil, fmt.Errorf("the user: %s does not exist", payment.User)
		}

		transaction := &Transaction{
			Owner:       payment.Owner,
			CreatedTime: util.GetCurrentTime(),
			Application: user.SignupApplication,
			Amount:      -payment.Price,
			Currency:    order.Currency,
			Payment:     payment.Name,
			Category:    TransactionCategoryPurchase,
			Type:        provider.Category,
			Subtype:     provider.Type,
			Provider:    provider.Name,
			Tag:         "User",
			User:        payment.User,
			State:       string(pp.PaymentStatePaid),
		}

		var affected bool
		affected, err = AddExternalPaymentTransaction(transaction, lang)
		if err != nil {
			return nil, err
		}
		if !affected {
			return nil, fmt.Errorf("failed to add transaction: %s", util.StructToJson(transaction))
		}

		hasRecharge := false
		totalPaidAmount := 0.0
		totalGrantedAmount := 0.0
		orderProductInfos := order.ProductInfos
		for _, productInfo := range orderProductInfos {
			if productInfo.IsRecharge {
				hasRecharge = true

				// Calculate paid and granted amounts for this product
				totalPaidAmount += productInfo.PaidAmount * float64(productInfo.Quantity)
				totalGrantedAmount += productInfo.GrantedAmount * float64(productInfo.Quantity)
			}
		}

		if hasRecharge {
			rechargeTransaction := &Transaction{
				Owner:         payment.Owner,
				CreatedTime:   util.GetCurrentTime(),
				Application:   user.SignupApplication,
				Amount:        totalPaidAmount + totalGrantedAmount,
				PaidAmount:    totalPaidAmount,
				GrantedAmount: totalGrantedAmount,
				Currency:      order.Currency,
				Payment:       payment.Name,
				Category:      TransactionCategoryRecharge,
				Type:          provider.Category,
				Subtype:       provider.Type,
				Provider:      provider.Name,
				Tag:           "User",
				User:          payment.User,
				State:         string(pp.PaymentStatePaid),
			}

			affected, err = AddExternalPaymentTransaction(rechargeTransaction, lang)
			if err != nil {
				return nil, err
			}
			if !affected {
				return nil, fmt.Errorf("failed to add recharge transaction: %s", util.StructToJson(rechargeTransaction))
			}
		}

		err = UpdateProductStock(orderProductInfos)
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
