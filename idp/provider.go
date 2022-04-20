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
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"golang.org/x/oauth2"
)

type UserInfoGetter interface {
	GetID() string
	GetUsername() string
	GetDisplayName() string
	GetEmail() string
	GetAvatarURL() string
	GetAllProperties() map[string]string
}

type UserInfo struct {
	Id          string `json:"id"`
	Username    string `json:"username"`
	DisplayName string `json:"displayName"`
	Email       string `json:"email"`
	AvatarUrl   string `json:"avatarUrl"`
}

func (u *UserInfo) GetID() string {
	return u.Id
}

func (u *UserInfo) GetUsername() string {
	return u.Username
}

func (u *UserInfo) GetDisplayName() string {
	return u.DisplayName
}

func (u *UserInfo) GetEmail() string {
	return u.Email
}

func (u *UserInfo) GetAvatarURL() string {
	return u.AvatarUrl
}

func (u *UserInfo) getAllProperties(cur interface{}) map[string]string {
	properties := make(map[string]interface{})
	result := make(map[string]string)
	
	data, _ := json.Marshal(cur)
	_ = json.Unmarshal(data, &properties)
	for k, v := range properties {
		vv, ok := v.(string)
		if !ok {
			vv = fmt.Sprintf("%v", v)
		}
		result[k] = vv
	}
	return result
}

func (u *UserInfo) GetAllProperties() map[string]string {
	return u.getAllProperties(u)
}


type IdProvider interface {
	SetHttpClient(client *http.Client)
	GetToken(code string) (*oauth2.Token, error)
	GetUserInfo(token *oauth2.Token) (UserInfoGetter, error)
}

func GetIdProvider(typ string, subType string, clientId string, clientSecret string, appId string, redirectUrl string, hostUrl string, authUrl string, tokenUrl string, userInfoUrl string) IdProvider {
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
	} else if typ == "Adfs" {
		return NewAdfsIdProvider(clientId, clientSecret, redirectUrl, hostUrl)
	} else if typ == "Baidu" {
		return NewBaiduIdProvider(clientId, clientSecret, redirectUrl)
	} else if typ == "Alipay" {
		return NewAlipayIdProvider(clientId, clientSecret, redirectUrl)
	} else if typ == "Custom" {
		return NewCustomIdProvider(clientId, clientSecret, redirectUrl, authUrl, tokenUrl, userInfoUrl)
	} else if typ == "Infoflow" {
		if subType == "Internal" {
			return NewInfoflowInternalIdProvider(clientId, clientSecret, appId, redirectUrl)
		} else if subType == "Third-party" {
			return NewInfoflowIdProvider(clientId, clientSecret, appId, redirectUrl)
		} else {
			return nil
		}
	} else if typ == "Casdoor" {
		return NewCasdoorIdProvider(clientId, clientSecret, redirectUrl, hostUrl)
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
