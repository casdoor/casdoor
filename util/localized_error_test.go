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

package util

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestMessageErrorEnglishAndLocalizedArguments(t *testing.T) {
	argument := "100%: general:User not found <example>"
	args := []interface{}{argument}
	err := MessageErrorf("user:User %s is unavailable", args...)
	args[0] = "changed"
	if got := err.Error(); got != "User "+argument+" is unavailable" {
		t.Fatalf("English diagnostic changed: %q", got)
	}
	calls := []string{}
	got := FormatMessageError(err, func(key string) string {
		calls = append(calls, key)
		return "使用者 %s 無法使用"
	})
	if got != "使用者 "+argument+" 無法使用" || len(calls) != 1 || calls[0] != "user:User %s is unavailable" {
		t.Fatalf("an argument was interpreted or modified: %q, %v", got, calls)
	}
}

func TestMessageErrorPreservesPercentLiteralAndCauses(t *testing.T) {
	plain := NewMessageError("check:100% complete")
	if plain.Error() != "100% complete" || FormatMessageError(plain, func(string) string { return "完成 100%" }) != "完成 100%" {
		t.Fatal("errors.New semantics were not preserved")
	}
	sentinel := errors.New("vendor diagnostic")
	wrapped := MessageErrorf("provider:Cannot load: %w", sentinel)
	if wrapped.Error() != fmt.Errorf("Cannot load: %w", sentinel).Error() || !errors.Is(wrapped, sentinel) {
		t.Fatal("fmt.Errorf %w semantics were not preserved")
	}
	if got := FormatMessageError(wrapped, func(string) string { return "無法載入：%w" }); got != "無法載入：vendor diagnostic" {
		t.Fatalf("unexpected presentation: %q", got)
	}
	attached := WrapMessageError(sentinel, "provider:Provider is unavailable")
	if attached.Error() != "Provider is unavailable" || !errors.Is(attached, sentinel) {
		t.Fatal("explicit cause was lost or leaked into the message")
	}
}

func TestMessageErrorDoesNotTranslateUnknownWrappersOrDiagnostics(t *testing.T) {
	inner := NewMessageError("user:User not found")
	unknown := fmt.Errorf("additional context: %w", inner)
	for _, err := range []error{unknown, errors.New("user:User not found"), errors.New("User not found")} {
		if IsMessageError(err) {
			t.Fatalf("unknown error was classified as first party: %T", err)
		}
		if got := FormatMessageError(err, func(string) string { t.Fatal("unexpected translation"); return "" }); got != err.Error() {
			t.Fatalf("unknown context was discarded: %q", got)
		}
	}
	if FormatMessageError(nil, func(string) string { t.Fatal("unexpected translation"); return "" }) != "" {
		t.Fatal("nil must remain empty")
	}
}

func TestMessageErrorNestedAndJoinedProvenance(t *testing.T) {
	sentinel := NewMessageError("mfa:Invalid code")
	outer := MessageErrorf("mfa:Verification failed: %w", sentinel)
	joined := JoinMessageErrors(", ", outer, nil, errors.New("opaque detail"))
	translations := map[string]string{"mfa:Invalid code": "驗證碼錯誤", "mfa:Verification failed: %w": "驗證失敗：%w"}
	got := FormatMessageError(joined, func(key string) string { return translations[key] })
	if got != "驗證失敗：驗證碼錯誤, opaque detail" || !errors.Is(joined, sentinel) {
		t.Fatalf("joined message or cause changed: %q", got)
	}
	if !strings.Contains(joined.Error(), "Verification failed: Invalid code") || !IsMessageError(joined) {
		t.Fatal("English provenance was not retained")
	}
	if JoinMessageErrors(", ", nil, nil) != nil {
		t.Fatal("an empty join must be nil")
	}
}
