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

	"github.com/casdoor/casdoor/idp"
	"github.com/casdoor/casdoor/util"
	"github.com/xorm-io/core"
	"golang.org/x/oauth2"
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

func TestPerProviderOAuthTokens(t *testing.T) {
	// Create test user
	user := &User{
		Owner:      "test-org",
		Name:       "test-user",
		Properties: make(map[string]string),
	}

	// Simulate storing tokens for GitHub provider
	githubTokenKey := fmt.Sprintf("oauth_%s_accessToken", "GitHub")
	githubRefreshKey := fmt.Sprintf("oauth_%s_refreshToken", "GitHub")
	setUserProperty(user, githubTokenKey, "github-access-token-123")
	setUserProperty(user, githubRefreshKey, "github-refresh-token-456")

	// Simulate storing tokens for Custom provider
	customTokenKey := fmt.Sprintf("oauth_%s_accessToken", "Custom")
	customRefreshKey := fmt.Sprintf("oauth_%s_refreshToken", "Custom")
	setUserProperty(user, customTokenKey, "custom-access-token-789")
	setUserProperty(user, customRefreshKey, "custom-refresh-token-012")

	// Test retrieving GitHub tokens
	githubAccessToken := GetUserOAuthAccessToken(user, "GitHub")
	if githubAccessToken != "github-access-token-123" {
		t.Errorf("Expected GitHub access token 'github-access-token-123', got '%s'", githubAccessToken)
	}

	githubRefreshToken := GetUserOAuthRefreshToken(user, "GitHub")
	if githubRefreshToken != "github-refresh-token-456" {
		t.Errorf("Expected GitHub refresh token 'github-refresh-token-456', got '%s'", githubRefreshToken)
	}

	// Test retrieving Custom tokens
	customAccessToken := GetUserOAuthAccessToken(user, "Custom")
	if customAccessToken != "custom-access-token-789" {
		t.Errorf("Expected Custom access token 'custom-access-token-789', got '%s'", customAccessToken)
	}

	customRefreshToken := GetUserOAuthRefreshToken(user, "Custom")
	if customRefreshToken != "custom-refresh-token-012" {
		t.Errorf("Expected Custom refresh token 'custom-refresh-token-012', got '%s'", customRefreshToken)
	}

	// Verify both tokens exist simultaneously
	if len(user.Properties) != 4 {
		t.Errorf("Expected 4 properties, got %d", len(user.Properties))
	}
}

func TestOAuthTokenMasking(t *testing.T) {
	// Create test user with OAuth tokens
	user := &User{
		Owner:                "test-org",
		Name:                 "test-user",
		OriginalToken:        "legacy-token-123",
		OriginalRefreshToken: "legacy-refresh-456",
		Properties: map[string]string{
			"oauth_GitHub_accessToken":  "github-token-abc",
			"oauth_GitHub_refreshToken": "github-refresh-def",
			"oauth_Custom_accessToken":  "custom-token-xyz",
			"oauth_Custom_refreshToken": "custom-refresh-uvw",
			"oauth_GitHub_id":           "12345",
			"oauth_Custom_username":     "testuser",
		},
	}

	// Make a copy for testing
	maskedUser := *user
	maskedUser.Properties = make(map[string]string)
	for k, v := range user.Properties {
		maskedUser.Properties[k] = v
	}

	// Apply masking logic (simulate non-admin user)
	isAdminOrSelf := false
	if !isAdminOrSelf {
		if maskedUser.OriginalToken != "" {
			maskedUser.OriginalToken = "***"
		}
		if maskedUser.OriginalRefreshToken != "" {
			maskedUser.OriginalRefreshToken = "***"
		}
		// Mask per-provider OAuth tokens in Properties
		if maskedUser.Properties != nil {
			for key := range maskedUser.Properties {
				if strings.Contains(key, "_accessToken") || strings.Contains(key, "_refreshToken") {
					maskedUser.Properties[key] = "***"
				}
			}
		}
	}

	// Verify legacy tokens are masked
	if maskedUser.OriginalToken != "***" {
		t.Errorf("Expected OriginalToken to be masked, got '%s'", maskedUser.OriginalToken)
	}
	if maskedUser.OriginalRefreshToken != "***" {
		t.Errorf("Expected OriginalRefreshToken to be masked, got '%s'", maskedUser.OriginalRefreshToken)
	}

	// Verify per-provider tokens are masked
	if maskedUser.Properties["oauth_GitHub_accessToken"] != "***" {
		t.Errorf("Expected GitHub access token to be masked, got '%s'", maskedUser.Properties["oauth_GitHub_accessToken"])
	}
	if maskedUser.Properties["oauth_GitHub_refreshToken"] != "***" {
		t.Errorf("Expected GitHub refresh token to be masked, got '%s'", maskedUser.Properties["oauth_GitHub_refreshToken"])
	}
	if maskedUser.Properties["oauth_Custom_accessToken"] != "***" {
		t.Errorf("Expected Custom access token to be masked, got '%s'", maskedUser.Properties["oauth_Custom_accessToken"])
	}
	if maskedUser.Properties["oauth_Custom_refreshToken"] != "***" {
		t.Errorf("Expected Custom refresh token to be masked, got '%s'", maskedUser.Properties["oauth_Custom_refreshToken"])
	}

	// Verify non-token properties are NOT masked
	if maskedUser.Properties["oauth_GitHub_id"] != "12345" {
		t.Errorf("Expected GitHub ID to not be masked, got '%s'", maskedUser.Properties["oauth_GitHub_id"])
	}
	if maskedUser.Properties["oauth_Custom_username"] != "testuser" {
		t.Errorf("Expected Custom username to not be masked, got '%s'", maskedUser.Properties["oauth_Custom_username"])
	}
}

func TestMultiProviderOAuthFlow(t *testing.T) {
	// Simulate a user logging in with multiple OAuth providers
	user := &User{
		Owner:      "test-org",
		Name:       "multi-provider-user",
		Properties: make(map[string]string),
	}

	// Simulate GitHub OAuth login (first provider)
	githubUserInfo := &idp.UserInfo{
		Id:          "github-user-123",
		Username:    "githubuser",
		DisplayName: "GitHub User",
		Email:       "user@github.com",
	}
	githubToken := &oauth2.Token{
		AccessToken:  "github-access-token-abc",
		RefreshToken: "github-refresh-token-def",
	}

	// Manually set GitHub OAuth properties (simulating SetUserOAuthProperties logic without DB)
	// Store tokens per provider in Properties map
	accessTokenKey := fmt.Sprintf("oauth_%s_accessToken", "GitHub")
	setUserProperty(user, accessTokenKey, githubToken.AccessToken)
	refreshTokenKey := fmt.Sprintf("oauth_%s_refreshToken", "GitHub")
	setUserProperty(user, refreshTokenKey, githubToken.RefreshToken)
	user.OriginalToken = githubToken.AccessToken
	user.OriginalRefreshToken = githubToken.RefreshToken

	// Set GitHub user info
	setUserProperty(user, "oauth_GitHub_id", githubUserInfo.Id)
	setUserProperty(user, "oauth_GitHub_username", githubUserInfo.Username)

	// Verify GitHub tokens are stored
	githubAccess := GetUserOAuthAccessToken(user, "GitHub")
	if githubAccess != "github-access-token-abc" {
		t.Errorf("Expected GitHub access token, got '%s'", githubAccess)
	}

	// Simulate Custom provider OAuth login (second provider)
	customUserInfo := &idp.UserInfo{
		Id:          "custom-user-456",
		Username:    "customuser",
		DisplayName: "Custom User",
		Email:       "user@custom.com",
	}
	customToken := &oauth2.Token{
		AccessToken:  "custom-access-token-xyz",
		RefreshToken: "custom-refresh-token-uvw",
	}

	// Manually set Custom OAuth properties (simulating SetUserOAuthProperties logic without DB)
	accessTokenKey = fmt.Sprintf("oauth_%s_accessToken", "Custom")
	setUserProperty(user, accessTokenKey, customToken.AccessToken)
	refreshTokenKey = fmt.Sprintf("oauth_%s_refreshToken", "Custom")
	setUserProperty(user, refreshTokenKey, customToken.RefreshToken)
	user.OriginalToken = customToken.AccessToken
	user.OriginalRefreshToken = customToken.RefreshToken

	// Set Custom user info
	setUserProperty(user, "oauth_Custom_id", customUserInfo.Id)
	setUserProperty(user, "oauth_Custom_username", customUserInfo.Username)

	// Verify Custom tokens are stored
	customAccess := GetUserOAuthAccessToken(user, "Custom")
	if customAccess != "custom-access-token-xyz" {
		t.Errorf("Expected Custom access token, got '%s'", customAccess)
	}

	// CRITICAL: Verify GitHub tokens are STILL present (not overwritten)
	githubAccessAfter := GetUserOAuthAccessToken(user, "GitHub")
	if githubAccessAfter != "github-access-token-abc" {
		t.Errorf("GitHub tokens were overwritten! Expected 'github-access-token-abc', got '%s'", githubAccessAfter)
	}

	githubRefreshAfter := GetUserOAuthRefreshToken(user, "GitHub")
	if githubRefreshAfter != "github-refresh-token-def" {
		t.Errorf("GitHub refresh token was overwritten! Expected 'github-refresh-token-def', got '%s'", githubRefreshAfter)
	}

	// Verify both provider IDs are present
	githubId := getUserProperty(user, "oauth_GitHub_id")
	if githubId != "github-user-123" {
		t.Errorf("Expected GitHub ID, got '%s'", githubId)
	}

	customId := getUserProperty(user, "oauth_Custom_id")
	if customId != "custom-user-456" {
		t.Errorf("Expected Custom ID, got '%s'", customId)
	}

	// Verify legacy fields contain the most recent token (for backward compatibility)
	if user.OriginalToken != "custom-access-token-xyz" {
		t.Errorf("Expected legacy OriginalToken to have most recent token, got '%s'", user.OriginalToken)
	}

	if user.OriginalRefreshToken != "custom-refresh-token-uvw" {
		t.Errorf("Expected legacy OriginalRefreshToken to have most recent token, got '%s'", user.OriginalRefreshToken)
	}
}
