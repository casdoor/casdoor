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
	"github.com/casdoor/casdoor/conf"
	"github.com/stripe/stripe-go/v74"
	stripeCheckout "github.com/stripe/stripe-go/v74/checkout/session"
	stripePrice "github.com/stripe/stripe-go/v74/price"
	stripeProduct "github.com/stripe/stripe-go/v74/product"
	"net/http"
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
	return pp, nil
}

func (pp *StripePaymentProvider) Pay(providerName string, productName string, payerName string, paymentName string, productDisplayName string, price float64, currency string, returnUrl string, notifyUrl string) (payUrl string, orderId string, err error) {
	stripe.Key = pp.SecretKey
	// Create a product
	description := &PaymentDescription{
		ProductName:        productName,
		ProductDisplayName: productDisplayName,
		ProviderName:       providerName,
	}
	productParams := &stripe.ProductParams{
		Name:        stripe.String(productDisplayName),
		Description: stripe.String(description.String()),
		DefaultPriceData: &stripe.ProductDefaultPriceDataParams{
			UnitAmountDecimal: stripe.Float64(price),
			Currency:          stripe.String(currency),
		},
	}
	sProduct, err := stripeProduct.New(productParams)
	if err != nil {
		return "", "", err
	}
	// Create a price for an existing product
	priceParams := &stripe.PriceParams{
		Currency:          stripe.String(currency),
		UnitAmountDecimal: stripe.Float64(price),
		Product:           stripe.String(sProduct.ID),
	}
	sPrice, err := stripePrice.New(priceParams)
	if err != nil {
		return "", "", err
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
		SuccessURL:        stripe.String(returnUrl),
		CancelURL:         stripe.String(returnUrl),
		ClientReferenceID: stripe.String(paymentName),
	}
	sCheckout, err := stripeCheckout.New(checkoutParams)
	if err != nil {
		return "", "", err
	}
	return sCheckout.URL, sCheckout.ID, nil
}

func (pp *StripePaymentProvider) Notify(request *http.Request, body []byte, authorityPublicKey string, outOrderId string) (*NotifyResult, error) {
	notifyResult := &NotifyResult{}
	sCheckout, err := stripeCheckout.Get(outOrderId, nil)
	if err != nil {
		return nil, err
	}
	switch sCheckout.PaymentStatus {
	case "paid":
		// skip
	case "unpaid":
		notifyResult.PaymentStatus = PaymentStateCreated
		return notifyResult, nil
	default:
		notifyResult.PaymentStatus = PaymentStateError
		notifyResult.NotifyMessage = fmt.Sprintf("unexpected payment status: %v", sCheckout.PaymentStatus)
		return notifyResult, nil
	}
	// Once payment is successful, the Checkout Session will contain a reference to the successful `PaymentIntent`
	intent := sCheckout.PaymentIntent
	description := &PaymentDescription{}
	description.FromString(intent.Description)
	notifyResult = &NotifyResult{
		PaymentStatus:      PaymentStatePaid,
		PaymentName:        sCheckout.ClientReferenceID,
		PaymentDescription: description,
		Price:              float64(intent.Amount) / 100,
		Currency:           string(intent.Currency),
		OutOrderId:         outOrderId,
	}
	return notifyResult, nil
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
