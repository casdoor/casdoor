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

	"golang.org/x/oauth2"
)

type FacebookIdProvider struct{}

func (idp *FacebookIdProvider) GetConfig() *oauth2.Config {
	var endpoint = oauth2.Endpoint{
		AuthURL:  "https://gitlab.com/oauth/authorize",
		TokenURL: "https://gitlab.com/oauth/token",
	}

	var config = &oauth2.Config{
		Scopes:   []string{"public_profile", "email"},
		Endpoint: endpoint,
	}

	return config
}

func (idp *FacebookIdProvider) GetUserInfo(httpClient *http.Client, token *oauth2.Token) (string, string, string, error) {

	type Url struct {
		Url string `json:"url"`
	}

	type Data struct {
		Data Url `json:"data"`
	}

	type userInfoFromFacebook struct {
		Email         string `json:"email"`
		Name          string `json:"name"`
		Is_guest_user string `json:"is_guest_user"`
		Picture       Data   `json:"picture"`
	}
	var fbUser userInfoFromFacebook
	req, err := http.NewRequest("GET", "https://facebook.com/v10.0.0/me?fields=name,email,is_guest_user,picture", nil)
	if err != nil {
		panic(err)
	}
	req.Header.Add("Authorization", "Bearer "+token.AccessToken)
	response, err := httpClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer response.Body.Close()
	contents, _ := ioutil.ReadAll(response.Body)

	err = json.Unmarshal(contents, &fbUser)
	if err != nil {
		panic(err)
	}

	return fbUser.Email, fbUser.Name, fbUser.Picture.Data.Url, nil
}