// Copyright 2020 The casbin Authors. All Rights Reserved.
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

import "fmt"

func CheckUserRegister(userId string, password string) string {
	if len(userId) == 0 || len(password) == 0 {
		return "username and password cannot be blank"
	} else if HasUser(userId) {
		return "username already exists"
	} else {
		return ""
	}
}

func CheckUserLogin(userId string, password string) string {
	if !HasUser(userId) {
		return "username does not exist, please sign up first"
	}

	if !IsPasswordCorrect(userId, password) {
		return "password incorrect"
	}

	return ""
}

func (user *User) getId() string {
	return fmt.Sprintf("%s/%s", user.Owner, user.Name)
}

func GetUserIdByField(application *Application, field string, value string) string {
	user := GetUserByField(application.Organization, field, value)
	if user != nil {
		return user.getId()
	}
	return ""
}
