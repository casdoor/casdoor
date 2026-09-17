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

package idp

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

// VkIdIdProvider implements VK ID (https://id.vk.ru), the OAuth 2.1 successor of
// the legacy "VK" provider. Apps registered in the VK ID dashboard cannot use the
// old oauth.vk.com endpoints anymore.
//
// VK ID deviates from plain OAuth 2.1 in two ways, which is why golang.org/x/oauth2
// cannot be used directly:
//   - the token request requires "device_id", which VK ID returns to the redirect URI
//     next to "code" and "state";
//   - user info is only served over POST, with the access token in the form body.
type VkIdIdProvider struct {
	Client *http.Client

	ClientId     string
	ClientSecret string
	RedirectUrl  string
	CodeVerifier string
	DeviceId     string
}

func NewVkIdIdProvider(idpInfo *ProviderInfo, redirectUrl string) *VkIdIdProvider {
	return &VkIdIdProvider{
		ClientId:     idpInfo.ClientId,
		ClientSecret: idpInfo.ClientSecret,
		RedirectUrl:  redirectUrl,
		CodeVerifier: idpInfo.CodeVerifier,
		DeviceId:     idpInfo.DeviceId,
	}
}

func (idp *VkIdIdProvider) SetHttpClient(client *http.Client) {
	idp.Client = client
}

type VkIdTokenResp struct {
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token"`
	IdToken          string `json:"id_token"`
	TokenType        string `json:"token_type"`
	ExpiresIn        int    `json:"expires_in"`
	UserId           int64  `json:"user_id"`
	Scope            string `json:"scope"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

// GetToken exchanges the authorization code for an access token.
// Endpoint: https://id.vk.ru/oauth2/auth
func (idp *VkIdIdProvider) GetToken(code string) (*oauth2.Token, error) {
	if idp.DeviceId == "" {
		return nil, fmt.Errorf("VK ID: the device_id parameter is empty, it is returned to the redirect URI along with the code")
	}

	params := url.Values{}
	params.Set("grant_type", "authorization_code")
	params.Set("code", code)
	params.Set("code_verifier", idp.CodeVerifier)
	params.Set("client_id", idp.ClientId)
	params.Set("device_id", idp.DeviceId)
	params.Set("redirect_uri", idp.RedirectUrl)
	if idp.ClientSecret != "" {
		// confidential apps authenticate the token request with their service key
		params.Set("service_token", idp.ClientSecret)
	}

	resp, err := idp.Client.PostForm("https://id.vk.ru/oauth2/auth", params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var tokenResp VkIdTokenResp
	err = json.Unmarshal(body, &tokenResp)
	if err != nil {
		return nil, err
	}

	if tokenResp.Error != "" {
		return nil, fmt.Errorf("VK ID: %s, %s", tokenResp.Error, tokenResp.ErrorDescription)
	}
	if tokenResp.AccessToken == "" {
		return nil, fmt.Errorf("VK ID: the access token is empty")
	}

	token := &oauth2.Token{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		TokenType:    "Bearer",
		Expiry:       time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
	}
	return token, nil
}

type VkIdUserInfoResp struct {
	User struct {
		UserId    string `json:"user_id"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Email     string `json:"email"`
		Phone     string `json:"phone"`
		Avatar    string `json:"avatar"`
	} `json:"user"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

// GetUserInfo fetches the user profile.
// Endpoint: https://id.vk.ru/oauth2/user_info, POST only, token in the form body.
func (idp *VkIdIdProvider) GetUserInfo(token *oauth2.Token) (*UserInfo, error) {
	params := url.Values{}
	params.Set("access_token", token.AccessToken)
	params.Set("client_id", idp.ClientId)

	resp, err := idp.Client.PostForm("https://id.vk.ru/oauth2/user_info", params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var infoResp VkIdUserInfoResp
	err = json.Unmarshal(body, &infoResp)
	if err != nil {
		return nil, err
	}

	if infoResp.Error != "" {
		return nil, fmt.Errorf("VK ID: %s, %s", infoResp.Error, infoResp.ErrorDescription)
	}
	if infoResp.User.UserId == "" {
		return nil, fmt.Errorf("VK ID: the user ID is empty")
	}

	user := infoResp.User
	displayName := strings.TrimSpace(fmt.Sprintf("%s %s", user.FirstName, user.LastName))
	if displayName == "" {
		displayName = user.UserId
	}

	userInfo := &UserInfo{
		Id:          user.UserId,
		Username:    fmt.Sprintf("vkid_%s", user.UserId),
		DisplayName: displayName,
		Email:       user.Email,
		Phone:       user.Phone,
		AvatarUrl:   user.Avatar,
	}
	return userInfo, nil
}
