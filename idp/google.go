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
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"golang.org/x/oauth2"
)

type GoogleIdProvider struct {
	Client *http.Client
	Config *oauth2.Config
}

func NewGoogleIdProvider(clientId string, clientSecret string, redirectUrl string) *GoogleIdProvider {
	idp := &GoogleIdProvider{}

	config := idp.getConfig()
	config.ClientID = clientId
	config.ClientSecret = clientSecret
	config.RedirectURL = redirectUrl
	idp.Config = config

	return idp
}

func (idp *GoogleIdProvider) SetHttpClient(client *http.Client) {
	idp.Client = client
}

func (idp *GoogleIdProvider) getConfig() *oauth2.Config {
	var endpoint = oauth2.Endpoint{
		AuthURL:  "https://accounts.google.com/o/oauth2/auth",
		TokenURL: "https://accounts.google.com/o/oauth2/token",
	}

	var config = &oauth2.Config{
		Scopes:   []string{"profile", "email"},
		Endpoint: endpoint,
	}

	return config
}

func (idp *GoogleIdProvider) GetToken(code string) (*oauth2.Token, error) {
	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, idp.Client)
	return idp.Config.Exchange(ctx, code)
}

//{
//	"id": "110613473084924141234",
//	"email": "jimgreen@gmail.com",
//	"verified_email": true,
//	"name": "Jim Green",
//	"given_name": "Jim",
//	"family_name": "Green",
//	"picture": "https://lh3.googleusercontent.com/-XdUIqdMkCWA/AAAAAAAAAAI/AAAAAAAAAAA/4252rscbv5M/photo.jpg",
//	"locale": "en"
//}

type GoogleUserInfo struct {
	Id            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	Locale        string `json:"locale"`
}

func (idp *GoogleIdProvider) GetUserInfo(token *oauth2.Token) (*UserInfo, error) {
	url := fmt.Sprintf("https://www.googleapis.com/oauth2/v2/userinfo?alt=json&access_token=%s", token.AccessToken)
	resp, err := idp.Client.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var googleUserInfo GoogleUserInfo
	err = json.Unmarshal(body, &googleUserInfo)
	if err != nil {
		return nil, err
	}

	if googleUserInfo.Email == "" {
		return nil, errors.New("google email is empty")
	}

	userInfo := UserInfo{
		Id:          googleUserInfo.Id,
		Username:    googleUserInfo.Email,
		DisplayName: googleUserInfo.Name,
		Email:       googleUserInfo.Email,
		AvatarUrl:   googleUserInfo.Picture,
	}
	return &userInfo, nil
}
