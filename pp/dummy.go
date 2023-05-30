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
	"fmt"
	"net/http"
)

type DummyPaymentProvider struct {
}

func NewDummyPaymentProvider() (*DummyPaymentProvider, error) {
	pp := &DummyPaymentProvider{}
	return pp, nil
}

func (pp *DummyPaymentProvider) Pay(providerName string, productName string, payerName string, paymentName string, productDisplayName string, price float64, returnUrl string, notifyUrl string) (string, error) {
	payUrl := fmt.Sprintf("/payments/%s/result", paymentName)
	return payUrl, nil
}

func (pp *DummyPaymentProvider) Notify(request *http.Request, body []byte, authorityPublicKey string) (string, string, float64, string, string, error) {
	return "", "", 0, "", "", nil
}

func (pp *DummyPaymentProvider) GetInvoice(paymentName string, personName string, personIdCard string, personEmail string, personPhone string, invoiceType string, invoiceTitle string, invoiceTaxId string) (string, error) {
	return "", nil
}

func (pp *DummyPaymentProvider) GetResponseError(err error) string {
	return ""
}
