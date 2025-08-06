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
	Extra       map[string]string
}

type ProviderInfo struct {
	Type          string
	SubType       string
	ClientId      string
	ClientSecret  string
	ClientId2     string
	ClientSecret2 string
	AppId         string
	HostUrl       string
	RedirectUrl   string
	DisableSsl    bool

	TokenURL    string
	AuthURL     string
	UserInfoURL string
	UserMapping map[string]string

	UseIdToFormEmail bool
	EmailSuffix      string
}

type IdProvider interface {
	SetHttpClient(client *http.Client)
	GetToken(code string) (*oauth2.Token, error)
	GetUserInfo(token *oauth2.Token) (*UserInfo, error)
}

func GetIdProvider(idpInfo *ProviderInfo, redirectUrl string) (IdProvider, error) {
	switch idpInfo.Type {
	case "GitHub":
		return NewGithubIdProvider(idpInfo.ClientId, idpInfo.ClientSecret, redirectUrl), nil
	case "Google":
		return NewGoogleIdProvider(idpInfo.ClientId, idpInfo.ClientSecret, redirectUrl), nil
	case "QQ":
		return NewQqIdProvider(idpInfo.ClientId, idpInfo.ClientSecret, redirectUrl), nil
	case "WeChat":
		return NewWeChatIdProvider(idpInfo.ClientId, idpInfo.ClientSecret, redirectUrl), nil
	case "Facebook":
		return NewFacebookIdProvider(idpInfo.ClientId, idpInfo.ClientSecret, redirectUrl), nil
	case "DingTalk":
		return NewDingTalkIdProvider(idpInfo.ClientId, idpInfo.ClientSecret, redirectUrl), nil
	case "Weibo":
		return NewWeiBoIdProvider(idpInfo.ClientId, idpInfo.ClientSecret, redirectUrl), nil
	case "Gitee":
		return NewGiteeIdProvider(idpInfo.ClientId, idpInfo.ClientSecret, redirectUrl), nil
	case "LinkedIn":
		return NewLinkedInIdProvider(idpInfo.ClientId, idpInfo.ClientSecret, redirectUrl), nil
	case "WeCom":
		if idpInfo.SubType == "Internal" {
			return NewWeComInternalIdProvider(idpInfo.ClientId, idpInfo.ClientSecret, redirectUrl, idpInfo.DisableSsl, idpInfo.UseIdToFormEmail, idpInfo.EmailSuffix), nil
		} else if idpInfo.SubType == "Third-party" {
			return NewWeComIdProvider(idpInfo.ClientId, idpInfo.ClientSecret, redirectUrl, idpInfo.DisableSsl, idpInfo.UseIdToFormEmail, idpInfo.EmailSuffix), nil
		} else {
			return nil, fmt.Errorf("WeCom provider subType: %s is not supported", idpInfo.SubType)
		}
	case "Lark":
		return NewLarkIdProvider(idpInfo.ClientId, idpInfo.ClientSecret, redirectUrl, idpInfo.DisableSsl), nil
	case "GitLab":
		return NewGitlabIdProvider(idpInfo.ClientId, idpInfo.ClientSecret, redirectUrl), nil
	case "ADFS":
		return NewAdfsIdProvider(idpInfo.ClientId, idpInfo.ClientSecret, redirectUrl, idpInfo.HostUrl), nil
	case "AzureADB2C":
		return NewAzureAdB2cProvider(idpInfo.ClientId, idpInfo.ClientSecret, redirectUrl, idpInfo.HostUrl, idpInfo.AppId), nil
	case "Baidu":
		return NewBaiduIdProvider(idpInfo.ClientId, idpInfo.ClientSecret, redirectUrl), nil
	case "Alipay":
		return NewAlipayIdProvider(idpInfo.ClientId, idpInfo.ClientSecret, redirectUrl), nil
	case "Custom":
		return NewCustomIdProvider(idpInfo, redirectUrl), nil
	case "Infoflow":
		if idpInfo.SubType == "Internal" {
			return NewInfoflowInternalIdProvider(idpInfo.ClientId, idpInfo.ClientSecret, idpInfo.AppId, redirectUrl), nil
		} else if idpInfo.SubType == "Third-party" {
			return NewInfoflowIdProvider(idpInfo.ClientId, idpInfo.ClientSecret, idpInfo.AppId, redirectUrl), nil
		} else {
			return nil, fmt.Errorf("Infoflow provider subType: %s is not supported", idpInfo.SubType)
		}
	case "Casdoor":
		return NewCasdoorIdProvider(idpInfo.ClientId, idpInfo.ClientSecret, redirectUrl, idpInfo.HostUrl), nil
	case "Okta":
		return NewOktaIdProvider(idpInfo.ClientId, idpInfo.ClientSecret, redirectUrl, idpInfo.HostUrl), nil
	case "Douyin":
		return NewDouyinIdProvider(idpInfo.ClientId, idpInfo.ClientSecret, redirectUrl), nil
	case "Kwai":
		return NewKwaiIdProvider(idpInfo.ClientId, idpInfo.ClientSecret, redirectUrl), nil
	case "Bilibili":
		return NewBilibiliIdProvider(idpInfo.ClientId, idpInfo.ClientSecret, redirectUrl), nil
	case "MetaMask":
		return NewMetaMaskIdProvider(), nil
	case "Web3Onboard":
		return NewWeb3OnboardIdProvider(), nil
	case "Twitter":
		return NewTwitterIdProvider(idpInfo.ClientId, idpInfo.ClientSecret, redirectUrl), nil
	default:
		if isGothSupport(idpInfo.Type) {
			return NewGothIdProvider(idpInfo.Type, idpInfo.ClientId, idpInfo.ClientSecret, idpInfo.ClientId2, idpInfo.ClientSecret2, redirectUrl, idpInfo.HostUrl)
		}
		return nil, fmt.Errorf("OAuth provider type: %s is not supported", idpInfo.Type)
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
