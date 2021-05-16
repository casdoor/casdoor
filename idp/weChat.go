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
	"bytes"
	"encoding/json"
	"fmt"
	"golang.org/x/oauth2"
	"io"
	"net/http"
	"net/url"
)

type WeChatIdProvider struct {
	Client *http.Client
	// Config:
	// - ClientID: 应用唯一标识, 用于请求CODE、通过CODE获取/刷新access_token，此信息前端可获取
	// - ClientSecret: 应用密钥，此信息保存在后端，前端不可获取【如保存在环境变量中】
	// - RedirectURL: 用户允许授权后，将会重定向到redirect_uri的网址上，并且带上code和state参数
	Config *oauth2.Config
}

func NewWeChatIdProvider(clientId string, clientSecret string, redirectUrl string) *WeChatIdProvider {
	idp := &WeChatIdProvider{}

	config := idp.getConfig(clientId, clientSecret, redirectUrl)
	idp.Config = config

	return idp
}

func (idp *WeChatIdProvider) SetHttpClient(client *http.Client) {
	idp.Client = client
}

// getConfig 返回一个Config的指针，其描述了一个典型的OAuth2.0流
func (idp *WeChatIdProvider) getConfig(clientId string, clientSecret string, redirectUrl string) *oauth2.Config {
	var endpoint = oauth2.Endpoint{
		TokenURL: "https://graph.qq.com/oauth2.0/token",
	}

	var config = &oauth2.Config{
		Scopes:       []string{"snsapi_login"},
		Endpoint:     endpoint,
		ClientID:     clientId,
		ClientSecret: clientSecret,
		RedirectURL:  redirectUrl,
	}

	return config
}

// GetAccessToken 通过code获取access_token (*获取code的操作在前端完成)
//具体参数等内容详见：https://developers.weixin.qq.com/doc/oplatform/Website_App/WeChat_Login/Wechat_Login.html
func (idp *WeChatIdProvider) GetAccessToken(code string) (*TencentAccessToken, error) {
	params := url.Values{}
	params.Add("grant_type", "authorization_code")
	params.Add("appid", idp.Config.ClientID)
	params.Add("secret", idp.Config.ClientSecret)
	params.Add("code", code)

	getAccessTokenUrl := fmt.Sprintf("https://api.weixin.qq.com/sns/oauth2/access_token?%s", params.Encode())
	tokenResponse, err := idp.Client.Get(getAccessTokenUrl)
	if err != nil {
		return nil, err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(tokenResponse.Body)

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(tokenResponse.Body)
	if err != nil {
		return nil, err
	}

	var tencentAccessToken TencentAccessToken
	if err = json.Unmarshal([]byte(buf.String()), &tencentAccessToken); err != nil {
		return nil, err
	}

	return &tencentAccessToken, nil
}

// GetUserInfo 根据之前获得的TencentAccessToken返回TencentUserInfo
// 具体参数等内容详见：https://developers.weixin.qq.com/doc/oplatform/Website_App/WeChat_Login/Authorized_Interface_Calling_UnionID.html
func (idp *WeChatIdProvider) GetUserInfo(tencentAccessToken *TencentAccessToken) (*TencentUserInfo, error) {
	var tencentUserInfo TencentUserInfo

	getUserInfoUrl := fmt.Sprintf("https://api.weixin.qq.com/sns/userinfo?access_token=%s&openid=%s", tencentAccessToken.AccessToken, tencentAccessToken.Openid)
	getUserInfoResponse, err := idp.Client.Get(getUserInfoUrl)
	if err != nil {
		return nil, err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(getUserInfoResponse.Body)

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(getUserInfoResponse.Body)
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal([]byte(buf.String()), &tencentUserInfo); err != nil {
		return nil, err
	}

	return &tencentUserInfo, nil
}
