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
	"io/ioutil"
	"net/http"
	"sync"

	"golang.org/x/oauth2"
)

type GithubIdProvider struct{}

func (idp *GithubIdProvider) GetConfig() *oauth2.Config {
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

func (idp *GithubIdProvider) getEmail(httpClient *http.Client, token *oauth2.Token) string {
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
	response, err := httpClient.Do(req)
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

func (idp *GithubIdProvider) getLoginAndAvatar(httpClient *http.Client, token *oauth2.Token) (string, string) {
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
	resp, err := httpClient.Do(req)
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

func (idp *GithubIdProvider) GetUserInfo(httpClient *http.Client, token *oauth2.Token) (string, string, string, error) {
	var email, username, avatarUrl string

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		email = idp.getEmail(httpClient, token)
		wg.Done()
	}()
	go func() {
		username, avatarUrl = idp.getLoginAndAvatar(httpClient, token)
		wg.Done()
	}()
	wg.Wait()

	return email, username, avatarUrl, nil
}
