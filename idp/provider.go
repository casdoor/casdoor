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
	"net/http"

	"golang.org/x/oauth2"
)

type UserInfo struct {
	Username  string
	Email     string
	AvatarUrl string
}

type TencentAccessToken struct {
	AccessToken  string `json:"access_token"`  //接口调用凭证
	ExpiresIn    int64  `json:"expires_in"`    //access_token接口调用凭证超时时间，单位（秒）
	RefreshToken string `json:"refresh_token"` //用户刷新access_token
	Openid       string `json:"openid"`        //授权用户唯一标识
	Scope        string `json:"scope"`         //用户授权的作用域，使用英文逗号分隔
	Unionid      string `json:"unionid"`       //当且仅当该网站应用已获得该用户的userinfo授权时，才会出现该字段。
}

type TencentUserInfo struct {
	openid     string   //普通用户的标识，对当前开发者帐号唯一
	nickname   string   //普通用户昵称
	sex        int      //普通用户性别，1为男性，2为女性
	province   string   //普通用户个人资料填写的省份
	city       string   //普通用户个人资料填写的城市
	country    string   //国家，如中国为CN
	headimgurl string   //用户头像，最后一个数值代表正方形头像大小（有0、46、64、96、132数值可选，0代表640*640正方形头像），用户没有头像时该项为空
	privilege  []string //用户特权信息，json数组，如微信沃卡用户为（chinaunicom）
	unionid    string   //用户统一标识。针对一个微信开放平台帐号下的应用，同一用户的unionid是唯一的。
}

type IdProvider interface {
	SetHttpClient(client *http.Client)
	GetToken(code string) (*oauth2.Token, error)
	GetUserInfo(token *oauth2.Token) (*UserInfo, error)
}

// TencentIdProvider qq以及微信的IdProvider接口
type TencentIdProvider interface {
	SetHttpClient(client *http.Client)
	GetAccessToken(code string) (*TencentAccessToken, error)
	GetUserInfo(tencentAccessToken *TencentAccessToken) (*TencentUserInfo, error)
}

func GetIdProvider(providerType string, clientId string, clientSecret string, redirectUrl string) IdProvider {
	if providerType == "GitHub" {
		return NewGithubIdProvider(clientId, clientSecret, redirectUrl)
	} else if providerType == "Google" {
		return NewGoogleIdProvider(clientId, clientSecret, redirectUrl)
	} else if providerType == "QQ" {
		return NewQqIdProvider(clientId, clientSecret, redirectUrl)
	}

	return nil
}

// GetTencentIdProvider 获取qq或微信的IdProvider
func GetTencentIdProvider(providerType string, clientId string, clientSecret string, redirectUrl string) TencentIdProvider {
	if providerType == "WeChat" {
		return NewWeChatIdProvider(clientId, clientSecret, redirectUrl)
	}

	return nil
}
