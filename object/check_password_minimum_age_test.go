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
	"testing"
	"time"

	"github.com/casdoor/casdoor/util"
)

func withFixedPolicyNow(t *testing.T, now time.Time, fn func()) {
	t.Helper()
	original := getPasswordPolicyNow
	getPasswordPolicyNow = func() time.Time {
		return now
	}
	t.Cleanup(func() {
		getPasswordPolicyNow = original
	})
	fn()
}

func TestCheckMinimumPasswordAgeDisabledWhenZero(t *testing.T) {
	user := &User{LastChangePasswordTime: util.GetCurrentTime()}
	org := &Organization{MinimumPasswordAgeHours: 0}
	if err := CheckMinimumPasswordAge(user, org, PasswordOpUserChange, "en"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestCheckMinimumPasswordAgeSkippedWhenLastChangeEmpty(t *testing.T) {
	user := &User{LastChangePasswordTime: ""}
	org := &Organization{MinimumPasswordAgeHours: 24}
	if err := CheckMinimumPasswordAge(user, org, PasswordOpUserChange, "en"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestCheckMinimumPasswordAgeRecoverySkipped(t *testing.T) {
	fixedNow := time.Date(2026, 8, 4, 12, 0, 0, 0, time.UTC)
	lastChange := fixedNow.Add(-time.Hour)
	user := &User{LastChangePasswordTime: util.Time2String(lastChange)}
	org := &Organization{MinimumPasswordAgeHours: 24}

	withFixedPolicyNow(t, fixedNow, func() {
		if err := CheckMinimumPasswordAge(user, org, PasswordOpRecovery, "en"); err != nil {
			t.Fatalf("expected no error for recovery, got %v", err)
		}
	})
}

func TestCheckMinimumPasswordAgeBlocksTooSoon(t *testing.T) {
	fixedNow := time.Date(2026, 8, 4, 12, 0, 0, 0, time.UTC)
	lastChange := fixedNow.Add(-time.Hour)
	user := &User{LastChangePasswordTime: util.Time2String(lastChange)}
	org := &Organization{MinimumPasswordAgeHours: 24}

	withFixedPolicyNow(t, fixedNow, func() {
		err := CheckMinimumPasswordAge(user, org, PasswordOpUserChange, "en")
		if err == nil {
			t.Fatal("expected error when minimum age not expired")
		}
	})
}

func TestCheckMinimumPasswordAgeAllowsAfterExpiry(t *testing.T) {
	fixedNow := time.Date(2026, 8, 4, 12, 0, 0, 0, time.UTC)
	lastChange := fixedNow.Add(-25 * time.Hour)
	user := &User{LastChangePasswordTime: util.Time2String(lastChange)}
	org := &Organization{MinimumPasswordAgeHours: 24}

	withFixedPolicyNow(t, fixedNow, func() {
		if err := CheckMinimumPasswordAge(user, org, PasswordOpUserChange, "en"); err != nil {
			t.Fatalf("expected no error after minimum age, got %v", err)
		}
	})
}

func TestShouldApplyPasswordChangePolicyStripsSpoofedTimestamp(t *testing.T) {
	oldUser := &User{Password: "old-hash", LastChangePasswordTime: "2020-01-01T00:00:00Z"}
	user := &User{Password: "new-secret", LastChangePasswordTime: "2099-01-01T00:00:00Z"}

	if !shouldApplyPasswordChangePolicy(oldUser, user, nil) {
		t.Fatal("expected policy to apply for untrusted update")
	}
	if user.LastChangePasswordTime != oldUser.LastChangePasswordTime {
		t.Fatalf("expected client timestamp stripped, got %q", user.LastChangePasswordTime)
	}
}

func TestShouldApplyPasswordChangePolicySkipsPrevalidated(t *testing.T) {
	oldUser := &User{Password: "old-hash", LastChangePasswordTime: "2020-01-01T00:00:00Z"}
	user := &User{Password: "new-hash", LastChangePasswordTime: "2026-08-04T12:00:00Z"}
	pwCtx := &PasswordUpdateContext{Prevalidated: true}

	if shouldApplyPasswordChangePolicy(oldUser, user, pwCtx) {
		t.Fatal("expected policy skip for prevalidated password change")
	}
	if user.LastChangePasswordTime != "2026-08-04T12:00:00Z" {
		t.Fatalf("prevalidated path must keep server-set timestamp, got %q", user.LastChangePasswordTime)
	}
}

func TestShouldApplyPasswordChangePolicySystemRehashKeepsOldTimestamp(t *testing.T) {
	oldUser := &User{Password: "old-hash", LastChangePasswordTime: "2020-01-01T00:00:00Z"}
	user := &User{Password: "new-hash", LastChangePasswordTime: "2026-08-04T12:00:00Z"}
	pwCtx := &PasswordUpdateContext{SystemRehash: true}

	if shouldApplyPasswordChangePolicy(oldUser, user, pwCtx) {
		t.Fatal("expected policy skip for system rehash")
	}
	if user.LastChangePasswordTime != oldUser.LastChangePasswordTime {
		t.Fatalf("system rehash must preserve old timestamp, got %q", user.LastChangePasswordTime)
	}
}

func TestCheckMinimumPasswordAgeAdminChangeEnforced(t *testing.T) {
	fixedNow := time.Date(2026, 8, 4, 12, 0, 0, 0, time.UTC)
	lastChange := fixedNow.Add(-30 * time.Minute)
	user := &User{LastChangePasswordTime: util.Time2String(lastChange)}
	org := &Organization{MinimumPasswordAgeHours: 1}

	withFixedPolicyNow(t, fixedNow, func() {
		err := CheckMinimumPasswordAge(user, org, PasswordOpAdminChange, "en")
		if err == nil {
			t.Fatal("expected error for admin change within minimum age")
		}
	})
}
