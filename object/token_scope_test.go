// Copyright 2024 The Casdoor Authors. All Rights Reserved.
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
	"sort"
	"strings"
	"testing"
)

func newApp(scopes ...string) *Application {
	items := make([]*ScopeItem, len(scopes))
	for i, s := range scopes {
		items[i] = &ScopeItem{Name: s}
	}
	return &Application{Scopes: items}
}

func sortedScope(s string) string {
	parts := strings.Fields(s)
	sort.Strings(parts)
	return strings.Join(parts, " ")
}

func TestExpandScope_NoAppScopes(t *testing.T) {
	app := &Application{}
	got, ok := ExpandScope("anything", app)
	if !ok || got != "anything" {
		t.Errorf("expected ('anything', true), got (%q, %v)", got, ok)
	}
}

func TestExpandScope_EmptyScope(t *testing.T) {
	app := newApp("read", "write")
	got, ok := ExpandScope("", app)
	if !ok || got != "" {
		t.Errorf("expected ('', true), got (%q, %v)", got, ok)
	}
}

func TestExpandScope_ExactMatch(t *testing.T) {
	app := newApp("read", "write", "admin")
	got, ok := ExpandScope("read write", app)
	if !ok {
		t.Fatalf("expected valid, got invalid")
	}
	if got != "read write" {
		t.Errorf("expected 'read write', got %q", got)
	}
}

func TestExpandScope_ExactMatchInvalid(t *testing.T) {
	app := newApp("read", "write")
	_, ok := ExpandScope("read delete", app)
	if ok {
		t.Error("expected invalid scope")
	}
}

func TestExpandScope_WildcardDot(t *testing.T) {
	app := newApp("payment.t1.read", "payment.t1.write", "payment.t2.read")
	got, ok := ExpandScope("payment.t1.*", app)
	if !ok {
		t.Fatalf("expected valid, got invalid")
	}
	if sortedScope(got) != "payment.t1.read payment.t1.write" {
		t.Errorf("expected 'payment.t1.read payment.t1.write', got %q", got)
	}
}

func TestExpandScope_WildcardNoMatch(t *testing.T) {
	app := newApp("payment.t1.read", "payment.t1.write")
	_, ok := ExpandScope("payment.t3.*", app)
	if ok {
		t.Error("expected invalid scope when wildcard matches nothing")
	}
}

func TestExpandScope_RegexPattern(t *testing.T) {
	app := newApp("payment.t1.read", "payment.t1.write", "payment.t2.read", "order.create")
	got, ok := ExpandScope("payment\\.t[12]\\.read", app)
	if !ok {
		t.Fatalf("expected valid, got invalid")
	}
	if sortedScope(got) != "payment.t1.read payment.t2.read" {
		t.Errorf("expected 'payment.t1.read payment.t2.read', got %q", got)
	}
}

func TestExpandScope_MixedLiteralAndRegex(t *testing.T) {
	app := newApp("payment.t1.read", "payment.t1.write", "order.create")
	got, ok := ExpandScope("order.create payment.t1.*", app)
	if !ok {
		t.Fatalf("expected valid, got invalid")
	}
	if sortedScope(got) != "order.create payment.t1.read payment.t1.write" {
		t.Errorf("expected 'order.create payment.t1.read payment.t1.write', got %q", got)
	}
}

func TestExpandScope_Deduplication(t *testing.T) {
	app := newApp("payment.t1.read", "payment.t1.write")
	got, ok := ExpandScope("payment.t1.read payment.t1.*", app)
	if !ok {
		t.Fatalf("expected valid, got invalid")
	}
	// payment.t1.read should appear only once
	if sortedScope(got) != "payment.t1.read payment.t1.write" {
		t.Errorf("expected 'payment.t1.read payment.t1.write', got %q", got)
	}
}

func TestExpandScope_InvalidRegex(t *testing.T) {
	app := newApp("read", "write")
	_, ok := ExpandScope("[invalid", app)
	if ok {
		t.Error("expected invalid scope for bad regex")
	}
}

func TestExpandScope_SecurityNoEscalation(t *testing.T) {
	app := newApp("payment.t1.read", "payment.t1.write")
	// .* should only match scopes configured in the app
	got, ok := ExpandScope(".*", app)
	if !ok {
		t.Fatalf("expected valid, got invalid")
	}
	if sortedScope(got) != "payment.t1.read payment.t1.write" {
		t.Errorf("expected all app scopes, got %q", got)
	}
}

func TestIsScopeValid_BackwardCompat(t *testing.T) {
	app := newApp("read", "write")
	if !IsScopeValid("read", app) {
		t.Error("expected valid")
	}
	if IsScopeValid("delete", app) {
		t.Error("expected invalid")
	}
}

func TestIsScopeValid_WildcardSupport(t *testing.T) {
	app := newApp("payment.t1.read", "payment.t1.write", "payment.t2.read")
	if !IsScopeValid("payment.t1.*", app) {
		t.Error("expected valid for wildcard matching configured scopes")
	}
	if IsScopeValid("payment.t3.*", app) {
		t.Error("expected invalid for wildcard matching no configured scopes")
	}
}

func TestExpandScope_DottedLiteralExactMatch(t *testing.T) {
	// payment.t1.read should match exactly, not as a regex
	app := newApp("payment.t1.read", "paymentXt1Xread")
	got, ok := ExpandScope("payment.t1.read", app)
	if !ok {
		t.Fatalf("expected valid, got invalid")
	}
	// Should only contain the exact match, not the regex-matched variant
	if got != "payment.t1.read" {
		t.Errorf("expected 'payment.t1.read', got %q", got)
	}
}

func TestIsRegexScope(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"read", false},
		{"payment.t1.read", true}, // '.' is a regex metacharacter
		{"payment.t1.*", true},
		{"payment\\.t1\\..*", true},
		{"scope[12]", true},
		{"scope+", true},
		{"scope?", true},
	}
	for _, tt := range tests {
		got := isRegexScope(tt.input)
		if got != tt.want {
			t.Errorf("isRegexScope(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}
