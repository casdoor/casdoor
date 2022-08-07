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
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

type InfoflowIdProvider struct {
	Client  *http.Client
	Config  *oauth2.Config
	AgentId string
	Ticket  string
}

func NewInfoflowIdProvider(clientId string, clientSecret string, appId string, redirectUrl string) *InfoflowIdProvider {
	idp := &InfoflowIdProvider{}

	config := idp.getConfig(clientId, clientSecret, redirectUrl)
	idp.Config = config
	idp.AgentId = appId
	return idp
}

func (idp *InfoflowIdProvider) SetHttpClient(client *http.Client) {
	idp.Client = client
}

func (idp *InfoflowIdProvider) getConfig(clientId string, clientSecret string, redirectUrl string) *oauth2.Config {
	config := &oauth2.Config{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		RedirectURL:  redirectUrl,
	}

	return config
}

type InfoflowToken struct {
	Errcode     int    `json:"errcode"`
	Errmsg      string `json:"errmsg"`
	AccessToken string `json:"suite_access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

// get more detail via: https://qy.baidu.com/doc/index.html#/third_serverapi/authority
func (idp *InfoflowIdProvider) GetToken(code string) (*oauth2.Token, error) {
	pTokenParams := &struct {
		SuiteId     string `json:"suite_id"`
		SuiteSecret string `json:"suite_secret"`
		SuiteTicket string `json:"suite_ticket"`
	}{idp.Config.ClientID, idp.Config.ClientSecret, idp.Ticket}
	data, err := idp.postWithBody(pTokenParams, "https://api.im.baidu.com/api/service/get_suite_token")

	pToken := &InfoflowToken{}
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

/*
{
    "errcode": 0,
    "errmsg": "ok",
    "userid": "lili",
    "name": "丽丽",
    "department": [1],
    "mobile": "13500088888",
    "email": "lili4@gzdev.com",
    "imid": 40000318,
    "hiuname": "lili4",
    "status": 1,
    "extattr": {
        "attrs": [
            {
                "name": "爱好",
                "value": "旅游"
            },
            {
                "name": "卡号",
                "value": "1234567234"
            }
        ]
    },
    "lm" : 14236463257
}
*/

type InfoflowUserResp struct {
	Errcode int    `json:"errcode"`
	Errmsg  string `json:"errmsg"`
	UserId  string `json:"UserId"`
}

type InfoflowUserInfo struct {
	Errcode int    `json:"errcode"`
	Errmsg  string `json:"errmsg"`
	Imid    string `json:"imid"`
	Name    string `json:"name"`
	Email   string `json:"email"`
}

// get more detail via: https://qy.baidu.com/doc/index.html#/third_serverapi/contacts?id=%e8%8e%b7%e5%8f%96%e6%88%90%e5%91%98
func (idp *InfoflowIdProvider) GetUserInfo(token *oauth2.Token) (*UserInfo, error) {
	// Get userid first
	accessToken := token.AccessToken
	code := token.Extra("code").(string)
	resp, err := idp.Client.Get(fmt.Sprintf("https://api.im.baidu.com/api/user/getuserinfo?access_token=%s&code=%s&agentid=%s", accessToken, code, idp.AgentId))
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	userResp := &InfoflowUserResp{}
	err = json.Unmarshal(data, userResp)
	if err != nil {
		return nil, err
	}
	if userResp.Errcode != 0 {
		return nil, fmt.Errorf("userIdResp.Errcode = %d, userIdResp.Errmsg = %s", userResp.Errcode, userResp.Errmsg)
	}
	// Use userid and accesstoken to get user information
	resp, err = idp.Client.Get(fmt.Sprintf("https://api.im.baidu.com/api/user/get?access_token=%s&userid=%s", accessToken, userResp.UserId))
	if err != nil {
		return nil, err
	}

	data, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	infoResp := &InfoflowUserInfo{}
	err = json.Unmarshal(data, infoResp)
	if err != nil {
		return nil, err
	}
	if infoResp.Errcode != 0 {
		return nil, fmt.Errorf("userInfoResp.errcode = %d, userInfoResp.errmsg = %s", infoResp.Errcode, infoResp.Errmsg)
	}
	userInfo := UserInfo{
		Id:          infoResp.Imid,
		Username:    infoResp.Name,
		DisplayName: infoResp.Name,
		Email:       infoResp.Email,
	}

	if userInfo.Id == "" {
		userInfo.Id = userInfo.Username
	}
	return &userInfo, nil
}

func (idp *InfoflowIdProvider) postWithBody(body interface{}, url string) ([]byte, error) {
	bs, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	r := strings.NewReader(string(bs))
	resp, err := idp.Client.Post(url, "application/json;charset=UTF-8", r)
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadAll(resp.Body)
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
