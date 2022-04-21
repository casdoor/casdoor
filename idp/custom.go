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
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	_ "net/url"
	_ "time"

	"golang.org/x/oauth2"
)

type CustomIdProvider struct {
	Client      *http.Client
	Config      *oauth2.Config
	UserInfoUrl string
}

func NewCustomIdProvider(clientId string, clientSecret string, redirectUrl string, authUrl string, tokenUrl string, userInfoUrl string) *CustomIdProvider {
	idp := &CustomIdProvider{}
	idp.UserInfoUrl = userInfoUrl

	var config = &oauth2.Config{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		RedirectURL:  redirectUrl,
		Endpoint: oauth2.Endpoint{
			AuthURL:  authUrl,
			TokenURL: tokenUrl,
		},
	}
	idp.Config = config

	return idp
}

func (idp *CustomIdProvider) SetHttpClient(client *http.Client) {
	idp.Client = client
}

func (idp *CustomIdProvider) GetToken(code string) (*oauth2.Token, error) {
	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, idp.Client)
	return idp.Config.Exchange(ctx, code)
}

type CustomUserInfo struct {
	Id          string `json:"sub"`
	Name        string `json:"name"`
	DisplayName string `json:"preferred_username"`
	Email       string `json:"email"`
	AvatarUrl   string `json:"picture"`
	Status      string `json:"status"`
	Msg         string `json:"msg"`
}

func (idp *CustomIdProvider) GetUserInfo(token *oauth2.Token) (*UserInfo, error) {
	ctUserinfo := &CustomUserInfo{}
	accessToken := token.AccessToken
	request, err := http.NewRequest("GET", idp.UserInfoUrl, nil)
	if err != nil {
		return nil, err
	}
	//add accessToken to request header
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	resp, err := idp.Client.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, ctUserinfo)
	if err != nil {
		return nil, err
	}

	if ctUserinfo.Status != "" {
		return nil, fmt.Errorf("err: %s", ctUserinfo.Msg)
	}

	userInfo := &UserInfo{
		Id:          ctUserinfo.Id,
		Username:    ctUserinfo.Name,
		DisplayName: ctUserinfo.DisplayName,
		Email:       ctUserinfo.Email,
		AvatarUrl:   ctUserinfo.AvatarUrl,
	}
	return userInfo, nil
}
