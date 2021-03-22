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
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"

	"golang.org/x/oauth2"
)

type QqIdProvider struct {
	ClientId     string
}

func (idp *QqIdProvider) GetConfig() *oauth2.Config {
	var endpoint = oauth2.Endpoint{
		TokenURL: "https://graph.qq.com/oauth2.0/token",
	}

	var config = &oauth2.Config{
		Scopes:   []string{"profile", "email"},
		Endpoint: endpoint,
	}

	return config
}

func (idp *QqIdProvider) GetUserInfo(httpClient *http.Client, token *oauth2.Token) (string, string, string, error) {
	var email, username, avatarUrl string

	type userInfoFromQq struct {
		Ret       int    `json:"ret"`
		Nickname  string `json:"nickname"`
		AvatarUrl string `json:"figureurl_qq_1"`
	}

	getOpenIdUrl := fmt.Sprintf("https://graph.qq.com/oauth2.0/me?access_token=%s", token)

	openIdResponse, err := httpClient.Get(getOpenIdUrl)
	if err != nil {
		panic(err)
	}
	defer openIdResponse.Body.Close()
	openIdContent, err := ioutil.ReadAll(openIdResponse.Body)

	openIdReg := regexp.MustCompile("\"openid\":\"(.*?)\"}")
	openIdRegRes := openIdReg.FindAllStringSubmatch(string(openIdContent), -1)
	openId := openIdRegRes[0][1]

	if openId == "" {
		return "", "", "", errors.New("openId is empty")
	}

	getUserInfoUrl := fmt.Sprintf("https://graph.qq.com/user/get_user_info?access_token=%s&oauth_consumer_key=%s&openid=%s", token, idp.ClientId, openId)
	getUserInfoResponse, err := httpClient.Get(getUserInfoUrl)
	if err != nil {
		panic(err)
	}
	defer getUserInfoResponse.Body.Close()
	userInfoContent, err := ioutil.ReadAll(getUserInfoResponse.Body)
	var userInfo userInfoFromQq
	err = json.Unmarshal(userInfoContent, &userInfo)
	if err != nil || userInfo.Ret != 0 {
		return "", "", "", err
	}

	email = ""
	username = userInfo.Nickname
	avatarUrl = userInfo.AvatarUrl

	return email, username, avatarUrl, nil
}
