// Copyright 2026 The Casdoor Authors. All Rights Reserved.
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

package action

import (
	"fmt"

	"github.com/casdoor/casdoor/pp"
)

type payReq struct {
	ProviderName       string  `json:"providerName"`
	ProductName        string  `json:"productName"`
	PayerName          string  `json:"payerName"`
	PayerId            string  `json:"payerId"`
	PayerEmail         string  `json:"payerEmail"`
	PaymentName        string  `json:"paymentName"`
	ProductDisplayName string  `json:"productDisplayName"`
	ProductDescription string  `json:"productDescription"`
	ProductImage       string  `json:"productImage"`
	Price              float64 `json:"price"`
	Currency           string  `json:"currency"`
	ReturnUrl          string  `json:"returnUrl"`
	NotifyUrl          string  `json:"notifyUrl"`
	PaymentEnv         string  `json:"paymentEnv"`
}

type payResp struct {
	PayUrl     string                 `json:"payUrl"`
	OrderId    string                 `json:"orderId"`
	AttachInfo map[string]interface{} `json:"attachInfo"`
}

type notifyReq struct {
	Body    []byte `json:"body"`
	OrderId string `json:"orderId"`
}

type notifyResult struct {
	PaymentName        string  `json:"paymentName"`
	PaymentStatus      string  `json:"paymentStatus"`
	NotifyMessage      string  `json:"notifyMessage"`
	ProductName        string  `json:"productName"`
	ProductDisplayName string  `json:"productDisplayName"`
	ProviderName       string  `json:"providerName"`
	Price              float64 `json:"price"`
	Currency           string  `json:"currency"`
	OrderId            string  `json:"orderId"`
}

// PaymentProvider is a payment provider whose code is an action module from the Marketplace.
type PaymentProvider struct {
	name   string
	module []byte
	hosts  []string
	config *Config
}

func NewPaymentProvider(name string, module []byte, hosts []string, config *Config) *PaymentProvider {
	return &PaymentProvider{name: name, module: module, hosts: hosts, config: config}
}

func (p *PaymentProvider) Pay(r *pp.PayReq) (*pp.PayResp, error) {
	req := &payReq{
		ProviderName:       r.ProviderName,
		ProductName:        r.ProductName,
		PayerName:          r.PayerName,
		PayerId:            r.PayerId,
		PayerEmail:         r.PayerEmail,
		PaymentName:        r.PaymentName,
		ProductDisplayName: r.ProductDisplayName,
		ProductDescription: r.ProductDescription,
		ProductImage:       r.ProductImage,
		Price:              r.Price,
		Currency:           r.Currency,
		ReturnUrl:          r.ReturnUrl,
		NotifyUrl:          r.NotifyUrl,
		PaymentEnv:         r.PaymentEnv,
	}

	resp := &payResp{}
	err := Call(p.name, p.module, p.hosts, "payment.pay", p.config, req, resp)
	if err != nil {
		return nil, err
	}
	if resp.PayUrl == "" || resp.OrderId == "" {
		return nil, fmt.Errorf("the action %s returned no payment URL or order ID", p.name)
	}
	return &pp.PayResp{PayUrl: resp.PayUrl, OrderId: resp.OrderId, AttachInfo: resp.AttachInfo}, nil
}

func (p *PaymentProvider) Notify(body []byte, orderId string) (*pp.NotifyResult, error) {
	res := &notifyResult{}
	err := Call(p.name, p.module, p.hosts, "payment.notify", p.config, &notifyReq{Body: body, OrderId: orderId}, res)
	if err != nil {
		return nil, err
	}

	state := pp.PaymentState(res.PaymentStatus)
	switch state {
	case pp.PaymentStatePaid, pp.PaymentStateCreated, pp.PaymentStateCanceled, pp.PaymentStateTimeout, pp.PaymentStateError:
	default:
		return nil, fmt.Errorf("the action %s returned an unknown payment state: %s", p.name, res.PaymentStatus)
	}

	return &pp.NotifyResult{
		PaymentName:        res.PaymentName,
		PaymentStatus:      state,
		NotifyMessage:      res.NotifyMessage,
		ProductName:        res.ProductName,
		ProductDisplayName: res.ProductDisplayName,
		ProviderName:       res.ProviderName,
		Price:              res.Price,
		Currency:           res.Currency,
		OrderId:            res.OrderId,
	}, nil
}

func (p *PaymentProvider) GetInvoice(paymentName string, personName string, personIdCard string, personEmail string, personPhone string, invoiceType string, invoiceTitle string, invoiceTaxId string) (string, error) {
	return "", nil
}

func (p *PaymentProvider) GetResponseError(err error) string {
	if err == nil {
		return "success"
	}
	return "fail"
}
