// Copyright 2025 The Casdoor Authors. All Rights Reserved.
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
	"testing"
)

// makeApp builds a minimal Application whose Scopes list is populated from names.
func makeApp(names ...string) *Application {
	scopes := make([]*ScopeItem, 0, len(names))
	for _, n := range names {
		scopes = append(scopes, &ScopeItem{Name: n})
	}
	return &Application{Scopes: scopes}
}

func TestExpandScopesPatternMatching(t *testing.T) {
	// Tests focused on individual pattern types, using single-scope apps so we can
	// assert match/no-match for specific (pattern, scope-name) pairs.
	tests := []struct {
		name      string
		pattern   string
		appScopes []string
		wantOK    bool
	}{
		// --- Exact match ---
		{"exact match", "payment.t1.read", []string{"payment.t1.read"}, true},
		{"exact no match", "payment.t1.read", []string{"payment.t1.write"}, false},

		// --- Glob wildcard ---
		{"glob t1 matches read", "payment.t1.*", []string{"payment.t1.read"}, true},
		{"glob t1 matches write", "payment.t1.*", []string{"payment.t1.write"}, true},
		{"glob t1 no match t2", "payment.t1.*", []string{"payment.t2.read"}, false},
		// path.Match separator is '/', not '.', so '*' crosses dot boundaries
		{"glob star crosses dots", "payment.*", []string{"payment.t1.read"}, true},
		{"glob star matches all", "*", []string{"anything"}, true},
		{"glob prefix", "pay*.read", []string{"payment.read"}, true},
		{"glob prefix no match", "pay*.read", []string{"payment.write"}, false},

		// --- Regex ---
		{"regex matches t1 read", "/payment\\.t[12]\\..*/", []string{"payment.t1.read"}, true},
		{"regex matches t2 write", "/payment\\.t[12]\\..*/", []string{"payment.t2.write"}, true},
		{"regex no match t3", "/payment\\.t[12]\\..*/", []string{"payment.t3.read"}, false},
		{"regex suffix read", "/.*read$/", []string{"payment.t1.read"}, true},
		{"regex suffix write", "/.*read$/", []string{"payment.t1.write"}, false},
		// Invalid regex → no match (treated as invalid scope)
		{"invalid regex", "/[invalid/", []string{"anything"}, false},
	}

	for _, tt := range tests {
		app := makeApp(tt.appScopes...)
		_, gotOK := ExpandScopes(tt.pattern, app)
		if gotOK != tt.wantOK {
			t.Errorf("[%s] ExpandScopes(%q) ok = %v, want %v", tt.name, tt.pattern, gotOK, tt.wantOK)
		}
	}
}

func TestExpandScopes(t *testing.T) {
	app := makeApp("payment.t1.read", "payment.t1.write", "payment.t2.read", "profile", "openid")

	tests := []struct {
		name      string
		scope     string
		wantScope string
		wantOK    bool
	}{
		// Backward-compat: no scopes configured → original string returned
		{
			name:      "no scopes configured",
			scope:     "anything",
			wantScope: "anything",
			wantOK:    true,
		},
		// Empty scope → always valid
		{
			name:      "empty scope",
			scope:     "",
			wantScope: "",
			wantOK:    true,
		},
		// Exact match
		{
			name:      "exact single scope",
			scope:     "profile",
			wantScope: "profile",
			wantOK:    true,
		},
		{
			name:      "exact multiple scopes",
			scope:     "profile openid",
			wantScope: "profile openid",
			wantOK:    true,
		},
		{
			name:      "unknown exact scope",
			scope:     "unknown",
			wantScope: "",
			wantOK:    false,
		},
		// Wildcard
		{
			name:      "wildcard matches subset",
			scope:     "payment.t1.*",
			wantScope: "payment.t1.read payment.t1.write",
			wantOK:    true,
		},
		{
			name:      "wildcard matches all payment scopes",
			scope:     "payment.*.*",
			wantScope: "payment.t1.read payment.t1.write payment.t2.read",
			wantOK:    true,
		},
		{
			name:      "wildcard no match",
			scope:     "payment.t3.*",
			wantScope: "",
			wantOK:    false,
		},
		// Regex
		{
			name:      "regex matches subset",
			scope:     "/payment\\.t[12]\\.read/",
			wantScope: "payment.t1.read payment.t2.read",
			wantOK:    true,
		},
		{
			name:      "regex no match",
			scope:     "/payment\\.t3\\..*/",
			wantScope: "",
			wantOK:    false,
		},
		// Mixed patterns in one request
		{
			name:      "wildcard and exact",
			scope:     "payment.t1.* profile",
			wantScope: "payment.t1.read payment.t1.write profile",
			wantOK:    true,
		},
		// Deduplication: same concrete scope requested by two patterns
		{
			name:      "deduplication across patterns",
			scope:     "payment.t1.read payment.t1.*",
			wantScope: "payment.t1.read payment.t1.write",
			wantOK:    true,
		},
	}

	noScopesApp := &Application{Scopes: nil}

	for _, tt := range tests {
		var targetApp *Application
		if tt.name == "no scopes configured" {
			targetApp = noScopesApp
		} else {
			targetApp = app
		}

		gotScope, gotOK := ExpandScopes(tt.scope, targetApp)
		if gotOK != tt.wantOK {
			t.Errorf("[%s] ExpandScopes(%q) ok = %v, want %v", tt.name, tt.scope, gotOK, tt.wantOK)
			continue
		}
		if gotScope != tt.wantScope {
			t.Errorf("[%s] ExpandScopes(%q) scope = %q, want %q", tt.name, tt.scope, gotScope, tt.wantScope)
		}
	}
}

func TestIsScopeValid(t *testing.T) {
	app := makeApp("read", "write", "admin")

	tests := []struct {
		scope string
		want  bool
	}{
		{"read", true},
		{"read write", true},
		{"read write admin", true},
		{"unknown", false},
		{"read unknown", false},
		{"read*", true},    // 'read*' matches 'read' (path.Match '*' also matches empty string)
		{"rea*", true},      // 'rea*' matches 'read' via path.Match
		{"read wri*", true}, // mix of exact and wildcard
		{"/rea.*/", true},   // regex matching 'read'
		{"/xyz.*/", false},  // regex matching nothing
	}

	for _, tt := range tests {
		got := IsScopeValid(tt.scope, app)
		if got != tt.want {
			t.Errorf("IsScopeValid(%q) = %v, want %v", tt.scope, got, tt.want)
		}
	}
}
