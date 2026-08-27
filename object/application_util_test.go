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

import "testing"

func TestRedirectUriMatchesPattern(t *testing.T) {
	tests := []struct {
		redirectUri string
		targetUri   string
		want        bool
	}{
		// Exact match
		{"https://login.example.com/callback", "https://login.example.com/callback", true},

		// Full URL pattern: exact host
		{"https://login.example.com/callback", "https://login.example.com/callback", true},
		{"https://login.example.com/other", "https://login.example.com/callback", false},

		// Full URL pattern: subdomain of configured host
		{"https://def.abc.com/callback", "abc.com", true},
		{"https://def.abc.com/callback", ".abc.com", true},
		{"https://def.abc.com/callback", ".abc.com/", true},
		{"https://deep.app.example.com/callback", "https://example.com/callback", true},

		// Full URL pattern: unrelated host must not match
		{"https://evil.com/callback", "https://example.com/callback", false},
		// Suffix collision: evilexample.com must not match example.com
		{"https://evilexample.com/callback", "https://example.com/callback", false},

		// Full URL pattern: scheme mismatch
		{"http://app.example.com/callback", "https://example.com/callback", false},

		// Full URL pattern: path mismatch
		{"https://app.example.com/other", "https://example.com/callback", false},

		// Scheme-less pattern: exact host
		{"https://login.example.com/callback", "login.example.com/callback", true},
		{"http://login.example.com/callback", "login.example.com/callback", true},

		// Scheme-less pattern: subdomain of configured host
		{"https://app.login.example.com/callback", "login.example.com/callback", true},

		// Scheme-less pattern: unrelated host must not match
		{"https://evil.com/callback", "login.example.com/callback", false},

		// Scheme-less pattern: query-string injection must not match
		{"https://evil.com/?r=https://login.example.com/callback", "login.example.com/callback", false},
		{"https://evil.com/page?redirect=https://login.example.com/callback", "login.example.com/callback", false},

		// Scheme-less pattern: path mismatch
		{"https://login.example.com/other", "login.example.com/callback", false},

		// Scheme-less pattern: non-http scheme must not match
		{"ftp://login.example.com/callback", "login.example.com/callback", false},

		// Empty target
		{"https://login.example.com/callback", "", false},
	}

	for _, tt := range tests {
		got := redirectUriMatchesPattern(tt.redirectUri, tt.targetUri)
		if got != tt.want {
			t.Errorf("redirectUriMatchesPattern(%q, %q) = %v, want %v", tt.redirectUri, tt.targetUri, got, tt.want)
		}
	}
}

// A signinMethods entry with Rule "Hide password" must not make the
// corresponding IsXEnabled() check return true — that rule exists
// specifically so an application can keep a method configured (e.g. to
// remember its settings) while actually disabling it. Before this fix,
// IsPasswordEnabled/IsLdapEnabled/IsFaceIdEnabled only checked whether an
// entry with the right Name existed, ignoring Rule entirely.
func TestSigninMethodHiddenByRuleIsNotEnabled(t *testing.T) {
	tests := []struct {
		name          string
		signinMethods []*SigninMethod
		enabledMethod func(*Application) bool
		wantEnabled   bool
	}{
		{
			name:          "Password hidden via rule is not enabled",
			signinMethods: []*SigninMethod{{Name: "Password", DisplayName: "Password", Rule: "Hide password"}},
			enabledMethod: (*Application).IsPasswordEnabled,
			wantEnabled:   false,
		},
		{
			name:          "Password with rule All is enabled",
			signinMethods: []*SigninMethod{{Name: "Password", DisplayName: "Password", Rule: "All"}},
			enabledMethod: (*Application).IsPasswordEnabled,
			wantEnabled:   true,
		},
		{
			name:          "LDAP hidden via rule is not enabled",
			signinMethods: []*SigninMethod{{Name: "LDAP", DisplayName: "LDAP", Rule: "Hide password"}},
			enabledMethod: (*Application).IsLdapEnabled,
			wantEnabled:   false,
		},
		{
			name:          "LDAP without hide rule is enabled",
			signinMethods: []*SigninMethod{{Name: "LDAP", DisplayName: "LDAP", Rule: "None"}},
			enabledMethod: (*Application).IsLdapEnabled,
			wantEnabled:   true,
		},
		{
			name:          "Face ID hidden via rule is not enabled",
			signinMethods: []*SigninMethod{{Name: "Face ID", DisplayName: "Face ID", Rule: "Hide password"}},
			enabledMethod: (*Application).IsFaceIdEnabled,
			wantEnabled:   false,
		},
		{
			name:          "Face ID without hide rule is enabled",
			signinMethods: []*SigninMethod{{Name: "Face ID", DisplayName: "Face ID", Rule: "None"}},
			enabledMethod: (*Application).IsFaceIdEnabled,
			wantEnabled:   true,
		},
		{
			name:          "Password hidden via the legacy rule value is not enabled",
			signinMethods: []*SigninMethod{{Name: "Password", DisplayName: "Password", Rule: "Hide-Password"}},
			enabledMethod: (*Application).IsPasswordEnabled,
			wantEnabled:   false,
		},
		{
			name:          "LDAP hidden via the legacy rule value is not enabled",
			signinMethods: []*SigninMethod{{Name: "LDAP", DisplayName: "LDAP", Rule: "Hide-Password"}},
			enabledMethod: (*Application).IsLdapEnabled,
			wantEnabled:   false,
		},
		{
			name:          "Face ID hidden via the legacy rule value is not enabled",
			signinMethods: []*SigninMethod{{Name: "Face ID", DisplayName: "Face ID", Rule: "Hide-Password"}},
			enabledMethod: (*Application).IsFaceIdEnabled,
			wantEnabled:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			application := &Application{SigninMethods: tt.signinMethods}
			got := tt.enabledMethod(application)
			if got != tt.wantEnabled {
				t.Errorf("got enabled=%v, want %v", got, tt.wantEnabled)
			}
		})
	}
}

// IsPasswordEnabled() falls back to the deprecated EnablePassword field when the
// application has no signin method at all, the reuse of HasSigninMethod() must not
// break that fallback.
func TestIsPasswordEnabledFallsBackToEnablePassword(t *testing.T) {
	application := &Application{EnablePassword: true}
	if !application.IsPasswordEnabled() {
		t.Errorf("got enabled=false, want true")
	}

	application = &Application{EnablePassword: false}
	if application.IsPasswordEnabled() {
		t.Errorf("got enabled=true, want false")
	}
}

// The "Hide password" rule was named "Hide-Password" when it was added, the legacy
// value stored by the applications configured before the rename is normalized on
// read so that the frontend and the backend agree on what is hidden.
func TestExtendApplicationWithSigninMethodsNormalizesLegacyHideRule(t *testing.T) {
	application := &Application{
		SigninMethods: []*SigninMethod{
			{Name: "Password", DisplayName: "Password", Rule: "Hide-Password"},
			{Name: "Verification code", DisplayName: "Verification code", Rule: "All"},
		},
	}

	err := extendApplicationWithSigninMethods(application)
	if err != nil {
		t.Fatalf("extendApplicationWithSigninMethods() error = %v", err)
	}

	if application.SigninMethods[0].Rule != "Hide password" {
		t.Errorf("got rule=%s, want %s", application.SigninMethods[0].Rule, "Hide password")
	}
	if application.SigninMethods[1].Rule != "All" {
		t.Errorf("got rule=%s, want %s", application.SigninMethods[1].Rule, "All")
	}
}
