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

//go:build !skipCi
// +build !skipCi

package object

import (
	"testing"

	"github.com/casdoor/casdoor/pp"
	"github.com/casdoor/casdoor/util"
)

func TestProduct(t *testing.T) {
	InitConfig()

	product, _ := GetProduct("admin/product_123")
	provider, _ := getProvider(product.Owner, "provider_pay_alipay")
	cert, _ := getCert(product.Owner, "cert-pay-alipay")
	pProvider, err := pp.GetPaymentProvider(provider.Type, provider.ClientId, provider.ClientSecret, provider.Host, cert.Certificate, cert.PrivateKey, cert.AuthorityPublicKey, cert.AuthorityRootPublicKey, provider.ClientId2)
	if err != nil {
		panic(err)
	}

	paymentName := util.GenerateTimeId()
	returnUrl := ""
	notifyUrl := ""
	payUrl, _, err := pProvider.Pay(provider.Name, product.Name, "alice", paymentName, product.DisplayName, product.Price, product.Currency, returnUrl, notifyUrl)
	if err != nil {
		panic(err)
	}

	println(payUrl)
}
