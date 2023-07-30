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
	"net/http"

	"github.com/casdoor/casdoor/util"
	"github.com/go-pay/gopay"
	"github.com/go-pay/gopay/alipay"
)

type AlipayPaymentProvider struct {
	Client *alipay.Client
}

func NewAlipayPaymentProvider(appId string, appCertificate string, appPrivateKey string, authorityPublicKey string, authorityRootPublicKey string) (*AlipayPaymentProvider, error) {
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

func (pp *AlipayPaymentProvider) Pay(providerName string, productName string, payerName string, paymentName string, productDisplayName string, price float64, currency string, returnUrl string, notifyUrl string) (string, string, error) {
	// pp.Client.DebugSwitch = gopay.DebugOn

	bm := gopay.BodyMap{}

	bm.Set("providerName", providerName)
	bm.Set("productName", productName)

	bm.Set("return_url", returnUrl)
	bm.Set("notify_url", notifyUrl)

	bm.Set("subject", productDisplayName)
	bm.Set("out_trade_no", paymentName)
	bm.Set("total_amount", getPriceString(price))

	payUrl, err := pp.Client.TradePagePay(context.Background(), bm)
	if err != nil {
		return "", "", err
	}
	return payUrl, "", nil
}

func (pp *AlipayPaymentProvider) Notify(request *http.Request, body []byte, authorityPublicKey string, orderId string) (*NotifyResult, error) {
	bm, err := alipay.ParseNotifyToBodyMap(request)
	if err != nil {
		return nil, err
	}

	providerName := bm.Get("providerName")
	productName := bm.Get("productName")

	productDisplayName := bm.Get("subject")
	paymentName := bm.Get("out_trade_no")
	price := util.ParseFloat(bm.Get("total_amount"))

	ok, err := alipay.VerifySignWithCert(authorityPublicKey, bm)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, err
	}
	notifyResult := &NotifyResult{
		ProductName:        productName,
		ProductDisplayName: productDisplayName,
		ProviderName:       providerName,
		OutOrderId:         orderId,
		PaymentStatus:      PaymentStatePaid,
		Price:              price,
		PaymentName:        paymentName,
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
