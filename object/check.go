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

package object

import (
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/casdoor/casdoor/cred"
	"github.com/casdoor/casdoor/form"
	"github.com/casdoor/casdoor/i18n"
	"github.com/casdoor/casdoor/util"
	goldap "github.com/go-ldap/ldap/v3"
)

const (
	DefaultFailedSigninLimit      = 5
	DefaultFailedSigninFrozenTime = 15
)

func CheckUserSignup(application *Application, organization *Organization, authForm *form.AuthForm, lang string) string {
	if organization == nil {
		return i18n.Translate(lang, "check:Organization does not exist")
	}

	if application.IsSignupItemVisible("Username") {
		if len(authForm.Username) <= 1 {
			return i18n.Translate(lang, "check:Username must have at least 2 characters")
		}
		if unicode.IsDigit(rune(authForm.Username[0])) {
			return i18n.Translate(lang, "check:Username cannot start with a digit")
		}
		if util.IsEmailValid(authForm.Username) {
			return i18n.Translate(lang, "check:Username cannot be an email address")
		}
		if util.ReWhiteSpace.MatchString(authForm.Username) {
			return i18n.Translate(lang, "check:Username cannot contain white spaces")
		}

		if msg := CheckUsername(authForm.Username, lang); msg != "" {
			return msg
		}

		if HasUserByField(organization.Name, "name", authForm.Username) {
			return i18n.Translate(lang, "check:Username already exists")
		}
		if HasUserByField(organization.Name, "email", authForm.Email) {
			return i18n.Translate(lang, "check:Email already exists")
		}
		if HasUserByField(organization.Name, "phone", authForm.Phone) {
			return i18n.Translate(lang, "check:Phone already exists")
		}
	}

	if application.IsSignupItemVisible("Password") {
		msg := CheckPasswordComplexityByOrg(organization, authForm.Password)
		if msg != "" {
			return msg
		}
	}

	if application.IsSignupItemVisible("Email") {
		if authForm.Email == "" {
			if application.IsSignupItemRequired("Email") {
				return i18n.Translate(lang, "check:Email cannot be empty")
			}
		} else {
			if HasUserByField(organization.Name, "email", authForm.Email) {
				return i18n.Translate(lang, "check:Email already exists")
			} else if !util.IsEmailValid(authForm.Email) {
				return i18n.Translate(lang, "check:Email is invalid")
			}
		}
	}

	if application.IsSignupItemVisible("Phone") {
		if authForm.Phone == "" {
			if application.IsSignupItemRequired("Phone") {
				return i18n.Translate(lang, "check:Phone cannot be empty")
			}
		} else {
			if HasUserByField(organization.Name, "phone", authForm.Phone) {
				return i18n.Translate(lang, "check:Phone already exists")
			} else if !util.IsPhoneAllowInRegin(authForm.CountryCode, organization.CountryCodes) {
				return i18n.Translate(lang, "check:Your region is not allow to signup by phone")
			} else if !util.IsPhoneValid(authForm.Phone, authForm.CountryCode) {
				return i18n.Translate(lang, "check:Phone number is invalid")
			}
		}
	}

	if application.IsSignupItemVisible("Display name") {
		if application.GetSignupItemRule("Display name") == "First, last" && (authForm.FirstName != "" || authForm.LastName != "") {
			if authForm.FirstName == "" {
				return i18n.Translate(lang, "check:FirstName cannot be blank")
			} else if authForm.LastName == "" {
				return i18n.Translate(lang, "check:LastName cannot be blank")
			}
		} else {
			if authForm.Name == "" {
				return i18n.Translate(lang, "check:DisplayName cannot be blank")
			} else if application.GetSignupItemRule("Display name") == "Real name" {
				if !isValidRealName(authForm.Name) {
					return i18n.Translate(lang, "check:DisplayName is not valid real name")
				}
			}
		}
	}

	if application.IsSignupItemVisible("Affiliation") {
		if authForm.Affiliation == "" {
			return i18n.Translate(lang, "check:Affiliation cannot be blank")
		}
	}

	for _, signupItem := range application.SignupItems {
		if signupItem.Regex == "" {
			continue
		}

		isString, value := form.GetAuthFormFieldValue(authForm, signupItem.Name)
		if !isString {
			continue
		}

		regexSignupItem, err := regexp.Compile(signupItem.Regex)
		if err != nil {
			return err.Error()
		}

		matched := regexSignupItem.MatchString(value)
		if !matched {
			return fmt.Sprintf(i18n.Translate(lang, "check:The value \"%s\" for signup field \"%s\" doesn't match the signup item regex of the application \"%s\""), value, signupItem.Name, application.Name)
		}
	}

	return ""
}

func CheckInvitationCode(application *Application, organization *Organization, authForm *form.AuthForm, lang string) (*Invitation, string) {
	if authForm.InvitationCode == "" {
		if application.IsSignupItemRequired("Invitation code") {
			return nil, i18n.Translate(lang, "check:Invitation code cannot be blank")
		} else {
			return nil, ""
		}
	}

	invitations, err := GetInvitations(organization.Name)
	if err != nil {
		return nil, err.Error()
	}
	errMsg := ""
	for _, invitation := range invitations {
		if invitation.Application != application.Name && invitation.Application != "All" {
			continue
		}
		if isValid, msg := invitation.IsInvitationCodeValid(application, authForm.InvitationCode, authForm.Username, authForm.Email, authForm.Phone, lang); isValid {
			return invitation, msg
		} else if msg != "" && errMsg == "" {
			errMsg = msg
		}
	}

	if errMsg != "" {
		return nil, errMsg
	} else {
		return nil, i18n.Translate(lang, "check:Invitation code is invalid")
	}
}

func CheckInvitationDefaultCode(code string, defaultCode string, lang string) error {
	if matched, err := util.IsInvitationCodeMatch(code, defaultCode); err != nil {
		return err
	} else if !matched {
		return fmt.Errorf(i18n.Translate(lang, "check:Default code does not match the code's matching rules"))
	}
	return nil
}

func checkSigninErrorTimes(user *User, lang string) error {
	failedSigninLimit, failedSigninFrozenTime, err := GetFailedSigninConfigByUser(user)
	if err != nil {
		return err
	}

	if user.SigninWrongTimes >= failedSigninLimit {
		lastSignWrongTime, _ := time.Parse(time.RFC3339, user.LastSigninWrongTime)
		passedTime := time.Now().UTC().Sub(lastSignWrongTime)
		minutes := failedSigninFrozenTime - int(passedTime.Minutes())

		// deny the login if the error times is greater than the limit and the last login time is less than the duration
		if minutes > 0 {
			return fmt.Errorf(i18n.Translate(lang, "check:You have entered the wrong password or code too many times, please wait for %d minutes and try again"), minutes)
		}

		// reset the error times
		user.SigninWrongTimes = 0

		_, err := UpdateUser(user.GetId(), user, []string{"signin_wrong_times"}, false)
		return err
	}

	return nil
}

func CheckPassword(user *User, password string, lang string, options ...bool) error {
	enableCaptcha := false
	if len(options) > 0 {
		enableCaptcha = options[0]
	}
	// check the login error times
	if !enableCaptcha {
		err := checkSigninErrorTimes(user, lang)
		if err != nil {
			return err
		}
	}

	organization, err := GetOrganizationByUser(user)
	if err != nil {
		return err
	}

	if organization == nil {
		return fmt.Errorf(i18n.Translate(lang, "check:Organization does not exist"))
	}

	if password == "" {
		return fmt.Errorf(i18n.Translate(lang, "check:Password cannot be empty"))
	}

	passwordType := user.PasswordType
	if passwordType == "" {
		passwordType = organization.PasswordType
	}
	credManager := cred.GetCredManager(passwordType)
	if credManager != nil {
		if organization.MasterPassword != "" {
			if password == organization.MasterPassword || credManager.IsPasswordCorrect(password, organization.MasterPassword, "", organization.PasswordSalt) {
				return resetUserSigninErrorTimes(user)
			}
		}

		if credManager.IsPasswordCorrect(password, user.Password, user.PasswordSalt, organization.PasswordSalt) {
			return resetUserSigninErrorTimes(user)
		}

		return recordSigninErrorInfo(user, lang, enableCaptcha)
	} else {
		return fmt.Errorf(i18n.Translate(lang, "check:unsupported password type: %s"), organization.PasswordType)
	}
}

func CheckPasswordComplexityByOrg(organization *Organization, password string) string {
	errorMsg := checkPasswordComplexity(password, organization.PasswordOptions)
	return errorMsg
}

func CheckPasswordComplexity(user *User, password string) string {
	organization, _ := GetOrganizationByUser(user)
	return CheckPasswordComplexityByOrg(organization, password)
}

func CheckLdapUserPassword(user *User, password string, lang string) error {
	ldaps, err := GetLdaps(user.Owner)
	if err != nil {
		return err
	}

	ldapLoginSuccess := false
	hit := false

	for _, ldapServer := range ldaps {
		conn, err := ldapServer.GetLdapConn()
		if err != nil {
			continue
		}

		searchReq := goldap.NewSearchRequest(ldapServer.BaseDn, goldap.ScopeWholeSubtree, goldap.NeverDerefAliases,
			0, 0, false, ldapServer.buildAuthFilterString(user), []string{}, nil)

		searchResult, err := conn.Conn.Search(searchReq)
		if err != nil {
			conn.Close()
			return err
		}

		if len(searchResult.Entries) == 0 {
			conn.Close()
			continue
		}
		if len(searchResult.Entries) > 1 {
			conn.Close()
			return fmt.Errorf(i18n.Translate(lang, "check:Multiple accounts with same uid, please check your ldap server"))
		}

		hit = true
		dn := searchResult.Entries[0].DN
		if err = conn.Conn.Bind(dn, password); err == nil {
			ldapLoginSuccess = true
			conn.Close()
			break
		}

		conn.Close()
	}

	if !ldapLoginSuccess {
		if !hit {
			return fmt.Errorf("user not exist")
		}
		return fmt.Errorf(i18n.Translate(lang, "check:LDAP user name or password incorrect"))
	}
	return resetUserSigninErrorTimes(user)
}

func CheckUserPassword(organization string, username string, password string, lang string, options ...bool) (*User, error) {
	enableCaptcha := false
	isSigninViaLdap := false
	isPasswordWithLdapEnabled := false
	if len(options) > 0 {
		enableCaptcha = options[0]
		isSigninViaLdap = options[1]
		isPasswordWithLdapEnabled = options[2]
	}
	user, err := GetUserByFields(organization, username)
	if err != nil {
		return nil, err
	}

	if user == nil || user.IsDeleted {
		return nil, fmt.Errorf(i18n.Translate(lang, "general:The user: %s doesn't exist"), util.GetId(organization, username))
	}

	if user.IsForbidden {
		return nil, fmt.Errorf(i18n.Translate(lang, "check:The user is forbidden to sign in, please contact the administrator"))
	}

	if isSigninViaLdap {
		if user.Ldap == "" {
			return nil, fmt.Errorf(i18n.Translate(lang, "check:The user: %s doesn't exist in LDAP server"), username)
		}
	}

	if user.Ldap != "" {
		if !isSigninViaLdap && !isPasswordWithLdapEnabled {
			return nil, fmt.Errorf(i18n.Translate(lang, "check:password or code is incorrect"))
		}

		// check the login error times
		if !enableCaptcha {
			err = checkSigninErrorTimes(user, lang)
			if err != nil {
				return nil, err
			}
		}

		// only for LDAP users
		err = CheckLdapUserPassword(user, password, lang)
		if err != nil {
			if err.Error() == "user not exist" {
				return nil, fmt.Errorf(i18n.Translate(lang, "check:The user: %s doesn't exist in LDAP server"), username)
			}

			return nil, recordSigninErrorInfo(user, lang, enableCaptcha)
		}
	} else {
		err = CheckPassword(user, password, lang, enableCaptcha)
		if err != nil {
			return nil, err
		}

		err = checkPasswordExpired(user, lang)
		if err != nil {
			return nil, err
		}
	}

	return user, nil
}

func CheckUserPermission(requestUserId, userId string, strict bool, lang string) (bool, error) {
	if requestUserId == "" {
		return false, fmt.Errorf(i18n.Translate(lang, "general:Please login first"))
	}

	userOwner := util.GetOwnerFromId(userId)

	if userId != "" {
		targetUser, err := GetUser(userId)
		if err != nil {
			return false, err
		}

		if targetUser == nil {
			if strings.HasPrefix(requestUserId, "built-in/") {
				return true, nil
			}

			return false, fmt.Errorf(i18n.Translate(lang, "general:The user: %s doesn't exist"), userId)
		}

		userOwner = targetUser.Owner
	}

	hasPermission := false
	if IsAppUser(requestUserId) {
		hasPermission = true
	} else {
		requestUser, err := GetUser(requestUserId)
		if err != nil {
			return false, err
		}

		if requestUser == nil {
			return false, fmt.Errorf(i18n.Translate(lang, "check:Session outdated, please login again"))
		}
		if requestUser.IsGlobalAdmin() {
			hasPermission = true
		} else if requestUserId == userId {
			hasPermission = true
		} else if userOwner == requestUser.Owner {
			if strict {
				hasPermission = requestUser.IsAdmin
			} else {
				hasPermission = true
			}
		}
	}

	return hasPermission, fmt.Errorf(i18n.Translate(lang, "auth:Unauthorized operation"))
}

func CheckLoginPermission(userId string, application *Application) (bool, error) {
	owner, _ := util.GetOwnerAndNameFromId(userId)
	if owner == "built-in" {
		return true, nil
	}

	permissions, err := GetPermissions(application.Organization)
	if err != nil {
		return false, err
	}

	allowPermissionCount := 0
	denyPermissionCount := 0
	allowCount := 0
	denyCount := 0
	for _, permission := range permissions {
		if !permission.IsEnabled || permission.State != "Approved" || permission.ResourceType != "Application" || !permission.isResourceHit(application.Name) {
			continue
		}

		if !permission.isUserHit(userId) && !permission.isRoleHit(userId) {
			if permission.Effect == "Allow" {
				allowPermissionCount += 1
			} else {
				denyPermissionCount += 1
			}
			continue
		}

		enforcer, err := getPermissionEnforcer(permission)
		if err != nil {
			return false, err
		}

		var isAllowed bool
		isAllowed, err = enforcer.Enforce(userId, application.Name, "Read")
		if err != nil {
			return false, err
		}

		if isAllowed {
			if permission.Effect == "Allow" {
				allowCount += 1
			}
		} else {
			if permission.Effect == "Deny" {
				denyCount += 1
			}
		}
	}

	// Deny-override, if one deny is found, then deny
	if denyCount > 0 {
		return false, nil
	} else if allowCount > 0 {
		return true, nil
	}

	// For no-allow and no-deny condition
	// If only allow permissions exist, we suppose it's Deny-by-default, aka no-allow means deny
	// Otherwise, it's Allow-by-default, aka no-deny means allow
	if allowPermissionCount > 0 && denyPermissionCount == 0 {
		return false, nil
	}
	return true, nil
}

func CheckUsername(username string, lang string) string {
	if username == "" {
		return i18n.Translate(lang, "check:Empty username.")
	} else if len(username) > 39 {
		return i18n.Translate(lang, "check:Username is too long (maximum is 39 characters).")
	}

	// https://stackoverflow.com/questions/58726546/github-username-convention-using-regex

	if !util.ReUserName.MatchString(username) {
		return i18n.Translate(lang, "check:The username may only contain alphanumeric characters, underlines or hyphens, cannot have consecutive hyphens or underlines, and cannot begin or end with a hyphen or underline.")
	}

	return ""
}

func CheckUsernameWithEmail(username string, lang string) string {
	if username == "" {
		return i18n.Translate(lang, "check:Empty username.")
	} else if len(username) > 39 {
		return i18n.Translate(lang, "check:Username is too long (maximum is 39 characters).")
	}

	// https://stackoverflow.com/questions/58726546/github-username-convention-using-regex

	if !util.ReUserNameWithEmail.MatchString(username) {
		return i18n.Translate(lang, "check:Username supports email format. Also The username may only contain alphanumeric characters, underlines or hyphens, cannot have consecutive hyphens or underlines, and cannot begin or end with a hyphen or underline. Also pay attention to the email format.")
	}
	return ""
}

func CheckUpdateUser(oldUser, user *User, lang string) string {
	if oldUser.Name != user.Name {
		organizationName := oldUser.Owner
		if organizationName == "" {
			organizationName = user.Owner
		}

		organization, err := getOrganization("admin", organizationName)
		if err != nil {
			return err.Error()
		}
		if organization == nil {
			return fmt.Sprintf(i18n.Translate(lang, "auth:The organization: %s does not exist"), organizationName)
		}

		if organization.UseEmailAsUsername {
			if msg := CheckUsernameWithEmail(user.Name, lang); msg != "" {
				return msg
			}
		} else {
			if msg := CheckUsername(user.Name, lang); msg != "" {
				return msg
			}
		}

		if HasUserByField(user.Owner, "name", user.Name) {
			return i18n.Translate(lang, "check:Username already exists")
		}
	}
	if oldUser.Email != user.Email {
		if HasUserByField(user.Owner, "email", user.Email) {
			return i18n.Translate(lang, "check:Email already exists")
		}
	}
	if oldUser.Phone != user.Phone {
		if HasUserByField(user.Owner, "phone", user.Phone) {
			return i18n.Translate(lang, "check:Phone already exists")
		}
	}
	if oldUser.IpWhitelist != user.IpWhitelist {
		if err := CheckIpWhitelist(user.IpWhitelist, lang); err != nil {
			return err.Error()
		}
	}

	return ""
}

func CheckToEnableCaptcha(application *Application, organization, username string) (bool, error) {
	if len(application.Providers) == 0 {
		return false, nil
	}

	for _, providerItem := range application.Providers {
		if providerItem.Provider == nil {
			continue
		}
		if providerItem.Provider.Category == "Captcha" {
			if providerItem.Rule == "Dynamic" {
				user, err := GetUserByFields(organization, username)
				if err != nil {
					return false, err
				}

				failedSigninLimit := application.FailedSigninLimit
				if failedSigninLimit == 0 {
					failedSigninLimit = DefaultFailedSigninLimit
				}

				return user != nil && user.SigninWrongTimes >= failedSigninLimit, nil
			}
			return providerItem.Rule == "Always", nil
		}
	}

	return false, nil
}
