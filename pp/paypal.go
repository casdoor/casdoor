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
	"errors"
	"net/http"
	"strconv"

	"github.com/go-pay/gopay"
	"github.com/go-pay/gopay/paypal"
	"github.com/go-pay/gopay/pkg/util"
)

type PaypalPaymentProvider struct {
	Client *paypal.Client
}

func NewPaypalPaymentProvider(clientID string, secret string) (*PaypalPaymentProvider, error) {
	pp := &PaypalPaymentProvider{}

	client, err := paypal.NewClient(clientID, secret, false)
	if err != nil {
		return nil, err
	}

	pp.Client = client
	return pp, nil
}

func (pp *PaypalPaymentProvider) Pay(providerName string, productName string, payerName string, paymentName string, productDisplayName string, price float64, currency string, returnUrl string, notifyUrl string) (string, string, error) {
	// pp.Client.DebugSwitch = gopay.DebugOn // Set log to terminal stdout

	priceStr := strconv.FormatFloat(price, 'f', 2, 64)
	var pus []*paypal.PurchaseUnit
	item := &paypal.PurchaseUnit{
		ReferenceId: util.GetRandomString(16),
		Amount: &paypal.Amount{
			CurrencyCode: currency,
			Value:        priceStr,
		},
		Description: joinAttachString([]string{productDisplayName, productName, providerName}),
	}
	pus = append(pus, item)

	bm := make(gopay.BodyMap)
	bm.Set("intent", "CAPTURE")
	bm.Set("purchase_units", pus)
	bm.SetBodyMap("payment_source", func(b1 gopay.BodyMap) {
		b1.SetBodyMap("paypal", func(b2 gopay.BodyMap) {
			b2.Set("brand_name", "Casdoor")
			b2.Set("return_url", returnUrl)
		})
	})
	ppRsp, err := pp.Client.CreateOrder(context.Background(), bm)
	if err != nil {
		return "", "", err
	}
	if ppRsp.Code != paypal.Success {
		return "", "", errors.New(ppRsp.Error)
	}

	return ppRsp.Response.Links[1].Href, ppRsp.Response.Id, nil
}

func (pp *PaypalPaymentProvider) Notify(request *http.Request, body []byte, authorityPublicKey string, orderId string) (string, string, float64, string, string, error) {
	ppRsp, err := pp.Client.OrderCapture(context.Background(), orderId, nil)
	if err != nil {
		return "", "", 0, "", "", err
	}
	if ppRsp.Code != paypal.Success {
		return "", "", 0, "", "", errors.New(ppRsp.Error)
	}

	paymentName := ppRsp.Response.Id
	price, err := strconv.ParseFloat(ppRsp.Response.PurchaseUnits[0].Amount.Value, 64)
	if err != nil {
		return "", "", 0, "", "", err
	}

	productDisplayName, productName, providerName, err := parseAttachString(ppRsp.Response.PurchaseUnits[0].Description)
	if err != nil {
		return "", "", 0, "", "", err
	}

	return productDisplayName, paymentName, price, productName, providerName, nil
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
