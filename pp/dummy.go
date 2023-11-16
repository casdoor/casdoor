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

type DummyPaymentProvider struct{}

func NewDummyPaymentProvider() (*DummyPaymentProvider, error) {
	pp := &DummyPaymentProvider{}
	return pp, nil
}

func (pp *DummyPaymentProvider) Pay(r *PayReq) (*PayResp, error) {
	return &PayResp{
		PayUrl: r.ReturnUrl,
	}, nil
}

func (pp *DummyPaymentProvider) Notify(body []byte, orderId string) (*NotifyResult, error) {
	return &NotifyResult{
		PaymentStatus: PaymentStatePaid,
	}, nil
}

func (pp *DummyPaymentProvider) GetInvoice(paymentName string, personName string, personIdCard string, personEmail string, personPhone string, invoiceType string, invoiceTitle string, invoiceTaxId string) (string, error) {
	return "", nil
}

func (pp *DummyPaymentProvider) GetResponseError(err error) string {
	return ""
}
