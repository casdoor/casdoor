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
	"context"
	"fmt"

	"github.com/casdoor/casdoor/conf"
	polargo "github.com/polarsource/polar-go"
	"github.com/polarsource/polar-go/models/components"
)

type PolarPaymentProvider struct {
	AccessToken string
	Server      string
	isProd      bool
	client      *polargo.Polar
}

// IsSandbox returns true if the provider is using sandbox environment
func (pp *PolarPaymentProvider) IsSandbox() bool {
	return !pp.isProd
}

// GetEnvironment returns the current environment string
func (pp *PolarPaymentProvider) GetEnvironment() string {
	return pp.Server
}

// GetServerURL returns the Polar API server URL
func (pp *PolarPaymentProvider) GetServerURL() string {
	if pp.isProd {
		return "https://api.polar.sh"
	}
	return "https://sandbox-api.polar.sh"
}

func NewPolarPaymentProvider(AccessToken string) (*PolarPaymentProvider, error) {
	isProd := false
	if conf.GetConfigString("runmode") == "prod" {
		isProd = true
	}

	server := "sandbox"
	if isProd {
		server = "production"
	}

	client := polargo.New(
		polargo.WithSecurity(AccessToken),
		polargo.WithServer(server),
	)

	pp := &PolarPaymentProvider{
		AccessToken: AccessToken,
		Server:      server,
		isProd:      isProd,
		client:      client,
	}
	return pp, nil
}

func (pp *PolarPaymentProvider) Pay(r *PayReq) (*PayResp, error) {
	description := joinAttachString([]string{r.ProductName, r.ProductDisplayName, r.ProviderName})

	// Create checkout request
	amount := priceFloat64ToInt64(r.Price)
	checkoutCreate := components.CheckoutCreate{
		SuccessURL: polargo.Pointer(r.ReturnUrl),
		Amount:     &amount,
		Metadata: map[string]components.CheckoutCreateMetadata{
			"product_description": components.CreateCheckoutCreateMetadataStr(description),
			"payment_name":        components.CreateCheckoutCreateMetadataStr(r.PaymentName),
			"product_name":        components.CreateCheckoutCreateMetadataStr(r.ProductName),
			"provider_name":       components.CreateCheckoutCreateMetadataStr(r.ProviderName),
			"environment":         components.CreateCheckoutCreateMetadataStr(pp.Server),
			"is_test": components.CreateCheckoutCreateMetadataInteger(func() int64 {
				if !pp.isProd {
					return 1
				} else {
					return 0
				}
			}()),
		},
	}

	// Set customer information if available
	if r.PayerEmail != "" {
		checkoutCreate.CustomerEmail = &r.PayerEmail
	}
	if r.PayerName != "" {
		checkoutCreate.CustomerName = &r.PayerName
	}

	// Create checkout session
	ctx := context.Background()
	res, err := pp.client.Checkouts.Create(ctx, checkoutCreate)
	if err != nil {
		return nil, fmt.Errorf("failed to create Polar checkout: %w", err)
	}

	if res.Checkout == nil {
		return nil, fmt.Errorf("no checkout returned from Polar API")
	}

	payResp := &PayResp{
		PayUrl:  res.Checkout.URL,
		OrderId: res.Checkout.ID,
	}
	return payResp, nil
}

func (pp *PolarPaymentProvider) Notify(body []byte, orderId string) (*NotifyResult, error) {
	notifyResult := &NotifyResult{}

	// Always fetch fresh data from Polar API instead of trusting webhook body
	ctx := context.Background()
	res, err := pp.client.Checkouts.Get(ctx, orderId)
	if err != nil {
		return nil, fmt.Errorf("failed to get Polar checkout: %w", err)
	}

	if res.Checkout == nil {
		return nil, fmt.Errorf("no checkout returned from Polar API")
	}

	checkout := res.Checkout

	// Map Polar status to Casdoor payment states
	switch checkout.Status {
	case components.CheckoutStatusOpen:
		notifyResult.PaymentStatus = PaymentStateCreated
		return notifyResult, nil
	case components.CheckoutStatusExpired:
		notifyResult.PaymentStatus = PaymentStateTimeout
		return notifyResult, nil
	case components.CheckoutStatusConfirmed:
		notifyResult.PaymentStatus = PaymentStatePaid
	case components.CheckoutStatusSucceeded:
		notifyResult.PaymentStatus = PaymentStatePaid
	default:
		notifyResult.PaymentStatus = PaymentStateError
		notifyResult.NotifyMessage = fmt.Sprintf("unexpected Polar checkout status: %v", checkout.Status)
		return notifyResult, nil
	}

	// Extract metadata
	var (
		productName        string
		productDisplayName string
		providerName       string
		paymentName        string
	)

	if description, ok := checkout.Metadata["product_description"]; ok && description.Str != nil {
		productName, productDisplayName, providerName, _ = parseAttachString(*description.Str)
	}

	if name, ok := checkout.Metadata["payment_name"]; ok && name.Str != nil {
		paymentName = *name.Str
	}

	notifyResult = &NotifyResult{
		PaymentName:   paymentName,
		PaymentStatus: PaymentStatePaid,

		ProductName:        productName,
		ProductDisplayName: productDisplayName,
		ProviderName:       providerName,

		Price:    priceInt64ToFloat64(checkout.Amount),
		Currency: checkout.Currency,

		OrderId: orderId,
	}
	return notifyResult, nil
}

func (pp *PolarPaymentProvider) GetInvoice(paymentName string, personName string, personIdCard string, personEmail string, personPhone string, invoiceType string, invoiceTitle string, invoiceTaxId string) (string, error) {
	return "", nil
}

func (pp *PolarPaymentProvider) GetResponseError(err error) string {
	if err == nil {
		return "success"
	} else {
		return "fail"
	}
}
