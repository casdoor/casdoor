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

	"github.com/casdoor/casdoor/notification"
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
// configured in the user's signup application
func SendSsoLogoutNotifications(user *User) error {
	if user == nil {
		return nil
	}

	// If user's signup application is empty, don't send notifications
	if user.SignupApplication == "" {
		return nil
	}

	// Get the user's signup application
	// Use GetApplicationByUser which properly handles SignupApplication field
	// that may contain just the application name without owner prefix
	application, err := GetApplicationByUser(user)
	if err != nil {
		return fmt.Errorf("failed to get signup application: %w", err)
	}

	if application == nil {
		return fmt.Errorf("signup application not found: %s", user.SignupApplication)
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

	// Send notifications to all notification providers in the signup application
	for _, providerItem := range application.Providers {
		if providerItem.Provider == nil {
			continue
		}

		// Only send to notification providers
		if providerItem.Provider.Category != "Notification" {
			continue
		}

		// Send the notification using the provider from the providerItem
		err = SendNotification(providerItem.Provider, content)
		if err != nil {
			return fmt.Errorf("failed to send SSO logout notification to provider %s/%s: %w", providerItem.Provider.Owner, providerItem.Provider.Name, err)
		}
	}

	return nil
}
