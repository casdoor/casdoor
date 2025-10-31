// Copyright 2024 The Casdoor Authors. All Rights Reserved.
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

	"github.com/casdoor/notify"
)

// wecomService encapsulates the WeCom webhook client
type wecomService struct {
	webhookURL string
}

// NewWeComProvider returns a new instance of a WeCom notification service
// WeCom (WeChat Work/企业微信) uses webhook for group chat notifications
// Reference: https://developer.work.weixin.qq.com/document/path/90236
func NewWeComProvider(webhookURL string) (notify.Notifier, error) {
	svc := &wecomService{
		webhookURL: webhookURL,
	}

	notifier := &wecomNotifier{service: svc}
	return notifier, nil
}

// wecomNotifier implements the notify.Notifier interface
type wecomNotifier struct {
	service *wecomService
}

// Send sends a text message to WeCom group chat via webhook
func (n *wecomNotifier) Send(ctx context.Context, subject, content string) error {
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

	resp, err := http.Post(n.service.webhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send WeCom message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("WeCom webhook returned status code: %d", resp.StatusCode)
	}

	return nil
}
