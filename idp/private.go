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

type PrivateIdProvider struct {
	Client      *http.Client
	Config      *oauth2.Config
	UserInfoApi string
}

func NewPrivateIdProvider(clientId string, clientSecret string, redirectUrl string, authPage string, tokenApi string, userInfoApi string) *PrivateIdProvider {
	idp := &PrivateIdProvider{}
	idp.UserInfoApi = userInfoApi

	var config = &oauth2.Config{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		RedirectURL:  redirectUrl,
		Endpoint: oauth2.Endpoint{
			AuthURL:  authPage,
			TokenURL: tokenApi,
		},
	}
	idp.Config = config

	return idp
}

func (idp *PrivateIdProvider) SetHttpClient(client *http.Client) {
	idp.Client = client
}

func (idp *PrivateIdProvider) GetToken(code string) (*oauth2.Token, error) {
	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, idp.Client)
	return idp.Config.Exchange(ctx, code)
}

func (idp *PrivateIdProvider) GetUserInfo(token *oauth2.Token) (*UserInfo, error) {
	cdUserinfo := &CasdoorUserInfo{}
	accessToken := token.AccessToken
	request, err := http.NewRequest("GET", idp.UserInfoApi, nil)
	if err != nil {
		return nil, err
	}
	//add accessToken to request header
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
