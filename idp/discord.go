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
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"golang.org/x/oauth2"
)

type DiscordIdProvider struct {
	Client *http.Client
	Config *oauth2.Config
}

func NewDiscordIdProvider(clientId string, clientSecret string, redirectUrl string) *DiscordIdProvider {
	idp := &DiscordIdProvider{}

	config := idp.getConfig(clientId, clientSecret, redirectUrl)
	idp.Config = config

	return idp
}

func (idp *DiscordIdProvider) SetHttpClient(client *http.Client) {
	idp.Client = client
}

func (idp *DiscordIdProvider) getConfig(clientId string, clientSecret string, redirectUrl string) *oauth2.Config {
	endpoint := oauth2.Endpoint{
		TokenURL:  "https://discord.com/api/oauth2/token",
		AuthURL:   "https://discord.com/api/oauth2/authorize",
		AuthStyle: oauth2.AuthStyleInParams,
	}

	config := &oauth2.Config{
		Scopes: []string{"openid guilds"},

		Endpoint:     endpoint,
		ClientID:     clientId,
		ClientSecret: clientSecret,
		RedirectURL:  redirectUrl,
	}

	return config
}

type DiscordAccessToken struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
}

func (idp *DiscordIdProvider) GetToken(code string) (*oauth2.Token, error) {
	params := url.Values{}
	params.Add("grant_type", "authorization_code")
	params.Add("client_id", idp.Config.ClientID)
	params.Add("client_secret", idp.Config.ClientSecret)
	params.Add("code", code)
	params.Add("redirect_uri", idp.Config.RedirectURL)

	accessTokenUrl := fmt.Sprintf("%s?%s", idp.Config.Endpoint.TokenURL, params.Encode())
	bs, _ := json.Marshal(params.Encode())
	req, _ := http.NewRequest("POST", accessTokenUrl, strings.NewReader(string(bs)))
	//req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.101 Safari/537.36")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	rbs, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	tokenResp := DiscordAccessToken{}
	if err = json.Unmarshal(rbs, &tokenResp); err != nil {
		return nil, err
	}

	token := &oauth2.Token{
		AccessToken:  tokenResp.AccessToken,
		TokenType:    tokenResp.TokenType,
		RefreshToken: tokenResp.RefreshToken,
		Expiry:       time.Unix(time.Now().Unix()+int64(tokenResp.ExpiresIn), 0),
	}

	return token, nil
}

func (idp *DiscordIdProvider) GetUserInfo(token *oauth2.Token) (*UserInfo, error) {
	session, err := discordgo.New("Bearer " + token.AccessToken)
	if err != nil {
		return nil, err
	}
	defer session.Close()

	user, err := session.User("@me")
	if err != nil {
		return nil, err
	}

	userInfo := UserInfo{
		Id:          user.ID,
		Username:    user.Username,
		DisplayName: user.Username,
		Email:       user.Email,
		AvatarUrl:   user.AvatarURL("128"),
		Extra:       map[string]string{},
	}

	guilds, err := session.UserGuilds(100, "", "")
	if err != nil {
		return nil, err
	}
	for guildId := range guilds {
		guild, err := session.Guild(strconv.Itoa(guildId))
		if err != nil {
			continue
		}
		userInfo.Extra[guild.ID] = guild.Name
	}

	return &userInfo, nil
}

func (idp *DiscordIdProvider) GetUrlResp(url string) (string, error) {
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
