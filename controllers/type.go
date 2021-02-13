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

package controllers

type userEmailFromGithub struct {
	Email      string `json:"email"`
	Primary    bool   `json:"primary"`
	Verified   bool   `json:"verified"`
	Visibility string `json:"visibility"`
}

type userInfoFromGithub struct {
	Login     string `json:"login"`
	AvatarUrl string `json:"avatar_url"`
}

type authResponse struct {
	IsAuthenticated bool   `json:"isAuthenticated"`
	IsSignedUp      bool   `json:"isSignedUp"`
	Email           string `json:"email"`
	Avatar          string `json:"avatar"`
	Addition        string `json:"addition"`
}
