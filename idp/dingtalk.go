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

	"golang.org/x/oauth2"
)

type DingTalkIdProvider struct {
	Client *http.Client
	Config *oauth2.Config
}

// NewDingTalkIdProvider ...
func NewDingTalkIdProvider(clientId string, clientSecret string, redirectUrl string) *DingTalkIdProvider {
	idp := &DingTalkIdProvider{}

	config := idp.getConfig(clientId, clientSecret, redirectUrl)
	idp.Config = config

	return idp
}

// SetHttpClient ...
func (idp *DingTalkIdProvider) SetHttpClient(client *http.Client) {
	idp.Client = client
}

// getConfig return a point of Config, which describes a typical 3-legged OAuth2 flow
func (idp *DingTalkIdProvider) getConfig(clientId string, clientSecret string, redirectUrl string) *oauth2.Config {
	endpoint := oauth2.Endpoint{
		AuthURL:  "https://api.dingtalk.com/v1.0/contact/users/me",
		TokenURL: "https://api.dingtalk.com/v1.0/oauth2/userAccessToken",
	}

	config := &oauth2.Config{
		// DingTalk not allow to set scopes,here it is just a placeholder,
		// convenient to use later
		Scopes: []string{"", ""},

		Endpoint:     endpoint,
		ClientID:     clientId,
		ClientSecret: clientSecret,
		RedirectURL:  redirectUrl,
	}

	return config
}

type DingTalkAccessToken struct {
	ErrCode     int    `json:"code"`
	ErrMsg      string `json:"message"`
	AccessToken string `json:"accessToken"` // Interface call credentials
	ExpiresIn   int64  `json:"expireIn"`    // access_token interface call credential timeout time, unit (seconds)
}

// GetToken use code get access_token (*operation of getting authCode ought to be done in front)
// get more detail via: https://open.dingtalk.com/document/orgapp-server/obtain-user-token
func (idp *DingTalkIdProvider) GetToken(code string) (*oauth2.Token, error) {
	pTokenParams := &struct {
		ClientId     string `json:"clientId"`
		ClientSecret string `json:"clientSecret"`
		Code         string `json:"code"`
		GrantType    string `json:"grantType"`
	}{idp.Config.ClientID, idp.Config.ClientSecret, code, "authorization_code"}

	data, err := idp.postWithBody(pTokenParams, idp.Config.Endpoint.TokenURL)
	if err != nil {
		return nil, err
	}

	pToken := &DingTalkAccessToken{}
	err = json.Unmarshal(data, pToken)
	if err != nil {
		return nil, err
	}

	if pToken.ErrCode != 0 {
		return nil, fmt.Errorf("pToken.Errcode = %d, pToken.Errmsg = %s", pToken.ErrCode, pToken.ErrMsg)
	}

	token := &oauth2.Token{
		AccessToken: pToken.AccessToken,
		Expiry:      time.Unix(time.Now().Unix()+pToken.ExpiresIn, 0),
	}
	return token, nil
}

/*
{
{
  "nick" : "zhangsan",
  "avatarUrl" : "https://xxx",
  "mobile" : "150xxxx9144",
  "openId" : "123",
  "unionId" : "z21HjQliSzpw0Yxxxx",
  "email" : "zhangsan@alibaba-inc.com",
  "stateCode" : "86"
}
*/

type DingTalkUserResponse struct {
	Nick      string `json:"nick"`
	OpenId    string `json:"openId"`
	UnionId   string `json:"unionId"`
	AvatarUrl string `json:"avatarUrl"`
	Email     string `json:"email"`
	Errmsg    string `json:"message"`
	Errcode   string `json:"code"`
}

// GetUserInfo Use  access_token to get UserInfo
// get more detail via: https://open.dingtalk.com/document/orgapp-server/dingtalk-retrieve-user-information
func (idp *DingTalkIdProvider) GetUserInfo(token *oauth2.Token) (*UserInfo, error) {
	dtUserInfo := &DingTalkUserResponse{}
	accessToken := token.AccessToken

	reqest, err := http.NewRequest("GET", idp.Config.Endpoint.AuthURL, nil)
	if err != nil {
		return nil, err
	}
	reqest.Header.Add("x-acs-dingtalk-access-token", accessToken)
	resp, err := idp.Client.Do(reqest)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, dtUserInfo)
	if err != nil {
		return nil, err
	}

	if dtUserInfo.Errmsg != "" {
		return nil, fmt.Errorf("userIdResp.Errcode = %s, userIdResp.Errmsg = %s", dtUserInfo.Errcode, dtUserInfo.Errmsg)
	}

	userInfo := UserInfo{
		Id:          dtUserInfo.OpenId,
		Username:    dtUserInfo.Nick,
		DisplayName: dtUserInfo.Nick,
		UnionId:     dtUserInfo.UnionId,
		Email:       dtUserInfo.Email,
		AvatarUrl:   dtUserInfo.AvatarUrl,
	}

	return &userInfo, nil
}

func (idp *DingTalkIdProvider) postWithBody(body interface{}, url string) ([]byte, error) {
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
