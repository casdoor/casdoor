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

	"github.com/casdoor/casdoor/cred"
	"github.com/casdoor/casdoor/util"
	goldap "github.com/go-ldap/ldap/v3"
)

var (
	reWhiteSpace     *regexp.Regexp
	reFieldWhiteList *regexp.Regexp
)

func init() {
	reWhiteSpace, _ = regexp.Compile(`\s`)
	reFieldWhiteList, _ = regexp.Compile(`^[A-Za-z0-9]+$`)
}

func CheckUserSignup(application *Application, organization *Organization, username string, password string, displayName string, email string, phone string, affiliation string) string {
	if organization == nil {
		return "organization does not exist"
	}

	if application.IsSignupItemVisible("Username") {
		if len(username) <= 1 {
			return "username must have at least 2 characters"
		} else if reWhiteSpace.MatchString(username) {
			return "username cannot contain white spaces"
		} else if HasUserByField(organization.Name, "name", username) {
			return "username already exists"
		}
	}

	if len(password) <= 5 {
		return "password must have at least 6 characters"
	}

	if application.IsSignupItemVisible("Email") {
		if email == "" {
			if application.IsSignupItemRequired("Email") {
				return "email cannot be empty"
			} else {
				return ""
			}
		}

		if HasUserByField(organization.Name, "email", email) {
			return "email already exists"
		} else if !util.IsEmailValid(email) {
			return "email is invalid"
		}
	}

	if application.IsSignupItemVisible("Phone") {
		if phone == "" {
			if application.IsSignupItemRequired("Phone") {
				return "phone cannot be empty"
			} else {
				return ""
			}
		}

		if HasUserByField(organization.Name, "phone", phone) {
			return "phone already exists"
		} else if organization.PhonePrefix == "86" && !util.IsPhoneCnValid(phone) {
			return "phone number is invalid"
		}
	}

	if application.IsSignupItemVisible("Display name") {
		if displayName == "" {
			return "displayName cannot be blank"
		} else if application.GetSignupItemRule("Display name") == "Personal" {
			if !isValidPersonalName(displayName) {
				return "displayName is not valid personal name"
			}
		}
	}

	if application.IsSignupItemVisible("Affiliation") {
		if affiliation == "" {
			return "affiliation cannot be blank"
		}
	}

	return ""
}

func CheckPassword(user *User, password string) string {
	organization := GetOrganizationByUser(user)
	if organization == nil {
		return "organization does not exist"
	}

	credManager := cred.GetCredManager(organization.PasswordType)
	if credManager != nil {
		if organization.MasterPassword != "" {
			if credManager.IsPasswordCorrect(password, organization.MasterPassword, "", organization.PasswordSalt) {
				return ""
			}
		}

		if credManager.IsPasswordCorrect(password, user.Password, user.PasswordSalt, organization.PasswordSalt) {
			return ""
		}
		return "password incorrect"
	} else {
		return fmt.Sprintf("unsupported password type: %s", organization.PasswordType)
	}
}

func checkLdapUserPassword(user *User, password string) (*User, string) {
	ldaps := GetLdaps(user.Owner)
	ldapLoginSuccess := false
	for _, ldapServer := range ldaps {
		conn, err := GetLdapConn(ldapServer.Host, ldapServer.Port, ldapServer.Admin, ldapServer.Passwd)
		if err != nil {
			continue
		}
		SearchFilter := fmt.Sprintf("(&(objectClass=posixAccount)(uid=%s))", user.Name)
		searchReq := goldap.NewSearchRequest(ldapServer.BaseDn,
			goldap.ScopeWholeSubtree, goldap.NeverDerefAliases, 0, 0, false,
			SearchFilter, []string{}, nil)
		searchResult, err := conn.Conn.Search(searchReq)
		if err != nil {
			return nil, err.Error()
		}

		if len(searchResult.Entries) == 0 {
			continue
		} else if len(searchResult.Entries) > 1 {
			return nil, "Error: multiple accounts with same uid, please check your ldap server"
		}

		dn := searchResult.Entries[0].DN
		if err := conn.Conn.Bind(dn, password); err == nil {
			ldapLoginSuccess = true
			break
		}
	}

	if !ldapLoginSuccess {
		return nil, "ldap user name or password incorrect"
	}
	return user, ""
}

func CheckUserPassword(organization string, username string, password string) (*User, string) {
	user := GetUserByFields(organization, username)
	if user == nil || user.IsDeleted == true {
		return nil, "the user does not exist, please sign up first"
	}

	if user.IsForbidden {
		return nil, "the user is forbidden to sign in, please contact the administrator"
	}
	//for ldap users
	if user.Ldap != "" {
		return checkLdapUserPassword(user, password)
	}

	msg := CheckPassword(user, password)
	if msg != "" {
		return nil, msg
	}

	return user, ""
}

func filterField(field string) bool {
	return reFieldWhiteList.MatchString(field)
}
