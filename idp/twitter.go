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
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

type TwitterIdProvider struct {
	Client *http.Client
	Config *oauth2.Config
}

func NewTwitterIdProvider(clientId string, clientSecret string, redirectUrl string) *TwitterIdProvider {
	idp := &TwitterIdProvider{}

	config := idp.getConfig(clientId, clientSecret, redirectUrl)
	idp.Config = config

	return idp
}

func (idp *TwitterIdProvider) SetHttpClient(client *http.Client) {
	idp.Client = client
}

// getConfig return a point of Config, which describes a typical 3-legged OAuth2 flow
func (idp *TwitterIdProvider) getConfig(clientId string, clientSecret string, redirectUrl string) *oauth2.Config {
	endpoint := oauth2.Endpoint{
		TokenURL: "https://api.twitter.com/2/oauth2/token",
	}

	config := &oauth2.Config{
		Scopes:       []string{"users.read", "tweet.read"},
		Endpoint:     endpoint,
		ClientID:     clientId,
		ClientSecret: clientSecret,
		RedirectURL:  redirectUrl,
	}

	return config
}

type TwitterAccessToken struct {
	AccessToken string `json:"access_token"` // Interface call credentials
	TokenType   string `json:"token_type"`   // Access token type
	ExpiresIn   int64  `json:"expires_in"`   // access_token interface call credential timeout time, unit (seconds)
}

type TwitterCheckToken struct {
	Data TwitterUserInfo `json:"data"`
}

// TwitterCheckTokenData
// Get more detail via: https://developers.Twitter.com/docs/Twitter-login/guides/advanced/manual-flow#checktoken
type TwitterCheckTokenData struct {
	UserId string `json:"user_id"`
}

// GetToken use code get access_token (*operation of getting code ought to be done in front)
// get more detail via: https://developers.Twitter.com/docs/Twitter-login/guides/advanced/manual-flow#confirm
func (idp *TwitterIdProvider) GetToken(code string) (*oauth2.Token, error) {
	params := url.Values{}
	// params.Add("client_id", idp.Config.ClientID)
	params.Add("redirect_uri", idp.Config.RedirectURL)
	params.Add("code_verifier", "casdoor-verifier")
	params.Add("code", code)
	params.Add("grant_type", "authorization_code")
	req, err := http.NewRequest("POST", "https://api.twitter.com/2/oauth2/token", strings.NewReader(params.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	e := base64.StdEncoding.EncodeToString([]byte(idp.Config.ClientID + ":" + idp.Config.ClientSecret))
	req.Header.Add("Authorization", "Basic "+e)
	accessTokenResp, err := idp.GetUrlResp(req)
	var TwitterAccessToken TwitterAccessToken
	if err = json.Unmarshal([]byte(accessTokenResp), &TwitterAccessToken); err != nil {
		return nil, err
	}

	token := oauth2.Token{
		AccessToken: TwitterAccessToken.AccessToken,
		TokenType:   TwitterAccessToken.TokenType,
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

type TwitterUserInfo struct {
	Id       string   `json:"id"`       // The app user's App-Scoped User ID. This ID is unique to the app and cannot be used by other apps.
	Name     string   `json:"name"`     // The person's full name.
	UserName string   `json:"username"` // The person's name formatted to correctly handle Chinese, Japanese, or Korean ordering.
	Picture  struct { // The person's profile picture.
		Data struct { // This struct is different as https://developers.Twitter.com/docs/graph-api/reference/user/picture/
			Height       int    `json:"height"`
			IsSilhouette bool   `json:"is_silhouette"`
			Url          string `json:"url"`
			Width        int    `json:"width"`
		} `json:"data"`
	} `json:"picture"`
	Email string `json:"email"` // The User's primary email address listed on their profile. This field will not be returned if no valid email address is available.
}

// GetUserInfo use TwitterAccessToken gotten before return TwitterUserInfo
// get more detail via: https://developers.Twitter.com/docs/graph-api/reference/user
func (idp *TwitterIdProvider) GetUserInfo(token *oauth2.Token) (*UserInfo, error) {
	var TwitterUserInfo TwitterUserInfo
	// accessToken := token.AccessToken

	req, err := http.NewRequest("GET", "https://api.twitter.com/2/users/me", nil)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization", "Bearer "+token.AccessToken)
	// req.URL.Query().Set("user.fields", "profile_image_url")
	// userIdUrl := fmt.Sprintf("https://graph.Twitter.com/me?access_token=%s", accessToken)
	userIdResp, err := idp.GetUrlResp(req)
	if err != nil {
		return nil, err
	}
	empTwitterCheckToken := &TwitterCheckToken{}
	if err = json.Unmarshal([]byte(userIdResp), &empTwitterCheckToken); err != nil {
		return nil, err
	}
	TwitterUserInfo = empTwitterCheckToken.Data

	userInfo := UserInfo{
		Id:          TwitterUserInfo.Id,
		Username:    TwitterUserInfo.UserName,
		DisplayName: TwitterUserInfo.Name,
		Email:       TwitterUserInfo.Email,
		AvatarUrl:   TwitterUserInfo.Picture.Data.Url,
	}
	return &userInfo, nil
}

func (idp *TwitterIdProvider) GetUrlResp(url *http.Request) (string, error) {
	resp, err := idp.Client.Do(url)
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
