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
	"context"
	"encoding/json"
	"fmt"

	"github.com/beego/beego/logs"
	"github.com/casdoor/casdoor/notification"
	"github.com/casdoor/casdoor/util"
	"github.com/casdoor/notify"
)

func getNotificationClient(provider *Provider) (notify.Notifier, error) {
	var client notify.Notifier
	client, err := notification.GetNotificationProvider(provider.Type, provider.ClientId, provider.ClientSecret, provider.ClientId2, provider.ClientSecret2, provider.AppId, provider.Receiver, provider.Method, provider.Title, provider.Metadata, provider.RegionId)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func SendNotification(provider *Provider, content string) error {
	client, err := getNotificationClient(provider)
	if err != nil {
		return err
	}

	err = client.Send(context.Background(), "", content)
	return err
}

// SendSsoLogoutNotifications sends logout notifications to all notification providers
// configured in applications within the user's organization
func SendSsoLogoutNotifications(user *User) error {
	if user == nil {
		return nil
	}

	// Get all applications in the user's organization
	applications, err := GetOrganizationApplications("admin", user.Owner)
	if err != nil {
		return fmt.Errorf("failed to get organization applications: %w", err)
	}

	// Prepare sanitized user data for notification
	// Only include safe, non-sensitive fields
	sanitizedData := map[string]interface{}{
		"owner":       user.Owner,
		"name":        user.Name,
		"displayName": user.DisplayName,
		"email":       user.Email,
		"phone":       user.Phone,
		"id":          user.Id,
		"event":       "sso-logout",
	}
	userData, err := json.Marshal(sanitizedData)
	if err != nil {
		return fmt.Errorf("failed to marshal user data: %w", err)
	}
	content := string(userData)

	// Send notifications to all notification providers in each application
	for _, app := range applications {
		for _, providerItem := range app.Providers {
			if providerItem.Provider == nil {
				continue
			}

			// Only send to notification providers
			if providerItem.Provider.Category != "Notification" {
				continue
			}

			// Get the full provider object
			provider, err := GetProvider(util.GetId(providerItem.Owner, providerItem.Name))
			if err != nil {
				logs.Info("Failed to get provider %s/%s for SSO logout notification: %v", providerItem.Owner, providerItem.Name, err)
				continue
			}

			if provider == nil {
				continue
			}

			// Send the notification
			err = SendNotification(provider, content)
			if err != nil {
				logs.Info("Failed to send SSO logout notification to provider %s/%s: %v", provider.Owner, provider.Name, err)
				continue
			}

			logs.Info("Successfully sent SSO logout notification to provider %s/%s for user %s/%s", provider.Owner, provider.Name, user.Owner, user.Name)
		}
	}

	return nil
}
