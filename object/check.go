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

	"github.com/casdoor/casdoor/util"
)

var reWhiteSpace *regexp.Regexp

func init() {
	reWhiteSpace, _ = regexp.Compile("\\s")
}

func CheckUserSignup(organization string, username string, password string, displayName string, email string, phonePrefix string, phone string, affiliation string) string {
	if len(username) == 0 {
		return "username cannot be blank"
	} else if len(password) == 0 {
		return "password cannot be blank"
	} else if getOrganization("admin", organization) == nil {
		return "organization does not exist"
	} else if reWhiteSpace.MatchString(username) {
		return "username cannot contain white spaces"
	} else if HasUserByField(organization, "name", username) {
		return "username already exists"
	} else if HasUserByField(organization, "email", email) {
		return "email already exists"
	} else if HasUserByField(organization, "phone", phone) {
		return "phone already exists"
	} else if displayName == "" {
		return "displayName cannot be blank"
	} else if affiliation == "" {
		return "affiliation cannot be blank"
	} else if !util.IsEmailValid(email) {
		return "email is invalid"
	} else if phonePrefix == "86" && !util.IsPhoneCnValid(phone) {
		return "phone number is invalid"
	} else {
		return ""
	}
}

func CheckUserLogin(organization string, username string, password string) (*User, string) {
	user := GetUserByFields(organization, username)
	if user == nil {
		return nil, "the user does not exist, please sign up first"
	}

	if user.Password != password {
		return nil, "password incorrect"
	}

	return user, ""
}

func (user *User) GetId() string {
	return fmt.Sprintf("%s/%s", user.Owner, user.Name)
}
