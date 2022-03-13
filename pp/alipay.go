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

	"github.com/go-pay/gopay"
	"github.com/go-pay/gopay/alipay"
)

type AlipayPaymentProvider struct {
	Client *alipay.Client
}

func NewAlipayPaymentProvider(appId string, appPublicKey string, appPrivateKey string, authorityPublicKey string, authorityRootPublicKey string) *AlipayPaymentProvider {
	pp := &AlipayPaymentProvider{}

	client, err := alipay.NewClient(appId, appPrivateKey, true)
	if err != nil {
		panic(err)
	}

	err = client.SetCertSnByContent([]byte(appPublicKey), []byte(authorityRootPublicKey), []byte(authorityPublicKey))
	if err != nil {
		panic(err)
	}

	pp.Client = client
	return pp
}

func (pp *AlipayPaymentProvider) Pay(productName string, productId string, providerId string, paymentId string, price float64, returnUrl string, notifyUrl string) (string, error) {
	//pp.Client.DebugSwitch = gopay.DebugOn

	bm := gopay.BodyMap{}

	bm.Set("return_url", returnUrl)
	bm.Set("notify_url", notifyUrl)

	bm.Set("subject", productName)
	bm.Set("out_trade_no", paymentId)
	bm.Set("total_amount", getPriceString(price))
	bm.Set("productId", productId)
	bm.Set("providerId", productId)

	payUrl, err := pp.Client.TradePagePay(context.Background(), bm)
	if err != nil {
		return "", err
	}
	return payUrl, nil
}
