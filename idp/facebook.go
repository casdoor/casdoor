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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/oauth2"
)

type FacebookIdProvider struct {
	Client *http.Client
	Config *oauth2.Config
}

func NewFacebookIdProvider(clientId string, clientSecret string, redirectUrl string) *FacebookIdProvider {
	idp := &FacebookIdProvider{}

	config := idp.getConfig(clientId, clientSecret, redirectUrl)
	idp.Config = config

	return idp
}

func (idp *FacebookIdProvider) SetHttpClient(client *http.Client) {
	idp.Client = client
}

// getConfig return a point of Config, which describes a typical 3-legged OAuth2 flow
func (idp *FacebookIdProvider) getConfig(clientId string, clientSecret string, redirectUrl string) *oauth2.Config {
	endpoint := oauth2.Endpoint{
		TokenURL: "https://graph.facebook.com/oauth/access_token",
	}

	config := &oauth2.Config{
		Scopes:       []string{"email,public_profile"},
		Endpoint:     endpoint,
		ClientID:     clientId,
		ClientSecret: clientSecret,
		RedirectURL:  redirectUrl,
	}

	return config
}

type FacebookAccessToken struct {
	AccessToken string `json:"access_token"` // Interface call credentials
	TokenType   string `json:"token_type"`   // Access token type
	ExpiresIn   int64  `json:"expires_in"`   // access_token interface call credential timeout time, unit (seconds)
}

type FacebookCheckToken struct {
	Data string `json:"data"`
}

// FacebookCheckTokenData
// Get more detail via: https://developers.facebook.com/docs/facebook-login/manually-build-a-login-flow#checktoken
type FacebookCheckTokenData struct {
	UserId string `json:"user_id"`
}

// GetToken use code get access_token (*operation of getting code ought to be done in front)
// get more detail via: https://developers.facebook.com/docs/facebook-login/manually-build-a-login-flow#confirm
func (idp *FacebookIdProvider) GetToken(code string) (*oauth2.Token, error) {
	params := url.Values{}
	params.Add("client_id", idp.Config.ClientID)
	params.Add("redirect_uri", idp.Config.RedirectURL)
	params.Add("client_secret", idp.Config.ClientSecret)
	params.Add("code", code)

	accessTokenUrl := fmt.Sprintf("https://graph.facebook.com/oauth/access_token?%s", params.Encode())

	accessTokenResp, err := idp.GetUrlResp(accessTokenUrl)
	if err != nil {
		return nil, err
	}

	var facebookAccessToken FacebookAccessToken
	if err = json.Unmarshal([]byte(accessTokenResp), &facebookAccessToken); err != nil {
		return nil, err
	}

	token := oauth2.Token{
		AccessToken: facebookAccessToken.AccessToken,
		TokenType:   "FacebookAccessToken",
		Expiry:      time.Time{},
	}

	return &token, nil
}

//{
//    "id": "123456789",
//    "name": "Example Name",
//    "name_format": "{first} {last}",
//    "picture": {
//        "data": {
//            "height": 50,
//            "is_silhouette": false,
//            "url": "https://example.com",
//            "width": 50
//        }
//    },
//    "email": "test@example.com"
//}

type FacebookUserInfo struct {
	Id         string   `json:"id"`          // The app user's App-Scoped User ID. This ID is unique to the app and cannot be used by other apps.
	Name       string   `json:"name"`        // The person's full name.
	NameFormat string   `json:"name_format"` // The person's name formatted to correctly handle Chinese, Japanese, or Korean ordering.
	Picture    struct { // The person's profile picture.
		Data struct { // This struct is different as https://developers.facebook.com/docs/graph-api/reference/user/picture/
			Height       int    `json:"height"`
			IsSilhouette bool   `json:"is_silhouette"`
			Url          string `json:"url"`
			Width        int    `json:"width"`
		} `json:"data"`
	} `json:"picture"`
	Email string `json:"email"` // The User's primary email address listed on their profile. This field will not be returned if no valid email address is available.
}

// GetUserInfo use FacebookAccessToken gotten before return FacebookUserInfo
// get more detail via: https://developers.facebook.com/docs/graph-api/reference/user
func (idp *FacebookIdProvider) GetUserInfo(token *oauth2.Token) (*UserInfo, error) {
	var facebookUserInfo FacebookUserInfo
	accessToken := token.AccessToken

	userIdUrl := fmt.Sprintf("https://graph.facebook.com/me?access_token=%s", accessToken)
	userIdResp, err := idp.GetUrlResp(userIdUrl)
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal([]byte(userIdResp), &facebookUserInfo); err != nil {
		return nil, err
	}

	userInfoUrl := fmt.Sprintf("https://graph.facebook.com/%s?fields=id,name,name_format,picture,email&access_token=%s", facebookUserInfo.Id, accessToken)
	userInfoResp, err := idp.GetUrlResp(userInfoUrl)
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal([]byte(userInfoResp), &facebookUserInfo); err != nil {
		return nil, err
	}

	userInfo := UserInfo{
		Id:          facebookUserInfo.Id,
		Username:    facebookUserInfo.Name,
		DisplayName: facebookUserInfo.Name,
		Email:       facebookUserInfo.Email,
		AvatarUrl:   facebookUserInfo.Picture.Data.Url,
	}
	return &userInfo, nil
}

func (idp *FacebookIdProvider) GetUrlResp(url string) (string, error) {
	resp, err := idp.Client.Get(url)
	if err != nil {
		return "", err
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
		return "", err
	}

	return buf.String(), nil
}
