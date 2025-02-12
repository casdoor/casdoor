// Copyright 2024 The Casdoor Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package email

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type SendgridEmailProvider struct {
	ApiKey   string
	Host     string
	Endpoint string
}

type SendgridResponseBody struct {
	Errors []struct {
		Message string      `json:"message"`
		Field   interface{} `json:"field"`
		Help    interface{} `json:"help"`
	} `json:"errors"`
}

func NewSendgridEmailProvider(apiKey string, host string, endpoint string) *SendgridEmailProvider {
	return &SendgridEmailProvider{ApiKey: apiKey, Host: host, Endpoint: endpoint}
}

func (s *SendgridEmailProvider) Send(fromAddress string, fromName string, toAddress string, subject string, content string) error {
	client := s.initSendgridClient()

	from := mail.NewEmail(fromName, fromAddress)
	to := mail.NewEmail("", toAddress)
	message := mail.NewSingleEmail(from, subject, to, "", content)

	resp, err := client.Send(message)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 300 {
		var responseBody SendgridResponseBody
		err = json.Unmarshal([]byte(resp.Body), &responseBody)
		if err != nil {
			return err
		}

		messages := []string{}
		for _, sendgridError := range responseBody.Errors {
			messages = append(messages, sendgridError.Message)
		}

		return fmt.Errorf("status code: %d, error message: %s", resp.StatusCode, messages)
	}

	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("status code: %d", resp.StatusCode)
	}

	return nil
}

func (s *SendgridEmailProvider) initSendgridClient() *sendgrid.Client {
	if s.Host == "" || s.Endpoint == "" {
		return sendgrid.NewSendClient(s.ApiKey)
	}

	request := sendgrid.GetRequest(s.ApiKey, s.Endpoint, s.Host)
	request.Method = "POST"

	return &sendgrid.Client{Request: request}
}
