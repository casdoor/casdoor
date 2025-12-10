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

package pp

import (
	"context"
	"fmt"

	"github.com/adyen/adyen-go-api-library/v11/src/adyen"
	"github.com/adyen/adyen-go-api-library/v11/src/checkout"
	"github.com/adyen/adyen-go-api-library/v11/src/common"
	"github.com/casdoor/casdoor/conf"
)

type AdyenPaymentProvider struct {
	Client          *adyen.APIClient
	MerchantAccount string
}

func NewAdyenPaymentProvider(apiKey string, merchantAccount string) (*AdyenPaymentProvider, error) {
	config := common.Config{
		ApiKey:      apiKey,
		Environment: common.TestEnv,
	}

	if conf.GetConfigString("runmode") == "prod" {
		config.Environment = common.LiveEnv
	}

	client := adyen.NewClient(&config)

	pp := &AdyenPaymentProvider{
		Client:          client,
		MerchantAccount: merchantAccount,
	}
	return pp, nil
}

func (pp *AdyenPaymentProvider) Pay(r *PayReq) (*PayResp, error) {
	ctx := context.Background()

	// Store product info in metadata for later retrieval
	description := joinAttachString([]string{r.ProductName, r.ProductDisplayName, r.ProviderName})

	// Convert price to amount in minor units (cents)
	amountValue := priceFloat64ToInt64(r.Price)

	// Create payment session request
	sessionReq := checkout.CreateCheckoutSessionRequest{
		Amount: checkout.Amount{
			Currency: r.Currency,
			Value:    amountValue,
		},
		MerchantAccount: pp.MerchantAccount,
		Reference:       r.PaymentName,
		ReturnUrl:       r.ReturnUrl,
		Metadata: &map[string]string{
			"payment_name":        r.PaymentName,
			"product_description": description,
		},
	}

	service := pp.Client.Checkout()
	req := service.PaymentsApi.SessionsInput()
	req = req.CreateCheckoutSessionRequest(sessionReq)

	res, httpRes, err := service.PaymentsApi.Sessions(ctx, req)
	if err != nil {
		return nil, err
	}
	if httpRes.StatusCode != 200 && httpRes.StatusCode != 201 {
		return nil, fmt.Errorf("adyen session creation failed with status: %d", httpRes.StatusCode)
	}

	payUrl := ""
	if res.Url != nil {
		payUrl = *res.Url
	}

	payResp := &PayResp{
		PayUrl:  payUrl,
		OrderId: res.Id,
	}
	return payResp, nil
}

func (pp *AdyenPaymentProvider) Notify(body []byte, orderId string) (*NotifyResult, error) {
	ctx := context.Background()

	// Get payment session result using session ID
	service := pp.Client.Checkout()
	req := service.PaymentsApi.GetResultOfPaymentSessionInput(orderId)

	res, httpRes, err := service.PaymentsApi.GetResultOfPaymentSession(ctx, req)
	if err != nil {
		return nil, err
	}
	if httpRes.StatusCode != 200 {
		return nil, fmt.Errorf("adyen session result request failed with status: %d", httpRes.StatusCode)
	}

	notifyResult := &NotifyResult{
		OrderId: orderId,
	}

	// Map Adyen session status to payment state
	if res.Status != nil {
		switch *res.Status {
		case "completed":
			notifyResult.PaymentStatus = PaymentStatePaid
		case "paymentPending", "active":
			notifyResult.PaymentStatus = PaymentStateCreated
			return notifyResult, nil
		case "canceled":
			notifyResult.PaymentStatus = PaymentStateCanceled
			notifyResult.NotifyMessage = "Payment cancelled"
			return notifyResult, nil
		case "refused":
			notifyResult.PaymentStatus = PaymentStateError
			notifyResult.NotifyMessage = "Payment refused"
			return notifyResult, nil
		case "expired":
			notifyResult.PaymentStatus = PaymentStateTimeout
			notifyResult.NotifyMessage = "Session expired"
			return notifyResult, nil
		default:
			notifyResult.PaymentStatus = PaymentStateError
			notifyResult.NotifyMessage = fmt.Sprintf("unexpected adyen session status: %s", *res.Status)
			return notifyResult, nil
		}
	}

	// Note: SessionResultResponse doesn't include detailed payment information like
	// amount, currency, or metadata. This information is stored when the payment is
	// created and retrieved from the database based on orderId (session ID).
	// The payment name, product details, price, and currency will be populated
	// by the calling code from the stored payment record.

	return notifyResult, nil
}

func (pp *AdyenPaymentProvider) GetInvoice(paymentName string, personName string, personIdCard string, personEmail string, personPhone string, invoiceType string, invoiceTitle string, invoiceTaxId string) (string, error) {
	// Adyen does not provide a direct API for invoice generation
	// Invoicing should be handled separately through Adyen's merchant portal or third-party systems
	return "", nil
}

func (pp *AdyenPaymentProvider) GetResponseError(err error) string {
	if err == nil {
		return "success"
	}
	// Return the error message for better debugging
	return fmt.Sprintf("fail: %s", err.Error())
}
