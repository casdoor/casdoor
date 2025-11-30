// Copyright 2024 The Casdoor Authors. All Rights Reserved.
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
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type FastSpringPaymentProvider struct {
	ApiUsername    string
	ApiPassword    string
	StorefrontPath string
}

func NewFastSpringPaymentProvider(apiUsername string, apiPassword string, storefrontPath string) (*FastSpringPaymentProvider, error) {
	pp := &FastSpringPaymentProvider{
		ApiUsername:    apiUsername,
		ApiPassword:    apiPassword,
		StorefrontPath: storefrontPath,
	}
	return pp, nil
}

type fastSpringSessionRequest struct {
	Account *fastSpringAccount `json:"account,omitempty"`
	Items   []fastSpringItem   `json:"items"`
	Tags    map[string]string  `json:"tags,omitempty"`
}

type fastSpringAccount struct {
	Contact *fastSpringContact `json:"contact,omitempty"`
}

type fastSpringContact struct {
	Email string `json:"email,omitempty"`
	First string `json:"first,omitempty"`
	Last  string `json:"last,omitempty"`
}

type fastSpringItem struct {
	Product  string             `json:"product"`
	Quantity int                `json:"quantity"`
	Pricing  *fastSpringPricing `json:"pricing,omitempty"`
}

type fastSpringPricing struct {
	Price map[string]float64 `json:"price,omitempty"`
}

type fastSpringSessionResponse struct {
	ID         string `json:"id"`
	Account    string `json:"account"`
	AccountURL string `json:"accountUrl"`
}

type fastSpringOrderResponse struct {
	ID                    string                `json:"id"`
	Reference             string                `json:"reference"`
	Total                 float64               `json:"total"`
	TotalDisplay          string                `json:"totalDisplay"`
	TotalInPayoutCurrency float64               `json:"totalInPayoutCurrency"`
	Currency              string                `json:"currency"`
	PayoutCurrency        string                `json:"payoutCurrency"`
	Completed             bool                  `json:"completed"`
	Changed               int64                 `json:"changed"`
	Tags                  map[string]string     `json:"tags"`
	Items                 []fastSpringOrderItem `json:"items"`
}

type fastSpringOrderItem struct {
	Product  string  `json:"product"`
	Quantity int     `json:"quantity"`
	Display  string  `json:"display"`
	Subtotal float64 `json:"subtotal"`
	Discount float64 `json:"discount"`
	Price    float64 `json:"price"`
}

func (pp *FastSpringPaymentProvider) Pay(r *PayReq) (*PayResp, error) {
	description := joinAttachString([]string{r.ProductName, r.ProductDisplayName, r.ProviderName})

	// Create session request
	sessionReq := fastSpringSessionRequest{
		Items: []fastSpringItem{
			{
				Product:  r.ProductName,
				Quantity: 1,
				Pricing: &fastSpringPricing{
					Price: map[string]float64{
						strings.ToUpper(r.Currency): r.Price,
					},
				},
			},
		},
		Tags: map[string]string{
			"payment_name":         r.PaymentName,
			"product_description":  description,
			"product_name":         r.ProductName,
			"product_display_name": r.ProductDisplayName,
			"provider_name":        r.ProviderName,
		},
	}

	if r.PayerEmail != "" || r.PayerName != "" {
		sessionReq.Account = &fastSpringAccount{
			Contact: &fastSpringContact{
				Email: r.PayerEmail,
				First: r.PayerName,
			},
		}
	}

	jsonData, err := json.Marshal(sessionReq)
	if err != nil {
		return nil, err
	}

	// Create HTTP request to FastSpring API
	req, err := http.NewRequest("POST", "https://api.fastspring.com/sessions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(pp.ApiUsername, pp.ApiPassword)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("fastspring API error: %s", string(body))
	}

	var sessionResp fastSpringSessionResponse
	if err := json.Unmarshal(body, &sessionResp); err != nil {
		return nil, err
	}

	// Build checkout URL
	checkoutURL := fmt.Sprintf("https://%s/session/%s", pp.StorefrontPath, sessionResp.ID)

	payResp := &PayResp{
		PayUrl:  checkoutURL,
		OrderId: sessionResp.ID,
	}
	return payResp, nil
}

func (pp *FastSpringPaymentProvider) Notify(body []byte, orderId string) (*NotifyResult, error) {
	// Fetch order details from FastSpring API
	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.fastspring.com/orders/%s", orderId), nil)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(pp.ApiUsername, pp.ApiPassword)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusNotFound {
		// Order not found - still pending
		return &NotifyResult{PaymentStatus: PaymentStateCreated}, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fastspring API error: %s", string(respBody))
	}

	var orderResp fastSpringOrderResponse
	if err := json.Unmarshal(respBody, &orderResp); err != nil {
		return nil, err
	}

	// Check if order is completed
	if !orderResp.Completed {
		return &NotifyResult{PaymentStatus: PaymentStateCreated}, nil
	}

	// Extract payment details from tags
	var (
		paymentName        string
		productName        string
		productDisplayName string
		providerName       string
	)

	if orderResp.Tags != nil {
		paymentName = orderResp.Tags["payment_name"]
		productName = orderResp.Tags["product_name"]
		productDisplayName = orderResp.Tags["product_display_name"]
		providerName = orderResp.Tags["provider_name"]
	}

	return &NotifyResult{
		PaymentName:   paymentName,
		PaymentStatus: PaymentStatePaid,

		ProductName:        productName,
		ProductDisplayName: productDisplayName,
		ProviderName:       providerName,

		Price:    orderResp.Total,
		Currency: orderResp.Currency,

		OrderId: orderId,
	}, nil
}

func (pp *FastSpringPaymentProvider) GetInvoice(paymentName string, personName string, personIdCard string, personEmail string, personPhone string, invoiceType string, invoiceTitle string, invoiceTaxId string) (string, error) {
	return "", nil
}

func (pp *FastSpringPaymentProvider) GetResponseError(err error) string {
	if err == nil {
		return "success"
	}
	return "fail"
}
