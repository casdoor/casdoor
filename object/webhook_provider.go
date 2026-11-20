// Copyright 2026 The Casdoor Authors. All Rights Reserved.
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
	"fmt"
	"html"
	"strings"

	"github.com/casdoor/casdoor/util"
)

const webhookReceiverUser = "$user"

func (webhook *Webhook) isProviderType() bool {
	return webhook.Type == "Email" || webhook.Type == "SMS" || webhook.Type == "Notification"
}

func isFailedRecord(record *Record) bool {
	return record.StatusCode >= 400 || strings.HasPrefix(record.Response, `{status:"error"`)
}

// getWebhookReceiverUser keeps only the fields the templates and "$user" need, as the user may be gone by the delivery
func getWebhookReceiverUser(record *Record) (*User, error) {
	user, err := getUser(record.Organization, record.User)
	if err != nil || user == nil {
		return nil, err
	}

	return &User{
		Owner:       user.Owner,
		Name:        user.Name,
		DisplayName: user.DisplayName,
		Email:       user.Email,
		Phone:       user.Phone,
		CountryCode: user.CountryCode,
	}, nil
}

func getWebhookProvider(webhook *Webhook) (*Provider, error) {
	provider, err := getProvider(webhook.Organization, webhook.Provider)
	if err != nil {
		return nil, err
	}

	if provider == nil && webhook.Organization == "built-in" {
		provider, err = getProvider("admin", webhook.Provider)
		if err != nil {
			return nil, err
		}
	}

	if provider == nil {
		return nil, fmt.Errorf("the provider: %s is not found in the organization: %s", webhook.Provider, webhook.Organization)
	}

	if provider.Category != webhook.Type {
		return nil, fmt.Errorf("the provider: %s is a %s provider, not a %s provider", provider.Name, provider.Category, webhook.Type)
	}

	return provider, nil
}

func getWebhookReceivers(webhook *Webhook, user *User) []string {
	res := []string{}
	for _, receiver := range webhook.Receivers {
		receiver = strings.TrimSpace(receiver)
		if receiver == webhookReceiverUser {
			receiver = ""
			if user != nil && webhook.Type == "Email" {
				receiver = user.Email
			} else if user != nil && webhook.Type == "SMS" && user.Phone != "" {
				receiver, _ = util.GetE164Number(user.Phone, user.GetCountryCode(""))
			}
		}

		if receiver != "" && !util.InSlice(res, receiver) {
			res = append(res, receiver)
		}
	}
	return res
}

func renderWebhookTemplate(text string, record *Record, user *User, escape bool) string {
	userName, displayName, email, phone := record.User, "", "", ""
	if user != nil {
		displayName, email, phone = user.DisplayName, user.Email, user.Phone
	}

	values := []string{
		"${organization}", record.Organization,
		"${user}", userName,
		"${user.displayName}", displayName,
		"${user.email}", email,
		"${user.phone}", phone,
		"${event}", record.Action,
		"${ip}", record.ClientIp,
		"${time}", record.CreatedTime,
	}
	if escape {
		for i := 1; i < len(values); i += 2 {
			values[i] = html.EscapeString(values[i])
		}
	}

	return strings.NewReplacer(values...).Replace(text)
}

func sendProviderWebhook(webhook *Webhook, record *Record, user *User) (int, string, error) {
	provider, err := getWebhookProvider(webhook)
	if err != nil {
		return 0, "", err
	}

	receivers := getWebhookReceivers(webhook, user)
	if len(receivers) == 0 && webhook.Type != "Notification" {
		return 0, "", fmt.Errorf("the webhook: %s has no receiver for the event: %s of the user: %s", webhook.GetId(), record.Action, record.User)
	}

	switch webhook.Type {
	case "Email":
		title := renderWebhookTemplate(webhook.Title, record, user, false)
		content := renderWebhookTemplate(webhook.Content, record, user, true)
		err = SendEmail(webhook.Organization, provider, title, content, receivers, "Casdoor")
	case "SMS":
		err = SendSms(provider, renderWebhookTemplate(webhook.Content, record, user, false), receivers...)
	case "Notification":
		content := renderWebhookTemplate(webhook.Content, record, user, false)
		if len(receivers) == 0 {
			err = SendNotification(provider, content, "")
		}
		for _, receiver := range receivers {
			if err = SendNotification(provider, content, receiver); err != nil {
				break
			}
		}
	}
	if err != nil {
		return 0, "", err
	}

	return 200, fmt.Sprintf("sent to %s", strings.Join(receivers, ", ")), nil
}
