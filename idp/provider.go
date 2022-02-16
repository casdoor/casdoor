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
	Email       string
	AvatarUrl   string
}

type IdProvider interface {
	SetHttpClient(client *http.Client)
	GetToken(code string) (*oauth2.Token, error)
	GetUserInfo(token *oauth2.Token) (*UserInfo, error)
}

func GetIdProvider(typ string, subType string, clientId string, clientSecret string, appId string, redirectUrl string) IdProvider {
	if typ == "GitHub" {
		return NewGithubIdProvider(clientId, clientSecret, redirectUrl)
	} else if typ == "Google" {
		return NewGoogleIdProvider(clientId, clientSecret, redirectUrl)
	} else if typ == "QQ" {
		return NewQqIdProvider(clientId, clientSecret, redirectUrl)
	} else if typ == "WeChat" {
		return NewWeChatIdProvider(clientId, clientSecret, redirectUrl)
	} else if typ == "Facebook" {
		return NewFacebookIdProvider(clientId, clientSecret, redirectUrl)
	} else if typ == "DingTalk" {
		return NewDingTalkIdProvider(clientId, clientSecret, redirectUrl)
	} else if typ == "Weibo" {
		return NewWeiBoIdProvider(clientId, clientSecret, redirectUrl)
	} else if typ == "Gitee" {
		return NewGiteeIdProvider(clientId, clientSecret, redirectUrl)
	} else if typ == "LinkedIn" {
		return NewLinkedInIdProvider(clientId, clientSecret, redirectUrl)
	} else if typ == "WeCom" {
		if subType == "Internal" {
			return NewWeComInternalIdProvider(clientId, clientSecret, redirectUrl)
		} else if subType == "Third-party" {
			return NewWeComIdProvider(clientId, clientSecret, redirectUrl)
		} else {
			return nil
		}
	} else if typ == "Lark" {
		return NewLarkIdProvider(clientId, clientSecret, redirectUrl)
	} else if typ == "GitLab" {
		return NewGitlabIdProvider(clientId, clientSecret, redirectUrl)
	} else if typ == "Baidu" {
		return NewBaiduIdProvider(clientId, clientSecret, redirectUrl)
	} else if typ == "Infoflow" {
		if subType == "Internal" {
			return NewInfoflowInternalIdProvider(clientId, clientSecret, appId, redirectUrl)
		} else if subType == "Third-party" {
			return NewInfoflowIdProvider(clientId, clientSecret, appId, redirectUrl)
		} else {
			return nil
		}
	} else if isGothSupport(typ) {
		return NewGothIdProvider(typ, clientId, clientSecret, redirectUrl)
	}

	return nil
}

var gothList = []string{"Apple", "AzureAd", "Slack", "Steam"}

func isGothSupport(provider string) bool {
	for _, value := range gothList {
		if strings.EqualFold(value, provider) {
			return true
		}
	}
	return false
}
