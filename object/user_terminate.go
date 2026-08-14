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

import "github.com/casdoor/casdoor/util"

// shouldTerminateUserAccess reports whether an update should end the user's
// sessions and tokens (account forbidden or soft-deleted).
// When columns is nil, all fields are considered (full update).
func shouldTerminateUserAccess(oldUser, newUser *User, columns []string) bool {
	if oldUser == nil || newUser == nil {
		return false
	}

	checkForbidden := columns == nil || util.InSlice(columns, "is_forbidden")
	if checkForbidden && !oldUser.IsForbidden && newUser.IsForbidden {
		return true
	}

	checkDeleted := columns == nil || util.InSlice(columns, "is_deleted")
	if checkDeleted && !oldUser.IsDeleted && newUser.IsDeleted {
		return true
	}

	return false
}

// TerminateUserAccess expires all OAuth tokens and deletes all sessions for the
// user, mirroring SSO logout-all. host is used for OIDC back-channel logout
// issuer resolution and may be empty (falls back to configured origin).
func TerminateUserAccess(organization, username, host string) error {
	tokens, err := GetTokensByUser(organization, username)
	if err != nil {
		return err
	}

	// Back-channel logout must run before ExpireTokenByUser: it only considers
	// tokens with expires_in > 0.
	SendBackchannelLogout(organization, username, "", host)

	_, err = ExpireTokenByUser(organization, username)
	if err != nil {
		return err
	}

	sessions, err := GetUserSessions(organization, username)
	if err != nil {
		return err
	}

	sessionIds := make([]string, 0)
	for _, session := range sessions {
		sessionIds = append(sessionIds, session.SessionId...)
	}

	_, err = DeleteAllUserSessions(organization, username)
	if err != nil {
		return err
	}

	user, err := getUser(organization, username)
	if err != nil {
		return err
	}
	if user != nil {
		// Notification failure must not undo forbid/delete: access is already revoked.
		_ = SendSsoLogoutNotifications(user, sessionIds, tokens)
	}

	return nil
}
