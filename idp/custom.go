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
	"encoding/base64"
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

	UserInfoURL  string
	TokenURL     string
	AuthURL      string
	UserMapping  map[string]string
	Scopes       []string
	CodeVerifier string
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

	idp.CodeVerifier = idpInfo.CodeVerifier
	return idp
}

func (idp *CustomIdProvider) SetHttpClient(client *http.Client) {
	idp.Client = client
}

func (idp *CustomIdProvider) GetToken(code string) (*oauth2.Token, error) {
	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, idp.Client)
	var oauth2Opts []oauth2.AuthCodeOption
	if idp.CodeVerifier != "" {
		oauth2Opts = append(oauth2Opts, oauth2.VerifierOption(idp.CodeVerifier))
	}
	return idp.Config.Exchange(ctx, code, oauth2Opts...)
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
	Phone       string `mapstructure:"phone"`
}

func parseIdTokenClaims(token *oauth2.Token) (map[string]interface{}, error) {
	rawIdToken, ok := token.Extra("id_token").(string)
	if !ok || rawIdToken == "" {
		return nil, fmt.Errorf("id_token not found in token response")
	}
	parts := strings.Split(rawIdToken, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid id_token format")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("failed to decode id_token payload: %v", err)
	}
	var claims map[string]interface{}
	if err = json.Unmarshal(payload, &claims); err != nil {
		return nil, fmt.Errorf("failed to parse id_token claims: %v", err)
	}
	return claims, nil
}

func userInfoFromIdTokenClaims(claims map[string]interface{}) (*UserInfo, error) {
	getString := func(key string) string {
		if v, ok := claims[key]; ok {
			if s, ok := v.(string); ok {
				return s
			}
		}
		return ""
	}

	sub := getString("sub")
	if sub == "" {
		return nil, fmt.Errorf("id_token missing required claim: sub")
	}

	username := getString("preferred_username")
	if username == "" {
		username = getString("name")
	}
	if username == "" {
		username = sub
	}

	return &UserInfo{
		Id:          sub,
		Username:    username,
		DisplayName: getString("name"),
		Email:       getString("email"),
		Phone:       getString("phone_number"),
		AvatarUrl:   getString("picture"),
	}, nil
}

func (idp *CustomIdProvider) GetUserInfo(token *oauth2.Token) (*UserInfo, error) {
	// When no UserInfo URL is configured, fall back to id_token claims (e.g. Telegram OIDC).
	if idp.UserInfoURL == "" {
		claims, err := parseIdTokenClaims(token)
		if err != nil {
			return nil, fmt.Errorf("UserInfoURL is empty and %v", err)
		}
		return userInfoFromIdTokenClaims(claims)
	}

	accessToken := token.AccessToken
	request, err := http.NewRequest("GET", idp.UserInfoURL, nil)
	if err != nil {
		return nil, err
	}

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
		Phone:       customUserinfo.Phone,
		AvatarUrl:   customUserinfo.AvatarUrl,
	}
	return userInfo, nil
}
