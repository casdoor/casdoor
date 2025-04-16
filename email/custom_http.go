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

package email

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/casdoor/casdoor/proxy"
)

type HttpEmailProvider struct {
	endpoint    string
	method      string
	httpHeaders map[string]string
	bodyMapping map[string]string
	contentType string
}

func NewHttpEmailProvider(endpoint string, method string, httpHeaders map[string]string, bodyMapping map[string]string, contentType string) *HttpEmailProvider {
	if contentType == "" {
		contentType = "application/x-www-form-urlencoded"
	}

	client := &HttpEmailProvider{
		endpoint:    endpoint,
		method:      method,
		httpHeaders: httpHeaders,
		bodyMapping: bodyMapping,
		contentType: contentType,
	}
	return client
}

func (c *HttpEmailProvider) Send(fromAddress string, fromName string, toAddress string, subject string, content string) error {
	var req *http.Request
	var err error

	fromNameField := "fromName"
	toAddressField := "toAddress"
	subjectField := "subject"
	contentField := "content"

	for k, v := range c.bodyMapping {
		switch k {
		case "fromName":
			fromNameField = v
		case "toAddress":
			toAddressField = v
		case "subject":
			subjectField = v
		case "content":
			contentField = v
		}
	}

	if c.method == "POST" || c.method == "PUT" || c.method == "DELETE" {
		bodyMap := make(map[string]string)
		bodyMap[fromNameField] = fromName
		bodyMap[toAddressField] = toAddress
		bodyMap[subjectField] = subject
		bodyMap[contentField] = content

		var fromValueBytes []byte
		if c.contentType == "application/json" {
			fromValueBytes, err = json.Marshal(bodyMap)
			if err != nil {
				return err
			}
			req, err = http.NewRequest(c.method, c.endpoint, bytes.NewBuffer(fromValueBytes))
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
		q.Add(fromNameField, fromName)
		q.Add(toAddressField, toAddress)
		q.Add(subjectField, subject)
		q.Add(contentField, content)
		req.URL.RawQuery = q.Encode()
	} else {
		return fmt.Errorf("HttpEmailProvider's Send() error, unsupported method: %s", c.method)
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
		return fmt.Errorf("HttpEmailProvider's Send() error, custom HTTP Email request failed with status: %s", resp.Status)
	}

	return err
}
