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

import "net/http"

type PaymentProvider interface {
	Pay(providerName string, productName string, payerName string, paymentName string, productDisplayName string, price float64, returnUrl string, notifyUrl string) (string, error)
	Notify(request *http.Request, body []byte, authorityPublicKey string) (string, string, float64, string, string, error)
	GetInvoice(paymentName string, personName string, personIdCard string, personEmail string, personPhone string, invoiceType string, invoiceTitle string, invoiceTaxId string) (string, error)
}

func GetPaymentProvider(typ string, appId string, clientSecret string, host string, appCertificate string, appPrivateKey string, authorityPublicKey string, authorityRootPublicKey string, clientId2 string) (PaymentProvider, error) {
	if typ == "Alipay" {
		newAlipayPaymentProvider, err := NewAlipayPaymentProvider(appId, appCertificate, appPrivateKey, authorityPublicKey, authorityRootPublicKey)
		if err != nil {
			return nil, err
		}
		return newAlipayPaymentProvider, nil
	} else if typ == "GC" {
		return NewGcPaymentProvider(appId, clientSecret, host), nil
	} else if typ == "WeChat Pay" {
		// appId, mchId, mchCertSerialNumber, apiV3Key, privateKey
		newWechatPaymentProvider, err := NewWechatPaymentProvider(clientId2, appId, appCertificate, clientSecret, appPrivateKey)
		if err != nil {
			return nil, err
		}
		return newWechatPaymentProvider, nil
	}

	return nil, nil
}
