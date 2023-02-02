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
	"github.com/markbates/goth/providers/auth0"
	"github.com/markbates/goth/providers/azureadv2"
	"github.com/markbates/goth/providers/battlenet"
	"github.com/markbates/goth/providers/bitbucket"
	"github.com/markbates/goth/providers/box"
	"github.com/markbates/goth/providers/cloudfoundry"
	"github.com/markbates/goth/providers/dailymotion"
	"github.com/markbates/goth/providers/deezer"
	"github.com/markbates/goth/providers/digitalocean"
	"github.com/markbates/goth/providers/discord"
	"github.com/markbates/goth/providers/dropbox"
	"github.com/markbates/goth/providers/eveonline"
	"github.com/markbates/goth/providers/facebook"
	"github.com/markbates/goth/providers/fitbit"
	"github.com/markbates/goth/providers/gitea"
	"github.com/markbates/goth/providers/github"
	"github.com/markbates/goth/providers/gitlab"
	"github.com/markbates/goth/providers/google"
	"github.com/markbates/goth/providers/heroku"
	"github.com/markbates/goth/providers/influxcloud"
	"github.com/markbates/goth/providers/instagram"
	"github.com/markbates/goth/providers/intercom"
	"github.com/markbates/goth/providers/kakao"
	"github.com/markbates/goth/providers/lastfm"
	"github.com/markbates/goth/providers/line"
	"github.com/markbates/goth/providers/linkedin"
	"github.com/markbates/goth/providers/mailru"
	"github.com/markbates/goth/providers/meetup"
	"github.com/markbates/goth/providers/microsoftonline"
	"github.com/markbates/goth/providers/naver"
	"github.com/markbates/goth/providers/nextcloud"
	"github.com/markbates/goth/providers/onedrive"
	"github.com/markbates/goth/providers/oura"
	"github.com/markbates/goth/providers/patreon"
	"github.com/markbates/goth/providers/paypal"
	"github.com/markbates/goth/providers/salesforce"
	"github.com/markbates/goth/providers/shopify"
	"github.com/markbates/goth/providers/slack"
	"github.com/markbates/goth/providers/soundcloud"
	"github.com/markbates/goth/providers/spotify"
	"github.com/markbates/goth/providers/steam"
	"github.com/markbates/goth/providers/strava"
	"github.com/markbates/goth/providers/stripe"
	"github.com/markbates/goth/providers/tiktok"
	"github.com/markbates/goth/providers/tumblr"
	"github.com/markbates/goth/providers/twitch"
	"github.com/markbates/goth/providers/twitterv2"
	"github.com/markbates/goth/providers/typetalk"
	"github.com/markbates/goth/providers/uber"
	"github.com/markbates/goth/providers/wepay"
	"github.com/markbates/goth/providers/xero"
	"github.com/markbates/goth/providers/yahoo"
	"github.com/markbates/goth/providers/yammer"
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
			Provider: azureadv2.New(clientId, clientSecret, redirectUrl, azureadv2.ProviderOptions{Tenant: "common"}),
			Session:  &azureadv2.Session{},
		}
	case "Auth0":
		idp = GothIdProvider{
			Provider: auth0.New(clientId, clientSecret, redirectUrl, "casdoor.auth0.com"),
			Session:  &auth0.Session{},
		}
	case "BattleNet":
		idp = GothIdProvider{
			Provider: battlenet.New(clientId, clientSecret, redirectUrl),
			Session:  &battlenet.Session{},
		}
	case "Bitbucket":
		idp = GothIdProvider{
			Provider: bitbucket.New(clientId, clientSecret, redirectUrl),
			Session:  &bitbucket.Session{},
		}
	case "Box":
		idp = GothIdProvider{
			Provider: box.New(clientId, clientSecret, redirectUrl),
			Session:  &box.Session{},
		}
	case "CloudFoundry":
		idp = GothIdProvider{
			Provider: cloudfoundry.New("", clientId, clientSecret, redirectUrl),
			Session:  &cloudfoundry.Session{},
		}
	case "Dailymotion":
		idp = GothIdProvider{
			Provider: dailymotion.New(clientId, clientSecret, redirectUrl),
			Session:  &dailymotion.Session{},
		}
	case "Deezer":
		idp = GothIdProvider{
			Provider: deezer.New(clientId, clientSecret, redirectUrl),
			Session:  &deezer.Session{},
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
	case "EveOnline":
		idp = GothIdProvider{
			Provider: eveonline.New(clientId, clientSecret, redirectUrl),
			Session:  &eveonline.Session{},
		}
	case "Fitbit":
		idp = GothIdProvider{
			Provider: fitbit.New(clientId, clientSecret, redirectUrl),
			Session:  &fitbit.Session{},
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
	case "InfluxCloud":
		idp = GothIdProvider{
			Provider: influxcloud.New(clientId, clientSecret, redirectUrl),
			Session:  &influxcloud.Session{},
		}
	case "Instagram":
		idp = GothIdProvider{
			Provider: instagram.New(clientId, clientSecret, redirectUrl),
			Session:  &instagram.Session{},
		}
	case "Intercom":
		idp = GothIdProvider{
			Provider: intercom.New(clientId, clientSecret, redirectUrl),
			Session:  &intercom.Session{},
		}
	case "Kakao":
		idp = GothIdProvider{
			Provider: kakao.New(clientId, clientSecret, redirectUrl),
			Session:  &kakao.Session{},
		}
	case "Lastfm":
		idp = GothIdProvider{
			Provider: lastfm.New(clientId, clientSecret, redirectUrl),
			Session:  &lastfm.Session{},
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
	case "Mailru":
		idp = GothIdProvider{
			Provider: mailru.New(clientId, clientSecret, redirectUrl),
			Session:  &mailru.Session{},
		}
	case "Meetup":
		idp = GothIdProvider{
			Provider: meetup.New(clientId, clientSecret, redirectUrl),
			Session:  &meetup.Session{},
		}
	case "MicrosoftOnline":
		idp = GothIdProvider{
			Provider: microsoftonline.New(clientId, clientSecret, redirectUrl),
			Session:  &microsoftonline.Session{},
		}
	case "Naver":
		idp = GothIdProvider{
			Provider: naver.New(clientId, clientSecret, redirectUrl),
			Session:  &naver.Session{},
		}
	case "Nextcloud":
		idp = GothIdProvider{
			Provider: nextcloud.New(clientId, clientSecret, redirectUrl),
			Session:  &nextcloud.Session{},
		}
	case "OneDrive":
		idp = GothIdProvider{
			Provider: onedrive.New(clientId, clientSecret, redirectUrl),
			Session:  &onedrive.Session{},
		}
	case "Oura":
		idp = GothIdProvider{
			Provider: oura.New(clientId, clientSecret, redirectUrl),
			Session:  &oura.Session{},
		}
	case "Patreon":
		idp = GothIdProvider{
			Provider: patreon.New(clientId, clientSecret, redirectUrl),
			Session:  &patreon.Session{},
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
	case "Soundcloud":
		idp = GothIdProvider{
			Provider: soundcloud.New(clientId, clientSecret, redirectUrl),
			Session:  &soundcloud.Session{},
		}
	case "Spotify":
		idp = GothIdProvider{
			Provider: spotify.New(clientId, clientSecret, redirectUrl),
			Session:  &spotify.Session{},
		}
	case "Steam":
		idp = GothIdProvider{
			Provider: steam.New(clientSecret, redirectUrl),
			Session:  &steam.Session{},
		}
	case "Strava":
		idp = GothIdProvider{
			Provider: strava.New(clientId, clientSecret, redirectUrl),
			Session:  &strava.Session{},
		}
	case "Stripe":
		idp = GothIdProvider{
			Provider: stripe.New(clientId, clientSecret, redirectUrl),
			Session:  &stripe.Session{},
		}
	case "TikTok":
		idp = GothIdProvider{
			Provider: tiktok.New(clientId, clientSecret, redirectUrl),
			Session:  &tiktok.Session{},
		}
	case "Tumblr":
		idp = GothIdProvider{
			Provider: tumblr.New(clientId, clientSecret, redirectUrl),
			Session:  &tumblr.Session{},
		}
	case "Twitch":
		idp = GothIdProvider{
			Provider: twitch.New(clientId, clientSecret, redirectUrl),
			Session:  &twitch.Session{},
		}
	case "Twitter":
		idp = GothIdProvider{
			Provider: twitterv2.New(clientId, clientSecret, redirectUrl),
			Session:  &twitterv2.Session{},
		}
	case "Typetalk":
		idp = GothIdProvider{
			Provider: typetalk.New(clientId, clientSecret, redirectUrl),
			Session:  &typetalk.Session{},
		}
	case "Uber":
		idp = GothIdProvider{
			Provider: uber.New(clientId, clientSecret, redirectUrl),
			Session:  &uber.Session{},
		}
	case "Wepay":
		idp = GothIdProvider{
			Provider: wepay.New(clientId, clientSecret, redirectUrl),
			Session:  &wepay.Session{},
		}
	case "Xero":
		idp = GothIdProvider{
			Provider: xero.New(clientId, clientSecret, redirectUrl),
			Session:  &xero.Session{},
		}
	case "Yahoo":
		idp = GothIdProvider{
			Provider: yahoo.New(clientId, clientSecret, redirectUrl),
			Session:  &yahoo.Session{},
		}
	case "Yammer":
		idp = GothIdProvider{
			Provider: yammer.New(clientId, clientSecret, redirectUrl),
			Session:  &yammer.Session{},
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
	default:
		return nil
	}

	return &idp
}

// SetHttpClient
// Goth's idp all implement the Client method, but since the goth.Provider interface does not provide to modify idp's client method, reflection is required
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
		// Need to construct variables supported by goth
		// to call the function to obtain accessToken
		value = url.Values{}
		value.Add("code", code)
		if idp.Provider.Name() == "twitterv2" || idp.Provider.Name() == "fitbit" {
			value.Add("oauth_verifier", "casdoor-verifier")
		}
	}
	accessToken, err := idp.Session.Authorize(idp.Provider, value)
	if err != nil {
		return nil, err
	}

	// Get ExpiresAt's value
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
	// Some idp return an empty Name
	// so construct the Name with firstname and lastname or nickname
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
		user.Username = user.Id
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
