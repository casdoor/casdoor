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
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"testing"

	"golang.org/x/oauth2"
)

func testAlipayPrivateKeyPEM(t *testing.T) string {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate test RSA key: %v", err)
	}
	der, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatalf("marshal test RSA key: %v", err)
	}
	return string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der}))
}

func TestAlipayGetUserInfoId(t *testing.T) {
	key := testAlipayPrivateKeyPEM(t)

	tests := []struct {
		name    string
		payload string
		wantID  string
	}{
		{
			name: "only alipay_open_id",
			payload: `{
				"alipay_user_info_share_response":{
					"code":"10000",
					"msg":"Success",
					"avatar":"https://example.com/avatar.png",
					"nick_name":"zhangsan",
					"alipay_open_id":"openid-new-app"
				},
				"sign":"dummy"
			}`,
			wantID: "openid-new-app",
		},
		{
			name: "legacy user_id",
			payload: `{
				"alipay_user_info_share_response":{
					"code":"10000",
					"msg":"Success",
					"avatar":"https://example.com/avatar.png",
					"nick_name":"zhangsan",
					"user_id":"2099222233334444"
				},
				"sign":"dummy"
			}`,
			wantID: "2099222233334444",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(tt.payload))
			}))
			defer srv.Close()

			idp := &AlipayIdProvider{
				Client: srv.Client(),
				Config: &oauth2.Config{
					ClientID:     "app-id",
					ClientSecret: key,
					Endpoint:     oauth2.Endpoint{TokenURL: srv.URL},
				},
			}

			got, err := idp.GetUserInfo(&oauth2.Token{AccessToken: "auth-token"})
			if err != nil {
				t.Fatalf("GetUserInfo: %v", err)
			}
			if got.Id != tt.wantID {
				t.Fatalf("Id = %q, want %q", got.Id, tt.wantID)
			}
		})
	}
}
