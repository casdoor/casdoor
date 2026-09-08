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
	"fmt"

	"github.com/casdoor/casdoor/cred"
	"github.com/casdoor/casdoor/i18n"
)

func resolvePasswordType(storedType, organizationType string) string {
	if storedType != "" {
		return storedType
	}
	return organizationType
}

func isPasswordMatchingHash(plainPassword, hashedPassword, passwordType, primarySalt, fallbackSalt string) (matched bool, typeUnsupported bool) {
	credManager := cred.GetCredManager(passwordType)
	if credManager == nil {
		return false, true
	}
	matched = credManager.IsPasswordCorrect(plainPassword, hashedPassword, primarySalt) ||
		credManager.IsPasswordCorrect(plainPassword, hashedPassword, fallbackSalt)
	return matched, false
}

func CheckPasswordHistory(user *User, newPassword string, organization *Organization, lang string) error {
	if user == nil || organization == nil {
		return nil
	}
	if user.Password == "" {
		return nil
	}

	currentReuseMessage := i18n.Translate(lang, "user:The new password must be different from your current password")
	historyReuseMessage := i18n.Translate(lang, "user:The new password must not match any of your recent passwords")

	currentType := resolvePasswordType(user.PasswordType, organization.PasswordType)
	matched, typeUnsupported := isPasswordMatchingHash(newPassword, user.Password, currentType, organization.PasswordSalt, user.PasswordSalt)
	if typeUnsupported {
		return fmt.Errorf(i18n.Translate(lang, "check:unsupported password type: %s"), currentType)
	}
	if matched {
		return fmt.Errorf("%s", currentReuseMessage)
	}

	if organization.PasswordHistoryCount <= 1 {
		return nil
	}

	for _, entry := range user.PasswordHistory {
		if entry.Hash == "" {
			continue
		}
		entryType := resolvePasswordType(entry.Type, organization.PasswordType)
		matched, typeUnsupported = isPasswordMatchingHash(newPassword, entry.Hash, entryType, organization.PasswordSalt, entry.Salt)
		if typeUnsupported {
			return fmt.Errorf(i18n.Translate(lang, "check:unsupported password type: %s"), entryType)
		}
		if matched {
			return fmt.Errorf("%s", historyReuseMessage)
		}
	}

	return nil
}

// UpdatePasswordHistory snapshots the current stored hash before UpdateUserPassword replaces it.
func UpdatePasswordHistory(user *User, organization *Organization) {
	if user == nil || organization == nil || organization.PasswordHistoryCount <= 1 {
		return
	}
	if user.Password == "" {
		return
	}

	entry := PasswordHistoryEntry{
		Hash: user.Password,
		Type: resolvePasswordType(user.PasswordType, organization.PasswordType),
		Salt: user.PasswordSalt,
	}
	history := append([]PasswordHistoryEntry{entry}, user.PasswordHistory...)

	maxHistory := organization.PasswordHistoryCount - 1
	if len(history) > maxHistory {
		history = history[:maxHistory]
	}
	user.PasswordHistory = history
}
