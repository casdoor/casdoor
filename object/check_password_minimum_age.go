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
	"errors"
	"time"

	"github.com/casdoor/casdoor/i18n"
	"github.com/casdoor/casdoor/util"
)

type PasswordChangeOperation int

const (
	PasswordOpUserChange PasswordChangeOperation = iota
	PasswordOpAdminChange
	PasswordOpRecovery
	PasswordOpInitialSet
	PasswordOpSyncImport
	PasswordOpSystemRehash
)

// PasswordUpdateContext marks trusted internal password persist paths for UpdateUser.
type PasswordUpdateContext struct {
	// Prevalidated: min-age/complexity/not-same checks done; password hashed; LastChangePasswordTime set.
	Prevalidated bool
	// SystemRehash: login-time password type migration; password hashed; keep LastChangePasswordTime.
	SystemRehash bool
}

func shouldEnforceMinimumPasswordAge(op PasswordChangeOperation) bool {
	switch op {
	case PasswordOpRecovery, PasswordOpInitialSet, PasswordOpSyncImport, PasswordOpSystemRehash:
		return false
	default:
		return true
	}
}

func CheckMinimumPasswordAge(user *User, org *Organization, op PasswordChangeOperation, lang string) error {
	if org == nil || org.MinimumPasswordAgeHours <= 0 {
		return nil
	}
	if !shouldEnforceMinimumPasswordAge(op) {
		return nil
	}
	if user == nil || user.LastChangePasswordTime == "" {
		return nil
	}

	lastTime := util.String2Time(user.LastChangePasswordTime)
	allowedAt := lastTime.Add(time.Duration(org.MinimumPasswordAgeHours) * time.Hour)
	if getPasswordPolicyNow().Before(allowedAt) {
		return errors.New(i18n.Translate(lang, "check:Password cannot be changed yet. The minimum password age has not expired."))
	}
	return nil
}

func MarkPasswordChanged(user *User) {
	user.LastChangePasswordTime = util.GetCurrentTime()
}

func isUserPasswordValueChanged(oldUser, user *User) bool {
	if user == nil || oldUser == nil {
		return false
	}
	if user.Password == "" || user.Password == "***" {
		return false
	}
	return user.Password != oldUser.Password
}

func passwordChangeOperationForUpdate(oldUser *User) PasswordChangeOperation {
	if oldUser.LastChangePasswordTime == "" {
		return PasswordOpInitialSet
	}
	return PasswordOpUserChange
}

// shouldApplyPasswordChangePolicy returns false when UpdateUser must skip policy (trusted paths).
func shouldApplyPasswordChangePolicy(oldUser, user *User, pwCtx *PasswordUpdateContext) bool {
	if pwCtx != nil && pwCtx.Prevalidated {
		return false
	}
	if pwCtx != nil && pwCtx.SystemRehash {
		user.LastChangePasswordTime = oldUser.LastChangePasswordTime
		return false
	}
	if isUserPasswordValueChanged(oldUser, user) {
		user.LastChangePasswordTime = oldUser.LastChangePasswordTime
	}
	return true
}

func applyPasswordChangePolicy(oldUser, user *User, columns []string, lang string) ([]string, error) {
	if !isUserPasswordValueChanged(oldUser, user) {
		return columns, nil
	}

	organization, err := GetOrganizationByUser(user)
	if err != nil {
		return columns, err
	}
	if organization == nil {
		return columns, errors.New(i18n.Translate(lang, "check:Organization does not exist"))
	}

	op := passwordChangeOperationForUpdate(oldUser)
	if err = CheckMinimumPasswordAge(oldUser, organization, op, lang); err != nil {
		return columns, err
	}

	if msg := CheckPasswordComplexity(user, user.Password, lang); msg != "" {
		return columns, errors.New(msg)
	}
	if !CheckPasswordNotSameAsCurrent(oldUser, user.Password, organization) {
		return columns, errors.New(i18n.Translate(lang, "user:The new password must be different from your current password"))
	}

	user.UpdateUserPassword(organization)
	MarkPasswordChanged(user)

	if len(columns) == 0 {
		return []string{"password", "password_salt", "password_type", "last_change_password_time"}, nil
	}

	if !util.InSlice(columns, "password") {
		columns = append(columns, "password", "password_salt", "password_type")
	}
	if !util.InSlice(columns, "last_change_password_time") {
		columns = append(columns, "last_change_password_time")
	}
	return columns, nil
}

// getPasswordPolicyNow is overridden in tests.
var getPasswordPolicyNow = func() time.Time {
	return time.Now()
}
