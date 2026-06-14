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
	"bytes"
	"time"

	"golang.org/x/oauth2"
)

// WeComInternalIdProvider
// This idp is using wecom internal application api as idp
type WeComInternalIdProvider struct {
	Client *http.Client
	Config *oauth2.Config

	UseIdAsName bool
}

func NewWeComInternalIdProvider(clientId string, clientSecret string, redirectUrl string, useIdAsName bool) *WeComInternalIdProvider {
	idp := &WeComInternalIdProvider{}

	config := idp.getConfig(clientId, clientSecret, redirectUrl)
	idp.Config = config
	idp.UseIdAsName = useIdAsName

	return idp
}

func (idp *WeComInternalIdProvider) SetHttpClient(client *http.Client) {
	idp.Client = client
}

func (idp *WeComInternalIdProvider) getConfig(clientId string, clientSecret string, redirectUrl string) *oauth2.Config {
	config := &oauth2.Config{
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
	Errcode  int    `json:"errcode"`
	Errmsg   string `json:"errmsg"`
	UserId   string `json:"UserId"`
	OpenId   string `json:"OpenId"`
	DeviceId string `json:"DeviceId"`

	// returned when scope is snsapi_userinfo
	UserName     string `json:"name"`
	UserAvatar   string `json:"avatar"`
	UserPosition string `json:"position"`
	UserEmail    string `json:"email"`

	// returned when scope is snsapi_privateinfo
	UserTicket string `json:"user_ticket"`
}

type WecomInternalUserDetail struct {
	Errcode int    `json:"errcode"`
	Errmsg  string `json:"errmsg"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Avatar  string `json:"avatar"`
	UserId  string `json:"userid"`
	Mobile  string `json:"mobile"`
}

type WecomInternalUserInfo struct {
	Errcode int    `json:"errcode"`
	Errmsg  string `json:"errmsg"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Avatar  string `json:"avatar"`
	OpenId  string `json:"open_userid"`
	UserId  string `json:"userid"`
}

func (idp *WeComInternalIdProvider) GetUserInfo(token *oauth2.Token) (*UserInfo, error) {
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

	userInfo := UserInfo{
		Id: userResp.UserId,
	}

	// snsapi_privateinfo scope returns user_ticket, use getuserdetail for full private info
	if userResp.UserTicket != "" {
		requestBody := map[string]string{"user_ticket": userResp.UserTicket}
		bs, _ := json.Marshal(requestBody)
		resp, err = idp.Client.Post(
			fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/user/getuserdetail?access_token=%s", accessToken),
			"application/json;charset=UTF-8",
			bytes.NewReader(bs))
		if err != nil {
			return nil, err
		}
		data, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		detailResp := &WecomInternalUserDetail{}
		err = json.Unmarshal(data, detailResp)
		if err != nil {
			return nil, err
		}
		if detailResp.Errcode != 0 {
			return nil, fmt.Errorf("getuserdetail.Errcode = %d, getuserdetail.Errmsg = %s", detailResp.Errcode, detailResp.Errmsg)
		}
		userInfo.Username = detailResp.Name
		userInfo.DisplayName = detailResp.Name
		userInfo.Email = detailResp.Email
		userInfo.AvatarUrl = detailResp.Avatar
	} else {
		// Fall back to user/get for basic info if no user_ticket
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
		userInfo.Username = infoResp.Name
		userInfo.DisplayName = infoResp.Name
		userInfo.Email = infoResp.Email
		userInfo.AvatarUrl = infoResp.Avatar
	}

	if userInfo.Id == "" {
		userInfo.Id = userInfo.Username
	}

	if idp.UseIdAsName {
		userInfo.Username = userInfo.Id
	}

	return &userInfo, nil
}
