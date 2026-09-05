// Copyright 2026 The Casdoor Authors. All Rights Reserved.
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
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/casdoor/casdoor/proxy"
)

const mailKiteSendUrl = "https://api.mailkite.dev/v1/send"

type MailKiteEmailProvider struct {
	ApiKey  string
	SendUrl string
}

type mailKiteSendRequest struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	Html    string   `json:"html"`
}

func NewMailKiteEmailProvider(apiKey string) *MailKiteEmailProvider {
	return &MailKiteEmailProvider{ApiKey: apiKey, SendUrl: mailKiteSendUrl}
}

func (m *MailKiteEmailProvider) Send(fromAddress string, fromName string, toAddresses []string, subject string, content string) error {
	from := fromAddress
	if fromName != "" {
		from = fmt.Sprintf("%s <%s>", fromName, fromAddress)
	}

	body, err := json.Marshal(mailKiteSendRequest{
		From:    from,
		To:      toAddresses,
		Subject: subject,
		Html:    content,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, m.SendUrl, bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+m.ApiKey)
	req.Header.Set("User-Agent", "Casdoor")

	resp, err := proxy.DefaultHttpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("MailKiteEmailProvider's Send() error, status code: %d, response: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	return nil
}
