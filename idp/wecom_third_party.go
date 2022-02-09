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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

type WeComIdProvider struct {
	Client *http.Client
	Config *oauth2.Config
}

func NewWeComIdProvider(clientId string, clientSecret string, redirectUrl string) *WeComIdProvider {
	idp := &WeComIdProvider{}

	config := idp.getConfig(clientId, clientSecret, redirectUrl)
	idp.Config = config

	return idp
}

func (idp *WeComIdProvider) SetHttpClient(client *http.Client) {
	idp.Client = client
}

// getConfig return a point of Config, which describes a typical 3-legged OAuth2 flow
func (idp *WeComIdProvider) getConfig(clientId string, clientSecret string, redirectUrl string) *oauth2.Config {
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

type WeComProviderToken struct {
	Errcode             int    `json:"errcode"`
	Errmsg              string `json:"errmsg"`
	ProviderAccessToken string `json:"provider_access_token"`
	ExpiresIn           int    `json:"expires_in"`
}

// GetToken use code get access_token (*operation of getting code ought to be done in front)
// get more detail via: https://developers.weixin.qq.com/doc/oplatform/Website_App/WeChat_Login/Wechat_Login.html
func (idp *WeComIdProvider) GetToken(code string) (*oauth2.Token, error) {
	pTokenParams := &struct {
		CorpId         string `json:"corpid"`
		ProviderSecret string `json:"provider_secret"`
	}{idp.Config.ClientID, idp.Config.ClientSecret}
	data, err := idp.postWithBody(pTokenParams, "https://qyapi.weixin.qq.com/cgi-bin/service/get_provider_token")

	pToken := &WeComProviderToken{}
	err = json.Unmarshal(data, pToken)
	if err != nil {
		return nil, err
	}
	if pToken.Errcode != 0 {
		return nil, fmt.Errorf("pToken.Errcode = %d, pToken.Errmsg = %s", pToken.Errcode, pToken.Errmsg)
	}

	token := &oauth2.Token{
		AccessToken: pToken.ProviderAccessToken,
		Expiry:      time.Unix(time.Now().Unix()+int64(pToken.ExpiresIn), 0),
	}

	raw := make(map[string]interface{})
	raw["code"] = code
	token = token.WithExtra(raw)

	return token, nil
}

/*
{
   "errcode":0,
   "errmsg":"ok",
   "usertype": 1,
   "user_info":{
       "userid":"xxxx",
       "open_userid":"xxx",
       "name":"xxxx",
       "avatar":"xxxx"
   },
   "corp_info":{
       "corpid":"wxCorpId",
    },
   "agent":[
       {"agentid":0,"auth_type":1},
       {"agentid":1,"auth_type":1},
       {"agentid":2,"auth_type":1}
   ],
   "auth_info":{
       "department":[
           {
               "id":2,
               "writable":true
           }
       ]
   }
}
*/

type WeComUserInfo struct {
	Errcode  int    `json:"errcode"`
	Errmsg   string `json:"errmsg"`
	Usertype int    `json:"usertype"`
	UserInfo struct {
		Userid     string `json:"userid"`
		OpenUserid string `json:"open_userid"`
		Name       string `json:"name"`
		Avatar     string `json:"avatar"`
	} `json:"user_info"`
	CorpInfo struct {
		Corpid string `json:"corpid"`
	} `json:"corp_info"`
	Agent []struct {
		Agentid  int `json:"agentid"`
		AuthType int `json:"auth_type"`
	} `json:"agent"`
	AuthInfo struct {
		Department []struct {
			Id       int  `json:"id"`
			Writable bool `json:"writable"`
		} `json:"department"`
	} `json:"auth_info"`
}

// GetUserInfo use WeComProviderToken gotten before return WeComUserInfo
// get more detail via: https://work.weixin.qq.com/api/doc/90001/90143/91125
func (idp *WeComIdProvider) GetUserInfo(token *oauth2.Token) (*UserInfo, error) {
	accessToken := token.AccessToken
	code := token.Extra("code").(string)

	requestBody := &struct {
		AuthCode string `json:"auth_code"`
	}{code}
	data, err := idp.postWithBody(requestBody, fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/service/get_login_info?access_token=%s", accessToken))
	if err != nil {
		return nil, err
	}

	wecomUserInfo := WeComUserInfo{}
	err = json.Unmarshal(data, &wecomUserInfo)
	if err != nil {
		return nil, err
	}
	if wecomUserInfo.Errcode != 0 {
		return nil, fmt.Errorf("wecomUserInfo.Errcode = %d, wecomUserInfo.Errmsg = %s", wecomUserInfo.Errcode, wecomUserInfo.Errmsg)
	}

	userInfo := UserInfo{
		Id:          wecomUserInfo.UserInfo.OpenUserid,
		Username:    wecomUserInfo.UserInfo.Name,
		DisplayName: wecomUserInfo.UserInfo.Name,
		AvatarUrl:   wecomUserInfo.UserInfo.Avatar,
	}
	return &userInfo, nil
}

func (idp *WeComIdProvider) postWithBody(body interface{}, url string) ([]byte, error) {
	bs, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	r := strings.NewReader(string(bs))
	resp, err := idp.Client.Post(url, "application/json;charset=UTF-8", r)
	if err != nil {
		return nil, err
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(resp.Body)

	return data, nil
}
