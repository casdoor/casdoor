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
	"net/http"
	"net/url"
	"time"

	"golang.org/x/oauth2"
)

// Wechat Official Accounts
type WeChatOAIdProvider struct {
	Client *http.Client
	Config *oauth2.Config
}

func NewWeChatOAIdProvider(clientId string, clientSecret string, redirectUrl string) *WeChatOAIdProvider {
	idp := &WeChatOAIdProvider{}

	config := idp.getConfig(clientId, clientSecret, redirectUrl)
	idp.Config = config

	return idp
}

func (idp *WeChatOAIdProvider) SetHttpClient(client *http.Client) {
	idp.Client = client
}

// getConfig return a point of Config, which describes a typical 3-legged OAuth2 flow
func (idp *WeChatOAIdProvider) getConfig(clientId string, clientSecret string, redirectUrl string) *oauth2.Config {
	endpoint := oauth2.Endpoint{
		TokenURL: "https://graph.qq.com/oauth2.0/token",
	}

	config := &oauth2.Config{
		Scopes:       []string{"snsapi_login"},
		Endpoint:     endpoint,
		ClientID:     clientId,
		ClientSecret: clientSecret,
		RedirectURL:  redirectUrl,
	}

	return config
}

func (idp *WeChatOAIdProvider) GetToken(code string) (*oauth2.Token, error) {
	params := url.Values{}
	params.Add("grant_type", "authorization_code")
	params.Add("appid", idp.Config.ClientID)
	params.Add("secret", idp.Config.ClientSecret)
	params.Add("code", code)

	token := oauth2.Token{
		AccessToken:  "AccessToken",
		TokenType:    "WeChatAccessToken",
		RefreshToken: "RefreshToken",
		Expiry:       time.Time{},
	}

	raw := make(map[string]interface{})
	raw["Openid"] = code

	return token.WithExtra(raw), nil
}

func (idp *WeChatOAIdProvider) GetUserInfo(token *oauth2.Token) (*UserInfo, error) {
	openid := token.Extra("Openid").(string)
	extra := make(map[string]string)

	extra["wechat_unionid"] = openid

	extra[BuildWechatOpenIdKey(idp.Config.ClientID)] = openid
	userInfo := UserInfo{
		Id:          openid,
		Username:    openid,
		DisplayName: openid,
		AvatarUrl:   "",
		Extra:       extra,
	}
	return &userInfo, nil
}
