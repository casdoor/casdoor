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
	endpoint     string
	method       string
	paramName    string
	template     string
	bodyTemplate string
	httpHeaders  map[string]string
	bodyMapping  map[string]string
	contentType  string
	enableProxy  bool
	httpClient   *http.Client
}

func newHttpSmsClient(endpoint, method, paramName, template, bodyTemplate string, httpHeaders map[string]string, bodyMapping map[string]string, contentType string, enableProxy bool, httpClient *http.Client) (*HttpSmsClient, error) {
	if template == "" {
		template = "%s"
	}
	if contentType == "" {
		contentType = "application/x-www-form-urlencoded"
	}
	client := &HttpSmsClient{
		endpoint:     endpoint,
		method:       method,
		paramName:    paramName,
		template:     template,
		bodyTemplate: bodyTemplate,
		httpHeaders:  httpHeaders,
		bodyMapping:  bodyMapping,
		contentType:  contentType,
		enableProxy:  enableProxy,
		httpClient:   httpClient,
	}
	return client, nil
}

func escapeJsonString(value string) string {
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(value)
	if err != nil {
		return value
	}
	quoted := strings.TrimSuffix(buf.String(), "\n")
	return quoted[1 : len(quoted)-1]
}

func (c *HttpSmsClient) renderBodyTemplate(phoneNumber string, phoneNumberNoPlus string, code string, content string) string {
	escape := url.QueryEscape
	if c.contentType == "application/json" {
		escape = escapeJsonString
	}

	replacer := strings.NewReplacer(
		"{phoneNumberNoPlus}", escape(phoneNumberNoPlus),
		"{phoneNumber}", escape(phoneNumber),
		"{code}", escape(code),
		"{content}", escape(content),
	)
	return replacer.Replace(c.bodyTemplate)
}

func (c *HttpSmsClient) SendMessage(param map[string]string, targetPhoneNumber ...string) error {
	phoneNumber := targetPhoneNumber[0]
	phoneNumberNoPlus := strings.TrimPrefix(phoneNumber, "+")
	code := param["code"]
	content := fmt.Sprintf(c.template, code)

	phoneNumberField := "phoneNumber"
	contentField := c.paramName
	for k, v := range c.bodyMapping {
		switch k {
		case "phoneNumber":
			phoneNumberField = v
		case "content":
			if contentField == "" {
				contentField = v
			}
		}
	}

	// Replace {mobile}, {phoneNumber}, {phoneNumberNoPlus} and {code} placeholders in the endpoint URL
	endpoint := c.endpoint
	hasMobilePlaceholder := strings.Contains(endpoint, "{mobile}")
	hasCodePlaceholder := strings.Contains(endpoint, "{code}")
	hasPhoneNumberPlaceholder := strings.Contains(endpoint, "{phoneNumber}")
	hasPhoneNumberNoPlusPlaceholder := strings.Contains(endpoint, "{phoneNumberNoPlus}")
	endpoint = strings.NewReplacer(
		"{phoneNumberNoPlus}", phoneNumberNoPlus,
		"{mobile}", phoneNumber,
		"{phoneNumber}", phoneNumber,
		"{code}", code,
	).Replace(endpoint)

	var req *http.Request
	var err error
	if c.method == "POST" || c.method == "PUT" || c.method == "DELETE" {
		if c.bodyTemplate != "" {
			body := c.renderBodyTemplate(phoneNumber, phoneNumberNoPlus, code, content)
			req, err = http.NewRequest(c.method, endpoint, strings.NewReader(body))
			if err != nil {
				return err
			}
		} else {
			bodyMap := make(map[string]string)
			bodyMap[phoneNumberField] = phoneNumber
			bodyMap[contentField] = content

			var bodyBytes []byte
			if c.contentType == "application/json" {
				bodyBytes, err = json.Marshal(bodyMap)
				if err != nil {
					return err
				}
				req, err = http.NewRequest(c.method, endpoint, bytes.NewBuffer(bodyBytes))
			} else {
				formValues := url.Values{}
				for k, v := range bodyMap {
					formValues.Add(k, v)
				}
				req, err = http.NewRequest(c.method, endpoint, strings.NewReader(formValues.Encode()))
			}

			if err != nil {
				return err
			}
		}

		req.Header.Set("Content-Type", c.contentType)
	} else if c.method == "GET" {
		req, err = http.NewRequest(c.method, endpoint, nil)
		if err != nil {
			return err
		}

		// Only append extra query params if the endpoint doesn't already use placeholders
		if !hasMobilePlaceholder && !hasCodePlaceholder && !hasPhoneNumberPlaceholder && !hasPhoneNumberNoPlusPlaceholder {
			q := req.URL.Query()
			q.Add(phoneNumberField, phoneNumber)
			q.Add(contentField, content)
			req.URL.RawQuery = q.Encode()
		}
	} else {
		return fmt.Errorf("HttpSmsClient's SendMessage() error, unsupported method: %s", c.method)
	}

	for k, v := range c.httpHeaders {
		req.Header.Set(k, v)
	}

	httpClient := proxy.DefaultHttpClient
	if c.enableProxy {
		httpClient = proxy.ProxyHttpClient
	}
	if c.httpClient != nil {
		httpClient = c.httpClient
	}
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
