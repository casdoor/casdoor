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

// Every history entry costs a hash comparison (bcrypt/argon2) on each password change.
const maxPasswordHistoryCount = 24

type PasswordHistoryEntry struct {
	Password     string `json:"password"`
	PasswordType string `json:"passwordType"`
	PasswordSalt string `json:"passwordSalt"`
}

// getPasswordHistoryCount returns how many recent passwords, the current one included, cannot be reused.
func getPasswordHistoryCount(organization *Organization) int {
	return max(min(organization.PasswordHistoryCount, maxPasswordHistoryCount), 1)
}

func isSamePassword(password string, hashedPassword string, passwordType string, salt string, organization *Organization) bool {
	if hashedPassword == "" {
		return false
	}

	if passwordType == "" {
		passwordType = organization.PasswordType
	}
	credManager := cred.GetCredManager(passwordType)
	if credManager == nil {
		return false
	}

	if credManager.IsPasswordCorrect(password, hashedPassword, salt) {
		return true
	}
	return salt != organization.PasswordSalt && credManager.IsPasswordCorrect(password, hashedPassword, organization.PasswordSalt)
}

func CheckPasswordReuse(user *User, newPassword string, organization *Organization, lang string) string {
	if isSamePassword(newPassword, user.Password, user.PasswordType, user.PasswordSalt, organization) {
		return i18n.Translate(lang, "user:The new password must be different from your current password")
	}

	count := getPasswordHistoryCount(organization)
	for i, entry := range user.PasswordHistory {
		if i >= count-1 {
			break
		}
		if isSamePassword(newPassword, entry.Password, entry.PasswordType, entry.PasswordSalt, organization) {
			return fmt.Sprintf(i18n.Translate(lang, "user:The new password must be different from your last %d passwords"), count)
		}
	}

	return ""
}

// AddPasswordHistory must be called before the new password replaces user.Password.
func (user *User) AddPasswordHistory(organization *Organization) {
	limit := getPasswordHistoryCount(organization) - 1
	if limit == 0 {
		user.PasswordHistory = nil
		return
	}

	history := user.PasswordHistory
	if user.Password != "" {
		passwordType := user.PasswordType
		if passwordType == "" {
			passwordType = organization.PasswordType
		}
		entry := &PasswordHistoryEntry{Password: user.Password, PasswordType: passwordType, PasswordSalt: user.PasswordSalt}
		history = append([]*PasswordHistoryEntry{entry}, history...)
	}

	if len(history) > limit {
		history = history[:limit]
	}
	user.PasswordHistory = history
}
