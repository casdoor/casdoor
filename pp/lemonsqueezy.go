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
	"strconv"

	"github.com/NdoleStudio/lemonsqueezy-go"
)

type LemonSqueezyPaymentProvider struct {
	Client    *lemonsqueezy.Client
	StoreID   int
	VariantID int
}

func NewLemonSqueezyPaymentProvider(apiKey string, storeID string, variantID string) (*LemonSqueezyPaymentProvider, error) {
	client := lemonsqueezy.New(
		lemonsqueezy.WithAPIKey(apiKey),
	)

	storeIDInt, err := strconv.Atoi(storeID)
	if err != nil {
		return nil, fmt.Errorf("invalid store ID: %v", err)
	}

	variantIDInt, err := strconv.Atoi(variantID)
	if err != nil {
		return nil, fmt.Errorf("invalid variant ID: %v", err)
	}

	pp := &LemonSqueezyPaymentProvider{
		Client:    client,
		StoreID:   storeIDInt,
		VariantID: variantIDInt,
	}
	return pp, nil
}

func (pp *LemonSqueezyPaymentProvider) Pay(r *PayReq) (*PayResp, error) {
	ctx := context.Background()

	// Store product info in custom data for later retrieval
	description := joinAttachString([]string{r.ProductName, r.ProductDisplayName, r.ProviderName})

	customData := map[string]any{
		"payment_name":         r.PaymentName,
		"product_description":  description,
		"product_name":         r.ProductName,
		"product_display_name": r.ProductDisplayName,
		"provider_name":        r.ProviderName,
	}

	checkoutData := lemonsqueezy.CheckoutCreateData{
		Email:  r.PayerEmail,
		Name:   r.PayerName,
		Custom: customData,
	}

	productOptions := lemonsqueezy.CheckoutCreateProductOptions{
		Name:        r.ProductDisplayName,
		Description: r.ProductDescription,
		RedirectURL: r.ReturnUrl,
	}

	// Convert price from float64 to cents (int)
	customPrice := int(priceFloat64ToInt64(r.Price))

	attributes := &lemonsqueezy.CheckoutCreateAttributes{
		CustomPrice:    &customPrice,
		CheckoutData:   checkoutData,
		ProductOptions: productOptions,
	}

	checkout, _, err := pp.Client.Checkouts.Create(ctx, pp.StoreID, pp.VariantID, attributes)
	if err != nil {
		return nil, err
	}

	if checkout == nil {
		return nil, fmt.Errorf("lemon squeezy checkout response is nil")
	}

	payResp := &PayResp{
		PayUrl:  checkout.Data.Attributes.URL,
		OrderId: checkout.Data.ID,
	}
	return payResp, nil
}

func (pp *LemonSqueezyPaymentProvider) Notify(body []byte, orderId string) (*NotifyResult, error) {
	ctx := context.Background()

	// Get checkout session status
	checkout, _, err := pp.Client.Checkouts.Get(ctx, orderId)
	if err != nil {
		return nil, err
	}

	if checkout == nil {
		return nil, fmt.Errorf("lemon squeezy checkout not found for order: %s", orderId)
	}

	// Lemon Squeezy checkouts don't have a status field for payment status
	// We need to check if there's an associated order
	// For now, we'll use the checkout URL to determine if it's still pending
	// If the checkout exists, payment may still be pending
	// A completed payment would create an order

	// Extract payment details from checkout custom data
	var (
		paymentName        string
		productName        string
		productDisplayName string
		providerName       string
	)

	if checkout.Data.Attributes.CheckoutData.Custom != nil {
		customData, ok := checkout.Data.Attributes.CheckoutData.Custom.(map[string]any)
		if ok {
			if v, exists := customData["payment_name"]; exists {
				if str, ok := v.(string); ok {
					paymentName = str
				}
			}
			if v, exists := customData["product_name"]; exists {
				if str, ok := v.(string); ok {
					productName = str
				}
			}
			if v, exists := customData["product_display_name"]; exists {
				if str, ok := v.(string); ok {
					productDisplayName = str
				}
			}
			if v, exists := customData["provider_name"]; exists {
				if str, ok := v.(string); ok {
					providerName = str
				}
			}
		}
	}

	// Get price from checkout custom_price
	var price float64
	if checkout.Data.Attributes.CustomPrice != nil {
		if customPrice, ok := checkout.Data.Attributes.CustomPrice.(float64); ok {
			price = priceInt64ToFloat64(int64(customPrice))
		}
	}

	// Lemon Squeezy checkout doesn't have an explicit status field
	// The checkout URL indicates it's still valid/pending
	// For a completed payment, we'd typically receive a webhook
	// Here we return Created status as the checkout is still valid
	return &NotifyResult{
		PaymentName:   paymentName,
		PaymentStatus: PaymentStateCreated,

		ProductName:        productName,
		ProductDisplayName: productDisplayName,
		ProviderName:       providerName,

		Price:    price,
		Currency: "USD", // Lemon Squeezy primarily uses USD

		OrderId: orderId,
	}, nil
}

func (pp *LemonSqueezyPaymentProvider) GetInvoice(paymentName string, personName string, personIdCard string, personEmail string, personPhone string, invoiceType string, invoiceTitle string, invoiceTaxId string) (string, error) {
	return "", nil
}

func (pp *LemonSqueezyPaymentProvider) GetResponseError(err error) string {
	if err == nil {
		return "success"
	}
	return "fail"
}
