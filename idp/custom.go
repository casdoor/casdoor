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
	"strconv"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

type CustomIdProvider struct {
	Client *http.Client
	Config *oauth2.Config

	UserInfoURL  string
	TokenURL     string
	AuthURL      string
	Issuer       string
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

// oidcProtocolClaims only matter to the OIDC protocol itself, they are not kept in the user's extra info
var oidcProtocolClaims = map[string]bool{
	"aud":       true,
	"at_hash":   true,
	"auth_time": true,
	"azp":       true,
	"c_hash":    true,
	"exp":       true,
	"iat":       true,
	"jti":       true,
	"nbf":       true,
	"nonce":     true,
	"s_hash":    true,
	"sid":       true,
}

// oidcClaimFallbacks are used when userMapping doesn't cover the user field
var oidcClaimFallbacks = map[string][]string{
	"id":          {"sub"},
	"username":    {"preferred_username", "name", "sub"},
	"displayName": {"name", "preferred_username"},
	"email":       {"email"},
	"phone":       {"phone_number"},
	"avatarUrl":   {"picture"},
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

func validateIdTokenClaims(claims map[string]interface{}, clientId string, issuer string) error {
	if sub, _ := claims["sub"].(string); sub == "" {
		return fmt.Errorf("id_token missing required claim: sub")
	}

	if issuer != "" {
		iss, _ := claims["iss"].(string)
		if strings.TrimSuffix(iss, "/") != issuer {
			return fmt.Errorf("id_token iss: %s doesn't match the issuer: %s", iss, issuer)
		}
	}

	if exp, ok := claims["exp"].(float64); ok {
		if time.Now().After(time.Unix(int64(exp), 0)) {
			return fmt.Errorf("id_token has expired")
		}
	}

	if clientId == "" {
		return nil
	}

	switch aud := claims["aud"].(type) {
	case string:
		if aud != clientId {
			return fmt.Errorf("id_token aud: %s doesn't match the client ID", aud)
		}
	case []interface{}:
		for _, item := range aud {
			if s, ok := item.(string); ok && s == clientId {
				return nil
			}
		}
		return fmt.Errorf("id_token aud doesn't contain the client ID")
	}
	return nil
}

func toStringValue(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case bool:
		return strconv.FormatBool(v)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case int:
		return strconv.Itoa(v)
	case int64:
		return strconv.FormatInt(v, 10)
	default:
		return ""
	}
}

// flattenClaims turns the merged OIDC claims into a flat string map, so that they can be stored in the user's extra info
func flattenClaims(claims map[string]interface{}) map[string]string {
	res := map[string]string{}
	for k, v := range claims {
		if v == nil || oidcProtocolClaims[k] {
			continue
		}

		s := toStringValue(v)
		if s == "" {
			data, err := json.Marshal(v)
			if err != nil {
				continue
			}
			s = string(data)
		}
		res[k] = s
	}
	return res
}

func (idp *CustomIdProvider) getClaimsFromUserInfoUrl(token *oauth2.Token) (map[string]interface{}, error) {
	request, err := http.NewRequest("GET", idp.UserInfoURL, nil)
	if err != nil {
		return nil, err
	}

	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))
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
	return dataMap, nil
}

func (idp *CustomIdProvider) mapUserInfo(claims map[string]interface{}) (*UserInfo, error) {
	dataMap := map[string]interface{}{}
	for k, v := range claims {
		dataMap[k] = v
	}

	for k, v := range idp.UserMapping {
		val, err := getNestedValue(claims, v)
		if err != nil {
			// only the ID is mandatory, the other claims may be absent for some users
			if k == "id" {
				return nil, fmt.Errorf("cannot find %s in user from custom provider: %v", v, err)
			}
			continue
		}
		dataMap[k] = val
	}

	getField := func(field string) string {
		if s := toStringValue(dataMap[field]); s != "" {
			return s
		}
		for _, claimName := range oidcClaimFallbacks[field] {
			if s := toStringValue(claims[claimName]); s != "" {
				return s
			}
		}
		return ""
	}

	userInfo := &UserInfo{
		Id:          getField("id"),
		Username:    getField("username"),
		DisplayName: getField("displayName"),
		Email:       getField("email"),
		Phone:       getField("phone"),
		AvatarUrl:   getField("avatarUrl"),
	}
	if userInfo.Id == "" {
		return nil, fmt.Errorf("cannot get the user ID from custom provider, please check the \"id\" field in userMapping")
	}
	if userInfo.Username == "" {
		userInfo.Username = userInfo.Id
	}
	return userInfo, nil
}

func (idp *CustomIdProvider) GetUserInfo(token *oauth2.Token) (*UserInfo, error) {
	claims := map[string]interface{}{}

	idTokenClaims, idTokenErr := parseIdTokenClaims(token)
	if idTokenErr == nil {
		err := validateIdTokenClaims(idTokenClaims, idp.Config.ClientID, idp.Issuer)
		if err != nil {
			return nil, err
		}
		for k, v := range idTokenClaims {
			claims[k] = v
		}
	} else if idp.Issuer != "" {
		return nil, idTokenErr
	} else if idp.UserInfoURL == "" {
		return nil, fmt.Errorf("UserInfoURL is empty and %v", idTokenErr)
	}

	// the UserInfo endpoint is more authoritative, so it overrides the id_token claims
	if idp.UserInfoURL != "" {
		userInfoClaims, err := idp.getClaimsFromUserInfoUrl(token)
		if err != nil {
			return nil, err
		}
		for k, v := range userInfoClaims {
			claims[k] = v
		}
	}

	userInfo, err := idp.mapUserInfo(claims)
	if err != nil {
		return nil, err
	}

	userInfo.Extra = flattenClaims(claims)
	return userInfo, nil
}
