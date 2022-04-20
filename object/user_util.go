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
	"reflect"
	"strings"

	"github.com/casdoor/casdoor/idp"
	"xorm.io/core"
)

func GetUserByField(organizationName string, field string, value string) *User {
	if field == "" || value == "" {
		return nil
	}

	user := User{Owner: organizationName}
	existed, err := adapter.Engine.Where(fmt.Sprintf("%s=?", field), value).Get(&user)
	if err != nil {
		panic(err)
	}

	if existed {
		return &user
	} else {
		return nil
	}
}

func HasUserByField(organizationName string, field string, value string) bool {
	return GetUserByField(organizationName, field, value) != nil
}

func GetUserByFields(organization string, field string) *User {
	// check username
	user := GetUserByField(organization, "name", field)
	if user != nil {
		return user
	}

	// check email
	user = GetUserByField(organization, "email", field)
	if user != nil {
		return user
	}

	// check phone
	user = GetUserByField(organization, "phone", field)
	if user != nil {
		return user
	}

	// check ID card
	user = GetUserByField(organization, "id_card", field)
	if user != nil {
		return user
	}

	return nil
}

func SetUserField(user *User, field string, value string) bool {
	if field == "password" {
		organization := GetOrganizationByUser(user)
		user.UpdateUserPassword(organization)
		value = user.Password
	}

	affected, err := adapter.Engine.Table(user).ID(core.PK{user.Owner, user.Name}).Update(map[string]interface{}{field: value})
	if err != nil {
		panic(err)
	}

	user = getUser(user.Owner, user.Name)
	user.UpdateUserHash()
	_, err = adapter.Engine.ID(core.PK{user.Owner, user.Name}).Cols("hash").Update(user)
	if err != nil {
		panic(err)
	}

	return affected != 0
}

func GetUserField(user *User, field string) string {
	// https://socketloop.com/tutorials/golang-how-to-get-struct-field-and-value-by-name
	u := reflect.ValueOf(user)
	f := reflect.Indirect(u).FieldByName(field)
	return f.String()
}

func setUserProperty(user *User, field string, value string) {
	if value == "" {
		delete(user.Properties, field)
	} else {
		user.Properties[field] = value
	}
}

func SetUserOAuthProperties(organization *Organization, user *User, providerType string, userInfo idp.UserInfoGetter) bool {
	for propertyName, propertyValue := range userInfo.GetAllProperties() {
		if propertyValue == "" {
			continue
		}
		propertyName := fmt.Sprintf("oauth_%s_%s", providerType, propertyName)
		setUserProperty(user, propertyName, propertyValue)
	}
	if userInfo.GetDisplayName() != "" && user.DisplayName == "" {
		user.DisplayName = userInfo.GetDisplayName()
	}
	if userInfo.GetEmail() != "" && user.DisplayName == "" {
		user.Email = userInfo.GetEmail()
	}
	if userInfo.GetAvatarURL() != "" && (user.Avatar == "" || user.Avatar == organization.DefaultAvatar) {
		user.Avatar = userInfo.GetAvatarURL()
	}

	affected := UpdateUserForAllFields(user.GetId(), user)
	return affected
}

func ClearUserOAuthProperties(user *User, providerType string) bool {
	for k := range user.Properties {
		prefix := fmt.Sprintf("oauth_%s_", providerType)
		if strings.HasPrefix(k, prefix) {
			delete(user.Properties, k)
		}
	}

	affected, err := adapter.Engine.ID(core.PK{user.Owner, user.Name}).Cols("properties").Update(user)
	if err != nil {
		panic(err)
	}

	return affected != 0
}
