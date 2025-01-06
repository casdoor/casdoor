// Copyright 2024 The Casdoor Authors. All Rights Reserved.
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
	"time"

	"golang.org/x/oauth2"
)

type KwaiIdProvider struct {
	Client *http.Client
	Config *oauth2.Config
}

func NewKwaiIdProvider(clientId string, clientSecret string, redirectUrl string) *KwaiIdProvider {
	idp := &KwaiIdProvider{}
	idp.Config = idp.getConfig(clientId, clientSecret, redirectUrl)
	return idp
}

func (idp *KwaiIdProvider) SetHttpClient(client *http.Client) {
	idp.Client = client
}

func (idp *KwaiIdProvider) getConfig(clientId string, clientSecret string, redirectUrl string) *oauth2.Config {
	endpoint := oauth2.Endpoint{
		TokenURL: "https://open.kuaishou.com/oauth2/access_token",
		AuthURL:  "https://open.kuaishou.com/oauth2/authorize", // qr code: /oauth2/connect
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

type KwaiTokenResp struct {
	Result                int      `json:"result"`
	ErrorMsg              string   `json:"error_msg"`
	AccessToken           string   `json:"access_token"`
	ExpiresIn             int      `json:"expires_in"`
	RefreshToken          string   `json:"refresh_token"`
	RefreshTokenExpiresIn int      `json:"refresh_token_expires_in"`
	OpenId                string   `json:"open_id"`
	Scopes                []string `json:"scopes"`
}

// GetToken use code to get access_token
func (idp *KwaiIdProvider) GetToken(code string) (*oauth2.Token, error) {
	params := map[string]string{
		"app_id":     idp.Config.ClientID,
		"app_secret": idp.Config.ClientSecret,
		"code":       code,
		"grant_type": "authorization_code",
	}
	tokenUrl := fmt.Sprintf("%s?app_id=%s&app_secret=%s&code=%s&grant_type=authorization_code",
		idp.Config.Endpoint.TokenURL, params["app_id"], params["app_secret"], params["code"])
	resp, err := idp.Client.Get(tokenUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var tokenResp KwaiTokenResp
	err = json.Unmarshal(body, &tokenResp)
	if err != nil {
		return nil, err
	}
	if tokenResp.Result != 1 {
		return nil, fmt.Errorf("get token error: %s", tokenResp.ErrorMsg)
	}

	token := &oauth2.Token{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		Expiry:       time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
	}

	raw := make(map[string]interface{})
	raw["open_id"] = tokenResp.OpenId
	token = token.WithExtra(raw)

	return token, nil
}

// More details: https://open.kuaishou.com/openapi/user_info
type KwaiUserInfo struct {
	Result   int    `json:"result"`
	ErrorMsg string `json:"error_msg"`
	UserInfo struct {
		Head string `json:"head"`
		Name string `json:"name"`
		Sex  string `json:"sex"`
		City string `json:"city"`
	} `json:"user_info"`
}

// GetUserInfo use token to get user profile
func (idp *KwaiIdProvider) GetUserInfo(token *oauth2.Token) (*UserInfo, error) {
	userInfoUrl := fmt.Sprintf("https://open.kuaishou.com/openapi/user_info?app_id=%s&access_token=%s",
		idp.Config.ClientID, token.AccessToken)

	resp, err := idp.Client.Get(userInfoUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var kwaiUserInfo KwaiUserInfo
	err = json.Unmarshal(body, &kwaiUserInfo)
	if err != nil {
		return nil, err
	}

	if kwaiUserInfo.Result != 1 {
		return nil, fmt.Errorf("get user info error: %s", kwaiUserInfo.ErrorMsg)
	}

	userInfo := &UserInfo{
		Id:          token.Extra("open_id").(string),
		Username:    kwaiUserInfo.UserInfo.Name,
		DisplayName: kwaiUserInfo.UserInfo.Name,
		AvatarUrl:   kwaiUserInfo.UserInfo.Head,
		Extra: map[string]string{
			"gender": kwaiUserInfo.UserInfo.Sex,
			"city":   kwaiUserInfo.UserInfo.City,
		},
	}

	return userInfo, nil
}
