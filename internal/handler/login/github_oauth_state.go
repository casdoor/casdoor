// Copyright 2020 The casbin Authors. All Rights Reserved.
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

package login

import (
	"crypto/rand"
	"encoding/base64"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"net/http"
	"os"
	"time"
)

const stateName = "github_oauth_state"

var githubOauthConfig = &oauth2.Config{
	RedirectURL:  "http://localhost:8000/auth/github/callback",
	ClientID:     os.Getenv("GITHUB_OAUTH_CLIENT_ID"),
	ClientSecret: os.Getenv("GITHUB_OAUTH_CLIENT_SECRET"),
	Scopes:       []string{},
	Endpoint:     github.Endpoint,
}

func getGithubOauthStateFromCookie(r *http.Request) (string, error) {
	c, err := r.Cookie(stateName)
	if err != nil {
		return "", err
	}
	return c.Value, nil
}

func generateGithubOauthStateToCookie(w http.ResponseWriter) (string, error) {
	var expiration = time.Now().Add(20 * time.Minute)

	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	state := base64.URLEncoding.EncodeToString(b)
	cookie := http.Cookie{Name: stateName, Value: state, Expires: expiration}
	http.SetCookie(w, &cookie)

	return state, nil
}
