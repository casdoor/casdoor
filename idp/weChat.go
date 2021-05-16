// Copyright 2021 The casbin Authors. All Rights Reserved.
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
	"golang.org/x/oauth2"
	"io"
	"net/http"
	"net/url"
	"time"
)

type WeChatIdProvider struct {
	Client *http.Client
	// Config:
	// - ClientID: Application unique identifier, used to request CODE, get/refresh access_token through CODE, this information can be obtained by the front end
	// - ClientSecret: Application key, this information is stored in the back-end, and the front-end is not available [such as stored in environment variables]
	// - RedirectURL: After the user allows authorization, it will be redirected to the URL of redirect_uri with the code and state parameters
	Config *oauth2.Config
}

type TencentAccessToken struct {
	AccessToken  string `json:"access_token"`  //Interface call credentials
	ExpiresIn    int64  `json:"expires_in"`    //access_token interface call credential timeout time, unit (seconds)
	RefreshToken string `json:"refresh_token"` //User refresh access_token
	Openid       string `json:"openid"`        //Unique ID of authorized user
	Scope        string `json:"scope"`         //The scope of user authorization, separated by commas. (,)
	Unionid      string `json:"unionid"`       //This field will appear if and only if the website application has been authorized by the user's UserInfo.
}

type TencentUserInfo struct {
	Openid     string   `json:"openid"`     //The ID of an ordinary user, which is unique to the current developer account
	Nickname   string   `json:"nickname"`   //Ordinary user nickname
	Sex        int      `json:"sex"`        //Ordinary user gender, 1 is male, 2 is female
	Province   string   `json:"province"`   //Province filled in by ordinary user's personal information
	City       string   `json:"city"`       //City filled in by general user's personal data
	Country    string   `json:"country"`    //Country, such as China is CN
	Headimgurl string   `json:"headimgurl"` //User avatar, the last value represents the size of the square avatar (there are optional values of 0, 46, 64, 96, 132, 0 represents a 640*640 square avatar), this item is empty when the user does not have a avatar
	Privilege  []string `json:"privilege"`  //User Privilege information, json array, such as Wechat Woka user (chinaunicom)
	Unionid    string   `json:"unionid"`    //Unified user identification. For an application under a WeChat open platform account, the unionid of the same user is unique.
}

func NewWeChatIdProvider(clientId string, clientSecret string, redirectUrl string) *WeChatIdProvider {
	idp := &WeChatIdProvider{}

	config := idp.getConfig(clientId, clientSecret, redirectUrl)
	idp.Config = config

	return idp
}

func (idp *WeChatIdProvider) SetHttpClient(client *http.Client) {
	idp.Client = client
}

// getConfig return a point of Config, which describes a typical 3-legged OAuth2 flow
func (idp *WeChatIdProvider) getConfig(clientId string, clientSecret string, redirectUrl string) *oauth2.Config {
	var endpoint = oauth2.Endpoint{
		TokenURL: "https://graph.qq.com/oauth2.0/token",
	}

	var config = &oauth2.Config{
		Scopes:       []string{"snsapi_login"},
		Endpoint:     endpoint,
		ClientID:     clientId,
		ClientSecret: clientSecret,
		RedirectURL:  redirectUrl,
	}

	return config
}

// GetToken use code get access_token (*operation of getting code ought to be done in front)
// get more detail via: https://developers.weixin.qq.com/doc/oplatform/Website_App/WeChat_Login/Wechat_Login.html
func (idp *WeChatIdProvider) GetToken(code string) (*oauth2.Token, error) {
	params := url.Values{}
	params.Add("grant_type", "authorization_code")
	params.Add("appid", idp.Config.ClientID)
	params.Add("secret", idp.Config.ClientSecret)
	params.Add("code", code)

	getAccessTokenUrl := fmt.Sprintf("https://api.weixin.qq.com/sns/oauth2/access_token?%s", params.Encode())
	tokenResponse, err := idp.Client.Get(getAccessTokenUrl)
	if err != nil {
		return nil, err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(tokenResponse.Body)

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(tokenResponse.Body)
	if err != nil {
		return nil, err
	}

	var tencentAccessToken TencentAccessToken
	if err = json.Unmarshal([]byte(buf.String()), &tencentAccessToken); err != nil {
		return nil, err
	}

	token := oauth2.Token{
		AccessToken:  tencentAccessToken.AccessToken,
		TokenType:    "WeChatAccessToken",
		RefreshToken: tencentAccessToken.RefreshToken,
		Expiry:       time.Time{},
	}

	raw := make(map[string]string)
	raw["Openid"] = tencentAccessToken.Openid
	token.WithExtra(raw)

	return &token, nil
}

// GetUserInfo use TencentAccessToken gotten before return TencentUserInfo
// get more detail via: https://developers.weixin.qq.com/doc/oplatform/Website_App/WeChat_Login/Authorized_Interface_Calling_UnionID.html
func (idp *WeChatIdProvider) GetUserInfo(token *oauth2.Token) (*UserInfo, error) {
	var tencentUserInfo TencentUserInfo
	accessToken := token.AccessToken
	openid := token.Extra("Openid")

	getUserInfoUrl := fmt.Sprintf("https://api.weixin.qq.com/sns/userinfo?access_token=%s&openid=%s", accessToken, openid)
	getUserInfoResponse, err := idp.Client.Get(getUserInfoUrl)
	if err != nil {
		return nil, err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(getUserInfoResponse.Body)

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(getUserInfoResponse.Body)
	if err != nil {
		return nil, err
	}
	if err = json.Unmarshal([]byte(buf.String()), &tencentUserInfo); err != nil {
		return nil, err
	}

	userInfo := UserInfo{
		Username:  tencentUserInfo.Nickname,
		Email:     "",
		AvatarUrl: tencentUserInfo.Headimgurl,
	}

	return &userInfo, nil
}
