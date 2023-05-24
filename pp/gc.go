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
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/casdoor/casdoor/util"
)

type GcPaymentProvider struct {
	Xmpch     string
	SecretKey string
	Host      string
}

type GcPayReqInfo struct {
	OrderDate string `json:"orderdate"`
	OrderNo   string `json:"orderno"`
	Amount    string `json:"amount"`
	Xmpch     string `json:"xmpch"`
	Body      string `json:"body"`
	ReturnUrl string `json:"return_url"`
	NotifyUrl string `json:"notify_url"`
	PayerId   string `json:"payerid"`
	PayerName string `json:"payername"`
	Remark1   string `json:"remark1"`
	Remark2   string `json:"remark2"`
}

type GcPayRespInfo struct {
	Jylsh     string `json:"jylsh"`
	Amount    string `json:"amount"`
	PayerId   string `json:"payerid"`
	PayerName string `json:"payername"`
	PayUrl    string `json:"payurl"`
}

type GcNotifyRespInfo struct {
	Xmpch      string  `json:"xmpch"`
	OrderDate  string  `json:"orderdate"`
	OrderNo    string  `json:"orderno"`
	Amount     float64 `json:"amount"`
	Jylsh      string  `json:"jylsh"`
	TradeNo    string  `json:"tradeno"`
	PayMethod  string  `json:"paymethod"`
	OrderState string  `json:"orderstate"`
	ReturnType string  `json:"return_type"`
	PayerId    string  `json:"payerid"`
	PayerName  string  `json:"payername"`
}

type GcRequestBody struct {
	Op          string `json:"op"`
	Xmpch       string `json:"xmpch"`
	Version     string `json:"version"`
	Data        string `json:"data"`
	RequestTime string `json:"requesttime"`
	Sign        string `json:"sign"`
}

type GcResponseBody struct {
	Op         string `json:"op"`
	Xmpch      string `json:"xmpch"`
	Version    string `json:"version"`
	ReturnCode string `json:"return_code"`
	ReturnMsg  string `json:"return_msg"`
	Data       string `json:"data"`
	NotifyTime string `json:"notifytime"`
	Sign       string `json:"sign"`
}

type GcInvoiceReqInfo struct {
	BusNo        string `json:"busno"`
	PayerName    string `json:"payername"`
	IdNum        string `json:"idnum"`
	PayerType    string `json:"payertype"`
	InvoiceTitle string `json:"invoicetitle"`
	Tin          string `json:"tin"`
	Phone        string `json:"phone"`
	Email        string `json:"email"`
}

type GcInvoiceRespInfo struct {
	BusNo     string `json:"busno"`
	State     string `json:"state"`
	EbillCode string `json:"ebillcode"`
	EbillNo   string `json:"ebillno"`
	CheckCode string `json:"checkcode"`
	Url       string `json:"url"`
	Content   string `json:"content"`
}

func NewGcPaymentProvider(clientId string, clientSecret string, host string) *GcPaymentProvider {
	pp := &GcPaymentProvider{}

	pp.Xmpch = clientId
	pp.SecretKey = clientSecret
	pp.Host = host
	return pp
}

func (pp *GcPaymentProvider) doPost(postBytes []byte) ([]byte, error) {
	client := &http.Client{}

	var resp *http.Response
	var err error

	contentType := "text/plain;charset=UTF-8"
	body := bytes.NewReader(postBytes)

	req, err := http.NewRequest("POST", pp.Host, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", contentType)

	resp, err = client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(resp.Body)

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return respBytes, nil
}

func (pp *GcPaymentProvider) Pay(providerName string, productName string, payerName string, paymentName string, productDisplayName string, price float64, returnUrl string, notifyUrl string) (string, error) {
	payReqInfo := GcPayReqInfo{
		OrderDate: util.GenerateSimpleTimeId(),
		OrderNo:   paymentName,
		Amount:    getPriceString(price),
		Xmpch:     pp.Xmpch,
		Body:      productDisplayName,
		ReturnUrl: returnUrl,
		NotifyUrl: notifyUrl,
		Remark1:   payerName,
		Remark2:   productName,
	}

	b, err := json.Marshal(payReqInfo)
	if err != nil {
		return "", err
	}

	body := GcRequestBody{
		Op:          "OrderCreate",
		Xmpch:       pp.Xmpch,
		Version:     "1.4",
		Data:        base64.StdEncoding.EncodeToString(b),
		RequestTime: util.GenerateSimpleTimeId(),
	}

	params := fmt.Sprintf("data=%s&op=%s&requesttime=%s&version=%s&xmpch=%s%s", body.Data, body.Op, body.RequestTime, body.Version, body.Xmpch, pp.SecretKey)
	body.Sign = strings.ToUpper(util.GetMd5Hash(params))

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	respBytes, err := pp.doPost(bodyBytes)
	if err != nil {
		return "", err
	}

	var respBody GcResponseBody
	err = json.Unmarshal(respBytes, &respBody)
	if err != nil {
		return "", err
	}

	if respBody.ReturnCode != "SUCCESS" {
		return "", fmt.Errorf("%s: %s", respBody.ReturnCode, respBody.ReturnMsg)
	}

	payRespInfoBytes, err := base64.StdEncoding.DecodeString(respBody.Data)
	if err != nil {
		return "", err
	}

	var payRespInfo GcPayRespInfo
	err = json.Unmarshal(payRespInfoBytes, &payRespInfo)
	if err != nil {
		return "", err
	}

	return payRespInfo.PayUrl, nil
}

func (pp *GcPaymentProvider) Notify(request *http.Request, body []byte, authorityPublicKey string, apiKey string) (string, string, float64, string, string, error) {
	reqBody := GcRequestBody{}
	m, err := url.ParseQuery(string(body))
	if err != nil {
		return "", "", 0, "", "", err
	}

	reqBody.Op = m["op"][0]
	reqBody.Xmpch = m["xmpch"][0]
	reqBody.Version = m["version"][0]
	reqBody.Data = m["data"][0]
	reqBody.RequestTime = m["requesttime"][0]
	reqBody.Sign = m["sign"][0]

	notifyReqInfoBytes, err := base64.StdEncoding.DecodeString(reqBody.Data)
	if err != nil {
		return "", "", 0, "", "", err
	}

	var notifyRespInfo GcNotifyRespInfo
	err = json.Unmarshal(notifyReqInfoBytes, &notifyRespInfo)
	if err != nil {
		return "", "", 0, "", "", err
	}

	providerName := ""
	productName := ""

	productDisplayName := ""
	paymentName := notifyRespInfo.OrderNo
	price := notifyRespInfo.Amount

	if notifyRespInfo.OrderState != "1" {
		return "", "", 0, "", "", fmt.Errorf("error order state: %s", notifyRespInfo.OrderDate)
	}

	return productDisplayName, paymentName, price, productName, providerName, nil
}

func (pp *GcPaymentProvider) GetInvoice(paymentName string, personName string, personIdCard string, personEmail string, personPhone string, invoiceType string, invoiceTitle string, invoiceTaxId string) (string, error) {
	payerType := "0"
	if invoiceType == "Organization" {
		payerType = "1"
	}

	invoiceReqInfo := GcInvoiceReqInfo{
		BusNo:        paymentName,
		PayerName:    personName,
		IdNum:        personIdCard,
		PayerType:    payerType,
		InvoiceTitle: invoiceTitle,
		Tin:          invoiceTaxId,
		Phone:        personPhone,
		Email:        personEmail,
	}

	b, err := json.Marshal(invoiceReqInfo)
	if err != nil {
		return "", err
	}

	body := GcRequestBody{
		Op:          "InvoiceEBillByOrder",
		Xmpch:       pp.Xmpch,
		Version:     "1.4",
		Data:        base64.StdEncoding.EncodeToString(b),
		RequestTime: util.GenerateSimpleTimeId(),
	}

	params := fmt.Sprintf("data=%s&op=%s&requesttime=%s&version=%s&xmpch=%s%s", body.Data, body.Op, body.RequestTime, body.Version, body.Xmpch, pp.SecretKey)
	body.Sign = strings.ToUpper(util.GetMd5Hash(params))

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	respBytes, err := pp.doPost(bodyBytes)
	if err != nil {
		return "", err
	}

	var respBody GcResponseBody
	err = json.Unmarshal(respBytes, &respBody)
	if err != nil {
		return "", err
	}

	if respBody.ReturnCode != "SUCCESS" {
		return "", fmt.Errorf("%s: %s", respBody.ReturnCode, respBody.ReturnMsg)
	}

	invoiceRespInfoBytes, err := base64.StdEncoding.DecodeString(respBody.Data)
	if err != nil {
		return "", err
	}

	var invoiceRespInfo GcInvoiceRespInfo
	err = json.Unmarshal(invoiceRespInfoBytes, &invoiceRespInfo)
	if err != nil {
		return "", err
	}

	if invoiceRespInfo.State == "0" {
		return "", fmt.Errorf("申请成功，开票中")
	}

	if invoiceRespInfo.Url == "" {
		return "", fmt.Errorf("invoice URL is empty")
	}

	return invoiceRespInfo.Url, nil
}
