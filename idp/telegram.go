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
	"io"
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
	userId, ok := authData["id"].(float64)
	if !ok {
		return nil, fmt.Errorf("invalid user id in auth data")
	}

	token := &oauth2.Token{
		AccessToken: fmt.Sprintf("telegram_%d", int64(userId)),
		TokenType:   "Bearer",
	}

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

	// Prepare data check string
	var dataCheckArr []string
	for key, value := range authData {
		if key == "hash" {
			continue
		}
		dataCheckArr = append(dataCheckArr, fmt.Sprintf("%s=%v", key, value))
	}
	sort.Strings(dataCheckArr)
	dataCheckString := strings.Join(dataCheckArr, "\n")

	// Calculate secret key
	secretKey := sha256.Sum256([]byte(idp.ClientSecret))

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

// TelegramUserInfo represents the user information from Telegram
// Response format from Telegram Login Widget:
//
//	{
//	  "id": 123456789,
//	  "first_name": "John",
//	  "last_name": "Doe",
//	  "username": "johndoe",
//	  "photo_url": "https://t.me/i/userpic/320/johndoe.jpg",
//	  "auth_date": 1234567890,
//	  "hash": "..."
//	}
type TelegramUserInfo struct {
	Id        int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name,omitempty"`
	Username  string `json:"username,omitempty"`
	PhotoUrl  string `json:"photo_url,omitempty"`
	AuthDate  int64  `json:"auth_date"`
}

func (idp *TelegramIdProvider) GetUserInfo(token *oauth2.Token) (*UserInfo, error) {
	// Extract user ID from token
	accessToken := token.AccessToken
	if !strings.HasPrefix(accessToken, "telegram_") {
		return nil, fmt.Errorf("invalid Telegram access token format")
	}

	// Get user info from Telegram API
	// Note: Telegram's Login Widget doesn't provide an API endpoint to fetch user info
	// The user data is passed during authentication and validated via hash
	// We need to fetch the user info using the Bot API
	userId := strings.TrimPrefix(accessToken, "telegram_")

	// Use Telegram Bot API to get user info
	apiUrl := fmt.Sprintf("https://api.telegram.org/bot%s/getChat?chat_id=%s", idp.ClientSecret, userId)

	resp, err := idp.Client.Get(apiUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info from Telegram: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	// Parse Telegram API response
	var apiResponse struct {
		Ok     bool `json:"ok"`
		Result struct {
			Id        int64  `json:"id"`
			FirstName string `json:"first_name"`
			LastName  string `json:"last_name"`
			Username  string `json:"username"`
			Photo     struct {
				SmallFileId string `json:"small_file_id"`
				BigFileId   string `json:"big_file_id"`
			} `json:"photo"`
		} `json:"result"`
		Description string `json:"description,omitempty"`
	}

	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, fmt.Errorf("failed to parse Telegram API response: %v", err)
	}

	if !apiResponse.Ok {
		return nil, fmt.Errorf("Telegram API error: %s", apiResponse.Description)
	}

	// Build display name
	displayName := apiResponse.Result.FirstName
	if apiResponse.Result.LastName != "" {
		displayName = displayName + " " + apiResponse.Result.LastName
	}

	// Get photo URL if available
	photoUrl := ""
	if apiResponse.Result.Photo.BigFileId != "" {
		// Get file path for the photo
		fileResp, err := idp.Client.Get(fmt.Sprintf("https://api.telegram.org/bot%s/getFile?file_id=%s",
			idp.ClientSecret, apiResponse.Result.Photo.BigFileId))
		if err == nil {
			defer fileResp.Body.Close()
			fileBody, _ := io.ReadAll(fileResp.Body)
			var fileResponse struct {
				Ok     bool `json:"ok"`
				Result struct {
					FilePath string `json:"file_path"`
				} `json:"result"`
			}
			if json.Unmarshal(fileBody, &fileResponse) == nil && fileResponse.Ok {
				photoUrl = fmt.Sprintf("https://api.telegram.org/file/bot%s/%s",
					idp.ClientSecret, fileResponse.Result.FilePath)
			}
		}
	}

	userInfo := UserInfo{
		Id:          strconv.FormatInt(apiResponse.Result.Id, 10),
		Username:    apiResponse.Result.Username,
		DisplayName: displayName,
		AvatarUrl:   photoUrl,
	}

	return &userInfo, nil
}
