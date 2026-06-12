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
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"strings"

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

func NewWechatPaymentProvider(mchId string, apiV3Key string, appId string, certificate string, privateKey string, wxPublicKey string, wxPublicKeyId string) (*WechatPaymentProvider, error) {
	// https://pay.weixin.qq.com/docs/merchant/products/native-payment/preparation.html
	// clientId         => mchId
	// clientSecret     => apiV3Key
	// clientId2        => appId
	// cert.certificate => merchant API certificate (its serial number is used for request signing)
	// cert.privateKey  => merchant API private key
	// content          => WeChat Pay public key (微信支付公钥, public-key mode only)
	// clientSecret2    => WeChat Pay public key ID (PUB_KEY_ID_..., public-key mode only)
	if appId == "" || mchId == "" || certificate == "" || apiV3Key == "" || privateKey == "" {
		return &WechatPaymentProvider{}, nil
	}

	// The certificate field may hold the merchant API certificate (PEM), from which the
	// serial number is extracted, or the certificate serial number string directly.
	serialNo := strings.TrimSpace(certificate)
	if strings.Contains(certificate, "BEGIN CERTIFICATE") {
		sn, err := getCertSerialNumber(certificate)
		if err != nil {
			return nil, err
		}
		serialNo = sn
	}

	clientV3, err := wechat.NewClientV3(mchId, serialNo, apiV3Key, privateKey)
	if err != nil {
		return nil, err
	}

	// Verify WeChat Pay responses/notifications:
	// - public-key mode (微信支付公钥): set the configured public key locally (no network call)
	// - platform-certificate mode (legacy): automatically fetch the platform certificates
	if wxPublicKey != "" && wxPublicKeyId != "" {
		err = clientV3.AutoVerifySignByPublicKey([]byte(wxPublicKey), wxPublicKeyId)
	} else {
		err = clientV3.AutoVerifySign()
	}
	if err != nil {
		return nil, err
	}

	pp := &WechatPaymentProvider{
		Client: clientV3,
		AppId:  appId,
	}

	return pp, nil
}

// getCertSerialNumber extracts the certificate serial number (uppercase hex) from a PEM-encoded certificate.
func getCertSerialNumber(certPem string) (string, error) {
	block, _ := pem.Decode([]byte(certPem))
	if block == nil {
		return "", fmt.Errorf("failed to decode the WeChat Pay merchant certificate PEM")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return "", err
	}
	return strings.ToUpper(cert.SerialNumber.Text(16)), nil
}

func (pp *WechatPaymentProvider) Pay(r *PayReq) (*PayResp, error) {
	bm := gopay.BodyMap{}
	desc := joinAttachString([]string{r.ProductDisplayName, r.ProductName, r.ProviderName})
	bm.Set("attach", desc)
	bm.Set("appid", pp.AppId)
	bm.Set("description", r.ProductDisplayName)
	bm.Set("notify_url", r.NotifyUrl)
	bm.Set("out_trade_no", r.PaymentName)
	bm.SetBodyMap("amount", func(bm gopay.BodyMap) {
		bm.Set("total", priceFloat64ToInt64(r.Price))
		bm.Set("currency", r.Currency)
	})
	// In Wechat browser, we use JSAPI
	if r.PaymentEnv == PaymentEnvWechatBrowser {
		if r.PayerId == "" {
			return nil, errors.New("failed to get the payer's openid, please retry login")
		}
		bm.SetBodyMap("payer", func(bm gopay.BodyMap) {
			bm.Set("openid", r.PayerId) // If the account is signup via Wechat, the PayerId is the Wechat OpenId e.g.oxW9O1ZDvgreSHuBSQDiQ2F055PI
		})
		jsapiRsp, err := pp.Client.V3TransactionJsapi(context.Background(), bm)
		if err != nil {
			return nil, err
		}
		if jsapiRsp.Code != wechat.Success {
			return nil, errors.New(jsapiRsp.Error)
		}
		// use RSA256 to sign the pay request
		params, err := pp.Client.PaySignOfJSAPI(pp.AppId, jsapiRsp.Response.PrepayId)
		if err != nil {
			return nil, err
		}
		payResp := &PayResp{
			PayUrl:  "",
			OrderId: r.PaymentName, // Wechat can use paymentName as the OutTradeNo to query order status
			AttachInfo: map[string]interface{}{
				"appId":     params.AppId,
				"timeStamp": params.TimeStamp,
				"nonceStr":  params.NonceStr,
				"package":   params.Package,
				"signType":  "RSA",
				"paySign":   params.PaySign,
			},
		}
		return payResp, nil
	} else if r.PaymentEnv == PaymentEnvWechatApp {
		appRsp, err := pp.Client.V3TransactionApp(context.Background(), bm)
		if err != nil {
			return nil, err
		}
		if appRsp.Code != wechat.Success {
			return nil, errors.New(appRsp.Error)
		}
		params, err := pp.Client.PaySignOfApp(pp.AppId, appRsp.Response.PrepayId)
		if err != nil {
			return nil, err
		}
		payResp := &PayResp{
			PayUrl:  "",
			OrderId: r.PaymentName,
			AttachInfo: map[string]interface{}{
				"appId":     params.Appid,
				"partnerId": params.Partnerid,
				"prepayId":  params.Prepayid,
				"package":   params.Package,
				"nonceStr":  params.Noncestr,
				"timeStamp": params.Timestamp,
				"sign":      params.Sign,
			},
		}
		return payResp, nil
	} else {
		// In other case, we use NativeAPI
		nativeRsp, err := pp.Client.V3TransactionNative(context.Background(), bm)
		if err != nil {
			return nil, err
		}
		if nativeRsp.Code != wechat.Success {
			return nil, errors.New(nativeRsp.Error)
		}
		payResp := &PayResp{
			PayUrl:  nativeRsp.Response.CodeUrl,
			OrderId: r.PaymentName, // Wechat can use paymentName as the OutTradeNo to query order status
		}
		return payResp, nil
	}
}

func (pp *WechatPaymentProvider) Notify(body []byte, orderId string) (*NotifyResult, error) {
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
