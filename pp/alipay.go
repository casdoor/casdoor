// Copyright 2022 The Casdoor Authors. All Rights Reserved.
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
	"encoding/json"
	"fmt"

	"github.com/go-pay/gopay"
	"github.com/go-pay/gopay/alipay"
)

type AlipayPaymentProvider struct {
	Client *alipay.Client
}

func NewAlipayPaymentProvider(appId string, appCertificate string, appPrivateKey string, authorityPublicKey string, authorityRootPublicKey string) (*AlipayPaymentProvider, error) {
	// clientId => appId
	// cert.Certificate => appCertificate
	// cert.PrivateKey => appPrivateKey
	// rootCert.Certificate => authorityPublicKey
	// rootCert.PrivateKey => authorityRootPublicKey
	pp := &AlipayPaymentProvider{}

	client, err := alipay.NewClient(appId, appPrivateKey, true)
	if err != nil {
		return nil, err
	}

	err = client.SetCertSnByContent([]byte(appCertificate), []byte(authorityRootPublicKey), []byte(authorityPublicKey))
	if err != nil {
		return nil, err
	}

	pp.Client = client
	return pp, nil
}

func (pp *AlipayPaymentProvider) Pay(r *PayReq) (*PayResp, error) {
	// pp.Client.DebugSwitch = gopay.DebugOn
	bm := gopay.BodyMap{}
	pp.Client.SetReturnUrl(r.ReturnUrl)
	pp.Client.SetNotifyUrl(r.NotifyUrl)
	bm.Set("subject", joinAttachString([]string{r.ProductName, r.ProductDisplayName, r.ProviderName}))
	bm.Set("out_trade_no", r.PaymentName)
	bm.Set("total_amount", priceFloat64ToString(r.Price))

	payUrl, err := pp.Client.TradePagePay(context.Background(), bm)
	if err != nil {
		return nil, err
	}
	payResp := &PayResp{
		PayUrl:  payUrl,
		OrderId: r.PaymentName,
	}
	return payResp, nil
}

func (pp *AlipayPaymentProvider) Notify(body []byte, orderId string) (*NotifyResult, error) {
	bm := gopay.BodyMap{}
	bm.Set("out_trade_no", orderId)
	aliRsp, err := pp.Client.TradeQuery(context.Background(), bm)
	notifyResult := &NotifyResult{}
	if err != nil {
		errRsp := &alipay.ErrorResponse{}
		unmarshalErr := json.Unmarshal([]byte(err.Error()), errRsp)
		if unmarshalErr != nil {
			return nil, err
		}
		if errRsp.SubCode == "ACQ.TRADE_NOT_EXIST" {
			notifyResult.PaymentStatus = PaymentStateCanceled
			return notifyResult, nil
		}
		return nil, err
	}
	switch aliRsp.Response.TradeStatus {
	case "WAIT_BUYER_PAY":
		notifyResult.PaymentStatus = PaymentStateCreated
		return notifyResult, nil
	case "TRADE_CLOSED":
		notifyResult.PaymentStatus = PaymentStateTimeout
		return notifyResult, nil
	case "TRADE_SUCCESS":
		// skip
	default:
		notifyResult.PaymentStatus = PaymentStateError
		notifyResult.NotifyMessage = fmt.Sprintf("unexpected alipay trade state: %v", aliRsp.Response.TradeStatus)
		return notifyResult, nil
	}
	productDisplayName, productName, providerName, _ := parseAttachString(aliRsp.Response.Subject)
	notifyResult = &NotifyResult{
		ProductName:        productName,
		ProductDisplayName: productDisplayName,
		ProviderName:       providerName,
		OrderId:            orderId,
		PaymentStatus:      PaymentStatePaid,
		Price:              priceStringToFloat64(aliRsp.Response.TotalAmount),
		PaymentName:        orderId,
	}
	return notifyResult, nil
}

func (pp *AlipayPaymentProvider) GetInvoice(paymentName string, personName string, personIdCard string, personEmail string, personPhone string, invoiceType string, invoiceTitle string, invoiceTaxId string) (string, error) {
	return "", nil
}

func (pp *AlipayPaymentProvider) GetResponseError(err error) string {
	if err == nil {
		return "success"
	} else {
		return "fail"
	}
}
