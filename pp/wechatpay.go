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
	"fmt"
	"net/http"

	"github.com/casdoor/casdoor/util"
	"github.com/go-pay/gopay"
	"github.com/go-pay/gopay/wechat/v3"
)

type WechatPaymentProvider struct {
	ClientV3 *wechat.ClientV3
	appId    string
}

func NewWechatPaymentProvider(appId string, mchId string, mchCertSerialNumber string, apiV3Key string, privateKey string) *WechatPaymentProvider {
	pp := &WechatPaymentProvider{appId: appId}

	clientV3, err := wechat.NewClientV3(mchId, mchCertSerialNumber, apiV3Key, privateKey)
	if err != nil {
		panic(err)
	}

	err = clientV3.AutoVerifySign()
	if err != nil {
		panic(err)
	}
	pp.ClientV3 = clientV3
	return pp
}

func (pp *WechatPaymentProvider) Pay(providerName string, productName string, payerName string, paymentName string, productDisplayName string, price float64, returnUrl string, notifyUrl string) (string, error) {
	// pp.Client.DebugSwitch = gopay.DebugOn

	bm := gopay.BodyMap{}

	bm.Set("providerName", providerName)
	bm.Set("productName", productName)

	bm.Set("return_url", returnUrl)
	bm.Set("notify_url", notifyUrl)

	bm.Set("body", productDisplayName)
	bm.Set("out_trade_no", paymentName)
	bm.Set("total_fee", getPriceString(price))

	wechatRsp, err := pp.ClientV3.V3TransactionJsapi(context.Background(), bm)
	if err != nil {
		return "", err
	}

	payUrl := fmt.Sprintf("https://open.weixin.qq.com/connect/oauth2/authorize?appid=%s&redirect_uri=%s&response_type=code&scope=snsapi_base&state=STATE#wechat_redirect", pp.appId, wechatRsp.Response.PrepayId)
	return payUrl, nil
}

func (pp *WechatPaymentProvider) Notify(request *http.Request, body []byte, authorityPublicKey string) (string, string, float64, string, string, error) {
	bm, err := wechat.V3ParseNotifyToBodyMap(request)
	if err != nil {
		return "", "", 0, "", "", err
	}

	providerName := bm.Get("providerName")
	productName := bm.Get("productName")

	productDisplayName := bm.Get("body")
	paymentName := bm.Get("out_trade_no")
	price := util.ParseFloat(bm.Get("total_fee"))

	notifyReq, err := wechat.V3ParseNotify(request)
	if err != nil {
		panic(err)
	}

	cert := pp.ClientV3.WxPublicKey()

	err = notifyReq.VerifySignByPK(cert)
	if err != nil {
		return "", "", 0, "", "", err
	}

	return productDisplayName, paymentName, price, productName, providerName, nil
}

func (pp *WechatPaymentProvider) GetInvoice(paymentName string, personName string, personIdCard string, personEmail string, personPhone string, invoiceType string, invoiceTitle string, invoiceTaxId string) (string, error) {
	return "", nil
}
