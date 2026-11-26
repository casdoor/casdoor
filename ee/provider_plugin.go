// Copyright 2026 Casdoor Inc. All Rights Reserved.
//
// This file is part of the commercial and enterprise editions of Casdoor and is licensed under the Casdoor Enterprise
// License, not the Apache License 2.0. You may not use this file except in compliance with that license, see
// ee/LICENSE.

package ee

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/casdoor/casdoor/action"
	"github.com/casdoor/casdoor/edition"
	"github.com/casdoor/casdoor/idp"
	"github.com/casdoor/casdoor/object"
	"golang.org/x/oauth2"
)

// Third-party sign-in and ID verification by actions of kind "idp" and "idv" installed from the Marketplace, such as
// the government and office IM platforms (Zhejiang Zhengding, WeLink, Yunzhijia) and the real-name checks of Tencent
// Cloud. They need an enterprise edition license, or the edition of a paid plugin made for the commercial edition.

const (
	idpTokenExtraKey = "casdoorPluginExtra"

	// PluginChallengeSessionKey keeps the challenge an idp plugin gave a browser until it signs in with it
	PluginChallengeSessionKey = "pluginChallenge"
	pluginChallengeLifetime   = 5 * time.Minute
)

type pluginIdProvider struct {
	module      *action.Module
	config      *action.Config
	redirectUri string
	challenge   string
}

// PluginChallenge is the challenge of an idp plugin that a browser has to answer, kept in its session
type PluginChallenge struct {
	Provider  string `json:"provider"`
	Challenge string `json:"challenge"`
	ExpiresAt int64  `json:"expiresAt"`
}

// requirePluginIdpEdition checks the license for an identity provider plugin: the edition a paid plugin was published
// for, the enterprise edition for any other
func requirePluginIdpEdition(provider *object.Provider) error {
	required := edition.Enterprise
	if provider.SubType != "" {
		pluginEdition, err := object.GetActionEdition(provider.Owner, provider.SubType)
		if err != nil {
			return err
		}
		if pluginEdition == edition.Commercial {
			required = edition.Commercial
		}
	}
	return edition.Require(required, edition.FeaturePluginIdp)
}

func getPluginIdpModule(provider *object.Provider) (*action.Module, error) {
	if err := requirePluginIdpEdition(provider); err != nil {
		return nil, err
	}
	if provider.SubType == "" {
		return nil, fmt.Errorf("choose the plugin of the provider: %s", provider.Name)
	}
	m, _, err := object.GetInstalledAction(provider.Owner, provider.SubType, "idp")
	return m, err
}

// GetPluginIdProvider is the idp.IdProvider of an OAuth provider of type Plugin. session is the challenge kept in the
// session of the browser for a plugin with a sign-in script, "" otherwise
func GetPluginIdProvider(provider *object.Provider, redirectUri string, session string) (idp.IdProvider, error) {
	m, err := getPluginIdpModule(provider)
	if err != nil {
		return nil, err
	}

	challenge := ""
	if session != "" {
		pending := &PluginChallenge{}
		if json.Unmarshal([]byte(session), pending) == nil && pending.Provider == provider.GetId() && time.Now().Unix() <= pending.ExpiresAt {
			challenge = pending.Challenge
		}
	}
	return &pluginIdProvider{module: m, config: provider.GetActionConfig(), redirectUri: redirectUri, challenge: challenge}, nil
}

// GetPluginChallenge asks the idp plugin of the provider for a challenge, such as the message for a wallet to sign, and
// returns it with what to keep in the session of the browser
func GetPluginChallenge(provider *object.Provider, application *object.Application, origin string, input map[string]string) (string, string, error) {
	m, err := getPluginIdpModule(provider)
	if err != nil {
		return "", "", err
	}

	name := application.DisplayName
	if name == "" {
		name = application.Name
	}
	challenge, err := action.GetIdpChallenge(m, provider.GetActionConfig(), origin, name, input)
	if err != nil {
		return "", "", err
	}
	if challenge == "" {
		return "", "", errors.New("the plugin returned no challenge")
	}

	pending := &PluginChallenge{Provider: provider.GetId(), Challenge: challenge, ExpiresAt: time.Now().Add(pluginChallengeLifetime).Unix()}
	data, err := json.Marshal(pending)
	if err != nil {
		return "", "", err
	}
	return challenge, string(data), nil
}

// GetPluginScript returns the sign-in script of the idp plugin of the provider
func GetPluginScript(provider *object.Provider) (string, error) {
	if provider.Type != object.ProviderTypePlugin || provider.Category != "OAuth" || provider.SubType == "" {
		return "", nil
	}
	if err := requirePluginIdpEdition(provider); err != nil {
		return "", err
	}
	return object.GetActionScript(provider.Owner, provider.SubType)
}

func (p *pluginIdProvider) SetHttpClient(client *http.Client) {}

func (p *pluginIdProvider) GetToken(code string) (*oauth2.Token, error) {
	token, err := action.GetIdpToken(p.module, p.config, code, p.redirectUri, p.challenge)
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
func GetPluginIdvProvider(provider *object.Provider, phone string) (*pluginIdvProvider, error) {
	if err := edition.Require(edition.Enterprise, edition.FeaturePluginIdv); err != nil {
		return nil, err
	}
	if provider.SubType == "" {
		return nil, fmt.Errorf("choose the plugin of the provider: %s", provider.Name)
	}
	m, _, err := object.GetInstalledAction(provider.Owner, provider.SubType, "idv")
	if err != nil {
		return nil, err
	}
	return &pluginIdvProvider{module: m, config: provider.GetActionConfig(), phone: phone}, nil
}

// GetReason is why the plugin said the identity didn't match
func (p *pluginIdvProvider) GetReason() string {
	return p.reason
}

// CheckProviderEdition refuses saving a provider of a kind that needs a license the instance doesn't have
func CheckProviderEdition(provider *object.Provider) error {
	if provider == nil || provider.Type != object.ProviderTypePlugin {
		return nil
	}
	if provider.Category == "OAuth" {
		return requirePluginIdpEdition(provider)
	}
	if provider.Category == "ID Verification" {
		return edition.Require(edition.Enterprise, edition.FeaturePluginIdv)
	}
	return nil
}
