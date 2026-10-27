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
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
)

// fakeWeChatTransport answers the WeChat endpoints used by the mini program provider.
type fakeWeChatTransport struct {
	mu               sync.Mutex
	stableTokenCalls int
	phoneCalls       int
	lastStableBody   map[string]interface{}
	lastAccessToken  string
	phoneResponse    string
}

func (f *fakeWeChatTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	body := ""
	switch req.URL.Path {
	case "/cgi-bin/stable_token":
		f.stableTokenCalls++
		data, _ := io.ReadAll(req.Body)
		_ = json.Unmarshal(data, &f.lastStableBody)
		body = `{"access_token":"token-` + string(rune('0'+f.stableTokenCalls)) + `","expires_in":7200}`
	case "/wxa/business/getuserphonenumber":
		f.phoneCalls++
		f.lastAccessToken = req.URL.Query().Get("access_token")
		body = f.phoneResponse
	default:
		body = `{"errcode":404,"errmsg":"unexpected path"}`
	}

	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(body)),
		Request:    req,
	}, nil
}

func newTestMiniProgramIdp(t *testing.T, appId string, transport *fakeWeChatTransport) *WeChatMiniProgramIdProvider {
	t.Helper()
	weChatMiniProgramAccessTokensMu.Lock()
	delete(weChatMiniProgramAccessTokens, appId)
	weChatMiniProgramAccessTokensMu.Unlock()

	idp := NewWeChatMiniProgramIdProvider(appId, "secret")
	idp.SetHttpClient(&http.Client{Transport: transport})
	return idp
}

func TestWeChatMiniProgramGetPhoneNumber(t *testing.T) {
	transport := &fakeWeChatTransport{
		phoneResponse: `{"errcode":0,"errmsg":"ok","phone_info":{"phoneNumber":"13800000000","purePhoneNumber":"13800000000","countryCode":"86"}}`,
	}
	idp := newTestMiniProgramIdp(t, "wx-test-phone", transport)

	info, err := idp.GetPhoneNumber("phone-code")
	if err != nil {
		t.Fatalf("GetPhoneNumber() error: %v", err)
	}
	if info.PurePhoneNumber != "13800000000" || info.E164() != "+8613800000000" || info.RegionCode() != "CN" {
		t.Fatalf("unexpected phone info: %+v, e164=%s, region=%s", info, info.E164(), info.RegionCode())
	}
	if transport.lastAccessToken != "token-1" {
		t.Fatalf("getuserphonenumber should use the stable access token, got %q", transport.lastAccessToken)
	}
	if force, ok := transport.lastStableBody["force_refresh"].(bool); !ok || force {
		t.Fatalf("stable_token must be requested with force_refresh=false, got %v", transport.lastStableBody["force_refresh"])
	}
}

func TestWeChatMiniProgramAccessTokenIsCached(t *testing.T) {
	transport := &fakeWeChatTransport{
		phoneResponse: `{"errcode":0,"errmsg":"ok","phone_info":{"purePhoneNumber":"13800000000","countryCode":"86"}}`,
	}
	idp := newTestMiniProgramIdp(t, "wx-test-cache", transport)

	for i := 0; i < 3; i++ {
		if _, err := idp.GetPhoneNumber("phone-code"); err != nil {
			t.Fatalf("GetPhoneNumber() error: %v", err)
		}
	}
	if transport.stableTokenCalls != 1 {
		t.Fatalf("access token should be fetched once and cached, fetched %d times", transport.stableTokenCalls)
	}
	if transport.phoneCalls != 3 {
		t.Fatalf("expected 3 phone number calls, got %d", transport.phoneCalls)
	}
}

func TestWeChatMiniProgramInvalidAccessTokenIsDropped(t *testing.T) {
	transport := &fakeWeChatTransport{
		phoneResponse: `{"errcode":40001,"errmsg":"invalid credential"}`,
	}
	idp := newTestMiniProgramIdp(t, "wx-test-invalid", transport)

	if _, err := idp.GetPhoneNumber("phone-code"); err == nil {
		t.Fatal("GetPhoneNumber() should fail when WeChat rejects the access token")
	}

	transport.phoneResponse = `{"errcode":0,"errmsg":"ok","phone_info":{"purePhoneNumber":"13800000000","countryCode":"86"}}`
	if _, err := idp.GetPhoneNumber("phone-code"); err != nil {
		t.Fatalf("GetPhoneNumber() error after refreshing the access token: %v", err)
	}
	if transport.stableTokenCalls != 2 || transport.lastAccessToken != "token-2" {
		t.Fatalf("a rejected access token should be refetched, stable_token calls=%d, last token=%q", transport.stableTokenCalls, transport.lastAccessToken)
	}
}

func TestWeChatMiniProgramGetPhoneNumberError(t *testing.T) {
	transport := &fakeWeChatTransport{
		phoneResponse: `{"errcode":40029,"errmsg":"invalid code"}`,
	}
	idp := newTestMiniProgramIdp(t, "wx-test-error", transport)

	if _, err := idp.GetPhoneNumber("expired-code"); err == nil || !strings.Contains(err.Error(), "40029") {
		t.Fatalf("expected an error carrying the WeChat errcode, got %v", err)
	}
	if transport.stableTokenCalls != 1 {
		t.Fatalf("an invalid phone code must not drop the cached access token, stable_token calls=%d", transport.stableTokenCalls)
	}
}

func TestWeChatMiniProgramPhoneInfoRegion(t *testing.T) {
	cases := []struct {
		countryCode string
		e164        string
		region      string
	}{
		{"86", "+8613800000000", "CN"},
		{"", "+8613800000000", "CN"},
		{"852", "+85213800000000", "HK"},
		{"1", "+113800000000", ""},
	}
	for _, c := range cases {
		info := &WeChatMiniProgramPhoneInfo{PurePhoneNumber: "13800000000", CountryCode: c.countryCode}
		if info.E164() != c.e164 || info.RegionCode() != c.region {
			t.Errorf("countryCode %q: got e164=%s region=%s, want %s %s", c.countryCode, info.E164(), info.RegionCode(), c.e164, c.region)
		}
	}
}
