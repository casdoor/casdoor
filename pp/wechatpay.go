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
	"errors"
	"fmt"
	"net/http"

	"github.com/casdoor/casdoor/util"
	"github.com/go-pay/gopay"
	"github.com/go-pay/gopay/wechat/v3"
)

type WechatPayNotifyResponse struct {
	Code    string `json:"Code"`
	Message string `json:"Message"`
}

type WechatPaymentProvider struct {
	Client *wechat.ClientV3
	AppId  string
}

func NewWechatPaymentProvider(mchId string, apiV3Key string, appId string, serialNo string, privateKey string) (*WechatPaymentProvider, error) {
	// https://pay.weixin.qq.com/docs/merchant/products/native-payment/preparation.html
	// clientId => mchId
	// clientSecret => apiV3Key
	// clientId2 => appId

	// appCertificate => serialNo
	// appPrivateKey => privateKey
	if appId == "" || mchId == "" || serialNo == "" || apiV3Key == "" || privateKey == "" {
		return &WechatPaymentProvider{}, nil
	}

	clientV3, err := wechat.NewClientV3(mchId, serialNo, apiV3Key, privateKey)
	if err != nil {
		return nil, err
	}

	platformCert, serialNo, err := clientV3.GetAndSelectNewestCert()
	if err != nil {
		return nil, err
	}
	pp := &WechatPaymentProvider{
		Client: clientV3.SetPlatformCert([]byte(platformCert), serialNo),
		AppId:  appId,
	}

	return pp, nil
}

func (pp *WechatPaymentProvider) Pay(providerName string, productName string, payerName string, paymentName string, productDisplayName string, price float64, currency string, returnUrl string, notifyUrl string) (string, string, error) {
	bm := gopay.BodyMap{}
	bm.Set("attach", joinAttachString([]string{productDisplayName, productName, providerName}))
	bm.Set("appid", pp.AppId)
	bm.Set("description", productDisplayName)
	bm.Set("notify_url", notifyUrl)
	bm.Set("out_trade_no", paymentName)
	bm.SetBodyMap("amount", func(bm gopay.BodyMap) {
		bm.Set("total", priceFloat64ToInt64(price))
		bm.Set("currency", currency)
	})

	nativeRsp, err := pp.Client.V3TransactionNative(context.Background(), bm)
	if err != nil {
		return "", "", err
	}
	if nativeRsp.Code != wechat.Success {
		return "", "", errors.New(nativeRsp.Error)
	}

	return nativeRsp.Response.CodeUrl, paymentName, nil // Wechat can use paymentName as the OutTradeNo to query order status
}

func (pp *WechatPaymentProvider) Notify(request *http.Request, body []byte, authorityPublicKey string, orderId string) (*NotifyResult, error) {
	notifyResult := &NotifyResult{}
	queryRsp, err := pp.Client.V3TransactionQueryOrder(context.Background(), wechat.OutTradeNo, orderId)
	if err != nil {
		return nil, err
	}
	if queryRsp.Code != wechat.Success {
		return nil, errors.New(queryRsp.Error)
	}

	switch queryRsp.Response.TradeState {
	case "SUCCESS":
		// skip
	case "CLOSED":
		notifyResult.PaymentStatus = PaymentStateCanceled
		return notifyResult, nil
	case "NOTPAY", "USERPAYING": // not-pad: waiting for user to pay; user-paying: user is paying
		notifyResult.PaymentStatus = PaymentStateCreated
		return notifyResult, nil
	default:
		notifyResult.PaymentStatus = PaymentStateError
		notifyResult.NotifyMessage = fmt.Sprintf("unexpected wechat trade state: %v", queryRsp.Response.TradeState)
		return notifyResult, nil
	}
	productDisplayName, productName, providerName, _ := parseAttachString(queryRsp.Response.Attach)
	notifyResult = &NotifyResult{
		ProductName:        productName,
		ProductDisplayName: productDisplayName,
		ProviderName:       providerName,
		OrderId:            orderId,
		Price:              priceInt64ToFloat64(int64(queryRsp.Response.Amount.Total)),
		PaymentStatus:      PaymentStatePaid,
		PaymentName:        queryRsp.Response.OutTradeNo,
	}
	return notifyResult, nil
}

func (pp *WechatPaymentProvider) GetInvoice(paymentName string, personName string, personIdCard string, personEmail string, personPhone string, invoiceType string, invoiceTitle string, invoiceTaxId string) (string, error) {
	return "", nil
}

func (pp *WechatPaymentProvider) GetResponseError(err error) string {
	response := &WechatPayNotifyResponse{
		Code:    "SUCCESS",
		Message: "",
	}

	if err != nil {
		response.Code = "FAIL"
		response.Message = err.Error()
	}

	return util.StructToJson(response)
}
