// Copyright 2022 The Casdoor Authors. All Rights Reserved.
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
	"net/url"
	"time"

	"golang.org/x/oauth2"
)

type DouyinIdProvider struct {
	Client *http.Client
	Config *oauth2.Config
}

func NewDouyinIdProvider(clientId string, clientSecret string, redirectUrl string) *DouyinIdProvider {
	idp := &DouyinIdProvider{}
	idp.Config = idp.getConfig(clientId, clientSecret, redirectUrl)
	return idp
}

func (idp *DouyinIdProvider) SetHttpClient(client *http.Client) {
	idp.Client = client
}

func (idp *DouyinIdProvider) getConfig(clientId string, clientSecret string, redirectUrl string) *oauth2.Config {
	endpoint := oauth2.Endpoint{
		TokenURL: "https://open.douyin.com/oauth/access_token",
		AuthURL:  "https://open.douyin.com/platform/oauth/connect",
	}

	config := &oauth2.Config{
		Scopes:       []string{"user_info"},
		Endpoint:     endpoint,
		ClientID:     clientId,
		ClientSecret: clientSecret,
		RedirectURL:  redirectUrl,
	}

	return config
}

// get more details via: https://open.douyin.com/platform/doc?doc=docs/openapi/account-permission/get-access-token
/*
{
  "data": {
    "access_token": "access_token",
    "description": "",
    "error_code": "0",
    "expires_in": "86400",
    "open_id": "aaa-bbb-ccc",
    "refresh_expires_in": "86400",
    "refresh_token": "refresh_token",
    "scope": "user_info"
  },
  "message": "<nil>"
}
*/

type DouyinTokenResp struct {
	Data struct {
		AccessToken  string `json:"access_token"`
		ExpiresIn    int64  `json:"expires_in"`
		OpenId       string `json:"open_id"`
		RefreshToken string `json:"refresh_token"`
		Scope        string `json:"scope"`
	} `json:"data"`
	Message string `json:"message"`
}

// GetToken use code to get access_token
// get more details via: https://open.douyin.com/platform/doc?doc=docs/openapi/account-permission/get-access-token
func (idp *DouyinIdProvider) GetToken(code string) (*oauth2.Token, error) {
	payload := url.Values{}
	payload.Set("code", code)
	payload.Set("grant_type", "authorization_code")
	payload.Set("client_key", idp.Config.ClientID)
	payload.Set("client_secret", idp.Config.ClientSecret)
	resp, err := idp.Client.PostForm(idp.Config.Endpoint.TokenURL, payload)
	if err != nil {
		return nil, err
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	tokenResp := &DouyinTokenResp{}
	err = json.Unmarshal(data, tokenResp)
	if err != nil {
		return nil, fmt.Errorf("fail to unmarshal token response: %s", err.Error())
	}

	token := &oauth2.Token{
		AccessToken:  tokenResp.Data.AccessToken,
		RefreshToken: tokenResp.Data.RefreshToken,
		Expiry:       time.Unix(time.Now().Unix()+tokenResp.Data.ExpiresIn, 0),
	}

	raw := make(map[string]interface{})
	raw["open_id"] = tokenResp.Data.OpenId
	token = token.WithExtra(raw)

	return token, nil
}

// get more details via: https://open.douyin.com/platform/doc?doc=docs/openapi/account-management/get-account-open-info
/*
{
  "data": {
    "avatar": "https://example.com/x.jpeg",
    "city": "上海",
    "country": "中国",
    "description": "",
    "e_account_role": "<nil>",
    "error_code": "0",
    "gender": "<nil>",
    "nickname": "张伟",
    "open_id": "0da22181-d833-447f-995f-1beefea5bef3",
    "province": "上海",
    "union_id": "1ad4e099-4a0c-47d1-a410-bffb4f2f64a4"
  }
}
*/

type DouyinUserInfo struct {
	Data struct {
		Avatar  string `json:"avatar"`
		City    string `json:"city"`
		Country string `json:"country"`
		// 0->unknown, 1->male, 2->female
		Gender   int64  `json:"gender"`
		Nickname string `json:"nickname"`
		OpenId   string `json:"open_id"`
		Province string `json:"province"`
	} `json:"data"`
}

// GetUserInfo use token to get user profile
func (idp *DouyinIdProvider) GetUserInfo(token *oauth2.Token) (*UserInfo, error) {
	body := &struct {
		AccessToken string `json:"access_token"`
		OpenId      string `json:"open_id"`
	}{token.AccessToken, token.Extra("open_id").(string)}
	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("GET", "https://open.douyin.com/oauth/userinfo/", bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Add("access-token", token.AccessToken)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	resp, err := idp.Client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var douyinUserInfo DouyinUserInfo
	err = json.Unmarshal(respBody, &douyinUserInfo)
	if err != nil {
		return nil, err
	}

	userInfo := UserInfo{
		Id:          douyinUserInfo.Data.OpenId,
		Username:    douyinUserInfo.Data.Nickname,
		DisplayName: douyinUserInfo.Data.Nickname,
		AvatarUrl:   douyinUserInfo.Data.Avatar,
	}
	return &userInfo, nil
}
