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

	"github.com/go-pay/gopay"
	"github.com/go-pay/gopay/wechat/v3"
)

const (
	wechatPayNotifySuccess = `{"Code": "SUCCESS", "Message": "成功"}`
	wechatPayNotifyFail    = `{"Code": "FAIL", "Message": "失败"}`
)

type WechatPaymentProvider struct {
	ClientV3 *wechat.ClientV3
	appId    string
}

func NewWechatPaymentProvider(appId string, mchId string, cert string, mchCertSerialNumber string, apiV3Key string, privateKey string) (*WechatPaymentProvider, error) {
	if appId == "" && mchId == "" && cert == "" && mchCertSerialNumber == "" && apiV3Key == "" && privateKey == "" {
		return &WechatPaymentProvider{}, nil
	}

	pp := &WechatPaymentProvider{appId: appId}

	clientV3, err := wechat.NewClientV3(mchId, mchCertSerialNumber, apiV3Key, privateKey)
	if err != nil {
		return nil, err
	}

	platformCert, serialNo, err := clientV3.GetAndSelectNewestCert()
	if err != nil {
		return nil, err
	}
	pp.ClientV3 = clientV3.SetPlatformCert([]byte(platformCert), serialNo)

	return pp, nil
}

func (pp *WechatPaymentProvider) Pay(providerName string, productName string, payerName string, paymentName string, productDisplayName string, price float64, returnUrl string, notifyUrl string) (string, error) {
	// pp.Client.DebugSwitch = gopay.DebugOn

	bm := gopay.BodyMap{}

	bm.Set("attach", getAttachString(productDisplayName, productName, providerName))

	bm.Set("appid", pp.appId)
	bm.Set("description", productDisplayName)
	bm.Set("notify_url", notifyUrl)

	bm.Set("out_trade_no", paymentName)
	bm.SetBodyMap("amount", func(bm gopay.BodyMap) {
		bm.Set("total", int(price*100))
		bm.Set("currency", "CNY")
	})

	wxRsp, err := pp.ClientV3.V3TransactionNative(context.Background(), bm)
	if err != nil {
		return "", err
	}
	if wxRsp.Code != wechat.Success {
		return "", fmt.Errorf("%s", wxRsp.Error)
	}

	return wxRsp.Response.CodeUrl, nil
}

func (pp *WechatPaymentProvider) Notify(request *http.Request, body []byte, authorityPublicKey string, apiKey string) (string, string, float64, string, string, error) {
	notifyReq, err := wechat.V3ParseNotify(request)
	if err != nil {
		panic(err)
	}

	cert := pp.ClientV3.WxPublicKey()
	err = notifyReq.VerifySignByPK(cert)
	if err != nil {
		return "", "", 0, "", "", err
	}

	result, err := notifyReq.DecryptCipherText(apiKey)
	if err != nil {
		return "", "", 0, "", "", err
	}

	paymentName := result.OutTradeNo
	price := float64(result.Amount.PayerTotal) / 100

	productDisplayName, productName, providerName, err := getInfoFromAttach(result.Attach)
	if err != nil {
		return "", "", 0, "", "", err
	}

	return productDisplayName, paymentName, price, productName, providerName, nil
}

func (pp *WechatPaymentProvider) GetInvoice(paymentName string, personName string, personIdCard string, personEmail string, personPhone string, invoiceType string, invoiceTitle string, invoiceTaxId string) (string, error) {
	return "", nil
}

func (pp *WechatPaymentProvider) GetResponseError(ok bool) string {
	if ok {
		return wechatPayNotifySuccess
	} else {
		return wechatPayNotifyFail
	}
}
