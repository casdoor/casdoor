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
	"strings"

	"github.com/casdoor/casdoor/idp"
	"github.com/casdoor/casdoor/pp"
	"github.com/casdoor/casdoor/util"
)

func PlaceOrder(owner string, reqProductInfos []ProductInfo, user *User, pricingName string, planName string) (*Order, error) {
	if len(reqProductInfos) == 0 {
		return nil, fmt.Errorf("order has no products")
	}

	productNames := make([]string, 0, len(reqProductInfos))
	reqInfoMap := make(map[string]ProductInfo, len(reqProductInfos))
	for _, reqInfo := range reqProductInfos {
		if reqInfo.Name == "" {
			return nil, fmt.Errorf("product name cannot be empty")
		}
		productNames = append(productNames, reqInfo.Name)
		reqInfoMap[reqInfo.Name] = reqInfo
	}

	products, err := getOrderProducts(owner, productNames)
	if err != nil {
		return nil, err
	}

	orderCurrency := products[0].Currency
	if orderCurrency == "" {
		orderCurrency = "USD"
	}

	if err := validateProductCurrencies(products, orderCurrency); err != nil {
		return nil, err
	}

	var productInfos []ProductInfo
	orderPrice := 0.0
	for _, product := range products {
		reqInfo := reqInfoMap[product.Name]

		var productPrice float64
		if product.IsRecharge {
			productPrice = reqInfo.Price
			if productPrice <= 0 {
				return nil, fmt.Errorf("the custom price should be greater than zero")
			}
		} else {
			productPrice = product.Price
		}
		productInfos = append(productInfos, ProductInfo{
			Name:        product.Name,
			DisplayName: product.DisplayName,
			Image:       product.Image,
			Detail:      product.Detail,
			Price:       productPrice,
			IsRecharge:  product.IsRecharge,
		})

		orderPrice += productPrice
	}

	orderName := fmt.Sprintf("order_%v", util.GenerateTimeId())
	order := &Order{
		Owner:        owner,
		Name:         orderName,
		DisplayName:  orderName,
		CreatedTime:  util.GetCurrentTime(),
		Products:     productNames,
		ProductInfos: productInfos,
		PricingName:  pricingName,
		PlanName:     planName,
		User:         user.Name,
		Payment:      "", // Payment will be set when user pays
		Price:        orderPrice,
		Currency:     orderCurrency,
		State:        "Created",
		Message:      "",
		StartTime:    util.GetCurrentTime(),
		EndTime:      "",
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

func PayOrder(providerName, host, paymentEnv string, order *Order, lang string) (payment *Payment, attachInfo map[string]interface{}, err error) {
	if order.State != "Created" {
		return nil, nil, fmt.Errorf("cannot pay for order: %s, current state is %s", order.GetId(), order.State)
	}
	productNames := order.Products
	products, err := getOrderProducts(order.Owner, productNames)
	if err != nil {
		return nil, nil, err
	}
	if len(products) == 0 {
		return nil, nil, fmt.Errorf("order has no products")
	}

	orderCurrency := order.Currency
	if orderCurrency == "" {
		orderCurrency = "USD"
	}

	if err := validateProductCurrencies(products, orderCurrency); err != nil {
		return nil, nil, err
	}

	user, err := GetUser(util.GetId(order.Owner, order.User))
	if err != nil {
		return nil, nil, err
	}
	if user == nil {
		return nil, nil, fmt.Errorf("the user: %s does not exist", order.User)
	}

	// For multi-product orders, the payment provider is determined by the first product
	baseProduct := products[0]
	provider, err := baseProduct.getProvider(providerName)
	if err != nil {
		return nil, nil, err
	}

	pProvider, err := GetPaymentProvider(provider)
	if err != nil {
		return nil, nil, err
	}

	owner := baseProduct.Owner
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

	if baseProduct.SuccessUrl != "" {
		returnUrl = fmt.Sprintf("%s?transactionOwner=%s&transactionName=%s", baseProduct.SuccessUrl, owner, paymentName)
	}

	displayNames := make([]string, len(products))
	descriptions := make([]string, len(products))
	for i, p := range products {
		displayNames[i] = p.DisplayName
		descriptions[i] = p.Description
	}
	reqProductName := strings.Join(productNames, ", ")
	reqProductDisplayName := strings.Join(displayNames, ", ")
	reqProductDescription := strings.Join(descriptions, ", ")

	payReq := &pp.PayReq{
		ProviderName:       providerName,
		ProductName:        reqProductName,
		PayerName:          payerName,
		PayerId:            user.Id,
		PayerEmail:         user.Email,
		PaymentName:        paymentName,
		ProductDisplayName: reqProductDisplayName,
		ProductDescription: reqProductDescription,
		ProductImage:       baseProduct.Image,
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
		Owner:       baseProduct.Owner,
		Name:        paymentName,
		CreatedTime: util.GetCurrentTime(),
		DisplayName: paymentName,

		Provider: provider.Name,
		Type:     provider.Type,

		Products:            productNames,
		ProductsDisplayName: reqProductDisplayName,
		Detail:              reqProductDescription,
		Currency:            order.Currency,
		Price:               order.Price,

		User:       user.Name,
		Order:      order.Name,
		PayUrl:     payResp.PayUrl,
		SuccessUrl: returnUrl,
		State:      pp.PaymentStateCreated,
		OutOrderId: payResp.OrderId,
	}

	if provider.Type == "Dummy" || provider.Type == "Balance" {
		payment.State = pp.PaymentStatePaid
	}

	affected, err := AddPayment(payment)
	if err != nil {
		return nil, nil, err
	}

	if !affected {
		return nil, nil, fmt.Errorf("failed to add payment: %s", util.StructToJson(payment))
	}

	if provider.Type == "Balance" {
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

		affected, err = AddInternalPaymentTransaction(transaction, lang)
		if err != nil {
			return nil, nil, err
		}
		if !affected {
			return nil, nil, fmt.Errorf("failed to add transaction: %s", util.StructToJson(transaction))
		}

		hasRecharge := false
		rechargeAmount := 0.0
		for _, productInfo := range order.ProductInfos {
			if productInfo.IsRecharge {
				hasRecharge = true
				rechargeAmount += productInfo.Price
			}
		}

		if hasRecharge {
			rechargeTransaction := &Transaction{
				Owner:       payment.Owner,
				CreatedTime: util.GetCurrentTime(),
				Application: user.SignupApplication,
				Amount:      rechargeAmount,
				Currency:    order.Currency,
				Payment:     payment.Name,
				Category:    TransactionCategoryRecharge,
				Type:        provider.Category,
				Subtype:     provider.Type,
				Provider:    provider.Name,
				Tag:         "User",
				User:        payment.User,
				State:       string(pp.PaymentStatePaid),
			}

			affected, err = AddInternalPaymentTransaction(rechargeTransaction, lang)
			if err != nil {
				return nil, nil, err
			}
			if !affected {
				return nil, nil, fmt.Errorf("failed to add recharge transaction: %s", util.StructToJson(rechargeTransaction))
			}
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
		err = UpdateProductStock(products)
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
