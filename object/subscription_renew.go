// Copyright 2026 The Casdoor Authors. All Rights Reserved.
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
	"time"

	"github.com/casdoor/casdoor/pp"
	"github.com/casdoor/casdoor/util"
	"github.com/xorm-io/core"
)

const subscriptionRenewalWindow = 24 * time.Hour

const (
	RenewalStateRenewing = "Renewing"
	RenewalStateRenewed  = "Renewed"
	RenewalStateFailed   = "Failed"
)

func isAutoRenewSupported(providerType string) bool {
	return providerType == "Balance" || providerType == "Stripe"
}

func getSubscriptionPaymentProvider(sub *Subscription) (*Payment, *Provider, error) {
	if sub.Payment == "" {
		return nil, nil, util.MessageErrorf("subscription:the subscription: %s has no payment to renew with", sub.GetId())
	}

	payment, err := getPayment(sub.Owner, sub.Payment)
	if err != nil {
		return nil, nil, err
	}
	if payment == nil {
		return nil, nil, util.MessageErrorf("payment:the payment: %s does not exist", sub.Payment)
	}
	if payment.State != pp.PaymentStatePaid {
		return nil, nil, util.MessageErrorf("subscription:the payment: %s is not paid", payment.Name)
	}

	provider, err := getProvider(payment.Owner, payment.Provider)
	if err != nil {
		return nil, nil, err
	}
	if provider == nil {
		return nil, nil, util.MessageErrorf("provider:the provider: %s does not exist", payment.Provider)
	}
	if !isAutoRenewSupported(provider.Type) {
		return nil, nil, util.MessageErrorf("subscription:auto renewal is not supported for the payment provider type: %s", provider.Type)
	}

	return payment, provider, nil
}

func getRenewablePaymentProvider(provider *Provider) (pp.RenewablePaymentProvider, error) {
	pProvider, err := GetPaymentProvider(provider)
	if err != nil {
		return nil, err
	}

	renewable, ok := pProvider.(pp.RenewablePaymentProvider)
	if !ok {
		return nil, util.MessageErrorf("subscription:auto renewal is not supported for the payment provider type: %s", provider.Type)
	}
	return renewable, nil
}

// checkSubscriptionAutoRenew refuses auto renewal for a subscription whose payment can't be charged again
func checkSubscriptionAutoRenew(sub *Subscription) error {
	payment, provider, err := getSubscriptionPaymentProvider(sub)
	if err != nil {
		return err
	}
	if provider.Type == "Balance" {
		return nil
	}

	renewable, err := getRenewablePaymentProvider(provider)
	if err != nil {
		return err
	}
	return renewable.CheckRenewal(payment.OutOrderId)
}

func SetSubscriptionAutoRenew(sub *Subscription, autoRenew bool) (bool, error) {
	if autoRenew {
		err := checkSubscriptionAutoRenew(sub)
		if err != nil {
			return false, err
		}
	}

	affected, err := ormer.Engine.ID(core.PK{sub.Owner, sub.Name}).Cols("auto_renew").Update(&Subscription{AutoRenew: autoRenew})
	if err != nil {
		return false, err
	}

	if autoRenew && sub.RenewalState == RenewalStateFailed {
		_, err = ormer.Engine.ID(core.PK{sub.Owner, sub.Name}).Where("renewal_state = ?", RenewalStateFailed).
			Cols("renewal_state", "renewal_message").Update(&Subscription{})
		if err != nil {
			return false, err
		}
	}

	return affected != 0, nil
}

func (sub *Subscription) needsRenewal(now time.Time) bool {
	if !sub.AutoRenew || sub.RenewalState != "" || sub.State != SubStateActive {
		return false
	}

	endTime, err := time.Parse(time.RFC3339, sub.EndTime)
	if err != nil {
		return false
	}
	return endTime.After(now) && endTime.Sub(now) <= subscriptionRenewalWindow
}

// claimSubscriptionRenewal marks the subscription as being renewed, only one caller can win it even across replicas
func claimSubscriptionRenewal(sub *Subscription) (bool, error) {
	affected, err := ormer.Engine.ID(core.PK{sub.Owner, sub.Name}).
		Where("(renewal_state = ? or renewal_state is null)", "").And("auto_renew = ?", true).
		Cols("renewal_state").Update(&Subscription{RenewalState: RenewalStateRenewing})
	if err != nil {
		return false, err
	}
	return affected != 0, nil
}

func renewSubscription(sub *Subscription) error {
	claimed, err := claimSubscriptionRenewal(sub)
	if err != nil || !claimed {
		return err
	}

	nextSub, renewErr := chargeSubscriptionRenewal(sub)
	if renewErr != nil {
		sub.RenewalState = RenewalStateFailed
		sub.RenewalMessage = renewErr.Error()
	} else {
		sub.RenewalState = RenewalStateRenewed
		sub.RenewalSubscription = nextSub.Name
		sub.RenewalMessage = ""
	}

	_, err = ormer.Engine.ID(core.PK{sub.Owner, sub.Name}).Cols("renewal_state", "renewal_subscription", "renewal_message").Update(sub)
	if err != nil {
		return err
	}
	return renewErr
}

func getRenewalPricingName(sub *Subscription, payment *Payment) (string, error) {
	if sub.Pricing != "" {
		return sub.Pricing, nil
	}

	order, err := getOrder(payment.Owner, payment.Order)
	if err != nil || order == nil {
		return "", err
	}
	for _, productInfo := range order.ProductInfos {
		if productInfo.PlanName == sub.Plan {
			return productInfo.PricingName, nil
		}
	}
	return "", nil
}

// chargeSubscriptionRenewal pays for the next period with the payment method of the current one and adds its subscription
func chargeSubscriptionRenewal(sub *Subscription) (*Subscription, error) {
	payment, provider, err := getSubscriptionPaymentProvider(sub)
	if err != nil {
		return nil, err
	}

	plan, err := getPlan(sub.Owner, sub.Plan)
	if err != nil {
		return nil, err
	}
	if plan == nil {
		return nil, util.MessageErrorf("order:the plan: %s does not exist", sub.Plan)
	}
	if !plan.IsEnabled {
		return nil, util.MessageErrorf("subscription:the plan: %s is disabled", plan.Name)
	}

	product, err := getProduct(sub.Owner, plan.Product)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, util.MessageErrorf("product:the product: %s does not exist", plan.Product)
	}

	user, err := getUser(payment.Owner, payment.User)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, util.MessageErrorf("order:the user: %s does not exist", payment.User)
	}

	startTime, err := time.Parse(time.RFC3339, sub.EndTime)
	if err != nil {
		return nil, err
	}

	pricingName, err := getRenewalPricingName(sub, payment)
	if err != nil {
		return nil, err
	}

	currency := product.Currency
	if currency == "" {
		currency = "USD"
	}

	productInfos := []ProductInfo{{
		Owner:       product.Owner,
		Name:        product.Name,
		DisplayName: product.DisplayName,
		Image:       product.Image,
		Detail:      product.Detail,
		Price:       product.Price,
		Currency:    product.Currency,
		Quantity:    1,
		PricingName: pricingName,
		PlanName:    plan.Name,
	}}

	paymentName := fmt.Sprintf("payment_%v", util.GenerateTimeId())
	orderName := fmt.Sprintf("order_%v", util.GenerateTimeId())
	order := &Order{
		Owner:        sub.Owner,
		Name:         orderName,
		DisplayName:  orderName,
		CreatedTime:  util.GetCurrentTime(),
		Products:     []string{product.Name},
		ProductInfos: productInfos,
		User:         payment.User,
		Payment:      paymentName,
		Price:        product.Price,
		Currency:     currency,
		State:        "Created",
		Message:      fmt.Sprintf("Auto renewal of subscription: %s", sub.Name),
	}
	renewalPayment := &Payment{
		Owner:       sub.Owner,
		Name:        paymentName,
		CreatedTime: util.GetCurrentTime(),
		DisplayName: paymentName,

		Provider: provider.Name,
		Type:     provider.Type,

		Products:            []string{product.Name},
		ProductsDisplayName: product.DisplayName,
		Currency:            currency,
		Price:               product.Price,

		User:    payment.User,
		Order:   orderName,
		State:   pp.PaymentStateCreated,
		Message: order.Message,
	}

	_, err = AddOrder(order)
	if err != nil {
		return nil, err
	}
	_, err = AddPayment(renewalPayment)
	if err != nil {
		return nil, err
	}

	chargeErr := chargeRenewalPayment(provider, payment, renewalPayment, user)
	if chargeErr != nil {
		renewalPayment.State = pp.PaymentStateError
		renewalPayment.Message = fmt.Sprintf("%s failed: %s", order.Message, chargeErr.Error())
		order.State = "Failed"
		order.Message = renewalPayment.Message
	} else {
		renewalPayment.State = pp.PaymentStatePaid
		order.State = "Paid"
	}
	order.UpdateTime = util.GetCurrentTime()

	_, err = ormer.Engine.ID(core.PK{renewalPayment.Owner, renewalPayment.Name}).Cols("state", "message", "out_order_id").Update(renewalPayment)
	if err != nil {
		return nil, err
	}
	_, err = ormer.Engine.ID(core.PK{order.Owner, order.Name}).Cols("state", "message", "update_time").Update(order)
	if err != nil {
		return nil, err
	}
	if chargeErr != nil {
		return nil, chargeErr
	}

	err = UpdateProductStock(productInfos)
	if err != nil {
		fmt.Printf("chargeSubscriptionRenewal() error for subscription %s: %v\n", sub.GetId(), err)
	}

	nextSub, err := newSubscriptionFrom(sub.Owner, sub.User, plan.Name, paymentName, plan.Period, startTime)
	if err != nil {
		return nil, err
	}
	nextSub.Group = sub.Group
	nextSub.Pricing = pricingName
	nextSub.AutoRenew = true

	_, err = AddSubscription(nextSub)
	if err != nil {
		return nil, err
	}

	err = nextSub.UpdateState()
	if err != nil {
		return nil, err
	}
	return nextSub, nil
}

func chargeRenewalPayment(provider *Provider, payment *Payment, renewalPayment *Payment, user *User) error {
	transaction := &Transaction{
		Owner:       renewalPayment.Owner,
		CreatedTime: util.GetCurrentTime(),
		Application: user.SignupApplication,
		Amount:      -renewalPayment.Price,
		Currency:    renewalPayment.Currency,
		Payment:     renewalPayment.Name,
		Category:    TransactionCategoryPurchase,
		Type:        provider.Category,
		Subtype:     provider.Type,
		Provider:    provider.Name,
		Tag:         "User",
		User:        renewalPayment.User,
		State:       string(pp.PaymentStatePaid),
	}

	if provider.Type == "Balance" {
		renewalPayment.OutOrderId = renewalPayment.GetId()
		affected, err := AddInternalPaymentTransaction(transaction, "en")
		if err != nil {
			return err
		}
		if !affected {
			return util.MessageErrorf("order:failed to add transaction: %s", util.StructToJson(transaction))
		}
		return nil
	}

	renewable, err := getRenewablePaymentProvider(provider)
	if err != nil {
		return err
	}

	payResp, err := renewable.Renew(&pp.RenewReq{
		PaymentName:        renewalPayment.Name,
		ProductDisplayName: renewalPayment.ProductsDisplayName,
		Price:              renewalPayment.Price,
		Currency:           renewalPayment.Currency,
		PreviousOrderId:    payment.OutOrderId,
	})
	if err != nil {
		return err
	}
	renewalPayment.OutOrderId = payResp.OrderId

	_, err = AddExternalPaymentTransaction(transaction, "en")
	if err != nil {
		fmt.Printf("chargeRenewalPayment() error for payment %s: %v\n", renewalPayment.GetId(), err)
	}
	return nil
}

func extendSubscriptionsWithPaymentType(subscriptions []*Subscription) error {
	ownerPayments := map[string][]string{}
	for _, sub := range subscriptions {
		if sub.Payment != "" {
			ownerPayments[sub.Owner] = append(ownerPayments[sub.Owner], sub.Payment)
		}
	}

	paymentTypes := map[string]string{}
	for owner, paymentNames := range ownerPayments {
		payments := []*Payment{}
		err := ormer.Engine.Cols("owner", "name", "type").Where("owner = ?", owner).In("name", paymentNames).Find(&payments)
		if err != nil {
			return err
		}
		for _, payment := range payments {
			paymentTypes[payment.GetId()] = payment.Type
		}
	}

	for _, sub := range subscriptions {
		sub.PaymentType = paymentTypes[util.GetId(sub.Owner, sub.Payment)]
	}
	return nil
}
