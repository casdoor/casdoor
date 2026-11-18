// Copyright 2026 Casdoor Inc. All Rights Reserved.
//
// This file is part of the commercial and enterprise editions of Casdoor and is licensed under the Casdoor Enterprise
// License, not the Apache License 2.0. You may not use this file except in compliance with that license, see
// ee/LICENSE.

package ee

import (
	"github.com/beego/beego/v2/core/logs"
	"github.com/casdoor/casdoor/action"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

// MigratePasswordOnSignin moves the password of a user imported without one (from Authing or another system that
// doesn't export password hashes) on the first sign-in: the organization's password migration action asks the old
// system, and Casdoor keeps the password from then on. Users who already have a password never reach the old system.
func MigratePasswordOnSignin(organization string, username string, password string) (*object.User, error) {
	policy := GetCachedSecurityPolicy(organization)
	if policy == nil || policy.PasswordMigrationPlugin == "" || password == "" {
		return nil, nil
	}

	user, err := object.GetUserByFields(organization, username)
	if err != nil || user == nil {
		return nil, err
	}
	if user.Password != "" || user.IsDeleted || user.IsForbidden || user.Ldap != "" {
		return nil, nil
	}

	m, _, err := object.GetInstalledAction(organization, policy.PasswordMigrationPlugin, "password-migration")
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

	org, err := object.LookupOrganization("admin", organization)
	if err != nil || org == nil {
		return nil, err
	}
	user.Password = password
	user.UpdateUserPassword(org)
	user.LastChangePasswordTime = util.GetCurrentTime()
	_, err = object.UpdateUser(user.GetId(), user, []string{"password", "password_salt", "password_type", "last_change_password_time"}, false)
	if err != nil {
		return nil, err
	}
	err = object.ResetUserSigninErrorTimes(user)
	if err != nil {
		return nil, err
	}

	logs.Info("migrated the password of the user: %s with the action: %s", user.GetId(), policy.PasswordMigrationPlugin)
	return user, nil
}
