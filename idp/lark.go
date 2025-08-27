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
	"time"

	"github.com/nyaruka/phonenumbers"
	"golang.org/x/oauth2"
)

type LarkIdProvider struct {
	Client     *http.Client
	Config     *oauth2.Config
	LarkDomain string
}

func NewLarkIdProvider(clientId string, clientSecret string, redirectUrl string, useGlobalEndpoint bool) *LarkIdProvider {
	idp := &LarkIdProvider{}

	if useGlobalEndpoint {
		idp.LarkDomain = "https://open.larksuite.com"
	} else {
		idp.LarkDomain = "https://open.feishu.cn"
	}

	config := idp.getConfig(clientId, clientSecret, redirectUrl)
	idp.Config = config
	return idp
}

func (idp *LarkIdProvider) SetHttpClient(client *http.Client) {
	idp.Client = client
}

// getConfig return a point of Config, which describes a typical 3-legged OAuth2 flow
func (idp *LarkIdProvider) getConfig(clientId string, clientSecret string, redirectUrl string) *oauth2.Config {
	endpoint := oauth2.Endpoint{
		TokenURL: idp.LarkDomain + "/open-apis/auth/v3/tenant_access_token/internal",
	}

	config := &oauth2.Config{
		Scopes:       []string{},
		Endpoint:     endpoint,
		ClientID:     clientId,
		ClientSecret: clientSecret,
		RedirectURL:  redirectUrl,
	}

	return config
}

/*
{
    "code": 0,
    "msg": "success",
    "tenant_access_token": "t-caecc734c2e3328a62489fe0648c4b98779515d3",
    "expire": 7140
}
*/

type LarkAccessToken struct {
	Code              int    `json:"code"`
	Msg               string `json:"msg"`
	TenantAccessToken string `json:"tenant_access_token"`
	Expire            int    `json:"expire"`
}

// GetToken use code get access_token (*operation of getting code ought to be done in front)
// get more detail via: https://docs.microsoft.com/en-us/linkedIn/shared/authentication/authorization-code-flow?context=linkedIn%2Fcontext&tabs=HTTPS
func (idp *LarkIdProvider) GetToken(code string) (*oauth2.Token, error) {
	params := &struct {
		AppID     string `json:"app_id"`
		AppSecret string `json:"app_secret"`
	}{idp.Config.ClientID, idp.Config.ClientSecret}

	data, err := idp.postWithBody(params, idp.Config.Endpoint.TokenURL)
	if err != nil {
		return nil, err
	}

	appToken := &LarkAccessToken{}
	err = json.Unmarshal(data, appToken)
	if err != nil {
		return nil, err
	}

	if appToken.Code != 0 {
		return nil, fmt.Errorf("GetToken() error, appToken.Code: %d, appToken.Msg: %s", appToken.Code, appToken.Msg)
	}

	t := &oauth2.Token{
		AccessToken: appToken.TenantAccessToken,
		TokenType:   "Bearer",
		Expiry:      time.Unix(time.Now().Unix()+int64(appToken.Expire), 0),
	}

	raw := make(map[string]interface{})
	raw["code"] = code
	t = t.WithExtra(raw)
	return t, nil
}

/*
{
    "code": 0,
    "msg": "success",
    "data": {
        "access_token": "u-6U1SbDiM6XIH2DcTCPyeub",
        "token_type": "Bearer",
        "expires_in": 7140,
        "name": "zhangsan",
        "en_name": "Three Zhang",
        "avatar_url": "www.feishu.cn/avatar/icon",
        "avatar_thumb": "www.feishu.cn/avatar/icon_thumb",
        "avatar_middle": "www.feishu.cn/avatar/icon_middle",
        "avatar_big": "www.feishu.cn/avatar/icon_big",
        "open_id": "ou-caecc734c2e3328a62489fe0648c4b98779515d3",
        "union_id": "on-d89jhsdhjsajkda7828enjdj328ydhhw3u43yjhdj",
        "email": "zhangsan@feishu.cn",
        "enterprise_email": "zhangsan@company.com",
        "user_id": "5d9bdxxx",
        "mobile": "+86130002883xx",
        "tenant_key": "736588c92lxf175d",
        "refresh_expires_in": 2591940,
        "refresh_token": "ur-t9HHgRCjMqGqIU9v05Zhos"
    }
}
*/

type LarkUserInfo struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		AccessToken      string `json:"access_token"`
		TokenType        string `json:"token_type"`
		ExpiresIn        int    `json:"expires_in"`
		Name             string `json:"name"`
		EnName           string `json:"en_name"`
		AvatarUrl        string `json:"avatar_url"`
		AvatarThumb      string `json:"avatar_thumb"`
		AvatarMiddle     string `json:"avatar_middle"`
		AvatarBig        string `json:"avatar_big"`
		OpenId           string `json:"open_id"`
		UnionId          string `json:"union_id"`
		Email            string `json:"email"`
		EnterpriseEmail  string `json:"enterprise_email"`
		UserId           string `json:"user_id"`
		Mobile           string `json:"mobile"`
		TenantKey        string `json:"tenant_key"`
		RefreshExpiresIn int    `json:"refresh_expires_in"`
		RefreshToken     string `json:"refresh_token"`
	} `json:"data"`
}

// GetUserInfo use LarkAccessToken gotten before return LinkedInUserInfo
// get more detail via: https://docs.microsoft.com/en-us/linkedin/consumer/integrations/self-serve/sign-in-with-linkedin?context=linkedin/consumer/context
func (idp *LarkIdProvider) GetUserInfo(token *oauth2.Token) (*UserInfo, error) {
	body := &struct {
		GrantType string `json:"grant_type"`
		Code      string `json:"code"`
	}{"authorization_code", token.Extra("code").(string)}

	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", idp.LarkDomain+"/open-apis/authen/v1/access_token", strings.NewReader(string(data)))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json;charset=UTF-8")
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)

	resp, err := idp.Client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	data, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var larkUserInfo LarkUserInfo
	err = json.Unmarshal(data, &larkUserInfo)
	if err != nil {
		return nil, err
	}

	// Use enterprise_email as fallback when email is empty
	email := larkUserInfo.Data.Email
	if email == "" {
		email = larkUserInfo.Data.EnterpriseEmail
	}

	var phoneNumber string
	var countryCode string
	if len(larkUserInfo.Data.Mobile) != 0 {
		phoneNumberParsed, err := phonenumbers.Parse(larkUserInfo.Data.Mobile, "")
		if err != nil {
			return nil, err
		}
		countryCode = phonenumbers.GetRegionCodeForNumber(phoneNumberParsed)
		phoneNumber = fmt.Sprintf("%d", phoneNumberParsed.GetNationalNumber())
	}

	userInfo := UserInfo{
		Id:          larkUserInfo.Data.OpenId,
		DisplayName: larkUserInfo.Data.Name,
		Username:    larkUserInfo.Data.UserId,
		Email:       email,
		AvatarUrl:   larkUserInfo.Data.AvatarUrl,
		Phone:       phoneNumber,
		CountryCode: countryCode,
	}
	return &userInfo, nil
}

func (idp *LarkIdProvider) postWithBody(body interface{}, url string) ([]byte, error) {
	bs, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	r := strings.NewReader(string(bs))
	resp, err := idp.Client.Post(url, "application/json;charset=UTF-8", r)
	if err != nil {
		return nil, err
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(resp.Body)
	return data, nil
}
