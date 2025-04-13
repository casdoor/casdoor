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
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/skip2/go-qrcode"
	"golang.org/x/oauth2"
)

var (
	WechatCacheMap map[string]WechatCacheMapValue
	Lock           sync.RWMutex
)

type WeChatIdProvider struct {
	Client *http.Client
	Config *oauth2.Config
}

type WechatCacheMapValue struct {
	IsScanned     bool
	WechatUnionId string
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

// getConfig return a point of Config, which describes a typical 3-legged OAuth2 flow
func (idp *WeChatIdProvider) getConfig(clientId string, clientSecret string, redirectUrl string) *oauth2.Config {
	endpoint := oauth2.Endpoint{
		TokenURL: "https://graph.qq.com/oauth2.0/token",
	}

	config := &oauth2.Config{
		Scopes:       []string{"snsapi_login"},
		Endpoint:     endpoint,
		ClientID:     clientId,
		ClientSecret: clientSecret,
		RedirectURL:  redirectUrl,
	}

	return config
}

type WechatAccessToken struct {
	AccessToken  string `json:"access_token"`  // Interface call credentials
	ExpiresIn    int64  `json:"expires_in"`    // access_token interface call credential timeout time, unit (seconds)
	RefreshToken string `json:"refresh_token"` // User refresh access_token
	Openid       string `json:"openid"`        // Unique ID of authorized user
	Scope        string `json:"scope"`         // The scope of user authorization, separated by commas. (,)
	Unionid      string `json:"unionid"`       // This field will appear if and only if the website application has been authorized by the user's UserInfo.
}

// GetToken use code get access_token (*operation of getting code ought to be done in front)
// get more detail via: https://developers.weixin.qq.com/doc/oplatform/Website_App/WeChat_Login/Wechat_Login.html
func (idp *WeChatIdProvider) GetToken(code string) (*oauth2.Token, error) {
	if strings.HasPrefix(code, "wechat_oa:") {
		token := oauth2.Token{
			AccessToken: code,
			TokenType:   "WeChatAccessToken",
			Expiry:      time.Time{},
		}
		return &token, nil
	}

	params := url.Values{}
	params.Add("grant_type", "authorization_code")
	params.Add("appid", idp.Config.ClientID)
	params.Add("secret", idp.Config.ClientSecret)
	params.Add("code", code)

	accessTokenUrl := fmt.Sprintf("https://api.weixin.qq.com/sns/oauth2/access_token?%s", params.Encode())
	tokenResponse, err := idp.Client.Get(accessTokenUrl)
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

	// {"errcode":40163,"errmsg":"code been used, rid: 6206378a-793424c0-2e4091cc"}
	if strings.Contains(buf.String(), "errcode") {
		return nil, fmt.Errorf(buf.String())
	}

	var wechatAccessToken WechatAccessToken
	if err = json.Unmarshal(buf.Bytes(), &wechatAccessToken); err != nil {
		return nil, err
	}

	token := oauth2.Token{
		AccessToken:  wechatAccessToken.AccessToken,
		TokenType:    "WeChatAccessToken",
		RefreshToken: wechatAccessToken.RefreshToken,
		Expiry:       time.Time{},
	}

	raw := make(map[string]string)
	raw["Openid"] = wechatAccessToken.Openid
	token.WithExtra(raw)

	return &token, nil
}

//{
//	"openid": "of_Hl5zVpyj0vwzIlAyIlnXe1234",
//	"nickname": "飞翔的企鹅",
//	"sex": 1,
//	"language": "zh_CN",
//	"city": "Shanghai",
//	"province": "Shanghai",
//	"country": "CN",
//	"headimgurl": "https:\/\/thirdwx.qlogo.cn\/mmopen\/vi_32\/Q0j4TwGTfTK6xc7vGca4KtibJib5dslRianc9VHt9k2N7fewYOl8fak7grRM7nS5V6HcvkkIkGThWUXPjDbXkQFYA\/132",
//	"privilege": [],
//	"unionid": "oxW9O1VAL8x-zfWP2hrqW9c81234"
//}

type WechatUserInfo struct {
	Openid     string   `json:"openid"`   // The ID of an ordinary user, which is unique to the current developer account
	Nickname   string   `json:"nickname"` // Ordinary user nickname
	Sex        int      `json:"sex"`      // Ordinary user gender, 1 is male, 2 is female
	Language   string   `json:"language"`
	City       string   `json:"city"`       // City filled in by general user's personal data
	Province   string   `json:"province"`   // Province filled in by ordinary user's personal information
	Country    string   `json:"country"`    // Country, such as China is CN
	Headimgurl string   `json:"headimgurl"` // User avatar, the last value represents the size of the square avatar (there are optional values of 0, 46, 64, 96, 132, 0 represents a 640*640 square avatar), this item is empty when the user does not have an avatar
	Privilege  []string `json:"privilege"`  // User Privilege information, json array, such as Wechat Woka user (chinaunicom)
	Unionid    string   `json:"unionid"`    // Unified user identification. For an application under a WeChat open platform account, the unionid of the same user is unique.
}

// GetUserInfo use WechatAccessToken gotten before return WechatUserInfo
// get more detail via: https://developers.weixin.qq.com/doc/oplatform/Website_App/WeChat_Login/Authorized_Interface_Calling_UnionID.html
func (idp *WeChatIdProvider) GetUserInfo(token *oauth2.Token) (*UserInfo, error) {
	var wechatUserInfo WechatUserInfo
	accessToken := token.AccessToken

	if strings.HasPrefix(accessToken, "wechat_oa:") {
		Lock.RLock()
		mapValue, ok := WechatCacheMap[accessToken[10:]]
		Lock.RUnlock()

		if !ok || mapValue.WechatUnionId == "" {
			return nil, fmt.Errorf("error ticket")
		}

		Lock.Lock()
		delete(WechatCacheMap, accessToken[10:])
		Lock.Unlock()

		userInfo := UserInfo{
			Id:          mapValue.WechatUnionId,
			Username:    "wx_user_" + mapValue.WechatUnionId,
			DisplayName: "wx_user_" + mapValue.WechatUnionId,
			AvatarUrl:   "",
		}
		return &userInfo, nil
	}

	openid := token.Extra("Openid")

	userInfoUrl := fmt.Sprintf("https://api.weixin.qq.com/sns/userinfo?access_token=%s&openid=%s", accessToken, openid)
	resp, err := idp.Client.Get(userInfoUrl)
	if err != nil {
		return nil, err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(resp.Body)

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return nil, err
	}
	if err = json.Unmarshal(buf.Bytes(), &wechatUserInfo); err != nil {
		return nil, err
	}

	id := wechatUserInfo.Unionid
	if id == "" {
		id = wechatUserInfo.Openid
	}

	extra := make(map[string]string)
	extra["wechat_unionid"] = wechatUserInfo.Openid
	// For WeChat, different appId corresponds to different openId
	extra[BuildWechatOpenIdKey(idp.Config.ClientID)] = wechatUserInfo.Openid
	userInfo := UserInfo{
		Id:          id,
		Username:    wechatUserInfo.Nickname,
		DisplayName: wechatUserInfo.Nickname,
		AvatarUrl:   wechatUserInfo.Headimgurl,
		Extra:       extra,
	}
	return &userInfo, nil
}

func BuildWechatOpenIdKey(appId string) string {
	return fmt.Sprintf("wechat_openid_%s", appId)
}

func GetWechatOfficialAccountAccessToken(clientId string, clientSecret string) (string, string, error) {
	accessTokenUrl := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s", clientId, clientSecret)
	request, err := http.NewRequest("GET", accessTokenUrl, nil)
	if err != nil {
		return "", "", err
	}

	client := new(http.Client)
	resp, err := client.Do(request)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}

	var data struct {
		ExpireIn    int    `json:"expires_in"`
		AccessToken string `json:"access_token"`
		ErrCode     int    `json:"errcode"`
		Errmsg      string `json:"errmsg"`
	}
	err = json.Unmarshal(respBytes, &data)
	if err != nil {
		return "", "", err
	}

	return data.AccessToken, data.Errmsg, nil
}

func GetWechatOfficialAccountQRCode(clientId string, clientSecret string, providerId string) (string, string, error) {
	accessToken, errMsg, err := GetWechatOfficialAccountAccessToken(clientId, clientSecret)
	if err != nil {
		return "", "", err
	}

	if errMsg != "" {
		return "", "", fmt.Errorf("Fail to fetch WeChat QRcode: %s", errMsg)
	}

	client := new(http.Client)

	weChatEndpoint := "https://api.weixin.qq.com/cgi-bin/qrcode/create"
	qrCodeUrl := fmt.Sprintf("%s?access_token=%s", weChatEndpoint, accessToken)
	params := fmt.Sprintf(`{"expire_seconds": 3600, "action_name": "QR_STR_SCENE", "action_info": {"scene": {"scene_str": "%s"}}}`, providerId)

	bodyData := bytes.NewReader([]byte(params))
	request, err := http.NewRequest("POST", qrCodeUrl, bodyData)
	if err != nil {
		return "", "", err
	}

	resp, err := client.Do(request)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}
	var data struct {
		Ticket        string `json:"ticket"`
		ExpireSeconds int    `json:"expire_seconds"`
		URL           string `json:"url"`
	}
	err = json.Unmarshal(respBytes, &data)
	if err != nil {
		return "", "", err
	}

	var png []byte
	png, err = qrcode.Encode(data.URL, qrcode.Medium, 256)
	base64Image := base64.StdEncoding.EncodeToString(png)
	return base64Image, data.Ticket, nil
}

func VerifyWechatSignature(token string, nonce string, timestamp string, signature string) bool {
	// verify the signature
	tmpArr := sort.StringSlice{token, timestamp, nonce}
	sort.Sort(tmpArr)

	tmpStr := ""
	for _, str := range tmpArr {
		tmpStr = tmpStr + str
	}

	b := sha1.Sum([]byte(tmpStr))
	res := hex.EncodeToString(b[:])
	return res == signature
}
