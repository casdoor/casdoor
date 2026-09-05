// Copyright 2021 The Casdoor Authors. All Rights Reserved.
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
	"testing"

	"github.com/markbates/goth"
)

// goth's own Apple provider never populates Name/FirstName/LastName/NickName
// on the goth.User it returns -- its Session only decodes the ID token,
// which carries email but never a name (see providers/apple/session.go
// upstream). So every real Apple FetchUser result looks like this: only
// UserID and Email are ever set.
func appleGothUser(email string) goth.User {
	return goth.User{
		Provider: "apple",
		UserID:   "001837.abc123def456.7890",
		Email:    email,
	}
}

func TestGetUserAppleWithoutCachedInfo(t *testing.T) {
	// The pre-fix behavior, and still the correct behavior for a RETURNING
	// Apple user: no form_post payload arrives on any authorization after the
	// first, so there is nothing in AppleUserInfoCache, and this must keep
	// falling back to inventing a username from the email exactly as before.
	state := "state-no-cache-entry"
	user := getUser(appleGothUser("abc123@privaterelay.appleid.com"), "apple", state)

	if user.Username != "abc123" {
		t.Errorf("Username = %q, want %q (derived from email)", user.Username, "abc123")
	}
	if user.Email != "abc123@privaterelay.appleid.com" {
		t.Errorf("Email = %q, want the ID token email unchanged", user.Email)
	}
}

func TestGetUserAppleConsumesCachedNameOnce(t *testing.T) {
	// The bug this fixes: on a FIRST authorization with the name/email scopes
	// requested, controllers.Callback has already stashed Apple's one-time
	// `user` payload under the OAuth state. getUser must prefer it over the
	// invented-from-email fallback, for BOTH DisplayName and Username.
	state := "state-first-authorization"
	AppleUserInfoCache.Store(state, AppleUserInfo{
		FirstName: "Stu",
		LastName:  "Alexander",
		Email:     "stu@example.com",
	})

	user := getUser(appleGothUser("abc123@privaterelay.appleid.com"), "apple", state)

	if user.DisplayName != "Stu Alexander" {
		t.Errorf("DisplayName = %q, want %q", user.DisplayName, "Stu Alexander")
	}
	if user.Username != "Stu Alexander" {
		t.Errorf("Username = %q, want %q", user.Username, "Stu Alexander")
	}
	if user.Email != "stu@example.com" {
		t.Errorf("Email = %q, want the cached callback email, not the (possibly private-relay) ID token one", user.Email)
	}

	// Apple never resends this payload -- a second call for the same state
	// (e.g. a retried or duplicated request) must not find it twice and must
	// fall back exactly as if it had never been cached.
	again := getUser(appleGothUser("abc123@privaterelay.appleid.com"), "apple", state)
	if again.Username != "abc123" {
		t.Errorf("second call: Username = %q, want %q (cache entry must be single-use)", again.Username, "abc123")
	}
}

func TestGetUserAppleIgnoresCacheUnderAnotherState(t *testing.T) {
	// A cache entry must only ever be picked up by the login attempt that
	// carries the SAME state it was stored under -- never by an unrelated one.
	AppleUserInfoCache.Store("state-for-someone-else", AppleUserInfo{
		FirstName: "Someone",
		LastName:  "Else",
	})

	user := getUser(appleGothUser("abc123@privaterelay.appleid.com"), "apple", "state-this-attempt")

	if user.Username != "abc123" {
		t.Errorf("Username = %q, want %q (must not adopt another state's cached name)", user.Username, "abc123")
	}
}
