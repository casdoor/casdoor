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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"golang.org/x/oauth2"
)

//This idp is using wecom internal application api as idp
type WeComInternalIdProvider struct {
	Client *http.Client
	Config *oauth2.Config
}

func NewWeComInternalIdProvider(clientId string, clientSecret string, redirectUrl string) *WeComInternalIdProvider {
	idp := &WeComInternalIdProvider{}

	config := idp.getConfig(clientId, clientSecret, redirectUrl)
	idp.Config = config

	return idp
}

func (idp *WeComInternalIdProvider) SetHttpClient(client *http.Client) {
	idp.Client = client
}

func (idp *WeComInternalIdProvider) getConfig(clientId string, clientSecret string, redirectUrl string) *oauth2.Config {
	var config = &oauth2.Config{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		RedirectURL:  redirectUrl,
	}

	return config
}

type WecomInterToken struct {
	Errcode     int    `json:"errcode"`
	Errmsg      string `json:"errmsg"`
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

// GetToken use code get access_token (*operation of getting code ought to be done in front)
// get more detail via: https://developer.work.weixin.qq.com/document/path/91039
func (idp *WeComInternalIdProvider) GetToken(code string) (*oauth2.Token, error) {
	pTokenParams := &struct {
		CorpId     string `json:"corpid"`
		Corpsecret string `json:"corpsecret"`
	}{idp.Config.ClientID, idp.Config.ClientSecret}
	resp, err := idp.Client.Get(fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid=%s&corpsecret=%s", pTokenParams.CorpId, pTokenParams.Corpsecret))
	if err != nil {
		return nil, err
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	pToken := &WecomInterToken{}
	err = json.Unmarshal(data, pToken)
	if err != nil {
		return nil, err
	}
	if pToken.Errcode != 0 {
		return nil, fmt.Errorf("pToken.Errcode = %d, pToken.Errmsg = %s", pToken.Errcode, pToken.Errmsg)
	}

	token := &oauth2.Token{
		AccessToken: pToken.AccessToken,
		Expiry:      time.Unix(time.Now().Unix()+int64(pToken.ExpiresIn), 0),
	}

	raw := make(map[string]interface{})
	raw["code"] = code
	token = token.WithExtra(raw)

	return token, nil
}

type WecomInternalUserResp struct {
	Errcode int    `json:"errcode"`
	Errmsg  string `json:"errmsg"`
	UserId  string `json:"UserId"`
	OpenId  string `json:"OpenId"`
}

type WecomInternalUserInfo struct {
	Errcode int    `json:"errcode"`
	Errmsg  string `json:"errmsg"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Avatar  string `json:"avatar"`
	OpenId  string `json:"open_userid"`
}

func (idp *WeComInternalIdProvider) GetUserInfo(token *oauth2.Token) (*UserInfo, error) {
	//Get userid first
	accessToken := token.AccessToken
	code := token.Extra("code").(string)
	resp, err := idp.Client.Get(fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/user/getuserinfo?access_token=%s&code=%s", accessToken, code))
	if err != nil {
		return nil, err
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	userResp := &WecomInternalUserResp{}
	err = json.Unmarshal(data, userResp)
	if err != nil {
		return nil, err
	}
	if userResp.Errcode != 0 {
		return nil, fmt.Errorf("userIdResp.Errcode = %d, userIdResp.Errmsg = %s", userResp.Errcode, userResp.Errmsg)
	}
	if userResp.OpenId != "" {
		return nil, fmt.Errorf("not an internal user")
	}
	//Use userid and accesstoken to get user information
	resp, err = idp.Client.Get(fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/user/get?access_token=%s&userid=%s", accessToken, userResp.UserId))
	if err != nil {
		return nil, err
	}

	data, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	infoResp := &WecomInternalUserInfo{}
	err = json.Unmarshal(data, infoResp)
	if err != nil {
		return nil, err
	}
	if infoResp.Errcode != 0 {
		return nil, fmt.Errorf("userInfoResp.errcode = %d, userInfoResp.errmsg = %s", infoResp.Errcode, infoResp.Errmsg)
	}
	userInfo := UserInfo{
		Id:          infoResp.OpenId,
		Username:    infoResp.Name,
		DisplayName: infoResp.Name,
		Email:       infoResp.Email,
		AvatarUrl:   infoResp.Avatar,
	}

	if userInfo.Id == "" {
		userInfo.Id = userInfo.Username
	}

	return &userInfo, nil
}
