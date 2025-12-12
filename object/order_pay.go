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

	"github.com/casdoor/casdoor/idp"
	"github.com/casdoor/casdoor/pp"
	"github.com/casdoor/casdoor/util"
)

func PlaceOrder(productId string, user *User, pricingName string, planName string, customPrice float64) (*Order, error) {
	product, err := GetProduct(productId)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, fmt.Errorf("the product: %s does not exist", productId)
	}

	if !product.IsRecharge && product.Quantity <= 0 {
		return nil, fmt.Errorf("the product: %s is out of stock", product.Name)
	}

	userBalanceCurrency := user.BalanceCurrency
	if userBalanceCurrency == "" {
		org, err := getOrganization("admin", user.Owner)
		if err == nil && org != nil && org.BalanceCurrency != "" {
			userBalanceCurrency = org.BalanceCurrency
		} else {
			userBalanceCurrency = "USD"
		}
	}

	productCurrency := product.Currency
	if productCurrency == "" {
		productCurrency = "USD"
	}

	var productPrice float64
	if product.IsRecharge {
		if customPrice <= 0 {
			return nil, fmt.Errorf("the custom price should be greater than zero")
		}
		productPrice = customPrice
	} else {
		productPrice = product.Price
	}
	price := ConvertCurrency(productPrice, productCurrency, userBalanceCurrency)

	orderName := fmt.Sprintf("order_%v", util.GenerateTimeId())
	order := &Order{
		Owner:       product.Owner,
		Name:        orderName,
		CreatedTime: util.GetCurrentTime(),
		DisplayName: fmt.Sprintf("Order for %s", product.DisplayName),
		ProductName: product.Name,
		Products:    []string{product.Name},
		PricingName: pricingName,
		PlanName:    planName,
		User:        user.Name,
		Payment:     "", // Payment will be set when user pays
		Price:       price,
		Currency:    userBalanceCurrency,
		State:       "Created",
		Message:     "",
		StartTime:   util.GetCurrentTime(),
		EndTime:     "",
	}

	affected, err := AddOrder(order)
	if err != nil {
		return nil, err
	}
	if !affected {
		return nil, fmt.Errorf("failed to add order: %s", util.StructToJson(order))
	}

	return order, nil
}

func PayOrder(providerName, host, paymentEnv string, order *Order) (payment *Payment, attachInfo map[string]interface{}, err error) {
	if order.State != "Created" {
		return nil, nil, fmt.Errorf("cannot pay for order: %s, current state is %s", order.GetId(), order.State)
	}

	productId := util.GetId(order.Owner, order.ProductName)
	product, err := GetProduct(productId)
	if err != nil {
		return nil, nil, err
	}
	if product == nil {
		return nil, nil, fmt.Errorf("the product: %s does not exist", productId)
	}

	if !product.IsRecharge && product.Quantity <= 0 {
		return nil, nil, fmt.Errorf("the product: %s is out of stock", product.Name)
	}

	user, err := GetUser(util.GetId(order.Owner, order.User))
	if err != nil {
		return nil, nil, err
	}
	if user == nil {
		return nil, nil, fmt.Errorf("the user: %s does not exist", order.User)
	}

	provider, err := product.getProvider(providerName)
	if err != nil {
		return nil, nil, err
	}

	pProvider, err := GetPaymentProvider(provider)
	if err != nil {
		return nil, nil, err
	}

	owner := product.Owner
	payerName := fmt.Sprintf("%s | %s", user.Name, user.DisplayName)
	paymentName := fmt.Sprintf("payment_%v", util.GenerateTimeId())

	originFrontend, originBackend := getOriginFromHost(host)
	returnUrl := fmt.Sprintf("%s/payments/%s/%s/result", originFrontend, owner, paymentName)
	notifyUrl := fmt.Sprintf("%s/api/notify-payment/%s/%s", originBackend, owner, paymentName)

	// Create a subscription when pricing and plan are provided
	// This allows both free users and paid users to subscribe to plans
	if order.PricingName != "" && order.PlanName != "" {
		plan, err := GetPlan(util.GetId(owner, order.PlanName))
		if err != nil {
			return nil, nil, err
		}
		if plan == nil {
			return nil, nil, fmt.Errorf("the plan: %s does not exist", order.PlanName)
		}

		sub, err := NewSubscription(owner, user.Name, plan.Name, paymentName, plan.Period)
		if err != nil {
			return nil, nil, err
		}

		affected, err := AddSubscription(sub)
		if err != nil {
			return nil, nil, err
		}
		if !affected {
			return nil, nil, fmt.Errorf("failed to add subscription: %s", sub.Name)
		}

		returnUrl = fmt.Sprintf("%s/buy-plan/%s/%s/result?subscription=%s", originFrontend, owner, order.PricingName, sub.Name)
	}

	if product.SuccessUrl != "" {
		returnUrl = fmt.Sprintf("%s?transactionOwner=%s&transactionName=%s", product.SuccessUrl, owner, paymentName)
	}

	payReq := &pp.PayReq{
		ProviderName:       providerName,
		ProductName:        product.Name,
		PayerName:          payerName,
		PayerId:            user.Id,
		PayerEmail:         user.Email,
		PaymentName:        paymentName,
		ProductDisplayName: product.DisplayName,
		ProductDescription: product.Description,
		ProductImage:       product.Image,
		Price:              order.Price,
		Currency:           order.Currency,
		ReturnUrl:          returnUrl,
		NotifyUrl:          notifyUrl,
		PaymentEnv:         paymentEnv,
	}

	if provider.Type == "WeChat Pay" {
		payReq.PayerId, err = getUserExtraProperty(user, "WeChat", idp.BuildWechatOpenIdKey(provider.ClientId2))
		if err != nil {
			return nil, nil, err
		}
	} else if provider.Type == "Balance" {
		payReq.PayerId = user.GetId()
	}

	payResp, err := pProvider.Pay(payReq)
	if err != nil {
		return nil, nil, err
	}

	payment = &Payment{
		Owner:       product.Owner,
		Name:        paymentName,
		CreatedTime: util.GetCurrentTime(),
		DisplayName: paymentName,

		Provider: provider.Name,
		Type:     provider.Type,

		ProductName:        product.Name,
		ProductDisplayName: product.DisplayName,
		Detail:             product.Detail,
		Tag:                product.Tag,
		Currency:           order.Currency,
		Price:              order.Price,
		IsRecharge:         product.IsRecharge,

		User:       user.Name,
		Order:      order.Name,
		PayUrl:     payResp.PayUrl,
		SuccessUrl: returnUrl,
		State:      pp.PaymentStateCreated,
		OutOrderId: payResp.OrderId,
	}

	transaction := &Transaction{
		Owner:       payment.Owner,
		Name:        payment.Name,
		CreatedTime: util.GetCurrentTime(),
		DisplayName: payment.DisplayName,
		Application: owner,
		Domain:      "",
		Amount:      payment.Price,
		Currency:    order.Currency,
		Payment:     payment.Name,
		State:       pp.PaymentStateCreated,
	}

	if product.IsRecharge {
		transaction.Category = "Recharge"
		transaction.Type = ""
		transaction.Subtype = ""
		transaction.Provider = ""
		transaction.Tag = "User"
		transaction.User = payment.User
		transaction.State = pp.PaymentStatePaid
	} else {
		transaction.Category = ""
		transaction.Type = provider.Category
		transaction.Subtype = provider.Type
		transaction.Provider = provider.Name
		transaction.Tag = product.Tag
		transaction.User = payment.User
	}

	if provider.Type == "Dummy" {
		payment.State = pp.PaymentStatePaid
		currency := payment.Currency
		if currency == "" {
			currency = "USD"
		}
		err = UpdateUserBalance(user.Owner, user.Name, payment.Price, currency, "en")
		if err != nil {
			return nil, nil, err
		}
	} else if provider.Type == "Balance" {
		convertedPrice := ConvertCurrency(order.Price, order.Currency, user.BalanceCurrency)
		if convertedPrice > user.Balance {
			return nil, nil, fmt.Errorf("insufficient user balance")
		}
		transaction.Amount = -transaction.Amount
		err = UpdateUserBalance(user.Owner, user.Name, -convertedPrice, user.BalanceCurrency, "en")
		if err != nil {
			return nil, nil, err
		}
		payment.State = pp.PaymentStatePaid
		transaction.State = pp.PaymentStatePaid
	}

	affected, err := AddPayment(payment)
	if err != nil {
		return nil, nil, err
	}

	if !affected {
		return nil, nil, fmt.Errorf("failed to add payment: %s", util.StructToJson(payment))
	}

	if product.IsRecharge || provider.Type == "Balance" {
		affected, _, err = AddTransaction(transaction, "en", false)
		if err != nil {
			return nil, nil, err
		}
		if !affected {
			return nil, nil, fmt.Errorf("failed to add transaction: %s", util.StructToJson(transaction))
		}
	}

	order.Payment = payment.Name
	if provider.Type == "Dummy" || provider.Type == "Balance" {
		order.State = "Paid"
		order.Message = "Payment successful"
		order.EndTime = util.GetCurrentTime()
	}

	// Update order state first to avoid inconsistency
	_, err = UpdateOrder(order.GetId(), order)
	if err != nil {
		return nil, nil, err
	}

	// Update product stock after order state is persisted (for instant payment methods)
	if provider.Type == "Dummy" || provider.Type == "Balance" {
		err = UpdateProductStock(product)
		if err != nil {
			return nil, nil, err
		}
	}

	return payment, payResp.AttachInfo, nil
}

func CancelOrder(order *Order) (bool, error) {
	if order.State != "Created" {
		return false, fmt.Errorf("cannot cancel order in state: %s", order.State)
	}

	order.State = "Canceled"
	order.Message = "Canceled by user"
	order.EndTime = util.GetCurrentTime()
	return UpdateOrder(order.GetId(), order)
}
