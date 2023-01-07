// Copyright 2023 The Casdoor Authors. All Rights Reserved.
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

package controllers

import (
	"encoding/json"

	"github.com/casdoor/casdoor/object"
)

func checkPermissionForUpdateUser(userId string, newUser object.User, c *ApiController) (bool, string) {
	oldUser := object.GetUser(userId)
	organization := object.GetOrganizationByUser(oldUser)
	var itemsChanged []*object.AccountItem

	if oldUser.Owner != newUser.Owner {
		item := object.GetAccountItemByName("Organization", organization)
		itemsChanged = append(itemsChanged, item)
	}
	if oldUser.Name != newUser.Name {
		item := object.GetAccountItemByName("Name", organization)
		itemsChanged = append(itemsChanged, item)
	}
	if oldUser.Id != newUser.Id {
		item := object.GetAccountItemByName("ID", organization)
		itemsChanged = append(itemsChanged, item)
	}
	if oldUser.DisplayName != newUser.DisplayName {
		item := object.GetAccountItemByName("Display name", organization)
		itemsChanged = append(itemsChanged, item)
	}
	if oldUser.Avatar != newUser.Avatar {
		item := object.GetAccountItemByName("Avatar", organization)
		itemsChanged = append(itemsChanged, item)
	}
	if oldUser.Type != newUser.Type {
		item := object.GetAccountItemByName("User type", organization)
		itemsChanged = append(itemsChanged, item)
	}
	// The password is *** when not modified
	if oldUser.Password != newUser.Password && newUser.Password != "***" {
		item := object.GetAccountItemByName("Password", organization)
		itemsChanged = append(itemsChanged, item)
	}
	if oldUser.Email != newUser.Email {
		item := object.GetAccountItemByName("Email", organization)
		itemsChanged = append(itemsChanged, item)
	}
	if oldUser.Phone != newUser.Phone {
		item := object.GetAccountItemByName("Phone", organization)
		itemsChanged = append(itemsChanged, item)
	}
	if oldUser.Region != newUser.Region {
		item := object.GetAccountItemByName("Country/Region", organization)
		itemsChanged = append(itemsChanged, item)
	}
	if oldUser.Location != newUser.Location {
		item := object.GetAccountItemByName("Location", organization)
		itemsChanged = append(itemsChanged, item)
	}
	if oldUser.Affiliation != newUser.Affiliation {
		item := object.GetAccountItemByName("Affiliation", organization)
		itemsChanged = append(itemsChanged, item)
	}
	if oldUser.Title != newUser.Title {
		item := object.GetAccountItemByName("Title", organization)
		itemsChanged = append(itemsChanged, item)
	}
	if oldUser.Homepage != newUser.Homepage {
		item := object.GetAccountItemByName("Homepage", organization)
		itemsChanged = append(itemsChanged, item)
	}
	if oldUser.Bio != newUser.Bio {
		item := object.GetAccountItemByName("Bio", organization)
		itemsChanged = append(itemsChanged, item)
	}
	if oldUser.Tag != newUser.Tag {
		item := object.GetAccountItemByName("Tag", organization)
		itemsChanged = append(itemsChanged, item)
	}
	if oldUser.SignupApplication != newUser.SignupApplication {
		item := object.GetAccountItemByName("Signup application", organization)
		itemsChanged = append(itemsChanged, item)
	}

	oldUserRolesJson, _ := json.Marshal(oldUser.Roles)
	newUserRolesJson, _ := json.Marshal(newUser.Roles)
	if string(oldUserRolesJson) != string(newUserRolesJson) {
		item := object.GetAccountItemByName("Roles", organization)
		itemsChanged = append(itemsChanged, item)
	}

	oldUserPermissionJson, _ := json.Marshal(oldUser.Permissions)
	newUserPermissionJson, _ := json.Marshal(newUser.Permissions)
	if string(oldUserPermissionJson) != string(newUserPermissionJson) {
		item := object.GetAccountItemByName("Permissions", organization)
		itemsChanged = append(itemsChanged, item)
	}

	oldUserPropertiesJson, _ := json.Marshal(oldUser.Properties)
	newUserPropertiesJson, _ := json.Marshal(newUser.Properties)
	if string(oldUserPropertiesJson) != string(newUserPropertiesJson) {
		item := object.GetAccountItemByName("Properties", organization)
		itemsChanged = append(itemsChanged, item)
	}

	if oldUser.IsAdmin != newUser.IsAdmin {
		item := object.GetAccountItemByName("Is admin", organization)
		itemsChanged = append(itemsChanged, item)
	}
	if oldUser.IsGlobalAdmin != newUser.IsGlobalAdmin {
		item := object.GetAccountItemByName("Is global admin", organization)
		itemsChanged = append(itemsChanged, item)
	}
	if oldUser.IsForbidden != newUser.IsForbidden {
		item := object.GetAccountItemByName("Is forbidden", organization)
		itemsChanged = append(itemsChanged, item)
	}
	if oldUser.IsDeleted != newUser.IsDeleted {
		item := object.GetAccountItemByName("Is deleted", organization)
		itemsChanged = append(itemsChanged, item)
	}

	for i := range itemsChanged {
		if pass, err := object.CheckAccountItemModifyRule(itemsChanged[i], c.getCurrentUser(), c.GetAcceptLanguage()); !pass {
			return pass, err
		}
	}
	return true, ""
}
