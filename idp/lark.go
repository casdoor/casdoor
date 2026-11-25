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
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/nyaruka/phonenumbers"
	"golang.org/x/oauth2"
)

type LarkIdProvider struct {
	Client     *http.Client
	Config     *oauth2.Config
	LarkDomain string
}

func NewLarkIdProvider(clientId string, clientSecret string, redirectUrl string, useGlobalEndpoint bool) *LarkIdProvider {
	idp := &LarkIdProvider{}

	if useGlobalEndpoint {
		idp.LarkDomain = "https://open.larksuite.com"
	} else {
		idp.LarkDomain = "https://open.feishu.cn"
	}

	config := idp.getConfig(clientId, clientSecret, redirectUrl)
	idp.Config = config
	return idp
}

func (idp *LarkIdProvider) SetHttpClient(client *http.Client) {
	idp.Client = client
}

// getConfig return a point of Config, which describes a typical 3-legged OAuth2 flow
func (idp *LarkIdProvider) getConfig(clientId string, clientSecret string, redirectUrl string) *oauth2.Config {
	endpoint := oauth2.Endpoint{
		TokenURL: idp.LarkDomain + "/open-apis/authen/v2/oauth/token",
	}

	config := &oauth2.Config{
		Scopes:       []string{},
		Endpoint:     endpoint,
		ClientID:     clientId,
		ClientSecret: clientSecret,
		RedirectURL:  redirectUrl,
	}

	return config
}

func (idp *LarkIdProvider) GetToken(code string) (*oauth2.Token, error) {
	params := struct {
		GrantType    string `json:"grant_type"`
		ClientID     string `json:"client_id"`
		ClientSecret string `json:"client_secret"`
		Code         string `json:"code"`
		RedirectURI  string `json:"redirect_uri,omitempty"`
	}{
		GrantType:    "authorization_code",
		ClientID:     idp.Config.ClientID,
		ClientSecret: idp.Config.ClientSecret,
		Code:         code,
		RedirectURI:  idp.Config.RedirectURL,
	}

	body, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	resp, err := idp.Client.Post(idp.Config.Endpoint.TokenURL, "application/json; charset=utf-8", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Code        int    `json:"code"`
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("invalid Lark token response (HTTP %d)", resp.StatusCode)
	}
	if resp.StatusCode != http.StatusOK || result.Code != 0 {
		return nil, fmt.Errorf("Lark token exchange failed (HTTP %d, code %d)", resp.StatusCode, result.Code)
	}
	if result.AccessToken == "" {
		return nil, fmt.Errorf("Lark token response is missing access_token")
	}

	token := &oauth2.Token{AccessToken: result.AccessToken, TokenType: "Bearer"}
	if result.ExpiresIn > 0 {
		token.Expiry = time.Now().Add(time.Duration(result.ExpiresIn) * time.Second)
	}
	return token, nil
}

func (idp *LarkIdProvider) GetUserInfo(token *oauth2.Token) (*UserInfo, error) {
	req, err := http.NewRequest(http.MethodGet, idp.LarkDomain+"/open-apis/authen/v1/user_info", nil)
	if err != nil {
		return nil, err
	}
	token.SetAuthHeader(req)

	resp, err := idp.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var larkUserInfo struct {
		Code int `json:"code"`
		Data struct {
			Name            string `json:"name"`
			AvatarUrl       string `json:"avatar_url"`
			OpenId          string `json:"open_id"`
			UnionId         string `json:"union_id"`
			Email           string `json:"email"`
			EnterpriseEmail string `json:"enterprise_email"`
			UserId          string `json:"user_id"`
			Mobile          string `json:"mobile"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&larkUserInfo); err != nil {
		return nil, fmt.Errorf("invalid Lark user info response (HTTP %d)", resp.StatusCode)
	}
	if resp.StatusCode != http.StatusOK || larkUserInfo.Code != 0 {
		return nil, fmt.Errorf("Lark user info request failed (HTTP %d, code %d)", resp.StatusCode, larkUserInfo.Code)
	}
	if larkUserInfo.Data.OpenId == "" {
		return nil, fmt.Errorf("Lark user info response is missing open_id")
	}

	// Use enterprise_email as fallback when email is empty
	email := larkUserInfo.Data.Email
	if email == "" {
		email = larkUserInfo.Data.EnterpriseEmail
	}

	// Use fallback mechanism for username: UserId -> UnionId -> OpenId
	username := larkUserInfo.Data.UserId
	if username == "" {
		username = larkUserInfo.Data.UnionId
	}
	if username == "" {
		username = larkUserInfo.Data.OpenId
	}

	var phoneNumber string
	var countryCode string
	if len(larkUserInfo.Data.Mobile) != 0 {
		phoneNumberParsed, err := phonenumbers.Parse(larkUserInfo.Data.Mobile, "")
		if err != nil {
			return nil, err
		}
		countryCode = phonenumbers.GetRegionCodeForNumber(phoneNumberParsed)
		phoneNumber = fmt.Sprintf("%d", phoneNumberParsed.GetNationalNumber())
	}

	userInfo := UserInfo{
		Id:          larkUserInfo.Data.OpenId,
		DisplayName: larkUserInfo.Data.Name,
		Username:    username,
		Email:       email,
		// enterprise directory email
		EmailVerified: email != "",
		AvatarUrl:     larkUserInfo.Data.AvatarUrl,
		Phone:         phoneNumber,
		PhoneVerified: phoneNumber != "",
		CountryCode:   countryCode,
	}
	return &userInfo, nil
}
