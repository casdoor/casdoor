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
	"path/filepath"
	"testing"

	"github.com/casdoor/casdoor/idp"
	"github.com/xorm-io/xorm"
	"golang.org/x/oauth2"
)

func TestSetUserOAuthPropertiesTokenPersistence(t *testing.T) {
	engine, err := xorm.NewEngine("sqlite", filepath.Join(t.TempDir(), "casdoor.db"))
	if err != nil {
		t.Fatal(err)
	}
	previousOrmer := ormer
	ormer = &Ormer{Engine: engine}
	t.Cleanup(func() {
		ormer = previousOrmer
		if err := engine.Close(); err != nil {
			t.Error(err)
		}
	})
	if err := engine.Sync2(new(User), new(Syncer)); err != nil {
		t.Fatal(err)
	}

	for _, tt := range []struct {
		provider string
		persist  bool
	}{
		{provider: "Lark", persist: false},
		{provider: "GitHub", persist: true},
	} {
		t.Run(tt.provider, func(t *testing.T) {
			user := &User{
				Owner:                "example",
				Name:                 tt.provider,
				OriginalToken:        "previous-access-token",
				OriginalRefreshToken: "previous-refresh-token",
			}
			if _, err := engine.Insert(user); err != nil {
				t.Fatal(err)
			}
			token := &oauth2.Token{AccessToken: "new-access-token", RefreshToken: "new-refresh-token"}
			info := &idp.UserInfo{Id: "external-subject", DisplayName: "Example User"}
			if changed, err := SetUserOAuthProperties(&Organization{}, user, tt.provider, info, token); err != nil || !changed {
				t.Fatalf("SetUserOAuthProperties changed=%t, err=%v", changed, err)
			}
			stored, err := getUser(user.Owner, user.Name)
			if err != nil || stored == nil {
				t.Fatalf("could not reload user: %v", err)
			}
			if stored.Properties["oauth_"+tt.provider+"_id"] != info.Id || stored.DisplayName != info.DisplayName {
				t.Fatal("third-party identity was not saved")
			}

			accessToken, hasAccessToken := stored.Properties[getOAuthTokenPropertyKey(tt.provider, "accessToken")]
			refreshToken, hasRefreshToken := stored.Properties[getOAuthTokenPropertyKey(tt.provider, "refreshToken")]
			if tt.persist {
				if accessToken != token.AccessToken || refreshToken != token.RefreshToken ||
					stored.OriginalToken != token.AccessToken || stored.OriginalRefreshToken != token.RefreshToken {
					t.Fatal("delegated credentials were not saved")
				}
			} else if hasAccessToken || hasRefreshToken ||
				stored.OriginalToken != "previous-access-token" || stored.OriginalRefreshToken != "previous-refresh-token" {
				t.Fatal("Lark sign-in persisted credentials or replaced the existing legacy credentials")
			}
		})
	}
}
