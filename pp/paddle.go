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
	"strings"

	"github.com/PaddleHQ/paddle-go-sdk"
	"github.com/casdoor/casdoor/conf"
)

type PaddlePaymentProvider struct {
	Client *paddle.SDK
}

func NewPaddlePaymentProvider(apiKey string) (*PaddlePaymentProvider, error) {
	var client *paddle.SDK
	var err error

	if conf.GetConfigString("runmode") == "prod" {
		client, err = paddle.New(apiKey)
	} else {
		client, err = paddle.NewSandbox(apiKey)
	}

	if err != nil {
		return nil, err
	}

	pp := &PaddlePaymentProvider{
		Client: client,
	}
	return pp, nil
}

func (pp *PaddlePaymentProvider) Pay(r *PayReq) (*PayResp, error) {
	ctx := context.Background()

	// Store product info in custom_data for later retrieval
	description := joinAttachString([]string{r.ProductName, r.ProductDisplayName, r.ProviderName})

	customData := paddle.CustomData{
		"payment_name":         r.PaymentName,
		"product_description":  description,
		"product_name":         r.ProductName,
		"product_display_name": r.ProductDisplayName,
		"provider_name":        r.ProviderName,
	}

	// Convert price to amount string in cents (lowest denomination)
	amountInCents := fmt.Sprintf("%d", priceFloat64ToInt64(r.Price))

	// Map currency string to paddle CurrencyCode
	currencyCode := paddle.CurrencyCode(strings.ToUpper(r.Currency))

	// Create a non-catalog price and product for this transaction
	items := []paddle.CreateTransactionItems{
		*paddle.NewCreateTransactionItemsNonCatalogPriceAndProduct(&paddle.NonCatalogPriceAndProduct{
			Quantity: 1,
			Price: paddle.TransactionPriceCreateWithProduct{
				Description: description,
				Name:        &r.ProductDisplayName,
				TaxMode:     paddle.TaxModeAccountSetting,
				UnitPrice: paddle.Money{
					Amount:       amountInCents,
					CurrencyCode: currencyCode,
				},
				Product: paddle.TransactionSubscriptionProductCreate{
					Name:        r.ProductDisplayName,
					Description: &r.ProductDescription,
					TaxCategory: paddle.TaxCategoryStandard,
					ImageURL:    &r.ProductImage,
					CustomData:  customData,
				},
			},
		}),
	}

	checkoutSettings := &paddle.TransactionCheckout{
		URL: &r.ReturnUrl,
	}

	req := &paddle.CreateTransactionRequest{
		Items:      items,
		CustomData: customData,
		Checkout:   checkoutSettings,
	}

	res, err := pp.Client.CreateTransaction(ctx, req)
	if err != nil {
		return nil, err
	}

	if res == nil {
		return nil, fmt.Errorf("paddle transaction response is nil")
	}

	// Get checkout URL from the transaction
	checkoutURL := ""
	if res.Checkout != nil && res.Checkout.URL != nil {
		checkoutURL = *res.Checkout.URL
	}

	payResp := &PayResp{
		PayUrl:  checkoutURL,
		OrderId: res.ID,
	}
	return payResp, nil
}

func (pp *PaddlePaymentProvider) Notify(body []byte, orderId string) (*NotifyResult, error) {
	ctx := context.Background()

	// Get transaction status
	req := &paddle.GetTransactionRequest{
		TransactionID: orderId,
	}
	res, err := pp.Client.GetTransaction(ctx, req)
	if err != nil {
		return nil, err
	}

	if res == nil {
		return nil, fmt.Errorf("paddle transaction not found for order: %s", orderId)
	}

	// Map Paddle status to payment state
	switch res.Status {
	case paddle.TransactionStatusDraft, paddle.TransactionStatusReady:
		return &NotifyResult{PaymentStatus: PaymentStateCreated}, nil
	case paddle.TransactionStatusCompleted, paddle.TransactionStatusPaid:
		// Payment successful, continue to extract payment details below
	case paddle.TransactionStatusBilled:
		// Billed but not yet paid
		return &NotifyResult{PaymentStatus: PaymentStateCreated}, nil
	case paddle.TransactionStatusCanceled:
		return &NotifyResult{PaymentStatus: PaymentStateCanceled, NotifyMessage: "Transaction canceled"}, nil
	case paddle.TransactionStatusPastDue:
		return &NotifyResult{PaymentStatus: PaymentStateError, NotifyMessage: "Payment past due"}, nil
	default:
		return &NotifyResult{PaymentStatus: PaymentStateError, NotifyMessage: fmt.Sprintf("unexpected paddle transaction status: %v", res.Status)}, nil
	}

	// Extract payment details from transaction for successful payment
	var (
		paymentName        string
		productName        string
		productDisplayName string
		providerName       string
	)

	if res.CustomData != nil {
		if v, ok := res.CustomData["payment_name"]; ok {
			if str, ok := v.(string); ok {
				paymentName = str
			}
		}
		if v, ok := res.CustomData["product_name"]; ok {
			if str, ok := v.(string); ok {
				productName = str
			}
		}
		if v, ok := res.CustomData["product_display_name"]; ok {
			if str, ok := v.(string); ok {
				productDisplayName = str
			}
		}
		if v, ok := res.CustomData["provider_name"]; ok {
			if str, ok := v.(string); ok {
				providerName = str
			}
		}
	}

	// Get price from transaction details
	var price float64
	var currency string

	if len(res.Details.LineItems) > 0 {
		// Get the total amount from transaction details
		price = priceStringToFloat64(res.Details.Totals.Total) / 100
		currency = string(res.CurrencyCode)
	}

	return &NotifyResult{
		PaymentName:   paymentName,
		PaymentStatus: PaymentStatePaid,

		ProductName:        productName,
		ProductDisplayName: productDisplayName,
		ProviderName:       providerName,

		Price:    price,
		Currency: currency,

		OrderId: orderId,
	}, nil
}

func (pp *PaddlePaymentProvider) GetInvoice(paymentName string, personName string, personIdCard string, personEmail string, personPhone string, invoiceType string, invoiceTitle string, invoiceTaxId string) (string, error) {
	return "", nil
}

func (pp *PaddlePaymentProvider) GetResponseError(err error) string {
	if err == nil {
		return "success"
	}
	return "fail"
}
