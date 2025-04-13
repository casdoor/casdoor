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
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/casdoor/casdoor/util"
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
	Mobile    string `json:"mobile"`
	StateCode string `json:"stateCode"`
}

// GetUserInfo Use  access_token to get UserInfo
// get more detail via: https://open.dingtalk.com/document/orgapp-server/dingtalk-retrieve-user-information
func (idp *DingTalkIdProvider) GetUserInfo(token *oauth2.Token) (*UserInfo, error) {
	dtUserInfo := &DingTalkUserResponse{}
	accessToken := token.AccessToken

	request, err := http.NewRequest("GET", idp.Config.Endpoint.AuthURL, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Add("x-acs-dingtalk-access-token", accessToken)
	resp, err := idp.Client.Do(request)
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

	countryCode, err := util.GetCountryCode(dtUserInfo.StateCode, dtUserInfo.Mobile)
	if err != nil {
		return nil, err
	}

	userInfo := UserInfo{
		Id:          dtUserInfo.OpenId,
		Username:    dtUserInfo.Nick,
		DisplayName: dtUserInfo.Nick,
		UnionId:     dtUserInfo.UnionId,
		Email:       dtUserInfo.Email,
		Phone:       dtUserInfo.Mobile,
		CountryCode: countryCode,
		AvatarUrl:   dtUserInfo.AvatarUrl,
	}

	corpAccessToken := idp.getInnerAppAccessToken()
	userId, err := idp.getUserId(userInfo.UnionId, corpAccessToken)
	if err != nil {
		return nil, err
	}

	corpMobile, corpEmail, jobNumber, err := idp.getUserCorpEmail(userId, corpAccessToken)
	if err == nil {
		if corpMobile != "" {
			userInfo.Phone = corpMobile
		}

		if corpEmail != "" {
			userInfo.Email = corpEmail
		}

		if jobNumber != "" {
			userInfo.Username = jobNumber
		}
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

func (idp *DingTalkIdProvider) getInnerAppAccessToken() string {
	body := make(map[string]string)
	body["appKey"] = idp.Config.ClientID
	body["appSecret"] = idp.Config.ClientSecret
	respBytes, err := idp.postWithBody(body, "https://api.dingtalk.com/v1.0/oauth2/accessToken")
	if err != nil {
		log.Println(err.Error())
	}

	var data struct {
		ExpireIn    int    `json:"expireIn"`
		AccessToken string `json:"accessToken"`
	}
	err = json.Unmarshal(respBytes, &data)
	if err != nil {
		log.Println(err.Error())
	}
	return data.AccessToken
}

func (idp *DingTalkIdProvider) getUserId(unionId string, accessToken string) (string, error) {
	body := make(map[string]string)
	body["unionid"] = unionId
	respBytes, err := idp.postWithBody(body, "https://oapi.dingtalk.com/topapi/user/getbyunionid?access_token="+accessToken)
	if err != nil {
		return "", err
	}

	var data struct {
		ErrCode    int    `json:"errcode"`
		ErrMessage string `json:"errmsg"`
		Result     struct {
			UserId string `json:"userid"`
		} `json:"result"`
	}
	err = json.Unmarshal(respBytes, &data)
	if err != nil {
		return "", err
	}
	if data.ErrCode == 60121 {
		return "", fmt.Errorf("该应用只允许本企业内部用户登录，您不属于该企业，无法登录")
	} else if data.ErrCode != 0 {
		return "", fmt.Errorf(data.ErrMessage)
	}
	return data.Result.UserId, nil
}

func (idp *DingTalkIdProvider) getUserCorpEmail(userId string, accessToken string) (string, string, string, error) {
	// https://open.dingtalk.com/document/isvapp/query-user-details
	body := make(map[string]string)
	body["userid"] = userId
	respBytes, err := idp.postWithBody(body, "https://oapi.dingtalk.com/topapi/v2/user/get?access_token="+accessToken)
	if err != nil {
		return "", "", "", err
	}

	var data struct {
		ErrMessage string `json:"errmsg"`
		Result     struct {
			Mobile    string `json:"mobile"`
			Email     string `json:"email"`
			JobNumber string `json:"job_number"`
		} `json:"result"`
	}
	err = json.Unmarshal(respBytes, &data)
	if err != nil {
		return "", "", "", err
	}
	if data.ErrMessage != "ok" {
		return "", "", "", fmt.Errorf(data.ErrMessage)
	}
	return data.Result.Mobile, data.Result.Email, data.Result.JobNumber, nil
}
