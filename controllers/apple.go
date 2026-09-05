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

package controllers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/beego/beego/v2/server/web/context"
	"github.com/casdoor/casdoor/util"
)

// Apple sends the real name only in the "user" field of the first
// authorization's form_post to /api/callback, so it waits in this cookie until
// the separate /api/login request that consumes it.
const (
	appleUserNameCookie    = "casdoor_apple_user"
	appleUserNameCookieAge = 300
)

// The payload is browser-relayed and unsigned, so only the name is read: the
// email must keep coming from the verified ID token.
type appleFormPostUser struct {
	Name struct {
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
	} `json:"name"`
}

func getAppleDisplayName(rawUser string) string {
	if rawUser == "" {
		return ""
	}

	user := appleFormPostUser{}
	err := json.Unmarshal([]byte(rawUser), &user)
	if err != nil {
		return ""
	}

	firstName := strings.TrimSpace(user.Name.FirstName)
	lastName := strings.TrimSpace(user.Name.LastName)
	if firstName == "" || lastName == "" {
		return firstName + lastName
	}

	if util.IsChinese(firstName) || util.IsChinese(lastName) {
		return fmt.Sprintf("%s%s", lastName, firstName)
	}
	return fmt.Sprintf("%s %s", firstName, lastName)
}

func setAppleDisplayNameCookie(ctx *context.Context, displayName string) {
	if displayName == "" {
		return
	}

	value := base64.RawURLEncoding.EncodeToString([]byte(displayName))
	ctx.SetCookie(appleUserNameCookie, value, appleUserNameCookieAge, "/", "", ctx.Input.Scheme() == "https", true, "Lax")
}

func takeAppleDisplayNameCookie(ctx *context.Context) string {
	value := ctx.GetCookie(appleUserNameCookie)
	if value == "" {
		return ""
	}

	ctx.SetCookie(appleUserNameCookie, "", -1, "/", "", ctx.Input.Scheme() == "https", true, "Lax")

	displayName, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return ""
	}
	return string(displayName)
}
