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
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"time"

	"github.com/casdoor/casdoor/util"
	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/amazon"
	"github.com/markbates/goth/providers/apple"
	"github.com/markbates/goth/providers/azuread"
	"github.com/markbates/goth/providers/bitbucket"
	"github.com/markbates/goth/providers/digitalocean"
	"github.com/markbates/goth/providers/discord"
	"github.com/markbates/goth/providers/dropbox"
	"github.com/markbates/goth/providers/facebook"
	"github.com/markbates/goth/providers/gitea"
	"github.com/markbates/goth/providers/github"
	"github.com/markbates/goth/providers/gitlab"
	"github.com/markbates/goth/providers/google"
	"github.com/markbates/goth/providers/heroku"
	"github.com/markbates/goth/providers/instagram"
	"github.com/markbates/goth/providers/kakao"
	"github.com/markbates/goth/providers/line"
	"github.com/markbates/goth/providers/linkedin"
	"github.com/markbates/goth/providers/microsoftonline"
	"github.com/markbates/goth/providers/paypal"
	"github.com/markbates/goth/providers/salesforce"
	"github.com/markbates/goth/providers/shopify"
	"github.com/markbates/goth/providers/slack"
	"github.com/markbates/goth/providers/steam"
	"github.com/markbates/goth/providers/tumblr"
	"github.com/markbates/goth/providers/twitter"
	"github.com/markbates/goth/providers/yahoo"
	"github.com/markbates/goth/providers/yandex"
	"github.com/markbates/goth/providers/zoom"
	"golang.org/x/oauth2"
)

type GothIdProvider struct {
	Provider goth.Provider
	Session  goth.Session
}

func NewGothIdProvider(providerType string, clientId string, clientSecret string, redirectUrl string) *GothIdProvider {
	var idp GothIdProvider
	switch providerType {
	case "Amazon":
		idp = GothIdProvider{
			Provider: amazon.New(clientId, clientSecret, redirectUrl),
			Session:  &amazon.Session{},
		}
	case "Apple":
		idp = GothIdProvider{
			Provider: apple.New(clientId, clientSecret, redirectUrl, nil),
			Session:  &apple.Session{},
		}
	case "AzureAD":
		idp = GothIdProvider{
			Provider: azuread.New(clientId, clientSecret, redirectUrl, nil),
			Session:  &azuread.Session{},
		}
	case "Bitbucket":
		idp = GothIdProvider{
			Provider: bitbucket.New(clientId, clientSecret, redirectUrl),
			Session:  &bitbucket.Session{},
		}
	case "DigitalOcean":
		idp = GothIdProvider{
			Provider: digitalocean.New(clientId, clientSecret, redirectUrl),
			Session:  &digitalocean.Session{},
		}
	case "Discord":
		idp = GothIdProvider{
			Provider: discord.New(clientId, clientSecret, redirectUrl),
			Session:  &discord.Session{},
		}
	case "Dropbox":
		idp = GothIdProvider{
			Provider: dropbox.New(clientId, clientSecret, redirectUrl),
			Session:  &dropbox.Session{},
		}
	case "Facebook":
		idp = GothIdProvider{
			Provider: facebook.New(clientId, clientSecret, redirectUrl),
			Session:  &facebook.Session{},
		}
	case "Gitea":
		idp = GothIdProvider{
			Provider: gitea.New(clientId, clientSecret, redirectUrl),
			Session:  &gitea.Session{},
		}
	case "GitHub":
		idp = GothIdProvider{
			Provider: github.New(clientId, clientSecret, redirectUrl),
			Session:  &github.Session{},
		}
	case "GitLab":
		idp = GothIdProvider{
			Provider: gitlab.New(clientId, clientSecret, redirectUrl),
			Session:  &gitlab.Session{},
		}
	case "Google":
		idp = GothIdProvider{
			Provider: google.New(clientId, clientSecret, redirectUrl),
			Session:  &google.Session{},
		}
	case "Heroku":
		idp = GothIdProvider{
			Provider: heroku.New(clientId, clientSecret, redirectUrl),
			Session:  &heroku.Session{},
		}
	case "Instagram":
		idp = GothIdProvider{
			Provider: instagram.New(clientId, clientSecret, redirectUrl),
			Session:  &instagram.Session{},
		}
	case "Kakao":
		idp = GothIdProvider{
			Provider: kakao.New(clientId, clientSecret, redirectUrl),
			Session:  &kakao.Session{},
		}
	case "Linkedin":
		idp = GothIdProvider{
			Provider: linkedin.New(clientId, clientSecret, redirectUrl),
			Session:  &linkedin.Session{},
		}
	case "Line":
		idp = GothIdProvider{
			Provider: line.New(clientId, clientSecret, redirectUrl),
			Session:  &line.Session{},
		}
	case "MicrosoftOnline":
		idp = GothIdProvider{
			Provider: microsoftonline.New(clientId, clientSecret, redirectUrl),
			Session:  &microsoftonline.Session{},
		}
	case "Paypal":
		idp = GothIdProvider{
			Provider: paypal.New(clientId, clientSecret, redirectUrl),
			Session:  &paypal.Session{},
		}
	case "SalesForce":
		idp = GothIdProvider{
			Provider: salesforce.New(clientId, clientSecret, redirectUrl),
			Session:  &salesforce.Session{},
		}
	case "Shopify":
		idp = GothIdProvider{
			Provider: shopify.New(clientId, clientSecret, redirectUrl),
			Session:  &shopify.Session{},
		}
	case "Slack":
		idp = GothIdProvider{
			Provider: slack.New(clientId, clientSecret, redirectUrl),
			Session:  &slack.Session{},
		}
	case "Steam":
		idp = GothIdProvider{
			Provider: steam.New(clientSecret, redirectUrl),
			Session:  &steam.Session{},
		}
	case "Tumblr":
		idp = GothIdProvider{
			Provider: tumblr.New(clientId, clientSecret, redirectUrl),
			Session:  &tumblr.Session{},
		}
	case "Twitter":
		idp = GothIdProvider{
			Provider: twitter.New(clientId, clientSecret, redirectUrl),
			Session:  &twitter.Session{},
		}
	case "Yahoo":
		idp = GothIdProvider{
			Provider: yahoo.New(clientId, clientSecret, redirectUrl),
			Session:  &yahoo.Session{},
		}
	case "Yandex":
		idp = GothIdProvider{
			Provider: yandex.New(clientId, clientSecret, redirectUrl),
			Session:  &yandex.Session{},
		}
	case "Zoom":
		idp = GothIdProvider{
			Provider: zoom.New(clientId, clientSecret, redirectUrl),
			Session:  &zoom.Session{},
		}
	}

	return &idp
}

//Goth's idp all implement the Client method, but since the goth.Provider interface does not provide to modify idp's client method, reflection is required
func (idp *GothIdProvider) SetHttpClient(client *http.Client) {
	idpClient := reflect.ValueOf(idp.Provider).Elem().FieldByName("HTTPClient")
	idpClient.Set(reflect.ValueOf(client))
}

func (idp *GothIdProvider) GetToken(code string) (*oauth2.Token, error) {
	var expireAt time.Time
	var value url.Values
	var err error
	if idp.Provider.Name() == "steam" {
		value, err = url.ParseQuery(code)
		returnUrl := reflect.ValueOf(idp.Session).Elem().FieldByName("CallbackURL")
		returnUrl.Set(reflect.ValueOf(value.Get("openid.return_to")))
		if err != nil {
			return nil, err
		}
	} else {
		//Need to construct variables supported by goth
		//to call the function to obtain accessToken
		value = url.Values{}
		value.Add("code", code)
	}
	accessToken, err := idp.Session.Authorize(idp.Provider, value)
	if err != nil {
		return nil, err
	}

	//Get ExpiresAt's value
	valueOfExpire := reflect.ValueOf(idp.Session).Elem().FieldByName("ExpiresAt")
	if valueOfExpire.IsValid() {
		expireAt = valueOfExpire.Interface().(time.Time)
	}
	token := oauth2.Token{
		AccessToken: accessToken,
		Expiry:      expireAt,
	}

	return &token, nil
}

func (idp *GothIdProvider) GetUserInfo(token *oauth2.Token) (*UserInfo, error) {
	gothUser, err := idp.Provider.FetchUser(idp.Session)
	if err != nil {
		return nil, err
	}
	return getUser(gothUser, idp.Provider.Name()), nil
}

func getUser(gothUser goth.User, provider string) *UserInfo {
	user := UserInfo{
		Id:          gothUser.UserID,
		Username:    gothUser.Name,
		DisplayName: gothUser.NickName,
		Email:       gothUser.Email,
		AvatarUrl:   gothUser.AvatarURL,
	}
	//Some idp return an empty Name
	//so construct the Name with firstname and lastname or nickname
	if user.Username == "" {
		if gothUser.FirstName != "" && gothUser.LastName != "" {
			user.Username = getName(gothUser.FirstName, gothUser.LastName)
		} else {
			user.Username = gothUser.NickName
		}
	}
	if user.DisplayName == "" {
		if gothUser.FirstName != "" && gothUser.LastName != "" {
			user.DisplayName = getName(gothUser.FirstName, gothUser.LastName)
		} else {
			user.DisplayName = user.Username
		}
	}
	if provider == "steam" {
		user.Username = user.DisplayName
		user.Email = ""
	}
	return &user
}

func getName(firstName, lastName string) string {
	if util.IsChinese(firstName) || util.IsChinese(lastName) {
		return fmt.Sprintf("%s%s", lastName, firstName)
	} else {
		return fmt.Sprintf("%s %s", firstName, lastName)
	}
}
