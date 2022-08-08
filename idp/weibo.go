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
	"strconv"
	"time"

	"golang.org/x/oauth2"
)

type WeiBoIdProvider struct {
	Client *http.Client
	Config *oauth2.Config
}

func NewWeiBoIdProvider(clientId string, clientSecret string, redirectUrl string) *WeiBoIdProvider {
	idp := &WeiBoIdProvider{}

	config := idp.getConfig(clientId, clientSecret, redirectUrl)
	idp.Config = config

	return idp
}

func (idp *WeiBoIdProvider) SetHttpClient(client *http.Client) {
	idp.Client = client
}

// getConfig return a point of Config, which describes a typical 3-legged OAuth2 flow
func (idp *WeiBoIdProvider) getConfig(clientId string, clientSecret string, redirectUrl string) *oauth2.Config {
	endpoint := oauth2.Endpoint{
		TokenURL: "https://api.weibo.com/oauth2/access_token",
	}

	config := &oauth2.Config{
		Scopes:       []string{""},
		Endpoint:     endpoint,
		ClientID:     clientId,
		ClientSecret: clientSecret,
		RedirectURL:  redirectUrl,
	}

	return config
}

type WeiboAccessToken struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	RemindIn    string `json:"remind_in"` // This parameter is about to be obsolete, developers please use expires_in
	Uid         string `json:"uid"`
}

// GetToken use code get access_token (*operation of getting code ought to be done in front)
// get more detail via: https://developers.weixin.qq.com/doc/oplatform/Website_App/WeChat_Login/Wechat_Login.html
func (idp *WeiBoIdProvider) GetToken(code string) (*oauth2.Token, error) {
	params := url.Values{}
	params.Add("grant_type", "authorization_code")
	params.Add("client_id", idp.Config.ClientID)
	params.Add("client_secret", idp.Config.ClientSecret)
	params.Add("code", code)
	params.Add("redirect_uri", idp.Config.RedirectURL)

	// accessTokenUrl := fmt.Sprintf("%s?%s", idp.Config.Endpoint.TokenURL, params.Encode())
	resp, err := idp.Client.PostForm(idp.Config.Endpoint.TokenURL, params)
	// resp, err := idp.GetUrlResp(accessTokenUrl)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(resp.Body)
	bs, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var weiboAccessToken WeiboAccessToken
	if err = json.Unmarshal(bs, &weiboAccessToken); err != nil {
		return nil, err
	}

	token := oauth2.Token{
		AccessToken: weiboAccessToken.AccessToken,
		TokenType:   "WeiboAccessToken",
		Expiry:      time.Unix(time.Now().Unix()+int64(weiboAccessToken.ExpiresIn), 0),
	}

	idp.Config.Scopes[0] = weiboAccessToken.Uid
	return &token, nil
}

/*
{
	"id": 1404376560,
	"screen_name": "zaku",
	"name": "zaku",
	"province": "11",
	"city": "5",
	"location": "北京 朝阳区",
	"description": "人生五十年，乃如梦如幻；有生斯有死，壮士复何憾。",
	"url": "http://blog.sina.com.cn/zaku",
	"profile_image_url": "http://tp1.sinaimg.cn/1404376560/50/0/1",
	"domain": "zaku",
	"gender": "m",
	"followers_count": 1204,
	"friends_count": 447,
	"statuses_count": 2908,
	"favourites_count": 0,
	"created_at": "Fri Aug 28 00:00:00 +0800 2009",
	"following": false,
	"allow_all_act_msg": false,
	"geo_enabled": true,
	"verified": false,
	"status": {
		"created_at": "Tue May 24 18:04:53 +0800 2011",
		"id": 11142488790,
		"text": "我的相机到了。",
		"source": "<a href="http://weibo.com" rel="nofollow">新浪微博</a>",
		"favorited": false,
		"truncated": false,
		"in_reply_to_status_id": "",
		"in_reply_to_user_id": "",
		"in_reply_to_screen_name": "",
		"geo": null,
		"mid": "5610221544300749636",
		"annotations": [],
		"reposts_count": 5,
		"comments_count": 8
	},
	"allow_all_comment": true,
	"avatar_large": "http://tp1.sinaimg.cn/1404376560/180/0/1",
	"verified_reason": "",
	"follow_me": false,
	"online_status": 0,
	"bi_followers_count": 215
}
*/

type WeiboUserinfo struct {
	Id              int    `json:"id"`
	ScreenName      string `json:"screen_name"`
	Name            string `json:"name"`
	Province        string `json:"province"`
	City            string `json:"city"`
	Location        string `json:"location"`
	Description     string `json:"description"`
	Url             string `json:"url"`
	ProfileImageUrl string `json:"profile_image_url"`
	Domain          string `json:"domain"`
	Gender          string `json:"gender"`
	FollowersCount  int    `json:"followers_count"`
	FriendsCount    int    `json:"friends_count"`
	StatusesCount   int    `json:"statuses_count"`
	FavouritesCount int    `json:"favourites_count"`
	CreatedAt       string `json:"created_at"`
	Following       bool   `json:"following"`
	AllowAllActMsg  bool   `json:"allow_all_act_msg"`
	GeoEnabled      bool   `json:"geo_enabled"`
	Verified        bool   `json:"verified"`
	Status          struct {
		CreatedAt           string        `json:"created_at"`
		Id                  int64         `json:"id"`
		Text                string        `json:"text"`
		Source              string        `json:"source"`
		Favorited           bool          `json:"favorited"`
		Truncated           bool          `json:"truncated"`
		InReplyToStatusId   string        `json:"in_reply_to_status_id"`
		InReplyToUserId     string        `json:"in_reply_to_user_id"`
		InReplyToScreenName string        `json:"in_reply_to_screen_name"`
		Geo                 interface{}   `json:"geo"`
		Mid                 string        `json:"mid"`
		Annotations         []interface{} `json:"annotations"`
		RepostsCount        int           `json:"reposts_count"`
		CommentsCount       int           `json:"comments_count"`
	} `json:"status"`
	AllowAllComment  bool   `json:"allow_all_comment"`
	AvatarLarge      string `json:"avatar_large"`
	VerifiedReason   string `json:"verified_reason"`
	FollowMe         bool   `json:"follow_me"`
	OnlineStatus     int    `json:"online_status"`
	BiFollowersCount int    `json:"bi_followers_count"`
}

// GetUserInfo use WeiboAccessToken gotten before return UserInfo
// get more detail via: https://open.weibo.com/wiki/2/users/show
func (idp *WeiBoIdProvider) GetUserInfo(token *oauth2.Token) (*UserInfo, error) {
	var weiboUserInfo WeiboUserinfo
	accessToken := token.AccessToken
	uid := idp.Config.Scopes[0]
	id, _ := strconv.Atoi(uid)

	userInfoUrl := fmt.Sprintf("https://api.weibo.com/2/users/show.json?access_token=%s&uid=%d", accessToken, id)
	resp, err := idp.GetUrlResp(userInfoUrl)
	if err != nil {
		return nil, err
	}
	if err = json.Unmarshal([]byte(resp), &weiboUserInfo); err != nil {
		return nil, err
	}

	// weibo user email need to get separately through this url, need user authorization.
	e := struct {
		Email string `json:"email"`
	}{}
	emailUrl := fmt.Sprintf("https://api.weibo.com/2/account/profile/email.json?access_token=%s", accessToken)
	resp, err = idp.GetUrlResp(emailUrl)
	if err != nil {
		return nil, err
	}
	if err = json.Unmarshal([]byte(resp), &e); err != nil {
		return nil, err
	}

	userInfo := UserInfo{
		Id:          strconv.Itoa(weiboUserInfo.Id),
		Username:    weiboUserInfo.Name,
		DisplayName: weiboUserInfo.Name,
		AvatarUrl:   weiboUserInfo.AvatarLarge,
		Email:       e.Email,
	}
	return &userInfo, nil
}

func (idp *WeiBoIdProvider) GetUrlResp(url string) (string, error) {
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
