// Copyright 2024 The Casdoor Authors. All Rights Reserved.
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

	polargo "github.com/polarsource/polar-go"
	"github.com/polarsource/polar-go/models/components"
)

type PolarPaymentProvider struct {
	Client *polargo.Polar
}

func NewPolarPaymentProvider(accessToken string) (*PolarPaymentProvider, error) {
	client := polargo.New(
		polargo.WithSecurity(accessToken),
	)

	pp := &PolarPaymentProvider{
		Client: client,
	}
	return pp, nil
}

func (pp *PolarPaymentProvider) Pay(r *PayReq) (*PayResp, error) {
	ctx := context.Background()

	// Store product info in metadata for later retrieval
	description := joinAttachString([]string{r.ProductName, r.ProductDisplayName, r.ProviderName})

	metadata := map[string]components.CheckoutCreateMetadata{
		"payment_name":         components.CreateCheckoutCreateMetadataStr(r.PaymentName),
		"product_description":  components.CreateCheckoutCreateMetadataStr(description),
		"product_name":         components.CreateCheckoutCreateMetadataStr(r.ProductName),
		"product_display_name": components.CreateCheckoutCreateMetadataStr(r.ProductDisplayName),
		"provider_name":        components.CreateCheckoutCreateMetadataStr(r.ProviderName),
	}

	checkoutCreate := components.CheckoutCreate{
		CustomerName:  polargo.Pointer(r.PayerName),
		CustomerEmail: polargo.Pointer(r.PayerEmail),
		SuccessURL:    polargo.Pointer(r.ReturnUrl),
		Metadata:      metadata,
		Amount:        polargo.Pointer(priceFloat64ToInt64(r.Price)),
	}

	res, err := pp.Client.Checkouts.Create(ctx, checkoutCreate)
	if err != nil {
		return nil, err
	}

	if res.Checkout == nil {
		return nil, fmt.Errorf("polar checkout response is nil")
	}

	payResp := &PayResp{
		PayUrl:  res.Checkout.URL,
		OrderId: res.Checkout.ID,
	}
	return payResp, nil
}

func (pp *PolarPaymentProvider) Notify(body []byte, orderId string) (*NotifyResult, error) {
	ctx := context.Background()

	// Get checkout session status
	res, err := pp.Client.Checkouts.Get(ctx, orderId)
	if err != nil {
		return nil, err
	}

	if res.Checkout == nil {
		return nil, fmt.Errorf("polar checkout not found for order: %s", orderId)
	}

	checkout := res.Checkout

	// Map Polar status to payment state
	switch checkout.Status {
	case components.CheckoutStatusOpen:
		return &NotifyResult{PaymentStatus: PaymentStateCreated}, nil
	case components.CheckoutStatusSucceeded:
		// Payment successful, continue to extract payment details below
	case components.CheckoutStatusConfirmed:
		// Payment confirmed but not yet succeeded
		return &NotifyResult{PaymentStatus: PaymentStateCreated}, nil
	case components.CheckoutStatusExpired:
		return &NotifyResult{PaymentStatus: PaymentStateTimeout}, nil
	case components.CheckoutStatusFailed:
		return &NotifyResult{PaymentStatus: PaymentStateError, NotifyMessage: "Payment failed"}, nil
	default:
		return &NotifyResult{PaymentStatus: PaymentStateError, NotifyMessage: fmt.Sprintf("unexpected polar checkout status: %v", checkout.Status)}, nil
	}

	// Extract payment details from checkout for successful payment
	var (
		paymentName        string
		productName        string
		productDisplayName string
		providerName       string
	)

	if checkout.Metadata != nil {
		if v, ok := checkout.Metadata["payment_name"]; ok && v.Str != nil {
			paymentName = *v.Str
		}
		if v, ok := checkout.Metadata["product_name"]; ok && v.Str != nil {
			productName = *v.Str
		}
		if v, ok := checkout.Metadata["product_display_name"]; ok && v.Str != nil {
			productDisplayName = *v.Str
		}
		if v, ok := checkout.Metadata["provider_name"]; ok && v.Str != nil {
			providerName = *v.Str
		}
	}

	return &NotifyResult{
		PaymentName:   paymentName,
		PaymentStatus: PaymentStatePaid,

		ProductName:        productName,
		ProductDisplayName: productDisplayName,
		ProviderName:       providerName,

		Price:    priceInt64ToFloat64(checkout.TotalAmount),
		Currency: checkout.Currency,

		OrderId: orderId,
	}, nil
}

func (pp *PolarPaymentProvider) GetInvoice(paymentName string, personName string, personIdCard string, personEmail string, personPhone string, invoiceType string, invoiceTitle string, invoiceTaxId string) (string, error) {
	return "", nil
}

func (pp *PolarPaymentProvider) GetResponseError(err error) string {
	if err == nil {
		return "success"
	}
	return "fail"
}
