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
	"io/ioutil"
	"net/http"
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
	PayerId   string `json:"payerid"`
	PayerName string `json:"payername"`
	Xmpch     string `json:"xmpch"`
	ReturnUrl string `json:"return_url"`
	NotifyUrl string `json:"notify_url"`
}

type GcPayRespInfo struct {
	Jylsh     string `json:"jylsh"`
	Amount    string `json:"amount"`
	PayerId   string `json:"payerid"`
	PayerName string `json:"payername"`
	PayUrl    string `json:"payurl"`
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

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return respBytes, nil
}

func (pp *GcPaymentProvider) Pay(productName string, productId string, providerId string, paymentId string, price float64, returnUrl string, notifyUrl string) (string, error) {
	payReqInfo := GcPayReqInfo{
		OrderDate: util.GenerateSimpleTimeId(),
		OrderNo:   util.GenerateTimeId(),
		Amount:    getPriceString(price),
		PayerId:   "",
		PayerName: "",
		Xmpch:     pp.Xmpch,
		ReturnUrl: returnUrl,
		NotifyUrl: notifyUrl,
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
