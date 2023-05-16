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
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/casdoor/casdoor/idp"
	"github.com/xorm-io/core"
)

func GetUserByField(organizationName string, field string, value string) *User {
	if field == "" || value == "" {
		return nil
	}

	user := User{Owner: organizationName}
	existed, err := adapter.Engine.Where(fmt.Sprintf("%s=?", strings.ToLower(field)), value).Get(&user)
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
	if strings.Contains(field, "@") {
		user = GetUserByField(organization, "email", field)
		if user != nil {
			return user
		}
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
	bean := make(map[string]interface{})
	if field == "password" {
		organization := GetOrganizationByUser(user)
		user.UpdateUserPassword(organization)
		bean[field] = user.Password
		bean["password_type"] = user.PasswordType
		bean["password_change_required"] = false
	} else {
		bean[strings.ToLower(field)] = value
	}

	affected, err := adapter.Engine.Table(user).ID(core.PK{user.Owner, user.Name}).Update(bean)
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
		if user.Properties == nil {
			user.Properties = make(map[string]string)
		}

		user.Properties[field] = value
	}
}

func SetUserOAuthProperties(organization *Organization, user *User, providerType string, userInfo *idp.UserInfo) bool {
	if userInfo.Id != "" {
		propertyName := fmt.Sprintf("oauth_%s_id", providerType)
		setUserProperty(user, propertyName, userInfo.Id)
	}
	if userInfo.Username != "" {
		propertyName := fmt.Sprintf("oauth_%s_username", providerType)
		setUserProperty(user, propertyName, userInfo.Username)
	}
	if userInfo.DisplayName != "" {
		propertyName := fmt.Sprintf("oauth_%s_displayName", providerType)
		setUserProperty(user, propertyName, userInfo.DisplayName)
		if user.DisplayName == "" {
			user.DisplayName = userInfo.DisplayName
		}
	} else if user.DisplayName == "" {
		if userInfo.Username != "" {
			user.DisplayName = userInfo.Username
		} else {
			user.DisplayName = userInfo.Id
		}
	}
	if userInfo.Email != "" {
		propertyName := fmt.Sprintf("oauth_%s_email", providerType)
		setUserProperty(user, propertyName, userInfo.Email)
		if user.Email == "" {
			user.Email = userInfo.Email
		}
	}

	if userInfo.UnionId != "" {
		propertyName := fmt.Sprintf("oauth_%s_unionId", providerType)
		setUserProperty(user, propertyName, userInfo.UnionId)
	}

	if userInfo.AvatarUrl != "" {
		propertyName := fmt.Sprintf("oauth_%s_avatarUrl", providerType)
		setUserProperty(user, propertyName, userInfo.AvatarUrl)
		if user.Avatar == "" || user.Avatar == organization.DefaultAvatar {
			user.Avatar = userInfo.AvatarUrl
		}
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

func CheckPermissionForUpdateUser(oldUser, newUser *User, isAdmin bool, lang string) (bool, string) {
	organization := GetOrganizationByUser(oldUser)
	var itemsChanged []*AccountItem

	if oldUser.Owner != newUser.Owner {
		item := GetAccountItemByName("Organization", organization)
		itemsChanged = append(itemsChanged, item)
	}
	if oldUser.Name != newUser.Name {
		item := GetAccountItemByName("Name", organization)
		itemsChanged = append(itemsChanged, item)
	}
	if oldUser.Id != newUser.Id {
		item := GetAccountItemByName("ID", organization)
		itemsChanged = append(itemsChanged, item)
	}
	if oldUser.DisplayName != newUser.DisplayName {
		item := GetAccountItemByName("Display name", organization)
		itemsChanged = append(itemsChanged, item)
	}
	if oldUser.Avatar != newUser.Avatar {
		item := GetAccountItemByName("Avatar", organization)
		itemsChanged = append(itemsChanged, item)
	}
	if oldUser.Type != newUser.Type {
		item := GetAccountItemByName("User type", organization)
		itemsChanged = append(itemsChanged, item)
	}
	// The password is *** when not modified
	if oldUser.Password != newUser.Password && newUser.Password != "***" {
		item := GetAccountItemByName("Password", organization)
		itemsChanged = append(itemsChanged, item)
	}
	if oldUser.Email != newUser.Email {
		item := GetAccountItemByName("Email", organization)
		itemsChanged = append(itemsChanged, item)
	}
	if oldUser.Phone != newUser.Phone {
		item := GetAccountItemByName("Phone", organization)
		itemsChanged = append(itemsChanged, item)
	}
	if oldUser.CountryCode != newUser.CountryCode {
		item := GetAccountItemByName("Country code", organization)
		itemsChanged = append(itemsChanged, item)
	}
	if oldUser.Region != newUser.Region {
		item := GetAccountItemByName("Country/Region", organization)
		itemsChanged = append(itemsChanged, item)
	}
	if oldUser.Location != newUser.Location {
		item := GetAccountItemByName("Location", organization)
		itemsChanged = append(itemsChanged, item)
	}
	if oldUser.Affiliation != newUser.Affiliation {
		item := GetAccountItemByName("Affiliation", organization)
		itemsChanged = append(itemsChanged, item)
	}
	if oldUser.Title != newUser.Title {
		item := GetAccountItemByName("Title", organization)
		itemsChanged = append(itemsChanged, item)
	}
	if oldUser.Homepage != newUser.Homepage {
		item := GetAccountItemByName("Homepage", organization)
		itemsChanged = append(itemsChanged, item)
	}
	if oldUser.Bio != newUser.Bio {
		item := GetAccountItemByName("Bio", organization)
		itemsChanged = append(itemsChanged, item)
	}
	if oldUser.Tag != newUser.Tag {
		item := GetAccountItemByName("Tag", organization)
		itemsChanged = append(itemsChanged, item)
	}
	if oldUser.SignupApplication != newUser.SignupApplication {
		item := GetAccountItemByName("Signup application", organization)
		itemsChanged = append(itemsChanged, item)
	}

	oldUserPropertiesJson, _ := json.Marshal(oldUser.Properties)
	newUserPropertiesJson, _ := json.Marshal(newUser.Properties)
	if string(oldUserPropertiesJson) != string(newUserPropertiesJson) {
		item := GetAccountItemByName("Properties", organization)
		itemsChanged = append(itemsChanged, item)
	}

	oldUserTwoFactorAuthJson, _ := json.Marshal(oldUser.MultiFactorAuths)
	newUserTwoFactorAuthJson, _ := json.Marshal(newUser.MultiFactorAuths)
	if string(oldUserTwoFactorAuthJson) != string(newUserTwoFactorAuthJson) {
		item := GetAccountItemByName("Multi-factor authentication", organization)
		itemsChanged = append(itemsChanged, item)
	}

	if oldUser.IsAdmin != newUser.IsAdmin {
		item := GetAccountItemByName("Is admin", organization)
		itemsChanged = append(itemsChanged, item)
	}
	if oldUser.IsGlobalAdmin != newUser.IsGlobalAdmin {
		item := GetAccountItemByName("Is global admin", organization)
		itemsChanged = append(itemsChanged, item)
	}
	if oldUser.IsForbidden != newUser.IsForbidden {
		item := GetAccountItemByName("Is forbidden", organization)
		itemsChanged = append(itemsChanged, item)
	}
	if oldUser.IsDeleted != newUser.IsDeleted {
		item := GetAccountItemByName("Is deleted", organization)
		itemsChanged = append(itemsChanged, item)
	}

	for i := range itemsChanged {
		if pass, err := CheckAccountItemModifyRule(itemsChanged[i], isAdmin, lang); !pass {
			return pass, err
		}
	}
	return true, ""
}

func (user *User) GetCountryCode(countryCode string) string {
	if countryCode != "" {
		return countryCode
	}

	if user != nil && user.CountryCode != "" {
		return user.CountryCode
	}

	if org := GetOrganizationByUser(user); org != nil && len(org.CountryCodes) > 0 {
		return org.CountryCodes[0]
	}
	return ""
}

func (user *User) IsAdminUser() bool {
	if user == nil {
		return false
	}

	return user.IsAdmin || user.IsGlobalAdmin
}
