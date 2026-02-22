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
	"encoding/json"
	"fmt"
)

type DummyPaymentProvider struct{}

type DummyOrderInfo struct {
	Price              float64 `json:"price"`
	Currency           string  `json:"currency"`
	ProductDisplayName string  `json:"productDisplayName"`
}

func NewDummyPaymentProvider() (*DummyPaymentProvider, error) {
	pp := &DummyPaymentProvider{}
	return pp, nil
}

func (pp *DummyPaymentProvider) Pay(r *PayReq) (*PayResp, error) {
	// Encode payment information in OrderId for later retrieval in Notify.
	// Note: This is a test/mock provider and the OrderId is only used internally for testing.
	// Real payment providers would receive this information from their external payment gateway.
	orderInfo := DummyOrderInfo{
		Price:              r.Price,
		Currency:           r.Currency,
		ProductDisplayName: "",
	}
	orderInfoBytes, err := json.Marshal(orderInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to encode order info: %w", err)
	}

	return &PayResp{
		PayUrl:  r.ReturnUrl,
		OrderId: string(orderInfoBytes),
	}, nil
}

func (pp *DummyPaymentProvider) Notify(body []byte, orderId string) (*NotifyResult, error) {
	// Decode payment information from OrderId
	var orderInfo DummyOrderInfo
	if orderId != "" {
		err := json.Unmarshal([]byte(orderId), &orderInfo)
		if err != nil {
			return nil, fmt.Errorf("failed to decode order info: %w", err)
		}
	}

	return &NotifyResult{
		PaymentStatus:      PaymentStatePaid,
		Price:              orderInfo.Price,
		Currency:           orderInfo.Currency,
		ProductDisplayName: orderInfo.ProductDisplayName,
	}, nil
}

func (pp *DummyPaymentProvider) GetInvoice(paymentName string, personName string, personIdCard string, personEmail string, personPhone string, invoiceType string, invoiceTitle string, invoiceTaxId string) (string, error) {
	return "", nil
}

func (pp *DummyPaymentProvider) GetResponseError(err error) string {
	return ""
}
