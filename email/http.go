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
	"fmt"
	"net/http"

	"github.com/casdoor/casdoor/proxy"
)

type HttpEmailProvider struct {
	endpoint string
	method   string
}

func NewHttpEmailProvider(endpoint string, method string) *HttpEmailProvider {
	client := &HttpEmailProvider{
		endpoint: endpoint,
		method:   method,
	}
	return client
}

func (c *HttpEmailProvider) Send(fromAddress string, fromName string, toAddress string, subject string, content string) error {
	req, err := http.NewRequest(c.method, c.endpoint, bytes.NewBufferString(content))
	if err != nil {
		return err
	}

	if c.method == "POST" {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.PostForm = map[string][]string{
			"fromName":  {fromName},
			"toAddress": {toAddress},
			"subject":   {subject},
			"content":   {content},
		}
	} else if c.method == "GET" {
		q := req.URL.Query()
		q.Add("fromName", fromName)
		q.Add("toAddress", toAddress)
		q.Add("subject", subject)
		q.Add("content", content)
		req.URL.RawQuery = q.Encode()
	} else {
		return fmt.Errorf("HttpEmailProvider's Send() error, unsupported method: %s", c.method)
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
