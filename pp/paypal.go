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
	"fmt"
	"net/http"
	"strconv"

	"github.com/casdoor/casdoor/conf"

	"github.com/go-pay/gopay"
	"github.com/go-pay/gopay/paypal"
	"github.com/go-pay/gopay/pkg/util"
)

type PaypalPaymentProvider struct {
	Client *paypal.Client
}

func NewPaypalPaymentProvider(clientID string, secret string) (*PaypalPaymentProvider, error) {
	pp := &PaypalPaymentProvider{}
	isProd := false
	if conf.GetConfigString("runmode") == "prod" {
		isProd = true
	}
	client, err := paypal.NewClient(clientID, secret, isProd)
	//if !isProd {
	//	client.DebugSwitch = gopay.DebugOn
	//}
	if err != nil {
		return nil, err
	}

	pp.Client = client
	return pp, nil
}

func (pp *PaypalPaymentProvider) Pay(providerName string, productName string, payerName string, paymentName string, productDisplayName string, price float64, currency string, returnUrl string, notifyUrl string) (string, string, error) {
	// https://github.com/go-pay/gopay/blob/main/doc/paypal.md
	units := make([]*paypal.PurchaseUnit, 0, 1)
	unit := &paypal.PurchaseUnit{
		ReferenceId: util.GetRandomString(16),
		Amount: &paypal.Amount{
			CurrencyCode: currency,                    // e.g."USD"
			Value:        priceFloat64ToString(price), // e.g."100.00"
		},
		Description: joinAttachString([]string{productDisplayName, productName, providerName}),
	}
	units = append(units, unit)

	bm := make(gopay.BodyMap)
	bm.Set("intent", "CAPTURE")
	bm.Set("purchase_units", units)
	bm.SetBodyMap("application_context", func(b gopay.BodyMap) {
		b.Set("brand_name", "Casdoor")
		b.Set("locale", "en-PT")
		b.Set("return_url", returnUrl)
		b.Set("cancel_url", returnUrl)
	})

	ppRsp, err := pp.Client.CreateOrder(context.Background(), bm)
	if err != nil {
		return "", "", err
	}
	if ppRsp.Code != paypal.Success {
		return "", "", errors.New(ppRsp.Error)
	}
	// {"id":"9BR68863NE220374S","status":"CREATED",
	// "links":[{"href":"https://api.sandbox.paypal.com/v2/checkout/orders/9BR68863NE220374S","rel":"self","method":"GET"},
	// 			{"href":"https://www.sandbox.paypal.com/checkoutnow?token=9BR68863NE220374S","rel":"approve","method":"GET"},
	// 			{"href":"https://api.sandbox.paypal.com/v2/checkout/orders/9BR68863NE220374S","rel":"update","method":"PATCH"},
	// 			{"href":"https://api.sandbox.paypal.com/v2/checkout/orders/9BR68863NE220374S/capture","rel":"capture","method":"POST"}]}
	return ppRsp.Response.Links[1].Href, ppRsp.Response.Id, nil
}

func (pp *PaypalPaymentProvider) Notify(request *http.Request, body []byte, authorityPublicKey string, orderId string) (*NotifyResult, error) {
	notifyResult := &NotifyResult{}
	captureRsp, err := pp.Client.OrderCapture(context.Background(), orderId, nil)
	if err != nil {
		return nil, err
	}
	if captureRsp.Code != paypal.Success {
		errDetail := captureRsp.ErrorResponse.Details[0]
		switch errDetail.Issue {
		// If order is already captured, just skip this type of error and check the order detail
		case "ORDER_ALREADY_CAPTURED":
			// skip
		case "ORDER_NOT_APPROVED":
			notifyResult.PaymentStatus = PaymentStateCanceled
			notifyResult.NotifyMessage = errDetail.Description
			return notifyResult, nil
		default:
			err = fmt.Errorf(errDetail.Description)
			return nil, err
		}
	}
	// Check the order detail
	detailRsp, err := pp.Client.OrderDetail(context.Background(), orderId, nil)
	if err != nil {
		return nil, err
	}
	if detailRsp.Code != paypal.Success {
		errDetail := detailRsp.ErrorResponse.Details[0]
		switch errDetail.Issue {
		case "ORDER_NOT_APPROVED":
			notifyResult.PaymentStatus = PaymentStateCanceled
			notifyResult.NotifyMessage = errDetail.Description
			return notifyResult, nil
		default:
			err = fmt.Errorf(errDetail.Description)
			return nil, err
		}
	}

	paymentName := detailRsp.Response.Id
	price, err := strconv.ParseFloat(detailRsp.Response.PurchaseUnits[0].Amount.Value, 64)
	if err != nil {
		return nil, err
	}
	currency := detailRsp.Response.PurchaseUnits[0].Amount.CurrencyCode
	productDisplayName, productName, providerName, err := parseAttachString(detailRsp.Response.PurchaseUnits[0].Description)
	if err != nil {
		return nil, err
	}
	// TODO: status better handler, e.g.`hanging`
	var paymentStatus PaymentState
	switch detailRsp.Response.Status { // CREATED、SAVED、APPROVED、VOIDED、COMPLETED、PAYER_ACTION_REQUIRED
	case "COMPLETED":
		paymentStatus = PaymentStatePaid
	default:
		paymentStatus = PaymentStateError
	}
	notifyResult = &NotifyResult{
		PaymentStatus:      paymentStatus,
		PaymentName:        paymentName,
		ProductName:        productName,
		ProductDisplayName: productDisplayName,
		ProviderName:       providerName,
		Price:              price,
		Currency:           currency,

		OrderId: orderId,
	}
	return notifyResult, nil
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
