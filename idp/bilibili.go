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
	"net/url"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

type BilibiliIdProvider struct {
	Client *http.Client
	Config *oauth2.Config
}

func NewBilibiliIdProvider(clientId string, clientSecret string, redirectUrl string) *BilibiliIdProvider {
	idp := &BilibiliIdProvider{}

	config := idp.getConfig(clientId, clientSecret, redirectUrl)
	idp.Config = config

	return idp
}

func (idp *BilibiliIdProvider) SetHttpClient(client *http.Client) {
	idp.Client = client
}

// getConfig return a point of Config, which describes a typical 3-legged OAuth2 flow
func (idp *BilibiliIdProvider) getConfig(clientId string, clientSecret string, redirectUrl string) *oauth2.Config {
	endpoint := oauth2.Endpoint{
		TokenURL: "https://api.bilibili.com/x/account-oauth2/v1/token",
		AuthURL:  "http://member.bilibili.com/arcopen/fn/user/account/info",
	}

	config := &oauth2.Config{
		Scopes:       []string{"", ""},
		Endpoint:     endpoint,
		ClientID:     clientId,
		ClientSecret: clientSecret,
		RedirectURL:  redirectUrl,
	}

	return config
}

type BilibiliProviderToken struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

type BilibiliIdProviderTokenResponse struct {
	Code    int                   `json:"code"`
	Message string                `json:"message"`
	TTL     int                   `json:"ttl"`
	Data    BilibiliProviderToken `json:"data"`
}

// GetToken
/*
{
    "code": 0,
    "message": "0",
    "ttl": 1,
    "data": {
         "access_token": "d30bedaa4d8eb3128cf35ddc1030e27d",
         "expires_in": 1630220614,
         "refresh_token": "WxFDKwqScZIQDm4iWmKDvetyFugM6HkX"
    }
}
*/
// GetToken use code get access_token (*operation of getting code ought to be done in front)
// get more detail via: https://openhome.bilibili.com/doc/4/eaf0e2b5-bde9-b9a0-9be1-019bb455701c
func (idp *BilibiliIdProvider) GetToken(code string) (*oauth2.Token, error) {
	pTokenParams := &struct {
		ClientId     string `json:"client_id"`
		ClientSecret string `json:"client_secret"`
		GrantType    string `json:"grant_type"`
		Code         string `json:"code"`
	}{
		idp.Config.ClientID,
		idp.Config.ClientSecret,
		"authorization_code",
		code,
	}

	data, err := idp.postWithBody(pTokenParams, idp.Config.Endpoint.TokenURL)
	if err != nil {
		return nil, err
	}

	response := &BilibiliIdProviderTokenResponse{}
	err = json.Unmarshal(data, response)
	if err != nil {
		return nil, err
	}

	if response.Code != 0 {
		return nil, fmt.Errorf("pToken.Errcode = %d, pToken.Errmsg = %s", response.Code, response.Message)
	}

	token := &oauth2.Token{
		AccessToken:  response.Data.AccessToken,
		Expiry:       time.Unix(time.Now().Unix()+int64(response.Data.ExpiresIn), 0),
		RefreshToken: response.Data.RefreshToken,
	}

	return token, nil
}

/*
{
    "code": 0,
    "message": "0",
    "ttl": 1,
    "data": {
        "name":"bilibili",
        "face":"http://i0.hdslb.com/bfs/face/e1c99895a9f9df4f260a70dc7e227bcb46cf319c.jpg",
        "openid":"9205eeaa1879skxys969ed47874f225c3"
    }
}
*/

type BilibiliUserInfo struct {
	Name   string `json:"name"`
	Face   string `json:"face"`
	OpenId string `json:"openid"`
}

type BilibiliUserInfoResponse struct {
	Code    int              `json:"code"`
	Message string           `json:"message"`
	TTL     int              `json:"ttl"`
	Data    BilibiliUserInfo `json:"data"`
}

// GetUserInfo Use  access_token to get UserInfo
// get more detail via: https://openhome.bilibili.com/doc/4/feb66f99-7d87-c206-00e7-d84164cd701c
func (idp *BilibiliIdProvider) GetUserInfo(token *oauth2.Token) (*UserInfo, error) {
	accessToken := token.AccessToken
	clientId := idp.Config.ClientID

	params := url.Values{}
	params.Add("client_id", clientId)
	params.Add("access_token", accessToken)

	userInfoUrl := fmt.Sprintf("%s?%s", idp.Config.Endpoint.AuthURL, params.Encode())

	resp, err := idp.Client.Get(userInfoUrl)
	if err != nil {
		return nil, err
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	bUserInfoResponse := &BilibiliUserInfoResponse{}
	if err = json.Unmarshal(data, bUserInfoResponse); err != nil {
		return nil, err
	}

	if bUserInfoResponse.Code != 0 {
		return nil, fmt.Errorf("userinfo.Errcode = %d, userinfo.Errmsg = %s", bUserInfoResponse.Code, bUserInfoResponse.Message)
	}

	userInfo := &UserInfo{
		Id:          bUserInfoResponse.Data.OpenId,
		Username:    bUserInfoResponse.Data.Name,
		DisplayName: bUserInfoResponse.Data.Name,
		AvatarUrl:   bUserInfoResponse.Data.Face,
	}

	return userInfo, nil
}

func (idp *BilibiliIdProvider) postWithBody(body interface{}, url string) ([]byte, error) {
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
