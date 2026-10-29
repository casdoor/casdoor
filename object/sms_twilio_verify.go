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

package object

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/casdoor/casdoor/proxy"
)

// TwilioVerify is the SMS provider type whose codes are generated, sent and checked by Twilio Verify,
// so its verification records keep an empty code until Twilio approves the one the user entered
const TwilioVerify = "Twilio Verify"

const twilioVerifyBaseUrl = "https://verify.twilio.com/v2/Services/"

type TwilioVerifyClient struct {
	accountSid string
	authToken  string
	serviceSid string
	httpClient *http.Client
}

type twilioVerifyResponse struct {
	Status  any    `json:"status"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func newTwilioVerifyClient(provider *Provider) (*TwilioVerifyClient, error) {
	if provider.ClientId == "" || provider.ClientSecret == "" || provider.AppId == "" {
		return nil, fmt.Errorf("the Account SID, Auth Token and Service SID of Twilio Verify are required")
	}

	httpClient := getTenantHttpClient(provider)
	if httpClient == nil {
		httpClient = proxy.DefaultHttpClient
		if provider.EnableProxy {
			httpClient = proxy.ProxyHttpClient
		}
	}

	client := &TwilioVerifyClient{
		accountSid: provider.ClientId,
		authToken:  provider.ClientSecret,
		serviceSid: provider.AppId,
		httpClient: httpClient,
	}
	return client, nil
}

func (c *TwilioVerifyClient) post(path string, form url.Values) (int, *twilioVerifyResponse, error) {
	endpoint := twilioVerifyBaseUrl + url.PathEscape(c.serviceSid) + "/" + path
	req, err := http.NewRequest("POST", endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return 0, nil, err
	}

	req.SetBasicAuth(c.accountSid, c.authToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, nil, err
	}

	res := &twilioVerifyResponse{}
	if err = json.Unmarshal(body, res); err != nil {
		return resp.StatusCode, nil, fmt.Errorf("Twilio Verify returned HTTP %d", resp.StatusCode)
	}
	return resp.StatusCode, res, nil
}

func (res *twilioVerifyResponse) error(statusCode int) error {
	if res.Message != "" {
		return fmt.Errorf("Twilio Verify error %d: %s", res.Code, res.Message)
	}
	return fmt.Errorf("Twilio Verify returned HTTP %d", statusCode)
}

// SendMessage starts a verification, Twilio generates the code itself so the code in param is not used
func (c *TwilioVerifyClient) SendMessage(param map[string]string, targetPhoneNumber ...string) error {
	if len(targetPhoneNumber) == 0 {
		return fmt.Errorf("missing parameter: targetPhoneNumber")
	}

	form := url.Values{}
	form.Set("To", targetPhoneNumber[0])
	form.Set("Channel", "sms")

	statusCode, res, err := c.post("Verifications", form)
	if err != nil {
		return err
	}
	if statusCode >= 300 {
		return res.error(statusCode)
	}
	return nil
}

// CheckCode reports whether Twilio approves the code, a missing verification (expired, already
// approved or out of attempts) counts as a wrong code
func (c *TwilioVerifyClient) CheckCode(phone string, code string) (bool, error) {
	form := url.Values{}
	form.Set("To", phone)
	form.Set("Code", code)

	statusCode, res, err := c.post("VerificationCheck", form)
	if err != nil {
		return false, err
	}
	if statusCode == http.StatusNotFound {
		return false, nil
	}
	if statusCode >= 300 {
		return false, res.error(statusCode)
	}

	status, _ := res.Status.(string)
	return status == "approved", nil
}

func checkTwilioVerifyCode(record *VerificationRecord, code string) (bool, error) {
	provider, err := getProvider(record.Owner, record.Provider)
	if err != nil {
		return false, err
	}
	if provider == nil || provider.Type != TwilioVerify {
		return false, fmt.Errorf("the Twilio Verify provider: %s of the verification code does not exist", record.Provider)
	}

	client, err := newTwilioVerifyClient(provider)
	if err != nil {
		return false, err
	}
	return client.CheckCode(record.Receiver, code)
}
