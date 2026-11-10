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

package action

import "strings"

// Module is an installed action: its name in the organization, the WebAssembly module and the hosts it may call
type Module struct {
	Name   string
	Module []byte
	Hosts  []string
}

type CheckPasswordResp struct {
	Ok     bool   `json:"ok"`
	Reason string `json:"reason"`
}

// CheckPassword runs an action of kind "password-check"
func CheckPassword(m *Module, config *Config, password string, username string, isSignin bool) (*CheckPasswordResp, error) {
	resp := &CheckPasswordResp{}
	err := Call(m.Name, m.Module, m.Hosts, "password.check", config, map[string]interface{}{
		"password": password,
		"username": username,
		"isSignin": isSignin,
	}, resp)
	return resp, err
}

// VerifyPassword runs an action of kind "password-migration"
func VerifyPassword(m *Module, config *Config, username string, password string) (bool, error) {
	resp := struct {
		Valid bool `json:"valid"`
	}{}
	err := Call(m.Name, m.Module, m.Hosts, "password.verify", config, map[string]string{
		"username": username,
		"password": password,
	}, &resp)
	return resp.Valid, err
}

// SendPush runs an action of kind "push"
func SendPush(m *Module, config *Config, receiver string, title string, description string, url string) error {
	return Call(m.Name, m.Module, m.Hosts, "push.send", config, map[string]string{
		"receiver":    receiver,
		"title":       title,
		"description": description,
		"url":         url,
	}, nil)
}

type SyncerUser struct {
	Id            string            `json:"id"`
	Name          string            `json:"name"`
	DisplayName   string            `json:"displayName"`
	Email         string            `json:"email"`
	EmailVerified bool              `json:"emailVerified"`
	Phone         string            `json:"phone"`
	CountryCode   string            `json:"countryCode"`
	Avatar        string            `json:"avatar"`
	Gender        string            `json:"gender"`
	Affiliation   string            `json:"affiliation"`
	Title         string            `json:"title"`
	IsForbidden   bool              `json:"isForbidden"`
	CreatedTime   string            `json:"createdTime"`
	Properties    map[string]string `json:"properties"`
	// the IDs of the user's departments, the Id of a SyncerGroup
	Groups []string `json:"groups"`
}

// SyncerGroup is a department of the external system, ParentId is empty for a top department
type SyncerGroup struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	ParentId    string `json:"parentId"`
}

// IsNotImplemented tells the error of calling a method that the action doesn't have, such as an optional one
func IsNotImplemented(err error) bool {
	return err != nil && strings.Contains(err.Error(), "the action does not implement")
}

// ListSyncerGroups runs the optional method "syncer.listGroups" of an action of kind "syncer"
func ListSyncerGroups(m *Module, config *Config) ([]*SyncerGroup, error) {
	groups := []*SyncerGroup{}
	err := Call(m.Name, m.Module, m.Hosts, "syncer.listGroups", config, map[string]string{}, &groups)
	return groups, err
}

type IdpToken struct {
	AccessToken  string            `json:"accessToken"`
	RefreshToken string            `json:"refreshToken"`
	ExpiresIn    int               `json:"expiresIn"`
	Extra        map[string]string `json:"extra"`
}

type IdpUser struct {
	Id          string            `json:"id"`
	Username    string            `json:"username"`
	DisplayName string            `json:"displayName"`
	Email       string            `json:"email"`
	Phone       string            `json:"phone"`
	CountryCode string            `json:"countryCode"`
	Avatar      string            `json:"avatar"`
	Extra       map[string]string `json:"extra"`
}

// GetIdpToken runs "idp.getToken" of an action of kind "idp", which exchanges the authorization code
func GetIdpToken(m *Module, config *Config, code string, redirectUri string) (*IdpToken, error) {
	token := &IdpToken{}
	err := Call(m.Name, m.Module, m.Hosts, "idp.getToken", config, map[string]string{"code": code, "redirectUri": redirectUri}, token)
	return token, err
}

// GetIdpUserInfo runs "idp.getUserInfo" of an action of kind "idp"
func GetIdpUserInfo(m *Module, config *Config, token *IdpToken) (*IdpUser, error) {
	user := &IdpUser{}
	err := Call(m.Name, m.Module, m.Hosts, "idp.getUserInfo", config, map[string]interface{}{"token": token}, user)
	return user, err
}

type IdvResult struct {
	Ok     bool   `json:"ok"`
	Reason string `json:"reason"`
}

// VerifyIdentity runs "idv.verify" of an action of kind "idv"
func VerifyIdentity(m *Module, config *Config, idCardType string, idCard string, realName string, phone string) (*IdvResult, error) {
	res := &IdvResult{}
	err := Call(m.Name, m.Module, m.Hosts, "idv.verify", config, map[string]string{
		"idCardType": idCardType,
		"idCard":     idCard,
		"realName":   realName,
		"phone":      phone,
	}, res)
	return res, err
}

// ListSyncerUsers runs an action of kind "syncer"
func ListSyncerUsers(m *Module, config *Config) ([]*SyncerUser, error) {
	users := []*SyncerUser{}
	err := Call(m.Name, m.Module, m.Hosts, "syncer.listUsers", config, map[string]string{}, &users)
	return users, err
}

// TestSyncer runs the connection test of an action of kind "syncer"
func TestSyncer(m *Module, config *Config) error {
	return Call(m.Name, m.Module, m.Hosts, "syncer.test", config, map[string]string{}, nil)
}
