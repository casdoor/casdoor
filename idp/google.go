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
	"errors"
	"io/ioutil"
	"net/http"

	"golang.org/x/oauth2"
)

type GoogleIdProvider struct {
	Client       *http.Client
	Config       *oauth2.Config
	ClientId     string
	ClientSecret string
	RedirectUrl  string
}

func NewGoogleIdProvider(clientId string, clientSecret string, redirectUrl string) *GoogleIdProvider {
	idp := &GoogleIdProvider{
		ClientId:     clientId,
		ClientSecret: clientSecret,
		RedirectUrl:  redirectUrl,
	}

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
	ctx := context.WithValue(oauth2.NoContext, oauth2.HTTPClient, idp.Client)
	return idp.Config.Exchange(ctx, code)
}

func (idp *GoogleIdProvider) GetUserInfo(token *oauth2.Token) (*UserInfo, error) {
	userInfo := &UserInfo{}

	type response struct {
		Picture string `json:"picture"`
		Email   string `json:"email"`
	}

	resp, err := idp.Client.Get("https://www.googleapis.com/oauth2/v2/userinfo?alt=json&access_token=" + token.AccessToken)
	defer resp.Body.Close()
	contents, err := ioutil.ReadAll(resp.Body)
	var userResponse response
	err = json.Unmarshal(contents, &userResponse)
	if err != nil {
		return nil, err
	}
	if userResponse.Email == "" {
		return userInfo, errors.New("google email is empty")
	}

	userInfo.Username = userResponse.Email
	userInfo.Email = userResponse.Email
	userInfo.AvatarUrl = userResponse.Picture
	return userInfo, nil
}
