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

package notification

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/casdoor/casdoor/proxy"
)

const customHttpNotificationTimeout = 30 * time.Second

type HttpNotificationClient struct {
	endpoint  string
	method    string
	paramName string
}

func NewCustomHttpProvider(endpoint string, method string, paramName string) (*HttpNotificationClient, error) {
	client := &HttpNotificationClient{
		endpoint:  endpoint,
		method:    method,
		paramName: paramName,
	}
	return client, nil
}

func (c *HttpNotificationClient) Send(ctx context.Context, subject string, content string) error {
	return c.SendWithRecipient(ctx, subject, content, "")
}

func (c *HttpNotificationClient) SendWithRecipient(ctx context.Context, subject string, content string, recipient string) error {
	if ctx == nil {
		ctx = context.Background()
	}

	var req *http.Request
	var err error
	if c.method == "POST" {
		formValues := url.Values{}
		formValues.Set(c.paramName, content)
		if recipient != "" {
			formValues.Set("recipient", recipient)
		}
		req, err = http.NewRequestWithContext(ctx, c.method, c.endpoint, strings.NewReader(formValues.Encode()))
		if err != nil {
			return err
		}

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	} else if c.method == "GET" {
		req, err = http.NewRequestWithContext(ctx, c.method, c.endpoint, nil)
		if err != nil {
			return err
		}

		q := req.URL.Query()
		q.Add(c.paramName, content)
		if recipient != "" {
			q.Add("recipient", recipient)
		}
		req.URL.RawQuery = q.Encode()
	} else {
		return fmt.Errorf("HttpNotificationClient's SendMessage() error, unsupported method: %s", c.method)
	}

	httpClient := getCustomHttpNotificationClient()
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HttpNotificationClient's SendMessage() error, custom HTTP Notification request failed with status: %s", resp.Status)
	}

	return err
}

func getCustomHttpNotificationClient() *http.Client {
	httpClient := proxy.DefaultHttpClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	client := *httpClient
	if client.Timeout == 0 {
		client.Timeout = customHttpNotificationTimeout
	}
	return &client
}
