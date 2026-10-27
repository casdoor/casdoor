// Copyright 2022 The Casdoor Authors. All Rights Reserved.
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
	"io"
	"net/http"
	"sync"
	"time"

	"golang.org/x/oauth2"
)

type WeChatMiniProgramIdProvider struct {
	Client *http.Client
	Config *oauth2.Config
}

func NewWeChatMiniProgramIdProvider(clientId string, clientSecret string) *WeChatMiniProgramIdProvider {
	idp := &WeChatMiniProgramIdProvider{}

	config := idp.getConfig(clientId, clientSecret)
	idp.Config = config
	idp.Client = &http.Client{}
	return idp
}

func (idp *WeChatMiniProgramIdProvider) SetHttpClient(client *http.Client) {
	idp.Client = client
}

func (idp *WeChatMiniProgramIdProvider) getConfig(clientId string, clientSecret string) *oauth2.Config {
	config := &oauth2.Config{
		ClientID:     clientId,
		ClientSecret: clientSecret,
	}

	return config
}

type WeChatMiniProgramSessionResponse struct {
	Openid     string `json:"openid"`
	SessionKey string `json:"session_key"`
	Unionid    string `json:"unionid"`
	Errcode    int    `json:"errcode"`
	Errmsg     string `json:"errmsg"`
}

func (idp *WeChatMiniProgramIdProvider) GetSessionByCode(code string) (*WeChatMiniProgramSessionResponse, error) {
	sessionUri := fmt.Sprintf("https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code", idp.Config.ClientID, idp.Config.ClientSecret, code)
	sessionResponse, err := idp.Client.Get(sessionUri)
	if err != nil {
		return nil, err
	}
	defer sessionResponse.Body.Close()
	data, err := io.ReadAll(sessionResponse.Body)
	if err != nil {
		return nil, err
	}
	var session WeChatMiniProgramSessionResponse
	err = json.Unmarshal(data, &session)
	if err != nil {
		return nil, err
	}
	if session.Errcode != 0 {
		return nil, fmt.Errorf("err: %s", session.Errmsg)
	}
	return &session, nil
}

// WeChatMiniProgramPhoneInfo is the phone_info returned by wxa/business/getuserphonenumber.
type WeChatMiniProgramPhoneInfo struct {
	PhoneNumber     string `json:"phoneNumber"`
	PurePhoneNumber string `json:"purePhoneNumber"`
	CountryCode     string `json:"countryCode"`
}

// E164 returns the phone number as +<country code><national number>, e.g. +8613800000000.
func (info *WeChatMiniProgramPhoneInfo) E164() string {
	countryCode := info.CountryCode
	if countryCode == "" {
		countryCode = "86"
	}
	return fmt.Sprintf("+%s%s", countryCode, info.PurePhoneNumber)
}

// RegionCode maps the WeChat country calling code to the region code stored in User.CountryCode.
func (info *WeChatMiniProgramPhoneInfo) RegionCode() string {
	switch info.CountryCode {
	case "", "86":
		return "CN"
	case "852":
		return "HK"
	case "853":
		return "MO"
	case "886":
		return "TW"
	default:
		return ""
	}
}

type weChatMiniProgramAccessToken struct {
	token    string
	expireAt time.Time
}

var (
	weChatMiniProgramAccessTokens   = map[string]weChatMiniProgramAccessToken{}
	weChatMiniProgramAccessTokensMu sync.Mutex
)

// getAccessToken uses cgi-bin/stable_token without force_refresh, so that it does not
// invalidate the access tokens held by other services that share the same mini program.
func (idp *WeChatMiniProgramIdProvider) getAccessToken() (string, error) {
	appId := idp.Config.ClientID
	weChatMiniProgramAccessTokensMu.Lock()
	cached, ok := weChatMiniProgramAccessTokens[appId]
	weChatMiniProgramAccessTokensMu.Unlock()
	if ok && time.Now().Before(cached.expireAt) {
		return cached.token, nil
	}

	body, err := json.Marshal(map[string]interface{}{
		"grant_type":    "client_credential",
		"appid":         appId,
		"secret":        idp.Config.ClientSecret,
		"force_refresh": false,
	})
	if err != nil {
		return "", err
	}
	resp, err := idp.Client.Post("https://api.weixin.qq.com/cgi-bin/stable_token", "application/json", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		Errcode     int    `json:"errcode"`
		Errmsg      string `json:"errmsg"`
	}
	if err = json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", err
	}
	if tokenResp.Errcode != 0 || tokenResp.AccessToken == "" {
		return "", fmt.Errorf("get access token: errcode %d, %s", tokenResp.Errcode, tokenResp.Errmsg)
	}

	// refresh one minute early
	expireAt := time.Now().Add(time.Duration(tokenResp.ExpiresIn-60) * time.Second)
	weChatMiniProgramAccessTokensMu.Lock()
	weChatMiniProgramAccessTokens[appId] = weChatMiniProgramAccessToken{token: tokenResp.AccessToken, expireAt: expireAt}
	weChatMiniProgramAccessTokensMu.Unlock()
	return tokenResp.AccessToken, nil
}

// GetPhoneNumber exchanges the code from the getPhoneNumber button for the user's phone number.
// https://developers.weixin.qq.com/miniprogram/dev/OpenApiDoc/user-info/phone-number/getPhoneNumber.html
func (idp *WeChatMiniProgramIdProvider) GetPhoneNumber(phoneCode string) (*WeChatMiniProgramPhoneInfo, error) {
	accessToken, err := idp.getAccessToken()
	if err != nil {
		return nil, err
	}

	body, err := json.Marshal(map[string]string{"code": phoneCode})
	if err != nil {
		return nil, err
	}
	resp, err := idp.Client.Post("https://api.weixin.qq.com/wxa/business/getuserphonenumber?access_token="+accessToken, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var phoneResp struct {
		PhoneInfo WeChatMiniProgramPhoneInfo `json:"phone_info"`
		Errcode   int                        `json:"errcode"`
		Errmsg    string                     `json:"errmsg"`
	}
	if err = json.NewDecoder(resp.Body).Decode(&phoneResp); err != nil {
		return nil, err
	}
	if phoneResp.Errcode != 0 || phoneResp.PhoneInfo.PurePhoneNumber == "" {
		if phoneResp.Errcode == 40001 || phoneResp.Errcode == 42001 {
			// the cached access token was invalidated: drop it so the next request fetches a new one
			weChatMiniProgramAccessTokensMu.Lock()
			delete(weChatMiniProgramAccessTokens, idp.Config.ClientID)
			weChatMiniProgramAccessTokensMu.Unlock()
		}
		return nil, fmt.Errorf("errcode %d, %s", phoneResp.Errcode, phoneResp.Errmsg)
	}
	return &phoneResp.PhoneInfo, nil
}
