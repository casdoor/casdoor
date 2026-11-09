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
	"github.com/beego/beego/v2/core/logs"
	"github.com/casdoor/casdoor/action"
	"github.com/casdoor/casdoor/util"
)

// MigratePasswordOnSignin moves the password of a user imported without one (from Authing or another system that
// doesn't export password hashes) on the first sign-in: the organization's password migration action asks the old
// system, and Casdoor keeps the password from then on. Users who already have a password never reach the old system.
func MigratePasswordOnSignin(organization string, username string, password string) (*User, error) {
	policy := GetCachedSecurityPolicy(organization)
	if policy == nil || policy.PasswordMigrationPlugin == "" || password == "" {
		return nil, nil
	}

	user, err := GetUserByFields(organization, username)
	if err != nil || user == nil {
		return nil, err
	}
	if user.Password != "" || user.IsDeleted || user.IsForbidden || user.Ldap != "" {
		return nil, nil
	}

	m, _, err := getInstalledAction(organization, policy.PasswordMigrationPlugin, "password-migration")
	if err != nil {
		logs.Warn("the password of the user: %s is not migrated: %s", user.GetId(), err.Error())
		return nil, nil
	}
	config := &action.Config{Host: policy.PasswordMigrationUrl, ClientSecret: policy.PasswordMigrationSecret}
	valid, err := action.VerifyPassword(m, config, user.Name, password)
	if err != nil {
		logs.Warn("the password of the user: %s is not migrated: %s", user.GetId(), err.Error())
		return nil, nil
	}
	if !valid {
		return nil, nil
	}

	org, err := getOrganization("admin", organization)
	if err != nil || org == nil {
		return nil, err
	}
	user.Password = password
	user.UpdateUserPassword(org)
	user.LastChangePasswordTime = util.GetCurrentTime()
	_, err = UpdateUser(user.GetId(), user, []string{"password", "password_salt", "password_type", "last_change_password_time"}, false)
	if err != nil {
		return nil, err
	}
	err = resetUserSigninErrorTimes(user)
	if err != nil {
		return nil, err
	}

	logs.Info("migrated the password of the user: %s with the action: %s", user.GetId(), policy.PasswordMigrationPlugin)
	return user, nil
}
