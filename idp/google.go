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
	"io/ioutil"
	"net/http"

	"golang.org/x/oauth2"
)

type GoogleIdProvider struct{}

func (idp *GoogleIdProvider) GetConfig() *oauth2.Config {
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

func (idp *GoogleIdProvider) GetUserInfo(httpClient *http.Client, token *oauth2.Token) (string, string, string, error) {
	var email, username, avatarUrl string

	type userInfoFromGoogle struct {
		Picture string `json:"picture"`
		Email   string `json:"email"`
	}

	resp, err := httpClient.Get("https://www.googleapis.com/oauth2/v2/userinfo?alt=json&access_token=" + token.AccessToken)
	defer resp.Body.Close()
	contents, err := ioutil.ReadAll(resp.Body)
	var tempUser userInfoFromGoogle
	err = json.Unmarshal(contents, &tempUser)
	if err != nil {
		panic(err)
	}
	email = tempUser.Email
	avatarUrl = tempUser.Picture

	if email == "" {
		return email, username, avatarUrl, errors.New("google email is empty, please try again")
	}

	return email, username, avatarUrl, nil
}
