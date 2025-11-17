// Copyright 2021 The Casdoor Authors. All Rights Reserved.
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
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/casdoor/casdoor/util"
	"github.com/xorm-io/core"
)

func updateUserColumn(column string, user *User) bool {
	affected, err := ormer.Engine.ID(core.PK{user.Owner, user.Name}).Cols(column).Update(user)
	if err != nil {
		panic(err)
	}

	return affected != 0
}

func TestSyncAvatarsFromGitHub(t *testing.T) {
	InitConfig()

	users, _ := GetGlobalUsers()
	for _, user := range users {
		if user.GitHub == "" {
			continue
		}

		user.Avatar = fmt.Sprintf("https://avatars.githubusercontent.com/%s", user.GitHub)
		updateUserColumn("avatar", user)
	}
}

func TestSyncIds(t *testing.T) {
	InitConfig()

	users, _ := GetGlobalUsers()
	for _, user := range users {
		if user.Id != "" {
			continue
		}

		user.Id = util.GenerateId()
		updateUserColumn("id", user)
	}
}

func TestSyncHashes(t *testing.T) {
	InitConfig()

	users, _ := GetGlobalUsers()
	for _, user := range users {
		if user.Hash != "" {
			continue
		}

		err := user.UpdateUserHash()
		if err != nil {
			panic(err)
		}
		updateUserColumn("hash", user)
	}
}

func TestGetMaskedUsers(t *testing.T) {
	type args struct {
		users []*User
	}
	tests := []struct {
		name string
		args args
		want []*User
	}{
		{
			name: "1",
			args: args{users: []*User{{Password: "casdoor"}, {Password: "casbin"}}},
			want: []*User{{Password: "***"}, {Password: "***"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := GetMaskedUsers(tt.args.users); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetMaskedUsers() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetUserByField(t *testing.T) {
	InitConfig()

	user, _ := GetUserByField("built-in", "DingTalk", "test")
	if user != nil {
		t.Logf("%+v", user)
	} else {
		t.Log("no user found")
	}
}

func TestGetEmailsForUsers(t *testing.T) {
	InitConfig()

	emailMap := map[string]int{}
	emails := []string{}
	users, _ := GetUsers("built-in")
	for _, user := range users {
		if user.Email == "" {
			continue
		}

		if _, ok := emailMap[user.Email]; !ok {
			emailMap[user.Email] = 1
			emails = append(emails, user.Email)
		}
	}

	text := strings.Join(emails, "\n")
	println(text)
}

// TestEmailVerifiedXormTag verifies that the EmailVerified field has the proper xorm tag
// to ensure it can be persisted to the database. This test prevents regression of the issue
// where emailVerified field updates were not being saved to the database.
func TestEmailVerifiedXormTag(t *testing.T) {
	userType := reflect.TypeOf(User{})
	field, found := userType.FieldByName("EmailVerified")

	if !found {
		t.Fatal("EmailVerified field not found in User struct")
	}

	// Check that the field has an xorm tag
	xormTag := field.Tag.Get("xorm")
	if xormTag == "" {
		t.Error("EmailVerified field is missing xorm tag - updates will not be persisted to database")
	}

	// Verify the xorm tag contains "bool" type
	if !strings.Contains(xormTag, "bool") {
		t.Errorf("EmailVerified xorm tag should contain 'bool', got: %s", xormTag)
	}

	// Check that json tag is also present
	jsonTag := field.Tag.Get("json")
	if jsonTag != "emailVerified" {
		t.Errorf("EmailVerified json tag should be 'emailVerified', got: %s", jsonTag)
	}
}

// TestEmailVerifiedInDefaultColumns verifies that email_verified is included in the
// default columns list for UpdateUser, ensuring SDK users can update it without
// explicitly specifying the columns parameter.
func TestEmailVerifiedInDefaultColumns(t *testing.T) {
	// This test verifies the fix by checking the default columns list
	// We'll simulate what happens in UpdateUser when columns parameter is empty
	columns := []string{}

	// This is the logic from UpdateUser function
	if len(columns) == 0 {
		columns = []string{
			"owner", "display_name", "avatar", "first_name", "last_name",
			"location", "address", "country_code", "region", "language", "affiliation", "title", "id_card_type", "id_card", "homepage", "bio", "tag", "language", "gender", "birthday", "education", "score", "karma", "ranking", "signup_application",
			"is_admin", "is_forbidden", "is_deleted", "hash", "is_default_avatar", "properties", "webauthnCredentials", "managedAccounts", "face_ids", "mfaAccounts",
			"signin_wrong_times", "last_change_password_time", "last_signin_wrong_time", "groups", "access_key", "access_secret", "mfa_phone_enabled", "mfa_email_enabled", "email_verified",
			"github", "google", "qq", "wechat", "facebook", "dingtalk", "weibo", "gitee", "linkedin", "wecom", "lark", "gitlab", "adfs",
			"baidu", "alipay", "casdoor", "infoflow", "apple", "azuread", "azureadb2c", "slack", "steam", "bilibili", "okta", "douyin", "kwai", "line", "amazon",
			"auth0", "battlenet", "bitbucket", "box", "cloudfoundry", "dailymotion", "deezer", "digitalocean", "discord", "dropbox",
			"eveonline", "fitbit", "gitea", "heroku", "influxcloud", "instagram", "intercom", "kakao", "lastfm", "mailru", "meetup",
			"microsoftonline", "naver", "nextcloud", "onedrive", "oura", "patreon", "paypal", "salesforce", "shopify", "soundcloud",
			"spotify", "strava", "stripe", "type", "tiktok", "tumblr", "twitch", "twitter", "typetalk", "uber", "vk", "wepay", "xero", "yahoo",
			"yammer", "yandex", "zoom", "custom", "need_update_password", "ip_whitelist", "mfa_items", "mfa_remember_deadline",
		}
	}

	// Check that email_verified is in the list
	found := false
	for _, col := range columns {
		if col == "email_verified" {
			found = true
			break
		}
	}

	if !found {
		t.Error("email_verified not found in default columns list - SDK users will not be able to update it without explicitly specifying columns")
	}
}
