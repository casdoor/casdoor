// Copyright 2021 The casbin Authors. All Rights Reserved.
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

	"github.com/casbin/casdoor/util"
)

var reWhiteSpace *regexp.Regexp

func init() {
	reWhiteSpace, _ = regexp.Compile("\\s")
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
		if HasUserByField(organization.Name, "email", email) {
			return "email already exists"
		} else if !util.IsEmailValid(email) {
			return "email is invalid"
		}
	}

	if application.IsSignupItemVisible("Phone") {
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

	if organization.PasswordType == "plain" {
		if password == user.Password {
			return ""
		} else {
			return "password incorrect"
		}
	} else if organization.PasswordType == "salt" {
		if password == user.Password || getSaltedPassword(password, organization.PasswordSalt) == user.Password {
			return ""
		} else {
			return "password incorrect"
		}
	} else {
		return fmt.Sprintf("unsupported password type: %s", organization.PasswordType)
	}
}

func CheckUserLogin(organization string, username string, password string) (*User, string) {
	user := GetUserByFields(organization, username)
	if user == nil {
		return nil, "the user does not exist, please sign up first"
	}

	if user.IsForbidden {
		return nil, "the user is forbidden to sign in, please contact the administrator"
	}

	msg := CheckPassword(user, password)
	if msg != "" {
		return nil, msg
	}

	return user, ""
}
