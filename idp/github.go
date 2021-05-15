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
	"io/ioutil"
	"net/http"
	"sync"

	"golang.org/x/oauth2"
)

type GithubIdProvider struct {
	Client *http.Client
	Config *oauth2.Config
}

func NewGithubIdProvider(clientId string, clientSecret string, redirectUrl string) *GithubIdProvider {
	idp := &GithubIdProvider{}

	config := idp.getConfig()
	config.ClientID = clientId
	config.ClientSecret = clientSecret
	config.RedirectURL = redirectUrl
	idp.Config = config

	return idp
}

func (idp *GithubIdProvider) SetHttpClient(client *http.Client) {
	idp.Client = client
}

func (idp *GithubIdProvider) getConfig() *oauth2.Config {
	var endpoint = oauth2.Endpoint{
		AuthURL:  "https://github.com/login/oauth/authorize",
		TokenURL: "https://github.com/login/oauth/access_token",
	}

	var config = &oauth2.Config{
		Scopes:   []string{"user:email", "read:user"},
		Endpoint: endpoint,
	}

	return config
}

func (idp *GithubIdProvider) GetToken(code string) (*oauth2.Token, error) {
	ctx := context.WithValue(oauth2.NoContext, oauth2.HTTPClient, idp.Client)
	return idp.Config.Exchange(ctx, code)
}

func (idp *GithubIdProvider) getEmail(token *oauth2.Token) string {
	res := ""

	type GithubEmail struct {
		Email      string `json:"email"`
		Primary    bool   `json:"primary"`
		Verified   bool   `json:"verified"`
		Visibility string `json:"visibility"`
	}
	var githubEmails []GithubEmail

	req, err := http.NewRequest("GET", "https://api.github.com/user/emails", nil)
	if err != nil {
		panic(err)
	}
	req.Header.Add("Authorization", "token "+token.AccessToken)
	response, err := idp.Client.Do(req)
	if err != nil {
		panic(err)
	}
	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)

	err = json.Unmarshal(contents, &githubEmails)
	if err != nil {
		panic(err)
	}
	for _, v := range githubEmails {
		if v.Primary == true {
			res = v.Email
			break
		}
	}
	return res
}

func (idp *GithubIdProvider) getLoginAndAvatar(token *oauth2.Token) (string, string) {
	type GithubUser struct {
		Login     string `json:"login"`
		AvatarUrl string `json:"avatar_url"`
	}
	var githubUser GithubUser

	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		panic(err)
	}
	req.Header.Add("Authorization", "token "+token.AccessToken)
	resp, err := idp.Client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	contents2, err := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(contents2, &githubUser)
	if err != nil {
		panic(err)
	}

	return githubUser.Login, githubUser.AvatarUrl
}

func (idp *GithubIdProvider) GetUserInfo(token *oauth2.Token) (*UserInfo, error) {
	userInfo := &UserInfo{}

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		userInfo.Email = idp.getEmail(token)
		wg.Done()
	}()
	go func() {
		userInfo.Username, userInfo.AvatarUrl = idp.getLoginAndAvatar(token)
		wg.Done()
	}()
	wg.Wait()

	return userInfo, nil
}
