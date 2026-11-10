// Copyright 2026 The Casdoor Authors. All Rights Reserved.
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

package object

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/casdoor/casdoor/action"
	"github.com/casdoor/casdoor/edition"
	"github.com/casdoor/casdoor/idp"
	"golang.org/x/oauth2"
)

// Third-party sign-in and ID verification by actions of kind "idp" and "idv" installed from the Marketplace, such as
// the government and office IM platforms (Zhejiang Zhengding, WeLink, Yunzhijia) and the real-name checks of Tencent
// Cloud. They need an enterprise edition license.

const idpTokenExtraKey = "casdoorPluginExtra"

type pluginIdProvider struct {
	module      *action.Module
	config      *action.Config
	redirectUri string
}

// GetPluginIdProvider is the idp.IdProvider of an OAuth provider of type Plugin
func GetPluginIdProvider(provider *Provider, redirectUri string) (idp.IdProvider, error) {
	if err := edition.Require(edition.Enterprise, edition.FeaturePluginIdp); err != nil {
		return nil, err
	}
	if provider.SubType == "" {
		return nil, fmt.Errorf("choose the plugin of the provider: %s", provider.Name)
	}
	m, _, err := getInstalledAction(provider.Owner, provider.SubType, "idp")
	if err != nil {
		return nil, err
	}
	return &pluginIdProvider{module: m, config: provider.getActionConfig(), redirectUri: redirectUri}, nil
}

func (p *pluginIdProvider) SetHttpClient(client *http.Client) {}

func (p *pluginIdProvider) GetToken(code string) (*oauth2.Token, error) {
	token, err := action.GetIdpToken(p.module, p.config, code, p.redirectUri)
	if err != nil {
		return nil, err
	}
	if token.AccessToken == "" {
		return nil, errors.New("the plugin returned no access token")
	}

	res := &oauth2.Token{AccessToken: token.AccessToken, RefreshToken: token.RefreshToken, TokenType: "Bearer"}
	if token.ExpiresIn > 0 {
		res.Expiry = time.Now().Add(time.Duration(token.ExpiresIn) * time.Second)
	}
	return res.WithExtra(map[string]interface{}{idpTokenExtraKey: token.Extra}), nil
}

func (p *pluginIdProvider) GetUserInfo(token *oauth2.Token) (*idp.UserInfo, error) {
	pluginToken := &action.IdpToken{AccessToken: token.AccessToken, RefreshToken: token.RefreshToken}
	if !token.Expiry.IsZero() {
		pluginToken.ExpiresIn = int(time.Until(token.Expiry).Seconds())
	}
	if extra, ok := token.Extra(idpTokenExtraKey).(map[string]string); ok {
		pluginToken.Extra = extra
	}

	user, err := action.GetIdpUserInfo(p.module, p.config, pluginToken)
	if err != nil {
		return nil, err
	}
	if user.Id == "" {
		return nil, errors.New("the plugin returned no user ID")
	}

	username := user.Username
	if username == "" {
		username = user.Id
	}
	return &idp.UserInfo{
		Id:          user.Id,
		Username:    username,
		DisplayName: user.DisplayName,
		Email:       user.Email,
		Phone:       user.Phone,
		CountryCode: user.CountryCode,
		AvatarUrl:   user.Avatar,
		Extra:       user.Extra,
	}, nil
}

type pluginIdvProvider struct {
	module *action.Module
	config *action.Config
	phone  string
	reason string
}

func (p *pluginIdvProvider) VerifyIdentity(idCardType string, idCard string, realName string) (bool, error) {
	res, err := action.VerifyIdentity(p.module, p.config, idCardType, idCard, realName, p.phone)
	if err != nil {
		return false, err
	}
	p.reason = res.Reason
	return res.Ok, nil
}

// GetPluginIdvProvider is the idv.IdvProvider of an ID Verification provider of type Plugin, the phone is sent too for
// the three-factor check of the carriers
func GetPluginIdvProvider(provider *Provider, phone string) (*pluginIdvProvider, error) {
	if err := edition.Require(edition.Enterprise, edition.FeaturePluginIdv); err != nil {
		return nil, err
	}
	if provider.SubType == "" {
		return nil, fmt.Errorf("choose the plugin of the provider: %s", provider.Name)
	}
	m, _, err := getInstalledAction(provider.Owner, provider.SubType, "idv")
	if err != nil {
		return nil, err
	}
	return &pluginIdvProvider{module: m, config: provider.getActionConfig(), phone: phone}, nil
}

// GetReason is why the plugin said the identity didn't match
func (p *pluginIdvProvider) GetReason() string {
	return p.reason
}

// CheckProviderEdition refuses saving a provider of a kind that needs a license the instance doesn't have
func CheckProviderEdition(provider *Provider) error {
	if provider == nil || provider.Type != ProviderTypePlugin {
		return nil
	}
	if provider.Category == "OAuth" {
		return edition.Require(edition.Enterprise, edition.FeaturePluginIdp)
	}
	if provider.Category == "ID Verification" {
		return edition.Require(edition.Enterprise, edition.FeaturePluginIdv)
	}
	return nil
}
