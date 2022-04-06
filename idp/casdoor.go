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
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/oauth2"
)

type CasdoorIdProvider struct {
	Client *http.Client
	Config *oauth2.Config
	Host   string
}

func NewCasdoorIdProvider(clientId string, clientSecret string, redirectUrl string, hostUrl string) *CasdoorIdProvider {
	idp := &CasdoorIdProvider{}
	config := idp.getConfig(hostUrl)
	config.ClientID = clientId
	config.ClientSecret = clientSecret
	config.RedirectURL = redirectUrl
	idp.Config = config
	idp.Host = hostUrl
	return idp
}

func (idp *CasdoorIdProvider) SetHttpClient(client *http.Client) {
	idp.Client = client
}

func (idp *CasdoorIdProvider) getConfig(hostUrl string) *oauth2.Config {
	return &oauth2.Config{
		Endpoint: oauth2.Endpoint{
			TokenURL: hostUrl + "/api/login/oauth/access_token",
		},
		Scopes: []string{"openid email profile"},
	}
}

type CasdoorToken struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

func (idp *CasdoorIdProvider) GetToken(code string) (*oauth2.Token, error) {
	resp, err := http.PostForm(idp.Config.Endpoint.TokenURL, url.Values{
		"client_id":     {idp.Config.ClientID},
		"client_secret": {idp.Config.ClientSecret},
		"code":          {code},
		"grant_type":    {"authorization_code"},
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}
	pToken := &CasdoorToken{}
	err = json.Unmarshal(body, pToken)
	if err != nil {
		return nil, err
	}

	//check if token is expired
	if pToken.ExpiresIn <= 0 {
		return nil, fmt.Errorf("%s", pToken.AccessToken)
	}
	token := &oauth2.Token{
		AccessToken: pToken.AccessToken,
		Expiry:      time.Unix(time.Now().Unix()+int64(pToken.ExpiresIn), 0),
	}
	return token, nil

}

/*
{
    "sub": "2f80c349-4beb-407f-b1f0-528aac0f1acd",
    "iss": "https://door.casbin.com",
    "aud": "7a11****0fa2172",
    "name": "admin",
    "preferred_username": "Admin",
    "email": "admin@example.com",
    "picture": "https://casbin.org/img/casbin.svg",
    "address": "Guangdong",
    "phone": "12345678910"
}
*/

type CasdoorUserInfo struct {
	Id          string `json:"sub"`
	Name        string `json:"name"`
	DisplayName string `json:"preferred_username"`
	Email       string `json:"email"`
	AvatarUrl   string `json:"picture"`
	Status      string `json:"status"`
	Msg         string `json:"msg"`
}

func (idp *CasdoorIdProvider) GetUserInfo(token *oauth2.Token) (*UserInfo, error) {
	cdUserinfo := &CasdoorUserInfo{}
	accessToken := token.AccessToken
	request, err := http.NewRequest("GET", fmt.Sprintf("%s/api/userinfo", idp.Host), nil)
	if err != nil {
		return nil, err
	}
	//add accesstoken to bearer token
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	resp, err := idp.Client.Do(request)
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, cdUserinfo)
	if err != nil {
		return nil, err
	}

	if cdUserinfo.Status != "" {
		return nil, fmt.Errorf("err: %s", cdUserinfo.Msg)
	}

	userInfo := &UserInfo{
		Id:          cdUserinfo.Id,
		Username:    cdUserinfo.Name,
		DisplayName: cdUserinfo.DisplayName,
		Email:       cdUserinfo.Email,
		AvatarUrl:   cdUserinfo.AvatarUrl,
	}
	return userInfo, nil

}
