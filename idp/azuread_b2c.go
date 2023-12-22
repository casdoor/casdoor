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

package idp

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/oauth2"
)

type AzureADB2CProvider struct {
	Client   *http.Client
	Config   *oauth2.Config
	Tenant   string
	UserFlow string
}

func NewAzureAdB2cProvider(clientId, clientSecret, redirectUrl, tenant string, userFlow string) *AzureADB2CProvider {
	return &AzureADB2CProvider{
		Config: &oauth2.Config{
			ClientID:     clientId,
			ClientSecret: clientSecret,
			RedirectURL:  redirectUrl,
			Endpoint: oauth2.Endpoint{
				AuthURL:  fmt.Sprintf("https://%s.b2clogin.com/%s.onmicrosoft.com/%s/oauth2/v2.0/authorize", tenant, tenant, userFlow),
				TokenURL: fmt.Sprintf("https://%s.b2clogin.com/%s.onmicrosoft.com/%s/oauth2/v2.0/token", tenant, tenant, userFlow),
			},
			Scopes: []string{"openid", "email"},
		},
		Tenant:   tenant,
		UserFlow: userFlow,
	}
}

func (p *AzureADB2CProvider) SetHttpClient(client *http.Client) {
	p.Client = client
}

type AzureadB2cToken struct {
	IdToken          string `json:"id_token"`
	TokenType        string `json:"token_type"`
	NotBefore        int    `json:"not_before"`
	IdTokenExpiresIn int    `json:"id_token_expires_in"`
	ProfileInfo      string `json:"profile_info"`
	Scope            string `json:"scope"`
}

func (p *AzureADB2CProvider) GetToken(code string) (*oauth2.Token, error) {
	payload := url.Values{}
	payload.Set("code", code)
	payload.Set("grant_type", "authorization_code")
	payload.Set("client_id", p.Config.ClientID)
	payload.Set("client_secret", p.Config.ClientSecret)
	payload.Set("redirect_uri", p.Config.RedirectURL)

	resp, err := p.Client.PostForm(p.Config.Endpoint.TokenURL, payload)
	if err != nil {
		return nil, err
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	pToken := &AzureadB2cToken{}
	err = json.Unmarshal(data, pToken)
	if err != nil {
		return nil, err
	}

	token := &oauth2.Token{
		AccessToken: pToken.IdToken,
		Expiry:      time.Unix(time.Now().Unix()+int64(pToken.IdTokenExpiresIn), 0),
	}
	return token, nil
}

func (p *AzureADB2CProvider) GetUserInfo(token *oauth2.Token) (*UserInfo, error) {
	userInfoEndpoint := fmt.Sprintf("https://%s.b2clogin.com/%s.onmicrosoft.com/%s/openid/v2.0/userinfo", p.Tenant, p.Tenant, p.UserFlow)
	req, err := http.NewRequest("GET", userInfoEndpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+token.AccessToken)

	resp, err := p.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error fetching user info: status code %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var userInfo UserInfo
	err = json.Unmarshal(bodyBytes, &userInfo)
	if err != nil {
		return nil, err
	}

	return &userInfo, nil
}
