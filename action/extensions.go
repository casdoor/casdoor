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
