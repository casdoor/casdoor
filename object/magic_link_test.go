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
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestMagicLinkTokenIsRandomAndHashed(t *testing.T) {
	token, err := GenerateMagicLinkToken()
	if err != nil {
		t.Fatalf("GenerateMagicLinkToken() error = %v", err)
	}
	if len(token) < 32 {
		t.Errorf("GenerateMagicLinkToken() = %q, want at least 32 characters", token)
	}

	other, err := GenerateMagicLinkToken()
	if err != nil {
		t.Fatalf("GenerateMagicLinkToken() error = %v", err)
	}
	if token == other {
		t.Errorf("GenerateMagicLinkToken() returned the same token twice: %q", token)
	}

	hash := HashMagicLinkToken(token)
	if len(hash) != 64 {
		t.Errorf("HashMagicLinkToken() = %q, want a 64 character hex digest", hash)
	}
	if strings.Contains(hash, token) {
		t.Errorf("HashMagicLinkToken() = %q, want the token not to be recoverable", hash)
	}
	if hash != HashMagicLinkToken(token) {
		t.Errorf("HashMagicLinkToken() is not stable for the same token")
	}
	if hash == HashMagicLinkToken(other) {
		t.Errorf("HashMagicLinkToken() collided for two different tokens")
	}
}

func TestMagicLinkApplicationDefaults(t *testing.T) {
	application := &Application{}

	if got := application.GetMagicLinkExpireMinutes(); got != MagicLinkDefaultExpireMinutes {
		t.Errorf("GetMagicLinkExpireMinutes() = %d, want %d", got, MagicLinkDefaultExpireMinutes)
	}
	if got := application.GetMagicLinkRateLimitEmail(); got != MagicLinkDefaultRateLimitEmail {
		t.Errorf("GetMagicLinkRateLimitEmail() = %d, want %d", got, MagicLinkDefaultRateLimitEmail)
	}
	if got := application.GetMagicLinkRateLimitIp(); got != MagicLinkDefaultRateLimitIp {
		t.Errorf("GetMagicLinkRateLimitIp() = %d, want %d", got, MagicLinkDefaultRateLimitIp)
	}
	if got := application.GetMagicLinkRateLimitApplication(); got != MagicLinkDefaultRateLimitApplication {
		t.Errorf("GetMagicLinkRateLimitApplication() = %d, want %d", got, MagicLinkDefaultRateLimitApplication)
	}

	if application.IsMagicLinkEnabled() {
		t.Errorf("IsMagicLinkEnabled() = true for a new application, want false")
	}
	application.EnableMagicLinkSignup = true
	if application.IsMagicLinkSignupEnabled() {
		t.Errorf("IsMagicLinkSignupEnabled() = true while sign-in is off, want false")
	}
	application.EnableMagicLink = true
	if !application.IsMagicLinkSignupEnabled() {
		t.Errorf("IsMagicLinkSignupEnabled() = false while both flags are on, want true")
	}
}

func TestValidateMagicLinkConfig(t *testing.T) {
	tests := []struct {
		name        string
		application *Application
		wantErr     bool
	}{
		{"nil application", nil, false},
		{"defaults", &Application{}, false},
		{"signup without signin", &Application{EnableMagicLinkSignup: true}, true},
		{"signup with signin", &Application{EnableMagicLink: true, EnableMagicLinkSignup: true}, false},
		{"ttl too short", &Application{MagicLinkExpireMinutes: MagicLinkMinExpireMinutes - 1}, true},
		{"ttl too long", &Application{MagicLinkExpireMinutes: MagicLinkMaxExpireMinutes + 1}, true},
		{"ttl in range", &Application{MagicLinkExpireMinutes: 30}, false},
		{"negative email limit", &Application{MagicLinkRateLimitEmail: -1}, true},
		{"negative ip limit", &Application{MagicLinkRateLimitIp: -1}, true},
		{"negative application limit", &Application{MagicLinkRateLimitApplication: -1}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMagicLinkConfig(tt.application)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateMagicLinkConfig() error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

func TestResolveMagicLinkExpireTime(t *testing.T) {
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name          string
		expireMinutes int
		want          time.Time
	}{
		{"default", 0, now.Add(MagicLinkDefaultExpireMinutes * time.Minute)},
		{"configured", 30, now.Add(30 * time.Minute)},
		{"clamped to the maximum", MagicLinkMaxExpireMinutes + 100, now.Add(MagicLinkMaxExpireMinutes * time.Minute)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			application := &Application{MagicLinkExpireMinutes: tt.expireMinutes}
			if got := ResolveMagicLinkExpireTime(application, now); !got.Equal(tt.want) {
				t.Errorf("ResolveMagicLinkExpireTime() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckMagicLinkRateLimitCounts(t *testing.T) {
	application := &Application{
		MagicLinkRateLimitEmail:       2,
		MagicLinkRateLimitIp:          3,
		MagicLinkRateLimitApplication: 4,
	}

	tests := []struct {
		name             string
		emailCount       int64
		ipCount          int64
		applicationCount int64
		wantErr          bool
	}{
		{"below every limit", 1, 2, 3, false},
		{"email limit reached", 2, 0, 0, true},
		{"ip limit reached", 0, 3, 0, true},
		{"application limit reached", 0, 0, 4, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkMagicLinkRateLimitCounts(application, tt.emailCount, tt.ipCount, tt.applicationCount)
			if (err != nil) != tt.wantErr {
				t.Errorf("checkMagicLinkRateLimitCounts() error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

func TestCheckMagicLinkState(t *testing.T) {
	now := time.Now().Unix()

	tests := []struct {
		name      string
		magicLink *MagicLink
		wantErr   bool
	}{
		{"created and valid", &MagicLink{Status: MagicLinkStatusCreated, ExpireAt: now + 60}, false},
		{"sent and valid", &MagicLink{Status: MagicLinkStatusSent, ExpireAt: now + 60}, false},
		{"expired by time", &MagicLink{Status: MagicLinkStatusSent, ExpireAt: now - 1}, true},
		{"expired by status", &MagicLink{Status: MagicLinkStatusExpired, ExpireAt: now + 60}, true},
		{"already used", &MagicLink{Status: MagicLinkStatusUsed, ExpireAt: now + 60}, true},
		{"revoked", &MagicLink{Status: MagicLinkStatusRevoked, ExpireAt: now + 60}, true},
		{"failed", &MagicLink{Status: MagicLinkStatusFailed, ExpireAt: now + 60}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckMagicLinkState(tt.magicLink, now)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckMagicLinkState() error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewMagicLink(t *testing.T) {
	application := &Application{Owner: "admin", Name: "app-built-in", Organization: "built-in"}
	expireAt := time.Now().Add(10 * time.Minute)
	token := "the-token"

	magicLink := NewMagicLink(application, "alice@example.com", "10.0.0.1", "admin/alice", token, map[string]string{
		"clientId":     "client-id",
		"responseType": "code",
		"redirectUri":  "https://example.com/callback",
		"state":        "the-state",
	}, expireAt)

	if magicLink.Owner != "built-in" {
		t.Errorf("Owner = %q, want the application's organization", magicLink.Owner)
	}
	if magicLink.Application != application.GetId() {
		t.Errorf("Application = %q, want %q", magicLink.Application, application.GetId())
	}
	if magicLink.Status != MagicLinkStatusCreated {
		t.Errorf("Status = %q, want %q", magicLink.Status, MagicLinkStatusCreated)
	}
	if magicLink.ExpireAt != expireAt.Unix() {
		t.Errorf("ExpireAt = %d, want %d", magicLink.ExpireAt, expireAt.Unix())
	}
	if magicLink.TokenHash != HashMagicLinkToken(token) {
		t.Errorf("TokenHash = %q, want the hash of the token", magicLink.TokenHash)
	}
	if magicLink.ResponseType != "code" {
		t.Errorf("ResponseType = %q, want %q", magicLink.ResponseType, "code")
	}

	// no OAuth request at all still means a plain sign-in
	plain := NewMagicLink(application, "alice@example.com", "10.0.0.1", "", "", map[string]string{}, time.Time{})
	if plain.ResponseType != "login" {
		t.Errorf("ResponseType = %q, want %q", plain.ResponseType, "login")
	}
	if plain.TokenHash != "" {
		t.Errorf("TokenHash = %q, want it to stay empty without a token", plain.TokenHash)
	}
	if plain.ExpireAt == 0 {
		t.Errorf("ExpireAt = 0, want the application's default TTL to be applied")
	}
}

func TestBuildMagicLinkUrl(t *testing.T) {
	magicLink := &MagicLink{
		ClientId:            "client-id",
		ResponseType:        "code",
		RedirectUri:         "https://example.com/callback",
		Scope:               "openid profile",
		State:               "the state",
		CodeChallengeMethod: "S256",
		CodeChallenge:       "the-challenge",
	}

	magicLinkUrl := BuildMagicLinkUrl(magicLink, "the-token", "example.com")

	parsed, err := url.Parse(magicLinkUrl)
	if err != nil {
		t.Fatalf("BuildMagicLinkUrl() returned an unparsable URL %q: %v", magicLinkUrl, err)
	}
	if parsed.Path != "/magic-link/callback" {
		t.Errorf("path = %q, want %q", parsed.Path, "/magic-link/callback")
	}

	query := parsed.Query()
	want := map[string]string{
		"token":                 "the-token",
		"clientId":              "client-id",
		"responseType":          "code",
		"redirectUri":           "https://example.com/callback",
		"scope":                 "openid profile",
		"state":                 "the state",
		"code_challenge_method": "S256",
		"code_challenge":        "the-challenge",
	}
	for key, value := range want {
		if got := query.Get(key); got != value {
			t.Errorf("query[%q] = %q, want %q", key, got, value)
		}
	}

	// the empty OAuth parameters are left out instead of being sent as blanks
	empty := BuildMagicLinkUrl(&MagicLink{ResponseType: "login"}, "the-token", "example.com")
	emptyQuery, err := url.Parse(empty)
	if err != nil {
		t.Fatalf("BuildMagicLinkUrl() returned an unparsable URL %q: %v", empty, err)
	}
	if _, ok := emptyQuery.Query()["redirectUri"]; ok {
		t.Errorf("query = %v, want no empty redirectUri", emptyQuery.Query())
	}
}

func TestGetDefaultMagicLinkEmailContent(t *testing.T) {
	content := GetDefaultMagicLinkEmailContent()

	for _, placeholder := range []string{"%link", "%expireTime"} {
		if !strings.Contains(content, placeholder) {
			t.Errorf("the default template does not contain %q", placeholder)
		}
	}

	replaced := strings.ReplaceAll(content, "%link", "https://example.com/magic-link/callback?token=x")
	replaced = strings.ReplaceAll(replaced, "%expireTime", "2026-01-01T12:00:00Z")
	if strings.Contains(replaced, "%link") || strings.Contains(replaced, "%expireTime") {
		t.Errorf("the default template still contains a placeholder after substitution")
	}
}
