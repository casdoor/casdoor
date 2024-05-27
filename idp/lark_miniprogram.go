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
	"io"
	"net/http"
	"strings"

	"golang.org/x/oauth2"
)

type LarkMiniProgramIdProvider struct {
	Client *http.Client
	Config *oauth2.Config
}

func NewLarkMiniProgramIdProvider(clientId, clientSecret string) *LarkMiniProgramIdProvider {
	return &LarkMiniProgramIdProvider{
		Config: &oauth2.Config{
			Scopes:       []string{},
			Endpoint:     oauth2.Endpoint{TokenURL: "https://open.feishu.cn/open-apis/auth/v3/app_access_token/internal"},
			ClientID:     clientId,
			ClientSecret: clientSecret,
		},
		Client: &http.Client{},
	}
}

func (idp *LarkMiniProgramIdProvider) SetHttpClient(client *http.Client) {
	idp.Client = client
}

/*
{
    "app_access_token": "t-g1044ghJRUIJJ5ZPPZMOHKWZISL33E4QSS3abcef",
    "code": 0,
    "expire": 7200,
    "msg": "ok",
    "tenant_access_token": "t-g1044ghJRUIJJ5ZPPZMOHKWZISL33E4QSS3abcef"
}
*/

type LarkMiniProgramAccessToken struct {
	Code              int    `json:"code"`
	Expire            int    `json:"expire"`
	Msg               string `json:"msg"`
	TenantAccessToken string `json:"tenant_access_token"`
	AppAccessToken    string `json:"app_access_token"`
}

// GetToken gets app_access_token or tenant_access_token
func (idp *LarkMiniProgramIdProvider) GetToken() (*LarkMiniProgramAccessToken, error) {
	params := map[string]string{
		"app_id":     idp.Config.ClientID,
		"app_secret": idp.Config.ClientSecret,
	}

	data, err := idp.sendRequest("POST", idp.Config.Endpoint.TokenURL, params, "")
	if err != nil {
		return nil, err
	}

	var appToken LarkMiniProgramAccessToken
	if err = json.Unmarshal(data, &appToken); err != nil {
		return nil, err
	}

	if appToken.Code != 0 {
		return nil, fmt.Errorf("GetToken() error, appToken.Code: %d, appToken.Msg: %s", appToken.Code, appToken.Msg)
	}

	return &appToken, nil
}

/*
{
    "code": 0,
    "msg": "success",
    "data": {
        "access_token": "u-5Dak9ZAxJ9tFUn8MaTD_BFM51FNdg5xzO0y010000HWb",
        "refresh_token": "ur-6EyFQZyplb9URrOx5NtT_HM53zrJg59HXwy040400G.e",
        "token_type": "Bearer",
        "expires_in": 7199,
        "refresh_expires_in": 2591999,
        "scope": "auth:user.id:read bitable:app"
    }
}
*/

type LarkMiniProgramUserAccessToken struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		AccessToken      string `json:"access_token"`
		RefreshToken     string `json:"refresh_token"`
		TokenType        string `json:"token_type"`
		ExpiresIn        int    `json:"expires_in"`
		RefreshExpiresIn int    `json:"refresh_expires_in"`
		Scope            string `json:"scope"`
	} `json:"data"`
}

// GetUserToken uses code to get access_token
func (idp *LarkMiniProgramIdProvider) GetUserToken(code string) (*LarkMiniProgramUserAccessToken, error) {
	appToken, err := idp.GetToken()
	if err != nil {
		return nil, err
	}

	body := map[string]string{
		"grant_type": "authorization_code",
		"code":       code,
	}

	data, err := idp.sendRequest("POST", "https://open.feishu.cn/open-apis/authen/v1/oidc/access_token", body, appToken.AppAccessToken)
	if err != nil {
		return nil, err
	}

	var userAccessToken LarkMiniProgramUserAccessToken
	if err = json.Unmarshal(data, &userAccessToken); err != nil {
		return nil, err
	}

	return &userAccessToken, nil
}

/*
{
    "code": 0,
    "msg": "success",
    "data": {
        "name": "zhangsan",
        "en_name": "zhangsan",
        "avatar_url": "www.feishu.cn/avatar/icon",
        "avatar_thumb": "www.feishu.cn/avatar/icon_thumb",
        "avatar_middle": "www.feishu.cn/avatar/icon_middle",
        "avatar_big": "www.feishu.cn/avatar/icon_big",
        "open_id": "ou-caecc734c2e3328a62489fe0648c4b98779515d3",
        "union_id": "on-d89jhsdhjsajkda7828enjdj328ydhhw3u43yjhdj",
        "email": "zhangsan@feishu.cn",
        "enterprise_email": "demo@mail.com",
        "user_id": "5d9bdxxx",
        "mobile": "+86130002883xx",
        "tenant_key": "736588c92lxf175d",
		"employee_no": "111222333"
    }
}
*/

type LarkMiniProgramUserInfo struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		Name            string `json:"name"`
		EnName          string `json:"en_name"`
		AvatarUrl       string `json:"avatar_url"`
		AvatarThumb     string `json:"avatar_thumb"`
		AvatarMiddle    string `json:"avatar_middle"`
		AvatarBig       string `json:"avatar_big"`
		OpenId          string `json:"open_id"`
		UnionId         string `json:"union_id"`
		Email           string `json:"email"`
		EnterpriseEmail string `json:"enterprise_email"`
		UserId          string `json:"user_id"`
		Mobile          string `json:"mobile"`
		TenantKey       string `json:"tenant_key"`
		EmployeeNo      string `json:"employee_no"`
	} `json:"data"`
}

// GetUserInfo uses LarkMiniProgramAccessToken to return LinkedInUserInfo
func (idp *LarkMiniProgramIdProvider) GetUserInfo(userToken *LarkMiniProgramUserAccessToken) (*LarkMiniProgramUserInfo, error) {
	data, err := idp.sendRequest("GET", "https://open.feishu.cn/open-apis/authen/v1/user_info", nil, userToken.Data.AccessToken)
	if err != nil {
		return nil, err
	}

	var userInfo LarkMiniProgramUserInfo
	if err = json.Unmarshal(data, &userInfo); err != nil {
		return nil, err
	}

	return &userInfo, nil
}

func (idp *LarkMiniProgramIdProvider) sendRequest(method, url string, body interface{}, accessToken string) ([]byte, error) {
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reader = strings.NewReader(string(data))
	}

	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json;charset=UTF-8")
	if accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}

	resp, err := idp.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}
