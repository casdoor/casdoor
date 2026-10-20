// Copyright 2023 The Casdoor Authors. All Rights Reserved.
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

package pp

import (
	"fmt"
	"strings"
	"time"

	"github.com/casdoor/casdoor/conf"
	"github.com/stripe/stripe-go/v74"
	stripeCheckout "github.com/stripe/stripe-go/v74/checkout/session"
	stripeIntent "github.com/stripe/stripe-go/v74/paymentintent"
	stripePrice "github.com/stripe/stripe-go/v74/price"
	stripeProduct "github.com/stripe/stripe-go/v74/product"
)

type StripePaymentProvider struct {
	PublishableKey string
	SecretKey      string
	isProd         bool
}

func NewStripePaymentProvider(PublishableKey, SecretKey string) (*StripePaymentProvider, error) {
	isProd := false
	if conf.GetConfigString("runmode") == "prod" {
		isProd = true
	}
	pp := &StripePaymentProvider{
		PublishableKey: PublishableKey,
		SecretKey:      SecretKey,
		isProd:         isProd,
	}
	stripe.Key = pp.SecretKey
	return pp, nil
}

func (pp *StripePaymentProvider) Pay(r *PayReq) (*PayResp, error) {
	// Create a temp product
	attachString := joinAttachString(r)
	productParams := &stripe.ProductParams{
		Name: stripe.String(r.ProductDisplayName),
		DefaultPriceData: &stripe.ProductDefaultPriceDataParams{
			UnitAmount: stripe.Int64(priceFloat64ToInt64(r.Price)),
			Currency:   stripe.String(r.Currency),
		},
	}
	if r.ProductDescription != "" {
		productParams.Description = stripe.String(r.ProductDescription)
	}

	sProduct, err := stripeProduct.New(productParams)
	if err != nil {
		return nil, err
	}
	// Create a price for an existing product
	priceParams := &stripe.PriceParams{
		Currency:   stripe.String(r.Currency),
		UnitAmount: stripe.Int64(priceFloat64ToInt64(r.Price)),
		Product:    stripe.String(sProduct.ID),
	}
	sPrice, err := stripePrice.New(priceParams)
	if err != nil {
		return nil, err
	}
	// Create a Checkout Session
	checkoutParams := &stripe.CheckoutSessionParams{
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(sPrice.ID),
				Quantity: stripe.Int64(1),
			},
		},
		Mode:              stripe.String(string(stripe.CheckoutSessionModePayment)),
		SuccessURL:        stripe.String(r.ReturnUrl),
		CancelURL:         stripe.String(r.ReturnUrl),
		ClientReferenceID: stripe.String(r.PaymentName),
		ExpiresAt:         stripe.Int64(time.Now().Add(30 * time.Minute).Unix()),
	}
	if r.SaveForRenewal {
		checkoutParams.CustomerCreation = stripe.String(string(stripe.CheckoutSessionCustomerCreationAlways))
		checkoutParams.PaymentIntentData = &stripe.CheckoutSessionPaymentIntentDataParams{
			SetupFutureUsage: stripe.String(string(stripe.PaymentIntentSetupFutureUsageOffSession)),
		}
		if r.PayerEmail != "" {
			checkoutParams.CustomerEmail = stripe.String(r.PayerEmail)
		}
	}
	checkoutParams.AddMetadata("product_description", attachString)
	sCheckout, err := stripeCheckout.New(checkoutParams)
	if err != nil {
		return nil, err
	}
	payResp := &PayResp{
		PayUrl:  sCheckout.URL,
		OrderId: sCheckout.ID,
	}
	return payResp, nil
}

func (pp *StripePaymentProvider) Notify(body []byte, orderId string) (*NotifyResult, error) {
	notifyResult := &NotifyResult{}
	sCheckout, err := stripeCheckout.Get(orderId, nil)
	if err != nil {
		return nil, err
	}
	switch sCheckout.Status {
	case "open":
		// The checkout session is still in progress. Payment processing has not started
		notifyResult.PaymentStatus = PaymentStateCreated
		return notifyResult, nil
	case "complete":
		// The checkout session is complete. Payment processing may still be in progress
	case "expired":
		// The checkout session has expired. No further processing will occur
		notifyResult.PaymentStatus = PaymentStateTimeout
		return notifyResult, nil
	default:
		notifyResult.PaymentStatus = PaymentStateError
		notifyResult.NotifyMessage = fmt.Sprintf("unexpected stripe checkout status: %v", sCheckout.Status)
		return notifyResult, nil
	}
	switch sCheckout.PaymentStatus {
	case "paid":
		// Skip
	case "unpaid":
		notifyResult.PaymentStatus = PaymentStateCreated
		return notifyResult, nil
	default:
		notifyResult.PaymentStatus = PaymentStateError
		notifyResult.NotifyMessage = fmt.Sprintf("unexpected stripe checkout payment status: %v", sCheckout.PaymentStatus)
		return notifyResult, nil
	}
	// Once payment is successful, the Checkout Session will contain a reference to the successful `PaymentIntent`
	sIntent, err := stripeIntent.Get(sCheckout.PaymentIntent.ID, nil)
	if err != nil {
		return nil, err
	}
	attachInfo, err := parseAttachString(sCheckout.Metadata["product_description"])
	if err != nil {
		return nil, err
	}
	notifyResult = &NotifyResult{
		PaymentName:   sCheckout.ClientReferenceID,
		PaymentStatus: PaymentStatePaid,

		ProductName:        attachInfo.ProductName,
		ProductDisplayName: attachInfo.ProductDisplayName,
		ProviderName:       attachInfo.ProviderName,

		Price:    priceInt64ToFloat64(sIntent.Amount),
		Currency: string(sIntent.Currency),

		OrderId: orderId,
	}
	return notifyResult, nil
}

// getSavedPaymentMethod returns the customer and the payment method that a checkout session or a renewal saved for off-session payments
func (pp *StripePaymentProvider) getSavedPaymentMethod(orderId string) (string, string, error) {
	intentId := orderId
	isCheckout := strings.HasPrefix(orderId, "cs_")
	if isCheckout {
		sCheckout, err := stripeCheckout.Get(orderId, nil)
		if err != nil {
			return "", "", err
		}
		if sCheckout.PaymentIntent == nil {
			return "", "", fmt.Errorf("the stripe checkout session: %s has no payment intent", orderId)
		}
		intentId = sCheckout.PaymentIntent.ID
	}

	sIntent, err := stripeIntent.Get(intentId, nil)
	if err != nil {
		return "", "", err
	}
	if sIntent.Status != stripe.PaymentIntentStatusSucceeded || sIntent.Customer == nil || sIntent.PaymentMethod == nil ||
		(isCheckout && sIntent.SetupFutureUsage != stripe.PaymentIntentSetupFutureUsageOffSession) {
		return "", "", fmt.Errorf("the stripe payment: %s did not save a payment method for renewals, please buy it again with auto renewal on", orderId)
	}
	return sIntent.Customer.ID, sIntent.PaymentMethod.ID, nil
}

func (pp *StripePaymentProvider) CheckRenewal(orderId string) error {
	_, _, err := pp.getSavedPaymentMethod(orderId)
	return err
}

func (pp *StripePaymentProvider) Renew(r *RenewReq) (*PayResp, error) {
	customerId, paymentMethodId, err := pp.getSavedPaymentMethod(r.PreviousOrderId)
	if err != nil {
		return nil, err
	}

	intentParams := &stripe.PaymentIntentParams{
		Amount:        stripe.Int64(priceFloat64ToInt64(r.Price)),
		Currency:      stripe.String(r.Currency),
		Customer:      stripe.String(customerId),
		PaymentMethod: stripe.String(paymentMethodId),
		Description:   stripe.String(r.ProductDisplayName),
		OffSession:    stripe.Bool(true),
		Confirm:       stripe.Bool(true),
	}
	intentParams.AddMetadata("payment_name", r.PaymentName)
	intentParams.SetIdempotencyKey(r.PaymentName)
	sIntent, err := stripeIntent.New(intentParams)
	if err != nil {
		if stripeErr, ok := err.(*stripe.Error); ok && stripeErr.Msg != "" {
			return nil, fmt.Errorf("%s", stripeErr.Msg)
		}
		return nil, err
	}
	if sIntent.Status != stripe.PaymentIntentStatusSucceeded {
		return nil, fmt.Errorf("the stripe payment intent: %s is %s", sIntent.ID, sIntent.Status)
	}

	return &PayResp{OrderId: sIntent.ID}, nil
}

func (pp *StripePaymentProvider) GetInvoice(paymentName string, personName string, personIdCard string, personEmail string, personPhone string, invoiceType string, invoiceTitle string, invoiceTaxId string) (string, error) {
	return "", nil
}

func (pp *StripePaymentProvider) GetResponseError(err error) string {
	if err == nil {
		return "success"
	} else {
		return "fail"
	}
}
