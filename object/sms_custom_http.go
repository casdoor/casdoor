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
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/casdoor/casdoor/proxy"
)

type HttpSmsClient struct {
	endpoint    string
	method      string
	paramName   string
	template    string
	httpHeaders map[string]string
	bodyMapping map[string]string
	contentType string
}

func newHttpSmsClient(endpoint, method, paramName, template string, httpHeaders map[string]string, bodyMapping map[string]string, contentType string) (*HttpSmsClient, error) {
	if template == "" {
		template = "%s"
	}
	if contentType == "" {
		contentType = "application/x-www-form-urlencoded"
	}
	client := &HttpSmsClient{
		endpoint:    endpoint,
		method:      method,
		paramName:   paramName,
		template:    template,
		httpHeaders: httpHeaders,
		bodyMapping: bodyMapping,
		contentType: contentType,
	}
	return client, nil
}

func (c *HttpSmsClient) SendMessage(param map[string]string, targetPhoneNumber ...string) error {
	phoneNumber := targetPhoneNumber[0]
	code := param["code"]
	content := fmt.Sprintf(c.template, code)

	phoneNumberField := "phoneNumber"
	contentField := c.paramName
	for k, v := range c.bodyMapping {
		switch k {
		case "phoneNumber":
			phoneNumberField = v
		case "content":
			contentField = v
		}
	}

	var req *http.Request
	var err error
	if c.method == "POST" || c.method == "PUT" || c.method == "DELETE" {
		bodyMap := make(map[string]string)
		bodyMap[phoneNumberField] = phoneNumber
		bodyMap[contentField] = content

		var bodyBytes []byte
		if c.contentType == "application/json" {
			bodyBytes, err = json.Marshal(bodyMap)
			if err != nil {
				return err
			}
			req, err = http.NewRequest(c.method, c.endpoint, bytes.NewBuffer(bodyBytes))
		} else {
			formValues := url.Values{}
			for k, v := range bodyMap {
				formValues.Add(k, v)
			}
			req, err = http.NewRequest(c.method, c.endpoint, strings.NewReader(formValues.Encode()))
		}

		if err != nil {
			return err
		}

		req.Header.Set("Content-Type", c.contentType)
	} else if c.method == "GET" {
		req, err = http.NewRequest(c.method, c.endpoint, nil)
		if err != nil {
			return err
		}

		q := req.URL.Query()
		q.Add(phoneNumberField, phoneNumber)
		q.Add(contentField, content)
		req.URL.RawQuery = q.Encode()
	} else {
		return fmt.Errorf("HttpSmsClient's SendMessage() error, unsupported method: %s", c.method)
	}

	for k, v := range c.httpHeaders {
		req.Header.Set(k, v)
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
