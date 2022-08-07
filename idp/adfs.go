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
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/lestrrat-go/jwx/jwt"
	"golang.org/x/oauth2"
)

type AdfsIdProvider struct {
	Client *http.Client
	Config *oauth2.Config
	Host   string
}

func NewAdfsIdProvider(clientId string, clientSecret string, redirectUrl string, hostUrl string) *AdfsIdProvider {
	idp := &AdfsIdProvider{}

	config := idp.getConfig(hostUrl)
	config.ClientID = clientId
	config.ClientSecret = clientSecret
	config.RedirectURL = redirectUrl
	idp.Config = config
	idp.Host = hostUrl
	return idp
}

func (idp *AdfsIdProvider) SetHttpClient(client *http.Client) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	idp.Client = client
	idp.Client.Transport = tr
}

func (idp *AdfsIdProvider) getConfig(hostUrl string) *oauth2.Config {
	endpoint := oauth2.Endpoint{
		AuthURL:  fmt.Sprintf("%s/adfs/oauth2/authorize", hostUrl),
		TokenURL: fmt.Sprintf("%s/adfs/oauth2/token", hostUrl),
	}

	config := &oauth2.Config{
		Endpoint: endpoint,
	}

	return config
}

type AdfsToken struct {
	IdToken   string `json:"id_token"`
	ExpiresIn int    `json:"expires_in"`
	ErrMsg    string `json:"error_description"`
}

// get more detail via: https://docs.microsoft.com/en-us/windows-server/identity/ad-fs/overview/ad-fs-openid-connect-oauth-flows-scenarios#request-an-access-token
func (idp *AdfsIdProvider) GetToken(code string) (*oauth2.Token, error) {
	payload := url.Values{}
	payload.Set("code", code)
	payload.Set("grant_type", "authorization_code")
	payload.Set("client_id", idp.Config.ClientID)
	payload.Set("redirect_uri", idp.Config.RedirectURL)
	resp, err := idp.Client.PostForm(idp.Config.Endpoint.TokenURL, payload)
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	pToken := &AdfsToken{}
	err = json.Unmarshal(data, pToken)
	if err != nil {
		return nil, fmt.Errorf("fail to unmarshal token response: %s", err.Error())
	}
	if pToken.ErrMsg != "" {
		return nil, fmt.Errorf("pToken.Errmsg = %s", pToken.ErrMsg)
	}

	token := &oauth2.Token{
		AccessToken: pToken.IdToken,
		Expiry:      time.Unix(time.Now().Unix()+int64(pToken.ExpiresIn), 0),
	}
	return token, nil
}

// Since the userinfo endpoint of ADFS only returns sub,
// the id_token is used to resolve the userinfo
func (idp *AdfsIdProvider) GetUserInfo(token *oauth2.Token) (*UserInfo, error) {
	resp, err := idp.Client.Get(fmt.Sprintf("%s/adfs/discovery/keys", idp.Host))
	if err != nil {
		return nil, err
	}
	keyset, err := jwk.Parse(resp.Body)
	if err != nil {
		return nil, err
	}
	tokenSrc := []byte(token.AccessToken)
	publicKey, _ := keyset.Keys[0].Materialize()
	id_token, _ := jwt.Parse(bytes.NewReader(tokenSrc), jwt.WithVerify(jwa.RS256, publicKey))
	sid, _ := id_token.Get("sid")
	upn, _ := id_token.Get("upn")
	name, _ := id_token.Get("unique_name")
	userinfo := &UserInfo{
		Id:          sid.(string),
		Username:    name.(string),
		DisplayName: name.(string),
		Email:       upn.(string),
	}
	return userinfo, nil
}
