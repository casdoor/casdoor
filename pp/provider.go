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

type PaymentState string

const (
	PaymentStatePaid     PaymentState = "Paid"
	PaymentStateCreated  PaymentState = "Created"
	PaymentStateCanceled PaymentState = "Canceled"
	PaymentStateTimeout  PaymentState = "Timeout"
	PaymentStateError    PaymentState = "Error"
)

const (
	PaymentEnvWechatBrowser = "WechatBrowser"
)

type PayReq struct {
	ProviderName       string
	ProductName        string
	PayerName          string
	PayerId            string
	PaymentName        string
	ProductDisplayName string
	Price              float64
	Currency           string

	ReturnUrl string
	NotifyUrl string

	PaymentEnv string
}

type PayResp struct {
	PayUrl     string
	OrderId    string
	AttachInfo map[string]interface{}
}

type NotifyResult struct {
	PaymentName   string
	PaymentStatus PaymentState
	NotifyMessage string

	ProductName        string
	ProductDisplayName string
	ProviderName       string
	Price              float64
	Currency           string

	OrderId string
}

type PaymentProvider interface {
	Pay(req *PayReq) (*PayResp, error)
	Notify(body []byte, orderId string) (*NotifyResult, error)
	GetInvoice(paymentName string, personName string, personIdCard string, personEmail string, personPhone string, invoiceType string, invoiceTitle string, invoiceTaxId string) (string, error)
	GetResponseError(err error) string
}
