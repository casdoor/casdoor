// Copyright 2021 The casbin Authors. All Rights Reserved.
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
	"net/http"

	"golang.org/x/oauth2"
)

type IdProvider interface {
	GetConfig() *oauth2.Config
	GetUserInfo(httpClient *http.Client, token *oauth2.Token) (string, string, string, error)
}

func GetIdProvider(providerType string, clientId string) IdProvider {
	if providerType == "github" {
		return &GithubIdProvider{}
	} else if providerType == "google" {
		return &GoogleIdProvider{}
	} else if providerType == "qq" {
		return &QqIdProvider{ClientId: clientId}
	}

	return nil
}
