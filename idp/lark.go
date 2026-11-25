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
	"io"
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
		TokenURL: idp.LarkDomain + "/open-apis/authen/v1/access_token",
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

type LarkUserData struct {
	AccessToken     string `json:"access_token"`
	ExpiresIn       int    `json:"expires_in"`
	RefreshToken    string `json:"refresh_token"`
	Name            string `json:"name"`
	AvatarUrl       string `json:"avatar_url"`
	OpenId          string `json:"open_id"`
	UnionId         string `json:"union_id"`
	Email           string `json:"email"`
	EnterpriseEmail string `json:"enterprise_email"`
	UserId          string `json:"user_id"`
	Mobile          string `json:"mobile"`
}

type larkResponse struct {
	Code              int          `json:"code"`
	Msg               string       `json:"msg"`
	TenantAccessToken string       `json:"tenant_access_token"`
	Data              LarkUserData `json:"data"`
}

func (idp *LarkIdProvider) request(method string, url string, bearer string, body interface{}) (*larkResponse, error) {
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json; charset=utf-8")
	}
	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}

	resp, err := idp.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	res := &larkResponse{}
	if err = json.NewDecoder(resp.Body).Decode(res); err != nil {
		return nil, fmt.Errorf("invalid Lark response from %s (HTTP %d): %v", url, resp.StatusCode, err)
	}
	if res.Code != 0 {
		return nil, fmt.Errorf("Lark request to %s failed, code: %d, msg: %s", url, res.Code, res.Msg)
	}
	return res, nil
}

func (idp *LarkIdProvider) GetToken(code string) (*oauth2.Token, error) {
	appParams := map[string]string{"app_id": idp.Config.ClientID, "app_secret": idp.Config.ClientSecret}
	appRes, err := idp.request(http.MethodPost, idp.LarkDomain+"/open-apis/auth/v3/tenant_access_token/internal", "", appParams)
	if err != nil {
		return nil, err
	}

	codeParams := map[string]string{"grant_type": "authorization_code", "code": code}
	userRes, err := idp.request(http.MethodPost, idp.Config.Endpoint.TokenURL, appRes.TenantAccessToken, codeParams)
	if err != nil {
		return nil, err
	}
	if userRes.Data.AccessToken == "" {
		return nil, fmt.Errorf("Lark token response is missing access_token")
	}

	token := &oauth2.Token{
		AccessToken:  userRes.Data.AccessToken,
		TokenType:    "Bearer",
		RefreshToken: userRes.Data.RefreshToken,
	}
	if userRes.Data.ExpiresIn > 0 {
		token.Expiry = time.Now().Add(time.Duration(userRes.Data.ExpiresIn) * time.Second)
	}

	data := userRes.Data
	return token.WithExtra(map[string]interface{}{"larkUserData": &data}), nil
}

func (idp *LarkIdProvider) GetUserInfo(token *oauth2.Token) (*UserInfo, error) {
	data, ok := token.Extra("larkUserData").(*LarkUserData)
	if !ok || data == nil {
		res, err := idp.request(http.MethodGet, idp.LarkDomain+"/open-apis/authen/v1/user_info", token.AccessToken, nil)
		if err != nil {
			return nil, err
		}
		data = &res.Data
	}
	if data.OpenId == "" {
		return nil, fmt.Errorf("Lark user info is missing open_id")
	}

	// Use enterprise_email as fallback when email is empty
	email := data.Email
	if email == "" {
		email = data.EnterpriseEmail
	}

	// Use fallback mechanism for username: UserId -> UnionId -> OpenId
	username := data.UserId
	if username == "" {
		username = data.UnionId
	}
	if username == "" {
		username = data.OpenId
	}

	var phoneNumber string
	var countryCode string
	if len(data.Mobile) != 0 {
		phoneNumberParsed, err := phonenumbers.Parse(data.Mobile, "")
		if err != nil {
			return nil, err
		}
		countryCode = phonenumbers.GetRegionCodeForNumber(phoneNumberParsed)
		phoneNumber = fmt.Sprintf("%d", phoneNumberParsed.GetNationalNumber())
	}

	userInfo := UserInfo{
		Id:          data.OpenId,
		DisplayName: data.Name,
		Username:    username,
		Email:       email,
		// enterprise directory email
		EmailVerified: email != "",
		AvatarUrl:     data.AvatarUrl,
		Phone:         phoneNumber,
		PhoneVerified: phoneNumber != "",
		CountryCode:   countryCode,
	}
	return &userInfo, nil
}
