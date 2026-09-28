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
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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
	idp.Client = newInsecureTlsHttpClient(client)
}

func newInsecureTlsHttpClient(client *http.Client) *http.Client {
	transport, ok := client.Transport.(*http.Transport)
	if !ok || transport == nil {
		transport = http.DefaultTransport.(*http.Transport)
	}
	transport = transport.Clone()
	transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	res := *client
	res.Transport = transport
	return &res
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

// GetToken
// get more detail via: https://docs.microsoft.com/en-us/windows-server/identity/ad-fs/overview/ad-fs-openid-connect-oauth-flows-scenarios#request-an-access-token
func (idp *AdfsIdProvider) GetToken(code string) (*oauth2.Token, error) {
	payload := url.Values{}
	payload.Set("code", code)
	payload.Set("grant_type", "authorization_code")
	payload.Set("client_id", idp.Config.ClientID)
	payload.Set("client_secret", idp.Config.ClientSecret)
	payload.Set("redirect_uri", idp.Config.RedirectURL)

	resp, err := idp.Client.PostForm(idp.Config.Endpoint.TokenURL, payload)
	if err != nil {
		return nil, err
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	pToken := &AdfsToken{}
	err = json.Unmarshal(data, pToken)
	if err != nil {
		return nil, err
	}
	if pToken.ErrMsg != "" {
		return nil, errors.New(pToken.ErrMsg)
	}

	token := &oauth2.Token{
		AccessToken: pToken.IdToken,
		Expiry:      time.Unix(time.Now().Unix()+int64(pToken.ExpiresIn), 0),
	}
	return token, nil
}

// GetUserInfo
// Since the userinfo endpoint of ADFS only returns sub,
// the id_token is used to resolve the userinfo
func (idp *AdfsIdProvider) GetUserInfo(token *oauth2.Token) (*UserInfo, error) {
	resp, err := idp.Client.Get(fmt.Sprintf("%s/adfs/discovery/keys", idp.Host))
	if err != nil {
		return nil, err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var respKeys struct {
		Keys []interface{} `json:"keys"`
	}

	if err := json.Unmarshal(body, &respKeys); err != nil {
		return nil, err
	}
	if len(respKeys.Keys) == 0 {
		return nil, errors.New("the ADFS discovery keys are empty")
	}

	respKey, err := json.Marshal(&(respKeys.Keys[0]))
	if err != nil {
		return nil, err
	}

	keyset, err := jwk.ParseKey(respKey)
	if err != nil {
		return nil, err
	}

	tokenSrc := []byte(token.AccessToken)
	publicKey, err := keyset.PublicKey()
	if err != nil {
		return nil, err
	}
	idToken, err := jwt.Parse(tokenSrc, jwt.WithVerify(jwa.RS256, publicKey))
	if err != nil {
		return nil, err
	}

	sid := getAdfsClaim(idToken, "sid")
	upn := getAdfsClaim(idToken, "upn")
	name := getAdfsClaim(idToken, "unique_name")
	if sid == "" {
		return nil, errors.New("the ADFS id_token has no sid claim")
	}

	userinfo := &UserInfo{
		Id:          sid,
		Username:    name,
		DisplayName: name,
		Email:       upn,
		// the UPN is the directory account itself
		EmailVerified: upn != "",
	}
	return userinfo, nil
}

func getAdfsClaim(token jwt.Token, name string) string {
	value, ok := token.Get(name)
	if !ok {
		return ""
	}
	res, _ := value.(string)
	return res
}
