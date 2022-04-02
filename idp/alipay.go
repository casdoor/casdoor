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
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

type AlipayIdProvider struct {
	Client *http.Client
	Config *oauth2.Config
}

// NewAlipayIdProvider ...
func NewAlipayIdProvider(clientId string, clientSecret string, redirectUrl string) *AlipayIdProvider {
	idp := &AlipayIdProvider{}

	config := idp.getConfig(clientId, clientSecret, redirectUrl)
	idp.Config = config

	return idp
}

// SetHttpClient ...
func (idp *AlipayIdProvider) SetHttpClient(client *http.Client) {
	idp.Client = client
}

// getConfig return a point of Config, which describes a typical 3-legged OAuth2 flow
func (idp *AlipayIdProvider) getConfig(clientId string, clientSecret string, redirectUrl string) *oauth2.Config {
	var endpoint = oauth2.Endpoint{
		AuthURL:  "https://openauth.alipay.com/oauth2/publicAppAuthorize.htm",
		TokenURL: "https://openapi.alipay.com/gateway.do",
	}

	var config = &oauth2.Config{
		Scopes:       []string{"", ""},
		Endpoint:     endpoint,
		ClientID:     clientId,
		ClientSecret: clientSecret,
		RedirectURL:  redirectUrl,
	}

	return config
}

type AlipayAccessToken struct {
	Response AlipaySystemOauthTokenResponse `json:"alipay_system_oauth_token_response"`
	Sign     string                         `json:"sign"`
}

type AlipaySystemOauthTokenResponse struct {
	AccessToken  string `json:"access_token"`
	AlipayUserId string `json:"alipay_user_id"`
	ExpiresIn    int    `json:"expires_in"`
	ReExpiresIn  int    `json:"re_expires_in"`
	RefreshToken string `json:"refresh_token"`
	UserId       string `json:"user_id"`
}

// GetToken use code to get access_token
func (idp *AlipayIdProvider) GetToken(code string) (*oauth2.Token, error) {
	pTokenParams := &struct {
		ClientId  string `json:"app_id"`
		CharSet   string `json:"charset"`
		Code      string `json:"code"`
		GrantType string `json:"grant_type"`
		Method    string `json:"method"`
		SignType  string `json:"sign_type"`
		TimeStamp string `json:"timestamp"`
		Version   string `json:"version"`
	}{idp.Config.ClientID, "utf-8", code, "authorization_code", "alipay.system.oauth.token", "RSA2", time.Now().Format("2006-01-02 15:04:05"), "1.0"}

	data, err := idp.postWithBody(pTokenParams, idp.Config.Endpoint.TokenURL)
	if err != nil {
		return nil, err
	}

	pToken := &AlipayAccessToken{}
	err = json.Unmarshal(data, pToken)
	if err != nil {
		return nil, err
	}

	token := &oauth2.Token{
		AccessToken: pToken.Response.AccessToken,
		Expiry:      time.Unix(time.Now().Unix()+int64(pToken.Response.ExpiresIn), 0),
	}
	return token, nil
}

/*
{
    "alipay_user_info_share_response":{
        "code":"10000",
        "msg":"Success",
        "avatar":"https:\/\/tfs.alipayobjects.com\/images\/partner\/T1.QxFXk4aXXXXXXXX",
        "nick_name":"zhangsan",
        "user_id":"2099222233334444"
    },
    "sign":"m8rWJeqfoa5tDQRRVnPhRHcpX7NZEgjIPTPF1QBxos6XXXXXXXXXXXXXXXXXXXXXXXXXX"
}
*/

type AlipayUserResponse struct {
	AlipayUserInfoShareResponse AlipayUserInfoShareResponse `json:"alipay_user_info_share_response"`
	Sign                        string                      `json:"sign"`
}

type AlipayUserInfoShareResponse struct {
	Code     string `json:"code"`
	Msg      string `json:"msg"`
	Avatar   string `json:"avatar"`
	NickName string `json:"nick_name"`
	UserId   string `json:"user_id"`
}

// GetUserInfo Use access_token to get UserInfo
func (idp *AlipayIdProvider) GetUserInfo(token *oauth2.Token) (*UserInfo, error) {
	atUserInfo := &AlipayUserResponse{}
	accessToken := token.AccessToken

	pTokenParams := &struct {
		ClientId  string `json:"app_id"`
		CharSet   string `json:"charset"`
		AuthToken string `json:"auth_token"`
		Method    string `json:"method"`
		SignType  string `json:"sign_type"`
		TimeStamp string `json:"timestamp"`
		Version   string `json:"version"`
	}{idp.Config.ClientID, "utf-8", accessToken, "alipay.user.info.share", "RSA2", time.Now().Format("2006-01-02 15:04:05"), "1.0"}
	data, err := idp.postWithBody(pTokenParams, idp.Config.Endpoint.TokenURL)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, atUserInfo)
	if err != nil {
		return nil, err
	}

	userInfo := UserInfo{
		Id:          atUserInfo.AlipayUserInfoShareResponse.UserId,
		Username:    atUserInfo.AlipayUserInfoShareResponse.NickName,
		DisplayName: atUserInfo.AlipayUserInfoShareResponse.NickName,
		AvatarUrl:   atUserInfo.AlipayUserInfoShareResponse.Avatar,
	}

	return &userInfo, nil
}

func (idp *AlipayIdProvider) postWithBody(body interface{}, targetUrl string) ([]byte, error) {
	bs, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	bodyJson := make(map[string]interface{})
	err = json.Unmarshal(bs, &bodyJson)
	if err != nil {
		return nil, err
	}

	formData := url.Values{}
	for k := range bodyJson {
		formData.Set(k, bodyJson[k].(string))
	}

	sign, err := rsaSignWithRSA256(getStringToSign(formData), idp.Config.ClientSecret)
	if err != nil {
		return nil, err
	}

	formData.Set("sign", sign)

	resp, err := idp.Client.PostForm(targetUrl, formData)
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadAll(resp.Body)

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

// get the string to sign, see https://opendocs.alipay.com/common/02kf5q
func getStringToSign(formData url.Values) string {
	keys := make([]string, 0, len(formData))
	for k := range formData {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	str := ""
	for _, k := range keys {
		if k == "sign" || formData[k][0] == "" {
			continue
		} else {
			str += "&" + k + "=" + formData[k][0]
		}
	}
	str = strings.Trim(str, "&")
	return str
}

// use privateKey to sign the content
func rsaSignWithRSA256(signContent string, privateKey string) (string, error) {
	privateKey = formatPrivateKey(privateKey)
	block, _ := pem.Decode([]byte(privateKey))
	if block == nil {
		panic("fail to parse privateKey")
	}

	h := sha256.New()
	h.Write([]byte(signContent))
	hashed := h.Sum(nil)

	privateKeyRSA, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return "", err
	}

	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKeyRSA.(*rsa.PrivateKey), crypto.SHA256, hashed)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(signature), nil
}

// privateKey in database is a string, format it to PEM style
func formatPrivateKey(privateKey string) string {
	// each line length is 64
	preFmtPrivateKey := ""
	for i := 0; ; {
		if i+64 <= len(privateKey) {
			preFmtPrivateKey = preFmtPrivateKey + privateKey[i:i+64] + "\n"
			i += 64
		} else {
			preFmtPrivateKey = preFmtPrivateKey + privateKey[i:]
			break
		}
	}
	privateKey = strings.Trim(preFmtPrivateKey, "\n")

	// add pkcs#8 BEGIN and END
	PemBegin := "-----BEGIN PRIVATE KEY-----\n"
	PemEnd := "\n-----END PRIVATE KEY-----"
	if !strings.HasPrefix(privateKey, PemBegin) {
		privateKey = PemBegin + privateKey
	}
	if !strings.HasSuffix(privateKey, PemEnd) {
		privateKey = privateKey + PemEnd
	}
	return privateKey
}
