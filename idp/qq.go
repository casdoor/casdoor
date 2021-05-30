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
	"net/url"
	"regexp"

	"golang.org/x/oauth2"
)

type QqIdProvider struct {
	Client *http.Client
	Config *oauth2.Config
}

func NewQqIdProvider(clientId string, clientSecret string, redirectUrl string) *QqIdProvider {
	idp := &QqIdProvider{}

	config := idp.getConfig()
	config.ClientID = clientId
	config.ClientSecret = clientSecret
	config.RedirectURL = redirectUrl
	idp.Config = config

	return idp
}

func (idp *QqIdProvider) SetHttpClient(client *http.Client) {
	idp.Client = client
}

func (idp *QqIdProvider) getConfig() *oauth2.Config {
	var endpoint = oauth2.Endpoint{
		TokenURL: "https://graph.qq.com/oauth2.0/token",
	}

	var config = &oauth2.Config{
		Scopes:   []string{"get_user_info"},
		Endpoint: endpoint,
	}

	return config
}

func (idp *QqIdProvider) GetToken(code string) (*oauth2.Token, error) {
	params := url.Values{}
	params.Add("grant_type", "authorization_code")
	params.Add("client_id", idp.Config.ClientID)
	params.Add("client_secret", idp.Config.ClientSecret)
	params.Add("code", code)
	params.Add("redirect_uri", idp.Config.RedirectURL)

	getAccessTokenUrl := fmt.Sprintf("https://graph.qq.com/oauth2.0/token?%s", params.Encode())
	resp, err := idp.Client.Get(getAccessTokenUrl)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	tokenContent, err := ioutil.ReadAll(resp.Body)

	re := regexp.MustCompile("token=(.*?)&")
	matched := re.FindAllStringSubmatch(string(tokenContent), -1)
	accessToken := matched[0][1]
	token := &oauth2.Token{
		AccessToken: accessToken,
		TokenType:   "Bearer",
	}
	return token, nil
}

func (idp *QqIdProvider) GetUserInfo(token *oauth2.Token) (*UserInfo, error) {
	getOpenIdUrl := fmt.Sprintf("https://graph.qq.com/oauth2.0/me?access_token=%s", token.AccessToken)
	resp, err := idp.Client.Get(getOpenIdUrl)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	openIdBody, err := ioutil.ReadAll(resp.Body)

	re := regexp.MustCompile("\"openid\":\"(.*?)\"}")
	matched := re.FindAllStringSubmatch(string(openIdBody), -1)
	openId := matched[0][1]
	if openId == "" {
		return nil, errors.New("openId is empty")
	}

	getUserInfoUrl := fmt.Sprintf("https://graph.qq.com/user/get_user_info?access_token=%s&oauth_consumer_key=%s&openid=%s", token.AccessToken, idp.Config.ClientID, openId)
	resp, err = idp.Client.Get(getUserInfoUrl)
	if err != nil {
		return nil, err
	}

	type response struct {
		Ret       int    `json:"ret"`
		Nickname  string `json:"nickname"`
		AvatarUrl string `json:"figureurl_qq_1"`
	}

	defer resp.Body.Close()
	userInfoContent, err := ioutil.ReadAll(resp.Body)
	var userResponse response
	err = json.Unmarshal(userInfoContent, &userResponse)
	if err != nil {
		return nil, err
	}
	if userResponse.Ret != 0 {
		return nil, errors.New(fmt.Sprintf("ret expected 0, got %d", userResponse.Ret))
	}

	userInfo := UserInfo{
		Username:    openId,
		DisplayName: userResponse.Nickname,
		AvatarUrl:   userResponse.AvatarUrl,
	}
	return &userInfo, nil
}
