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
	"github.com/xorm-io/core"
)

// Every history entry costs a hash comparison (bcrypt/argon2) on each password change.
const maxPasswordHistoryCount = 24

type PasswordHistoryEntry struct {
	Password     string `json:"password"`
	PasswordType string `json:"passwordType"`
	PasswordSalt string `json:"passwordSalt"`
}

// PasswordHistory lives in its own table on purpose: the `user` table already sits at
// InnoDB's 8126-byte row-size limit on MySQL 8.4.9+/9.7.0+, so adding any column to it
// fails there (casdoor/casdoor#5844).
type PasswordHistory struct {
	Owner   string                  `xorm:"varchar(100) notnull pk" json:"owner"`
	Name    string                  `xorm:"varchar(255) notnull pk" json:"name"`
	Entries []*PasswordHistoryEntry `xorm:"mediumtext" json:"entries"`
}

func getPasswordHistory(owner string, name string) (*PasswordHistory, bool, error) {
	history := PasswordHistory{Owner: owner, Name: name}
	existed, err := ormer.Engine.Get(&history)
	if err != nil {
		return nil, false, err
	}
	if !existed {
		return &PasswordHistory{Owner: owner, Name: name}, false, nil
	}
	return &history, true, nil
}

func savePasswordHistory(history *PasswordHistory, existed bool) error {
	var err error
	if len(history.Entries) == 0 {
		err = DeletePasswordHistoryByUser(history.Owner, history.Name)
	} else if existed {
		_, err = ormer.Engine.ID(core.PK{history.Owner, history.Name}).AllCols().Update(history)
	} else {
		_, err = ormer.Engine.Insert(history)
	}
	return err
}

func DeletePasswordHistoryByUser(owner string, name string) error {
	_, err := ormer.Engine.ID(core.PK{owner, name}).Delete(&PasswordHistory{})
	return err
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

func CheckPasswordReuse(user *User, newPassword string, organization *Organization, lang string) (string, error) {
	if isSamePassword(newPassword, user.Password, user.PasswordType, user.PasswordSalt, organization) {
		return i18n.Translate(lang, "user:The new password must be different from your current password"), nil
	}

	count := getPasswordHistoryCount(organization)
	if count == 1 {
		return "", nil
	}

	history, _, err := getPasswordHistory(user.Owner, user.Name)
	if err != nil {
		return "", err
	}
	for i, entry := range history.Entries {
		if i >= count-1 {
			break
		}
		if isSamePassword(newPassword, entry.Password, entry.PasswordType, entry.PasswordSalt, organization) {
			return fmt.Sprintf(i18n.Translate(lang, "user:The new password must be different from your last %d passwords"), count), nil
		}
	}

	return "", nil
}

// AddPasswordHistory records user.Password, so it must be called before the new password replaces it.
func (user *User) AddPasswordHistory(organization *Organization) error {
	limit := getPasswordHistoryCount(organization) - 1
	if limit == 0 {
		return DeletePasswordHistoryByUser(user.Owner, user.Name)
	}

	history, existed, err := getPasswordHistory(user.Owner, user.Name)
	if err != nil {
		return err
	}

	if user.Password != "" {
		passwordType := user.PasswordType
		if passwordType == "" {
			passwordType = organization.PasswordType
		}
		entry := &PasswordHistoryEntry{Password: user.Password, PasswordType: passwordType, PasswordSalt: user.PasswordSalt}
		history.Entries = append([]*PasswordHistoryEntry{entry}, history.Entries...)
	}

	if len(history.Entries) > limit {
		history.Entries = history.Entries[:limit]
	}
	return savePasswordHistory(history, existed)
}
