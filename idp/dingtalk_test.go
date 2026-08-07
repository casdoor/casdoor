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
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"golang.org/x/oauth2"
)

type roundTripFunc func(request *http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

func TestDingTalkGetUserInfoIncludesTitle(t *testing.T) {
	client := &http.Client{
		Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
			var responseBody string

			switch request.URL.Host + request.URL.Path {
			case "api.dingtalk.com/v1.0/contact/users/me":
				if token := request.Header.Get("x-acs-dingtalk-access-token"); token != "user-token" {
					t.Fatalf("unexpected user access token: %q", token)
				}
				responseBody = `{
					"nick": "Alice",
					"openId": "open-id",
					"unionId": "union-id",
					"avatarUrl": "https://example.com/avatar.png",
					"email": "alice@example.com",
					"mobile": "13800138000",
					"stateCode": "86"
				}`
			case "api.dingtalk.com/v1.0/oauth2/accessToken":
				responseBody = `{"accessToken": "corp-token", "expireIn": 7200}`
			case "oapi.dingtalk.com/topapi/user/getbyunionid":
				responseBody = `{
					"errcode": 0,
					"errmsg": "ok",
					"result": {"userid": "user-id"}
				}`
			case "oapi.dingtalk.com/topapi/v2/user/get":
				responseBody = `{
					"errcode": 0,
					"errmsg": "ok",
					"result": {
						"mobile": "13900139000",
						"email": "alice.corp@example.com",
						"unionid": "union-id",
						"title": "Senior Engineer"
					}
				}`
			default:
				return nil, fmt.Errorf("unexpected request: %s %s", request.Method, request.URL)
			}

			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(responseBody)),
				Header:     make(http.Header),
			}, nil
		}),
	}

	provider := NewDingTalkIdProvider("client-id", "client-secret", "https://example.com/callback")
	provider.SetHttpClient(client)

	userInfo, err := provider.GetUserInfo(&oauth2.Token{AccessToken: "user-token"})
	if err != nil {
		t.Fatalf("GetUserInfo() error = %v", err)
	}

	if got := userInfo.Extra["title"]; got != "Senior Engineer" {
		t.Errorf("GetUserInfo() title = %q, want %q", got, "Senior Engineer")
	}
}
