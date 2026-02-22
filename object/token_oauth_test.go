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

func makeApp(scopeNames ...string) *Application {
	scopes := make([]*ScopeItem, len(scopeNames))
	for i, name := range scopeNames {
		scopes[i] = &ScopeItem{Name: name}
	}
	return &Application{Scopes: scopes}
}

func TestMatchScopePattern_ExactMatch(t *testing.T) {
	configured := []string{"read", "write", "admin"}

	got := matchScopePattern("read", configured)
	if len(got) != 1 || got[0] != "read" {
		t.Fatalf("exact match: expected [read], got %v", got)
	}
}

func TestMatchScopePattern_RegexMatch(t *testing.T) {
	configured := []string{"payment.t1.read", "payment.t1.write", "payment.t2.read"}

	got := matchScopePattern("payment\\.t1\\..*", configured)
	if len(got) != 2 {
		t.Fatalf("regex match: expected 2 matches, got %v", got)
	}
}

func TestMatchScopePattern_WildcardStar(t *testing.T) {
	configured := []string{"payment.t1.read", "payment.t1.write", "payment.t2.read"}

	// "payment.t1.*" as a regex: unescaped '.' matches any single char and '*'
	// means zero-or-more of the preceding element ('.'), so "payment.t1.*"
	// matches anything that starts with "payment" + any_char + "t1".
	// Against the configured list this matches "payment.t1.read" and
	// "payment.t1.write" but not "payment.t2.read".
	got := matchScopePattern("payment.t1.*", configured)
	if len(got) != 2 {
		t.Fatalf("wildcard star: expected 2 matches, got %v", got)
	}
}

func TestMatchScopePattern_NoMatch(t *testing.T) {
	configured := []string{"read", "write"}
	got := matchScopePattern("delete", configured)
	if len(got) != 0 {
		t.Fatalf("no match: expected [], got %v", got)
	}
}

func TestMatchScopePattern_InvalidRegex(t *testing.T) {
	configured := []string{"read"}
	got := matchScopePattern("[invalid", configured)
	if len(got) != 0 {
		t.Fatalf("invalid regex: expected [], got %v", got)
	}
}

func TestExpandScope_ExactScopes(t *testing.T) {
	app := makeApp("read", "write", "admin")

	expanded, valid := ExpandScope("read write", app)
	if !valid {
		t.Fatal("expected valid")
	}
	if expanded != "read write" {
		t.Fatalf("expected 'read write', got %q", expanded)
	}
}

func TestExpandScope_WildcardExpansion(t *testing.T) {
	app := makeApp("payment.t1.read", "payment.t1.write", "payment.t2.read")

	// "payment.t1.*" should expand to "payment.t1.read payment.t1.write"
	expanded, valid := ExpandScope("payment.t1.*", app)
	if !valid {
		t.Fatal("expected valid")
	}
	if expanded != "payment.t1.read payment.t1.write" {
		t.Fatalf("unexpected expansion: %q", expanded)
	}
}

func TestExpandScope_RegexExpansion(t *testing.T) {
	app := makeApp("payment.t1.read", "payment.t1.write", "payment.t2.read")

	expanded, valid := ExpandScope("payment\\.t1\\..*", app)
	if !valid {
		t.Fatal("expected valid")
	}
	if expanded != "payment.t1.read payment.t1.write" {
		t.Fatalf("unexpected expansion: %q", expanded)
	}
}

func TestExpandScope_InvalidScope(t *testing.T) {
	app := makeApp("read", "write")

	_, valid := ExpandScope("delete", app)
	if valid {
		t.Fatal("expected invalid scope")
	}
}

func TestExpandScope_WildcardNoMatch(t *testing.T) {
	app := makeApp("payment.t1.read", "payment.t1.write")

	_, valid := ExpandScope("payment.t2.*", app)
	if valid {
		t.Fatal("expected invalid scope when wildcard matches nothing")
	}
}

func TestExpandScope_EmptyScope(t *testing.T) {
	app := makeApp("read", "write")

	expanded, valid := ExpandScope("", app)
	if !valid {
		t.Fatal("empty scope should be valid")
	}
	if expanded != "" {
		t.Fatalf("expected empty string, got %q", expanded)
	}
}

func TestExpandScope_NoAppScopes(t *testing.T) {
	app := &Application{Scopes: nil}

	expanded, valid := ExpandScope("anything", app)
	if !valid {
		t.Fatal("should be valid when app has no configured scopes")
	}
	if expanded != "anything" {
		t.Fatalf("expected unchanged scope, got %q", expanded)
	}
}

func TestExpandScope_DeduplicatesScopes(t *testing.T) {
	app := makeApp("read", "write")

	// Requesting the same scope twice should deduplicate
	expanded, valid := ExpandScope("read read", app)
	if !valid {
		t.Fatal("expected valid")
	}
	if expanded != "read" {
		t.Fatalf("expected deduplicated 'read', got %q", expanded)
	}
}

func TestExpandScope_MixedExactAndWildcard(t *testing.T) {
	app := makeApp("payment.t1.read", "payment.t1.write", "payment.t2.read", "admin")

	expanded, valid := ExpandScope("admin payment.t1.*", app)
	if !valid {
		t.Fatal("expected valid")
	}
	if expanded != "admin payment.t1.read payment.t1.write" {
		t.Fatalf("unexpected expansion: %q", expanded)
	}
}

func TestIsScopeValid_ExactMatch(t *testing.T) {
	app := makeApp("read", "write")
	if !IsScopeValid("read write", app) {
		t.Fatal("expected valid")
	}
}

func TestIsScopeValid_WildcardValid(t *testing.T) {
	app := makeApp("payment.t1.read", "payment.t1.write")
	if !IsScopeValid("payment.t1.*", app) {
		t.Fatal("expected valid for wildcard that matches configured scopes")
	}
}

func TestIsScopeValid_WildcardNoMatch(t *testing.T) {
	app := makeApp("payment.t1.read")
	if IsScopeValid("payment.t2.*", app) {
		t.Fatal("expected invalid for wildcard that matches nothing")
	}
}
