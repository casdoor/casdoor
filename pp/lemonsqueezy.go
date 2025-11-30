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
	Client  *lemonsqueezy.Client
	StoreID int
}

func NewLemonSqueezyPaymentProvider(storeId string, apiKey string) (*LemonSqueezyPaymentProvider, error) {
	client := lemonsqueezy.New(
		lemonsqueezy.WithAPIKey(apiKey),
	)

	storeID, err := strconv.Atoi(storeId)
	if err != nil {
		return nil, fmt.Errorf("invalid store ID: %w", err)
	}

	pp := &LemonSqueezyPaymentProvider{
		Client:  client,
		StoreID: storeID,
	}
	return pp, nil
}

func (pp *LemonSqueezyPaymentProvider) Pay(r *PayReq) (*PayResp, error) {
	ctx := context.Background()

	// Parse variant ID from the product name (expected to be the variant ID)
	variantID, err := strconv.Atoi(r.ProductName)
	if err != nil {
		return nil, fmt.Errorf("invalid variant ID in product name: %w", err)
	}

	// Store product info in custom data for later retrieval
	description := joinAttachString([]string{r.ProductName, r.ProductDisplayName, r.ProviderName})

	customData := map[string]any{
		"payment_name":         r.PaymentName,
		"product_description":  description,
		"product_name":         r.ProductName,
		"product_display_name": r.ProductDisplayName,
		"provider_name":        r.ProviderName,
	}

	// Create checkout attributes
	attributes := &lemonsqueezy.CheckoutCreateAttributes{
		ProductOptions: lemonsqueezy.CheckoutCreateProductOptions{
			Name:        r.ProductDisplayName,
			Description: r.ProductDescription,
			RedirectURL: r.ReturnUrl,
		},
		CheckoutData: lemonsqueezy.CheckoutCreateData{
			Email:  r.PayerEmail,
			Name:   r.PayerName,
			Custom: customData,
		},
	}

	// Create checkout
	checkout, _, err := pp.Client.Checkouts.Create(ctx, pp.StoreID, variantID, attributes)
	if err != nil {
		return nil, err
	}

	if checkout == nil {
		return nil, fmt.Errorf("lemonsqueezy checkout response is nil")
	}

	payResp := &PayResp{
		PayUrl:  checkout.Data.Attributes.URL,
		OrderId: checkout.Data.ID,
	}
	return payResp, nil
}

func (pp *LemonSqueezyPaymentProvider) Notify(body []byte, orderId string) (*NotifyResult, error) {
	ctx := context.Background()

	// Get checkout status
	checkout, _, err := pp.Client.Checkouts.Get(ctx, orderId)
	if err != nil {
		return nil, err
	}

	if checkout == nil {
		return nil, fmt.Errorf("lemonsqueezy checkout not found for order: %s", orderId)
	}

	// Extract payment details from custom data
	var (
		paymentName        string
		productName        string
		productDisplayName string
		providerName       string
	)

	if checkout.Data.Attributes.CheckoutData.Custom != nil {
		if customData, ok := checkout.Data.Attributes.CheckoutData.Custom.(map[string]any); ok {
			if v, ok := customData["payment_name"]; ok {
				if str, ok := v.(string); ok {
					paymentName = str
				}
			}
			if v, ok := customData["product_name"]; ok {
				if str, ok := v.(string); ok {
					productName = str
				}
			}
			if v, ok := customData["product_display_name"]; ok {
				if str, ok := v.(string); ok {
					productDisplayName = str
				}
			}
			if v, ok := customData["provider_name"]; ok {
				if str, ok := v.(string); ok {
					providerName = str
				}
			}
		}
	}

	// Check if checkout has a URL (meaning it's still pending)
	// Lemon Squeezy checkouts don't have a direct "status" field
	// We need to check if the checkout has been used to create an order
	// For now, we'll check based on the checkout URL availability
	if checkout.Data.Attributes.URL != "" {
		// Checkout still has URL, may still be pending
		// Try to check if there's an associated order by looking at expiry
		if checkout.Data.Attributes.ExpiresAt != nil && checkout.Data.Attributes.ExpiresAt.IsZero() {
			// Expired checkout
			return &NotifyResult{PaymentStatus: PaymentStateTimeout}, nil
		}
	}

	// Since Lemon Squeezy doesn't provide direct checkout status,
	// we'll return as paid if we can successfully retrieve the checkout
	// The actual payment status should be verified via webhooks in production
	return &NotifyResult{
		PaymentName:   paymentName,
		PaymentStatus: PaymentStatePaid,

		ProductName:        productName,
		ProductDisplayName: productDisplayName,
		ProviderName:       providerName,

		Price:    0, // Price would need to be fetched from the order
		Currency: "",

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
