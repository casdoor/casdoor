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

package idp

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"golang.org/x/oauth2"
)

func TestLarkAuthorizationCodeLogin(t *testing.T) {
	const (
		redirectURL  = "https://example.test/callback"
		tokenPath    = "/open-apis/authen/v2/oauth/token"
		userInfoPath = "/open-apis/authen/v1/user_info"
	)
	for _, tt := range []struct {
		name     string
		global   bool
		domain   string
		userID   string
		username string
	}{
		{"Feishu", false, "https://open.feishu.cn", "", "example-union-id"},
		{"Lark", true, "https://open.larksuite.com", "example-user-id", "example-user-id"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			provider := NewLarkIdProvider("example-app-id", "example-app-secret", redirectURL, tt.global)
			if provider.LarkDomain != tt.domain || provider.Config.Endpoint.TokenURL != tt.domain+tokenPath {
				t.Fatalf("unexpected OAuth domain or token endpoint: %q, %q", provider.LarkDomain, provider.Config.Endpoint.TokenURL)
			}

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				switch r.URL.Path {
				case tokenPath:
					if r.Method != http.MethodPost || r.Header.Get("Authorization") != "" || !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
						t.Error("authorization code exchange should be a JSON POST using the client credentials in the body")
					}
					var body map[string]string
					want := map[string]string{
						"grant_type":    "authorization_code",
						"client_id":     "example-app-id",
						"client_secret": "example-app-secret",
						"code":          "example-code",
						"redirect_uri":  redirectURL,
					}
					if err := json.NewDecoder(r.Body).Decode(&body); err != nil || !reflect.DeepEqual(body, want) {
						t.Errorf("unexpected code exchange request: %v", err)
					}
					fmt.Fprint(w, `{"code":0,"access_token":"example-user-token","expires_in":3600,"refresh_token":"unused-refresh-token"}`)
				case userInfoPath:
					if r.Method != http.MethodGet || r.Header.Get("Authorization") != "Bearer example-user-token" {
						t.Error("user info should be read with the user access token")
					}
					fmt.Fprintf(w, `{"code":0,"data":{"open_id":"example-open-id","union_id":"example-union-id","user_id":%q,"name":"Example User","enterprise_email":"user@example.test","mobile":"+16502530000"}}`, tt.userID)
				default:
					t.Errorf("unexpected login endpoint: %s", r.URL.Path)
					http.NotFound(w, r)
				}
			}))
			defer server.Close()
			provider.Config.Endpoint.TokenURL = strings.Replace(provider.Config.Endpoint.TokenURL, provider.LarkDomain, server.URL, 1)
			provider.LarkDomain = server.URL
			provider.SetHttpClient(server.Client())

			token, err := provider.GetToken("example-code")
			if err != nil || !token.Valid() || token.RefreshToken != "" {
				t.Fatalf("code exchange did not return a usable short-lived credential: %v", err)
			}
			info, err := provider.GetUserInfo(token)
			if err != nil {
				t.Fatal(err)
			}
			if info.Id != "example-open-id" || info.Username != tt.username || info.DisplayName != "Example User" ||
				info.Email != "user@example.test" || !info.EmailVerified || info.Phone != "6502530000" || !info.PhoneVerified {
				t.Fatalf("Unexpected identity mapping: %+v", info)
			}
		})
	}
}

func TestLarkUserInfoRequiresOpenID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"code":0,"data":{"user_id":"example-user-id"}}`)
	}))
	defer server.Close()
	provider := &LarkIdProvider{Client: server.Client(), LarkDomain: server.URL}
	if info, err := provider.GetUserInfo(&oauth2.Token{AccessToken: "example-user-token"}); err == nil || info != nil {
		t.Fatalf("expected login to reject identity with no open_id")
	}
}
