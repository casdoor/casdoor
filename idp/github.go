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
	"strconv"
	"time"

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
	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, idp.Client)
	return idp.Config.Exchange(ctx, code)
}

//{
//	"login": "jimgreen",
//	"id": 3781234,
//	"node_id": "MDQ6VXNlcjM3O123456=",
//	"avatar_url": "https://avatars.githubusercontent.com/u/3781234?v=4",
//	"gravatar_id": "",
//	"url": "https://api.github.com/users/jimgreen",
//	"html_url": "https://github.com/jimgreen",
//	"followers_url": "https://api.github.com/users/jimgreen/followers",
//	"following_url": "https://api.github.com/users/jimgreen/following{/other_user}",
//	"gists_url": "https://api.github.com/users/jimgreen/gists{/gist_id}",
//	"starred_url": "https://api.github.com/users/jimgreen/starred{/owner}{/repo}",
//	"subscriptions_url": "https://api.github.com/users/jimgreen/subscriptions",
//	"organizations_url": "https://api.github.com/users/jimgreen/orgs",
//	"repos_url": "https://api.github.com/users/jimgreen/repos",
//	"events_url": "https://api.github.com/users/jimgreen/events{/privacy}",
//	"received_events_url": "https://api.github.com/users/jimgreen/received_events",
//	"type": "User",
//	"site_admin": false,
//	"name": "Jim Green",
//	"company": "Casbin",
//	"blog": "https://casbin.org",
//	"location": "Bay Area",
//	"email": "jimgreen@gmail.com",
//	"hireable": true,
//	"bio": "My bio",
//	"twitter_username": null,
//	"public_repos": 45,
//	"public_gists": 3,
//	"followers": 123,
//	"following": 31,
//	"created_at": "2016-03-06T13:16:13Z",
//	"updated_at": "2020-05-30T12:15:29Z",
//	"private_gists": 0,
//	"total_private_repos": 12,
//	"owned_private_repos": 12,
//	"disk_usage": 46331,
//	"collaborators": 5,
//	"two_factor_authentication": true,
//	"plan": {
//		"name": "free",
//		"space": 976562499,
//		"collaborators": 0,
//		"private_repos": 10000
//	}
//}

type GitHubUserInfo struct {
	Login                   string      `json:"login"`
	Id                      int         `json:"id"`
	NodeId                  string      `json:"node_id"`
	AvatarUrl               string      `json:"avatar_url"`
	GravatarId              string      `json:"gravatar_id"`
	Url                     string      `json:"url"`
	HtmlUrl                 string      `json:"html_url"`
	FollowersUrl            string      `json:"followers_url"`
	FollowingUrl            string      `json:"following_url"`
	GistsUrl                string      `json:"gists_url"`
	StarredUrl              string      `json:"starred_url"`
	SubscriptionsUrl        string      `json:"subscriptions_url"`
	OrganizationsUrl        string      `json:"organizations_url"`
	ReposUrl                string      `json:"repos_url"`
	EventsUrl               string      `json:"events_url"`
	ReceivedEventsUrl       string      `json:"received_events_url"`
	Type                    string      `json:"type"`
	SiteAdmin               bool        `json:"site_admin"`
	Name                    string      `json:"name"`
	Company                 string      `json:"company"`
	Blog                    string      `json:"blog"`
	Location                string      `json:"location"`
	Email                   string      `json:"email"`
	Hireable                bool        `json:"hireable"`
	Bio                     string      `json:"bio"`
	TwitterUsername         interface{} `json:"twitter_username"`
	PublicRepos             int         `json:"public_repos"`
	PublicGists             int         `json:"public_gists"`
	Followers               int         `json:"followers"`
	Following               int         `json:"following"`
	CreatedAt               time.Time   `json:"created_at"`
	UpdatedAt               time.Time   `json:"updated_at"`
	PrivateGists            int         `json:"private_gists"`
	TotalPrivateRepos       int         `json:"total_private_repos"`
	OwnedPrivateRepos       int         `json:"owned_private_repos"`
	DiskUsage               int         `json:"disk_usage"`
	Collaborators           int         `json:"collaborators"`
	TwoFactorAuthentication bool        `json:"two_factor_authentication"`
	Plan                    struct {
		Name          string `json:"name"`
		Space         int    `json:"space"`
		Collaborators int    `json:"collaborators"`
		PrivateRepos  int    `json:"private_repos"`
	} `json:"plan"`
}

func (idp *GithubIdProvider) GetUserInfo(token *oauth2.Token) (*UserInfo, error) {
	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		panic(err)
	}
	req.Header.Add("Authorization", "token "+token.AccessToken)
	resp, err := idp.Client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var githubUserInfo GitHubUserInfo
	err = json.Unmarshal(body, &githubUserInfo)
	if err != nil {
		return nil, err
	}

	userInfo := UserInfo{
		Id:          strconv.Itoa(githubUserInfo.Id),
		Username:    githubUserInfo.Login,
		DisplayName: githubUserInfo.Name,
		Email:       githubUserInfo.Email,
		AvatarUrl:   githubUserInfo.AvatarUrl,
	}
	return &userInfo, nil
}
