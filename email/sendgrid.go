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
	"strings"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type SendgridEmailProvider struct {
	ApiKey string
}

type SendgridResponseBody struct {
	Errors []struct {
		Message string      `json:"message"`
		Field   interface{} `json:"field"`
		Help    interface{} `json:"help"`
	} `json:"errors"`
}

func NewSendgridEmailProvider(apiKey string) *SendgridEmailProvider {
	return &SendgridEmailProvider{ApiKey: apiKey}
}

func (s *SendgridEmailProvider) Send(fromAddress string, fromName, toAddress string, subject string, content string) error {
	from := mail.NewEmail(fromName, fromAddress)
	to := mail.NewEmail("", toAddress)
	message := mail.NewSingleEmail(from, subject, to, "", content)
	client := sendgrid.NewSendClient(s.ApiKey)
	response, err := client.Send(message)
	if err != nil {
		return err
	}

	if response.StatusCode >= 300 {
		var responseBody SendgridResponseBody
		err = json.Unmarshal([]byte(response.Body), &responseBody)
		if err != nil {
			return err
		}

		messages := []string{}
		for _, sendgridError := range responseBody.Errors {
			messages = append(messages, sendgridError.Message)
		}

		return fmt.Errorf("SendGrid status code: %d, error message: %s", response.StatusCode, strings.Join(messages, " | "))
	}

	return nil
}
