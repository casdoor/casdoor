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
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/casdoor/casdoor/util"
	"github.com/mitchellh/mapstructure"
	"golang.org/x/oauth2"
)

type CustomIdProvider struct {
	Client *http.Client
	Config *oauth2.Config

	UserInfoURL string
	TokenURL    string
	AuthURL     string
	UserMapping map[string]string
	Scopes      []string
}

func NewCustomIdProvider(idpInfo *ProviderInfo, redirectUrl string) *CustomIdProvider {
	idp := &CustomIdProvider{}

	idp.Config = &oauth2.Config{
		ClientID:     idpInfo.ClientId,
		ClientSecret: idpInfo.ClientSecret,
		RedirectURL:  redirectUrl,
		Endpoint: oauth2.Endpoint{
			AuthURL:  idpInfo.AuthURL,
			TokenURL: idpInfo.TokenURL,
		},
	}
	idp.UserInfoURL = idpInfo.UserInfoURL
	idp.UserMapping = idpInfo.UserMapping

	return idp
}

func (idp *CustomIdProvider) SetHttpClient(client *http.Client) {
	idp.Client = client
}

func (idp *CustomIdProvider) GetToken(code string) (*oauth2.Token, error) {
	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, idp.Client)
	return idp.Config.Exchange(ctx, code)
}

func getNestedValue(data map[string]interface{}, path string) (interface{}, error) {
	keys := strings.Split(path, ".")
	var val interface{} = data

	for _, key := range keys {
		m, ok := val.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("path '%s' is not valid: %s is not a map", path, key)
		}

		val, ok = m[key]
		if !ok {
			return nil, fmt.Errorf("key '%s' not found in path '%s'", key, path)
		}
	}

	return val, nil
}

type CustomUserInfo struct {
	Id          string `mapstructure:"id"`
	Username    string `mapstructure:"username"`
	DisplayName string `mapstructure:"displayName"`
	Email       string `mapstructure:"email"`
	AvatarUrl   string `mapstructure:"avatarUrl"`
}

func (idp *CustomIdProvider) GetUserInfo(token *oauth2.Token) (*UserInfo, error) {
	accessToken := token.AccessToken
	request, err := http.NewRequest("GET", idp.UserInfoURL, nil)
	if err != nil {
		return nil, err
	}

	// add accessToken to request header
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	resp, err := idp.Client.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var dataMap map[string]interface{}
	err = json.Unmarshal(data, &dataMap)
	if err != nil {
		return nil, err
	}

	requiredFields := []string{"id", "username", "displayName"}
	for _, field := range requiredFields {
		_, ok := idp.UserMapping[field]
		if !ok {
			return nil, fmt.Errorf("cannot find %s in userMapping, please check your configuration in custom provider", field)
		}
	}

	// map user info
	for k, v := range idp.UserMapping {
		val, err := getNestedValue(dataMap, v)
		if err != nil {
			return nil, fmt.Errorf("cannot find %s in user from custom provider: %v", v, err)
		}
		dataMap[k] = val
	}

	// try to parse id to string
	id, err := util.ParseIdToString(dataMap["id"])
	if err != nil {
		return nil, err
	}
	dataMap["id"] = id

	customUserinfo := &CustomUserInfo{}
	err = mapstructure.Decode(dataMap, customUserinfo)
	if err != nil {
		return nil, err
	}

	userInfo := &UserInfo{
		Id:          customUserinfo.Id,
		Username:    customUserinfo.Username,
		DisplayName: customUserinfo.DisplayName,
		Email:       customUserinfo.Email,
		AvatarUrl:   customUserinfo.AvatarUrl,
	}
	return userInfo, nil
}
