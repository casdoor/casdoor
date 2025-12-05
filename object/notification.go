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
	"time"

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

// SsoLogoutNotification represents the structure of a session-level SSO logout notification
// This includes session information and a signature for authentication
type SsoLogoutNotification struct {
	// User information
	Owner       string `json:"owner"`
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	Id          string `json:"id"`

	// Event type
	Event string `json:"event"`

	// Session-level information for targeted logout
	SessionIds       []string `json:"sessionIds"`       // List of session IDs being logged out
	AccessTokenHashes []string `json:"accessTokenHashes"` // Hashes of access tokens being expired

	// Authentication fields to prevent malicious logout requests
	Nonce     string `json:"nonce"`     // Random nonce for replay protection
	Timestamp int64  `json:"timestamp"` // Unix timestamp of the notification
	Signature string `json:"signature"` // HMAC-SHA256 signature for verification
}

// GetTokensByUser retrieves all tokens for a specific user
func GetTokensByUser(owner, username string) ([]*Token, error) {
	tokens := []*Token{}
	err := ormer.Engine.Where("organization = ? and user = ?", owner, username).Find(&tokens)
	if err != nil {
		return nil, err
	}
	return tokens, nil
}

// generateLogoutSignature generates an HMAC-SHA256 signature for the logout notification
// The signature is computed over the critical fields to prevent tampering
func generateLogoutSignature(clientSecret string, owner string, name string, nonce string, timestamp int64, sessionIds []string, accessTokenHashes []string) string {
	// Create a deterministic string from all fields that need to be verified
	sessionIdsStr := ""
	for _, sid := range sessionIds {
		sessionIdsStr += sid + ","
	}
	tokenHashesStr := ""
	for _, hash := range accessTokenHashes {
		tokenHashesStr += hash + ","
	}

	data := fmt.Sprintf("%s|%s|%s|%d|%s|%s", owner, name, nonce, timestamp, sessionIdsStr, tokenHashesStr)
	return util.GetHmacSha256(clientSecret, data)
}

// SendSsoLogoutNotifications sends logout notifications to all notification providers
// configured in the user's signup application. The notification includes session-level
// information and a signature for authentication.
// Parameters:
//   - user: The user being logged out
//   - sessionIds: List of session IDs being logged out (can be nil for all sessions)
//   - tokens: List of tokens being expired (can be nil if no tokens are being expired)
func SendSsoLogoutNotifications(user *User, sessionIds []string, tokens []*Token) error {
	if user == nil {
		return nil
	}

	// If user's signup application is empty, don't send notifications
	if user.SignupApplication == "" {
		return nil
	}

	// Get the user's signup application
	application, err := GetApplication(user.SignupApplication)
	if err != nil {
		return fmt.Errorf("failed to get signup application: %w", err)
	}

	if application == nil {
		return fmt.Errorf("signup application not found: %s", user.SignupApplication)
	}

	// Extract access token hashes from tokens
	accessTokenHashes := make([]string, 0, len(tokens))
	for _, token := range tokens {
		if token.AccessTokenHash != "" {
			accessTokenHashes = append(accessTokenHashes, token.AccessTokenHash)
		}
	}

	// Generate nonce and timestamp for replay protection
	nonce := util.GenerateId()
	timestamp := time.Now().Unix()

	// Generate signature using the application's client secret
	signature := generateLogoutSignature(
		application.ClientSecret,
		user.Owner,
		user.Name,
		nonce,
		timestamp,
		sessionIds,
		accessTokenHashes,
	)

	// Prepare the notification data
	notification := SsoLogoutNotification{
		Owner:             user.Owner,
		Name:              user.Name,
		DisplayName:       user.DisplayName,
		Email:             user.Email,
		Phone:             user.Phone,
		Id:                user.Id,
		Event:             "sso-logout",
		SessionIds:        sessionIds,
		AccessTokenHashes: accessTokenHashes,
		Nonce:             nonce,
		Timestamp:         timestamp,
		Signature:         signature,
	}

	notificationData, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("failed to marshal notification data: %w", err)
	}
	content := string(notificationData)

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

// VerifySsoLogoutSignature verifies the signature of an SSO logout notification
// This should be called by applications receiving logout notifications
func VerifySsoLogoutSignature(clientSecret string, notification *SsoLogoutNotification) bool {
	expectedSignature := generateLogoutSignature(
		clientSecret,
		notification.Owner,
		notification.Name,
		notification.Nonce,
		notification.Timestamp,
		notification.SessionIds,
		notification.AccessTokenHashes,
	)
	return notification.Signature == expectedSignature
}
