// Copyright 2025 The Casdoor Authors. All Rights Reserved.
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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	notify "github.com/casdoor/notify2"
)

// wecomService encapsulates the WeCom webhook client
type wecomService struct {
	webhookURL string
}

// wecomResponse represents the response from WeCom webhook API
type wecomResponse struct {
	Errcode int    `json:"errcode"`
	Errmsg  string `json:"errmsg"`
}

// NewWeComProvider returns a new instance of a WeCom notification service
// WeCom (WeChat Work) uses webhook for group chat notifications
// Reference: https://developer.work.weixin.qq.com/document/path/90236
func NewWeComProvider(webhookURL string) (notify.Notifier, error) {
	wecomSrv := &wecomService{
		webhookURL: webhookURL,
	}

	notifier := notify.New()
	notifier.UseServices(wecomSrv)

	return notifier, nil
}

// Send sends a text message to WeCom group chat via webhook
func (s *wecomService) Send(ctx context.Context, subject, content string) error {
	text := subject
	if content != "" {
		text = subject + "\n" + content
	}

	// WeCom webhook message format
	message := map[string]interface{}{
		"msgtype": "text",
		"text": map[string]string{
			"content": text,
		},
	}

	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal WeCom message: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.webhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create WeCom request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send WeCom message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("WeCom webhook returned HTTP status code: %d", resp.StatusCode)
	}

	// Parse WeCom API response
	var wecomResp wecomResponse
	if err := json.NewDecoder(resp.Body).Decode(&wecomResp); err != nil {
		return fmt.Errorf("failed to decode WeCom response: %w", err)
	}

	// Check WeCom API error code
	if wecomResp.Errcode != 0 {
		return fmt.Errorf("WeCom API error: errcode=%d, errmsg=%s", wecomResp.Errcode, wecomResp.Errmsg)
	}

	return nil
}
