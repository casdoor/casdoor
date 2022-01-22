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
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"golang.org/x/oauth2"
)

type BaiduIdProvider struct {
	Client *http.Client
	Config *oauth2.Config
}

func NewBaiduIdProvider(clientId string, clientSecret string, redirectUrl string) *BaiduIdProvider {
	idp := &BaiduIdProvider{}

	config := idp.getConfig()
	config.ClientID = clientId
	config.ClientSecret = clientSecret
	config.RedirectURL = redirectUrl
	idp.Config = config

	return idp
}

func (idp *BaiduIdProvider) SetHttpClient(client *http.Client) {
	idp.Client = client
}

func (idp *BaiduIdProvider) getConfig() *oauth2.Config {
	var endpoint = oauth2.Endpoint{
		AuthURL:  "https://openapi.baidu.com/oauth/2.0/authorize",
		TokenURL: "https://openapi.baidu.com/oauth/2.0/token",
	}

	var config = &oauth2.Config{
		Scopes:   []string{"email"},
		Endpoint: endpoint,
	}

	return config
}

func (idp *BaiduIdProvider) GetToken(code string) (*oauth2.Token, error) {
	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, idp.Client)
	return idp.Config.Exchange(ctx, code)
}

/*
{
    "userid":"2097322476",
    "username":"wl19871011",
    "realname":"阳光",
    "userdetail":"喜欢自由",
    "birthday":"1987-01-01",
    "marriage":"恋爱",
    "sex":"男",
    "blood":"O",
    "constellation":"射手",
    "figure":"小巧",
    "education":"大学/专科",
    "trade":"计算机/电子产品",
    "job":"未知",
    "birthday_year":"1987",
    "birthday_month":"01",
    "birthday_day":"01",
}
*/

type BaiduUserInfo struct {
	OpenId   string `json:"openid"`
	Username string `json:"username"`
	Portrait string `json:"portrait"`
}

func (idp *BaiduIdProvider) GetUserInfo(token *oauth2.Token) (*UserInfo, error) {
	resp, err := idp.Client.Get(fmt.Sprintf("https://openapi.baidu.com/rest/2.0/passport/users/getInfo?access_token=%s", token.AccessToken))
	if err != nil {
		return nil, err
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	baiduUser := BaiduUserInfo{}
	if err = json.Unmarshal(data, &baiduUser); err != nil {
		return nil, err
	}

	userInfo := UserInfo{
		Id:        baiduUser.OpenId,
		Username:  baiduUser.Username,
		AvatarUrl: fmt.Sprintf("https://himg.bdimg.com/sys/portrait/item/%s", baiduUser.Portrait),
	}
	return &userInfo, nil
}
