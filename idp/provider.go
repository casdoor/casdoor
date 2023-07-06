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
	"net/http"
	"strings"

	"golang.org/x/oauth2"
)

type UserInfo struct {
	Id          string
	Username    string
	DisplayName string
	UnionId     string
	Email       string
	Phone       string
	CountryCode string
	AvatarUrl   string
}

type ProviderInfo struct {
	Type         string
	SubType      string
	ClientId     string
	ClientSecret string
	AppId        string
	HostUrl      string
	RedirectUrl  string

	TokenURL    string
	AuthURL     string
	UserInfoURL string
	UserMapping map[string]string
}

type IdProvider interface {
	SetHttpClient(client *http.Client)
	GetToken(code string) (*oauth2.Token, error)
	GetUserInfo(token *oauth2.Token) (*UserInfo, error)
}

func GetIdProvider(idpInfo *ProviderInfo, redirectUrl string) IdProvider {
	switch idpInfo.Type {
	case "GitHub":
		return NewGithubIdProvider(idpInfo.ClientId, idpInfo.ClientSecret, redirectUrl)
	case "Google":
		return NewGoogleIdProvider(idpInfo.ClientId, idpInfo.ClientSecret, redirectUrl)
	case "QQ":
		return NewQqIdProvider(idpInfo.ClientId, idpInfo.ClientSecret, redirectUrl)
	case "WeChat":
		return NewWeChatIdProvider(idpInfo.ClientId, idpInfo.ClientSecret, redirectUrl)
	case "Facebook":
		return NewFacebookIdProvider(idpInfo.ClientId, idpInfo.ClientSecret, redirectUrl)
	case "DingTalk":
		return NewDingTalkIdProvider(idpInfo.ClientId, idpInfo.ClientSecret, redirectUrl)
	case "Weibo":
		return NewWeiBoIdProvider(idpInfo.ClientId, idpInfo.ClientSecret, redirectUrl)
	case "Gitee":
		return NewGiteeIdProvider(idpInfo.ClientId, idpInfo.ClientSecret, redirectUrl)
	case "LinkedIn":
		return NewLinkedInIdProvider(idpInfo.ClientId, idpInfo.ClientSecret, redirectUrl)
	case "WeCom":
		if idpInfo.SubType == "Internal" {
			return NewWeComInternalIdProvider(idpInfo.ClientId, idpInfo.ClientSecret, redirectUrl)
		} else if idpInfo.SubType == "Third-party" {
			return NewWeComIdProvider(idpInfo.ClientId, idpInfo.ClientSecret, redirectUrl)
		} else {
			return nil
		}
	case "Lark":
		return NewLarkIdProvider(idpInfo.ClientId, idpInfo.ClientSecret, redirectUrl)
	case "GitLab":
		return NewGitlabIdProvider(idpInfo.ClientId, idpInfo.ClientSecret, redirectUrl)
	case "Adfs":
		return NewAdfsIdProvider(idpInfo.ClientId, idpInfo.ClientSecret, redirectUrl, idpInfo.HostUrl)
	case "Baidu":
		return NewBaiduIdProvider(idpInfo.ClientId, idpInfo.ClientSecret, redirectUrl)
	case "Alipay":
		return NewAlipayIdProvider(idpInfo.ClientId, idpInfo.ClientSecret, redirectUrl)
	case "Custom":
		return NewCustomIdProvider(idpInfo, redirectUrl)
	case "Infoflow":
		if idpInfo.SubType == "Internal" {
			return NewInfoflowInternalIdProvider(idpInfo.ClientId, idpInfo.ClientSecret, idpInfo.AppId, redirectUrl)
		} else if idpInfo.SubType == "Third-party" {
			return NewInfoflowIdProvider(idpInfo.ClientId, idpInfo.ClientSecret, idpInfo.AppId, redirectUrl)
		} else {
			return nil
		}
	case "Casdoor":
		return NewCasdoorIdProvider(idpInfo.ClientId, idpInfo.ClientSecret, redirectUrl, idpInfo.HostUrl)
	case "Okta":
		return NewOktaIdProvider(idpInfo.ClientId, idpInfo.ClientSecret, redirectUrl, idpInfo.HostUrl)
	case "Douyin":
		return NewDouyinIdProvider(idpInfo.ClientId, idpInfo.ClientSecret, redirectUrl)
	case "Bilibili":
		return NewBilibiliIdProvider(idpInfo.ClientId, idpInfo.ClientSecret, redirectUrl)
	default:
		if isGothSupport(idpInfo.Type) {
			return NewGothIdProvider(idpInfo.Type, idpInfo.ClientId, idpInfo.ClientSecret, redirectUrl, idpInfo.HostUrl)
		}
		return nil
	}
}

var gothList = []string{
	"Apple",
	"AzureAD",
	"Slack",
	"Steam",
	"Line",
	"Amazon",
	"Auth0",
	"BattleNet",
	"Bitbucket",
	"Box",
	"CloudFoundry",
	"Dailymotion",
	"Deezer",
	"DigitalOcean",
	"Discord",
	"Dropbox",
	"EveOnline",
	"Fitbit",
	"Gitea",
	"Heroku",
	"InfluxCloud",
	"Instagram",
	"Intercom",
	"Kakao",
	"Lastfm",
	"Mailru",
	"Meetup",
	"MicrosoftOnline",
	"Naver",
	"Nextcloud",
	"OneDrive",
	"Oura",
	"Patreon",
	"Paypal",
	"SalesForce",
	"Shopify",
	"Soundcloud",
	"Spotify",
	"Strava",
	"Stripe",
	"TikTok",
	"Tumblr",
	"Twitch",
	"Twitter",
	"Typetalk",
	"Uber",
	"VK",
	"Wepay",
	"Xero",
	"Yahoo",
	"Yammer",
	"Yandex",
	"Zoom",
}

func isGothSupport(provider string) bool {
	for _, value := range gothList {
		if strings.EqualFold(value, provider) {
			return true
		}
	}
	return false
}
