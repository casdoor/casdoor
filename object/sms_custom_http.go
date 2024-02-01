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

package object

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/casdoor/casdoor/proxy"
)

type HttpSmsClient struct {
	endpoint  string
	method    string
	paramName string
	template  string
}

func newHttpSmsClient(endpoint, method, paramName, template string) (*HttpSmsClient, error) {
	if template == "" {
		template = "%s"
	}
	client := &HttpSmsClient{
		endpoint:  endpoint,
		method:    method,
		paramName: paramName,
		template:  template,
	}
	return client, nil
}

func (c *HttpSmsClient) SendMessage(param map[string]string, targetPhoneNumber ...string) error {
	phoneNumber := targetPhoneNumber[0]
	code := param["code"]
	content := fmt.Sprintf(c.template, code)

	var req *http.Request
	var err error
	if c.method == "POST" {
		formValues := url.Values{}
		formValues.Set("phoneNumber", phoneNumber)
		formValues.Set(c.paramName, content)
		req, err = http.NewRequest(c.method, c.endpoint, strings.NewReader(formValues.Encode()))
		if err != nil {
			return err
		}

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	} else if c.method == "GET" {
		req, err = http.NewRequest(c.method, c.endpoint, nil)
		if err != nil {
			return err
		}

		q := req.URL.Query()
		q.Add("phoneNumber", phoneNumber)
		q.Add(c.paramName, content)
		req.URL.RawQuery = q.Encode()
	} else {
		return fmt.Errorf("HttpSmsClient's SendMessage() error, unsupported method: %s", c.method)
	}

	httpClient := proxy.DefaultHttpClient
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HttpSmsClient's SendMessage() error, custom HTTP SMS request failed with status: %s", resp.Status)
	}

	return err
}
