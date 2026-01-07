// Copyright 2021 The Casdoor Authors. All Rights Reserved.
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

package idp

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/oauth2"
)

type TelegramIdProvider struct {
	Client       *http.Client
	ClientId     string
	ClientSecret string
	RedirectUrl  string
}

func NewTelegramIdProvider(clientId string, clientSecret string, redirectUrl string) *TelegramIdProvider {
	idp := &TelegramIdProvider{
		ClientId:     clientId,
		ClientSecret: clientSecret,
		RedirectUrl:  redirectUrl,
	}

	return idp
}

func (idp *TelegramIdProvider) SetHttpClient(client *http.Client) {
	idp.Client = client
}

// GetToken validates the Telegram auth data and returns a token
// Telegram uses a widget-based authentication, not standard OAuth2
// The "code" parameter contains the JSON-encoded auth data from Telegram
func (idp *TelegramIdProvider) GetToken(code string) (*oauth2.Token, error) {
	// Decode the auth data from the code parameter
	var authData map[string]interface{}
	if err := json.Unmarshal([]byte(code), &authData); err != nil {
		return nil, fmt.Errorf("failed to parse Telegram auth data: %v", err)
	}

	// Verify the data authenticity
	if err := idp.verifyTelegramAuth(authData); err != nil {
		return nil, fmt.Errorf("failed to verify Telegram auth data: %v", err)
	}

	// Create a token with the user ID as access token
	userId, ok := telegramAsInt64(authData["id"])
	if !ok {
		return nil, fmt.Errorf("invalid user id in auth data")
	}

	// Store the complete auth data in the token for later retrieval
	authDataJson, err := json.Marshal(authData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal auth data: %v", err)
	}

	token := &oauth2.Token{
		AccessToken: fmt.Sprintf("telegram_%d", userId),
		TokenType:   "Bearer",
	}

	// Store auth data in token extras to avoid additional API calls
	token = token.WithExtra(map[string]interface{}{
		"telegram_auth_data": string(authDataJson),
	})

	return token, nil
}

// verifyTelegramAuth verifies the authenticity of Telegram auth data
// According to Telegram docs: https://core.telegram.org/widgets/login#checking-authorization
func (idp *TelegramIdProvider) verifyTelegramAuth(authData map[string]interface{}) error {
	// Extract hash from auth data
	hash, ok := authData["hash"].(string)
	if !ok {
		return fmt.Errorf("hash not found in auth data")
	}
	hash = strings.TrimSpace(hash)

	// Prepare data check string
	var dataCheckArr []string
	for key, value := range authData {
		if key == "hash" {
			continue
		}
		dataCheckArr = append(dataCheckArr, fmt.Sprintf("%s=%s", key, telegramAsString(value)))
	}
	sort.Strings(dataCheckArr)
	dataCheckString := strings.Join(dataCheckArr, "\n")

	// Calculate secret key
	clientSecret := strings.TrimSpace(idp.ClientSecret)
	secretKey := sha256.Sum256([]byte(clientSecret))

	// Calculate hash
	h := hmac.New(sha256.New, secretKey[:])
	h.Write([]byte(dataCheckString))
	calculatedHash := hex.EncodeToString(h.Sum(nil))

	// Compare hashes
	if calculatedHash != hash {
		return fmt.Errorf("data verification failed")
	}

	return nil
}

func (idp *TelegramIdProvider) GetUserInfo(token *oauth2.Token) (*UserInfo, error) {
	// Extract auth data from token
	authDataStr, ok := token.Extra("telegram_auth_data").(string)
	if !ok {
		return nil, fmt.Errorf("telegram auth data not found in token")
	}

	// Parse the auth data
	var authData map[string]interface{}
	if err := json.Unmarshal([]byte(authDataStr), &authData); err != nil {
		return nil, fmt.Errorf("failed to parse auth data: %v", err)
	}

	// Extract user information from auth data
	userId, ok := telegramAsInt64(authData["id"])
	if !ok {
		return nil, fmt.Errorf("invalid user id in auth data")
	}

	firstName, _ := authData["first_name"].(string)
	lastName, _ := authData["last_name"].(string)
	username, _ := authData["username"].(string)
	photoUrl, _ := authData["photo_url"].(string)

	// Build display name with fallback
	displayName := strings.TrimSpace(firstName + " " + lastName)
	if displayName == "" {
		displayName = username
	}
	if displayName == "" {
		displayName = strconv.FormatInt(userId, 10)
	}

	userInfo := UserInfo{
		Id:          strconv.FormatInt(userId, 10),
		Username:    username,
		DisplayName: displayName,
		AvatarUrl:   photoUrl,
	}

	return &userInfo, nil
}

func telegramAsInt64(v interface{}) (int64, bool) {
	switch t := v.(type) {
	case float64:
		if t != math.Trunc(t) {
			return 0, false
		}
		if t > float64(math.MaxInt64) || t < float64(math.MinInt64) {
			return 0, false
		}
		return int64(t), true
	case string:
		i, err := strconv.ParseInt(t, 10, 64)
		if err != nil {
			return 0, false
		}
		return i, true
	default:
		return 0, false
	}
}

func telegramAsString(v interface{}) string {
	switch t := v.(type) {
	case string:
		return t
	case float64:
		if t == math.Trunc(t) && t <= float64(math.MaxInt64) && t >= float64(math.MinInt64) {
			return strconv.FormatInt(int64(t), 10)
		}
		return strconv.FormatFloat(t, 'g', -1, 64)
	default:
		return fmt.Sprint(v)
	}
}
