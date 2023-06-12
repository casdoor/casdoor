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
	"net/http"

	"github.com/plutov/paypal/v4"
)

type PaypalPaymentProvider struct {
	Client *paypal.Client
}

func NewPaypalPaymentProvider(clientID string, secret string) (*PaypalPaymentProvider, error) {
	pp := &PaypalPaymentProvider{}

	client, err := paypal.NewClient(clientID, secret, paypal.APIBaseSandBox)
	if err != nil {
		return nil, err
	}

	pp.Client = client
	return pp, nil
}

func (pp *PaypalPaymentProvider) Pay(providerName string, productName string, payerName string, paymentName string, productDisplayName string, price float64, returnUrl string, notifyUrl string) (string, error) {
	// pp.Client.SetLog(os.Stdout) // Set log to terminal stdout

	receiverEmail := "sb-tmsqa26118644@business.example.com"

	amount := paypal.AmountPayout{
		Value:    fmt.Sprintf("%.2f", price),
		Currency: "USD",
	}

	description := fmt.Sprintf("%s-%s", providerName, productName)

	payout := paypal.Payout{
		SenderBatchHeader: &paypal.SenderBatchHeader{
			EmailSubject: description,
		},
		Items: []paypal.PayoutItem{
			{
				RecipientType: "EMAIL",
				Receiver:      receiverEmail,
				Amount:        &amount,
				Note:          description,
				SenderItemID:  description,
			},
		},
	}

	_, err := pp.Client.GetAccessToken(context.Background())
	if err != nil {
		return "", err
	}

	payoutResponse, err := pp.Client.CreatePayout(context.Background(), payout)
	if err != nil {
		return "", err
	}

	payUrl := payoutResponse.Links[0].Href
	return payUrl, nil
}

func (pp *PaypalPaymentProvider) Notify(request *http.Request, body []byte, authorityPublicKey string) (string, string, float64, string, string, error) {
	// The PayPal SDK does not directly support IPN verification.
	// So, you need to implement this part according to PayPal's IPN guide.
	return "", "", 0, "", "", nil
}

func (pp *PaypalPaymentProvider) GetInvoice(paymentName string, personName string, personIdCard string, personEmail string, personPhone string, invoiceType string, invoiceTitle string, invoiceTaxId string) (string, error) {
	return "", nil
}

func (pp *PaypalPaymentProvider) GetResponseError(err error) string {
	if err == nil {
		return "success"
	} else {
		return "fail"
	}
}
