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
	"fmt"
	"net/http"

	"github.com/casdoor/casdoor/proxy"
)

type HttpSmsClient struct {
	endpoint  string
	method    string
	paramName string
	text      string
}

func newHttpSmsClient(endpoint string, method string, paramName string, text string) (*HttpSmsClient, error) {
	client := &HttpSmsClient{
		endpoint:  endpoint,
		method:    method,
		paramName: paramName,
		text:      text,
	}
	return client, nil
}

func (c *HttpSmsClient) SendMessage(param map[string]string, targetPhoneNumber ...string) error {
	var err error

	content := param["code"]
	httpClient := proxy.DefaultHttpClient

	req, err := http.NewRequest(c.method, c.endpoint, bytes.NewBufferString(content))
	if err != nil {
		return err
	}

	if c.method == "POST" {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.PostForm = map[string][]string{
			c.paramName: {content},
		}
	} else if c.method == "GET" {
		q := req.URL.Query()
		q.Add(c.paramName, content)
		req.URL.RawQuery = q.Encode()
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("SendMessage() error, custom HTTP SMS request failed with status: %s", resp.Status)
	}

	return err
}
