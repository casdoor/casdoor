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

package controllers

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/beego/beego/utils/pagination"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

// GetGlobalUsers
// @Title GetGlobalUsers
// @Tag User API
// @Description get global users
// @Success 200 {array} object.User The Response object
// @router /get-global-users [get]
func (c *ApiController) GetGlobalUsers() {
	limit := c.Input().Get("pageSize")
	page := c.Input().Get("p")
	field := c.Input().Get("field")
	value := c.Input().Get("value")
	sortField := c.Input().Get("sortField")
	sortOrder := c.Input().Get("sortOrder")
	if limit == "" || page == "" {
		c.Data["json"] = object.GetMaskedUsers(object.GetGlobalUsers())
		c.ServeJSON()
	} else {
		limit := util.ParseInt(limit)
		paginator := pagination.SetPaginator(c.Ctx, limit, int64(object.GetGlobalUserCount(field, value)))
		users := object.GetPaginationGlobalUsers(paginator.Offset(), limit, field, value, sortField, sortOrder)
		users = object.GetMaskedUsers(users)
		c.ResponseOk(users, paginator.Nums())
	}
}

// GetUsers
// @Title GetUsers
// @Tag User API
// @Description
// @Param   owner     query    string  true        "The owner of users"
// @Success 200 {array} object.User The Response object
// @router /get-users [get]
func (c *ApiController) GetUsers() {
	owner := c.Input().Get("owner")
	limit := c.Input().Get("pageSize")
	page := c.Input().Get("p")
	field := c.Input().Get("field")
	value := c.Input().Get("value")
	sortField := c.Input().Get("sortField")
	sortOrder := c.Input().Get("sortOrder")
	if limit == "" || page == "" {
		c.Data["json"] = object.GetMaskedUsers(object.GetUsers(owner))
		c.ServeJSON()
	} else {
		limit := util.ParseInt(limit)
		paginator := pagination.SetPaginator(c.Ctx, limit, int64(object.GetUserCount(owner, field, value)))
		users := object.GetPaginationUsers(owner, paginator.Offset(), limit, field, value, sortField, sortOrder)
		users = object.GetMaskedUsers(users)
		c.ResponseOk(users, paginator.Nums())
	}
}

// GetUser
// @Title GetUser
// @Tag User API
// @Description get user
// @Param   id     query    string  true         "The id of the user"
// @Param   owner  query    string  false        "The owner of the user"
// @Param   email  query    string  false 	     "The email of the user"
// @Param   phone  query    string  false 	     "The phone of the user"
// @Success 200 {object} object.User The Response object
// @router /get-user [get]
func (c *ApiController) GetUser() {
	id := c.Input().Get("id")
	email := c.Input().Get("email")
	phone := c.Input().Get("phone")
	userId := c.Input().Get("userId")

	owner := c.Input().Get("owner")
	if owner == "" {
		owner, _ = util.GetOwnerAndNameFromId(id)
	}

	organization := object.GetOrganization(fmt.Sprintf("%s/%s", "admin", owner))
	if !organization.IsProfilePublic {
		requestUserId := c.GetSessionUsername()
		hasPermission, err := object.CheckUserPermission(requestUserId, id, owner, false, c.GetAcceptLanguage())
		if !hasPermission {
			c.ResponseError(err.Error())
			return
		}
	}

	var user *object.User
	switch {
	case email != "":
		user = object.GetUserByEmail(owner, email)
	case phone != "":
		user = object.GetUserByPhone(owner, phone)
	case userId != "":
		user = object.GetUserByUserId(owner, userId)
	default:
		user = object.GetUser(id)
	}

	object.ExtendUserWithRolesAndPermissions(user)

	c.Data["json"] = object.GetMaskedUser(user)
	c.ServeJSON()
}

func checkPermissionForUpdateUser(id string, newUser object.User, c *ApiController) (bool, string) {
	oldUser := object.GetUser(id)
	org := object.GetOrganizationByUser(oldUser)
	var itemsChanged []*object.AccountItem

	if oldUser.Owner != newUser.Owner {
		item := object.GetAccountItemByName("Organization", org)
		itemsChanged = append(itemsChanged, item)
	}
	if oldUser.Name != newUser.Name {
		item := object.GetAccountItemByName("Name", org)
		itemsChanged = append(itemsChanged, item)
	}
	if oldUser.Id != newUser.Id {
		item := object.GetAccountItemByName("ID", org)
		itemsChanged = append(itemsChanged, item)
	}
	if oldUser.DisplayName != newUser.DisplayName {
		item := object.GetAccountItemByName("Display name", org)
		itemsChanged = append(itemsChanged, item)
	}
	if oldUser.Avatar != newUser.Avatar {
		item := object.GetAccountItemByName("Avatar", org)
		itemsChanged = append(itemsChanged, item)
	}
	if oldUser.Type != newUser.Type {
		item := object.GetAccountItemByName("User type", org)
		itemsChanged = append(itemsChanged, item)
	}
	// The password is *** when not modified
	if oldUser.Password != newUser.Password && newUser.Password != "***" {
		item := object.GetAccountItemByName("Password", org)
		itemsChanged = append(itemsChanged, item)
	}
	if oldUser.Email != newUser.Email {
		item := object.GetAccountItemByName("Email", org)
		itemsChanged = append(itemsChanged, item)
	}
	if oldUser.Phone != newUser.Phone {
		item := object.GetAccountItemByName("Phone", org)
		itemsChanged = append(itemsChanged, item)
	}
	if oldUser.Region != newUser.Region {
		item := object.GetAccountItemByName("Country/Region", org)
		itemsChanged = append(itemsChanged, item)
	}
	if oldUser.Location != newUser.Location {
		item := object.GetAccountItemByName("Location", org)
		itemsChanged = append(itemsChanged, item)
	}
	if oldUser.Affiliation != newUser.Affiliation {
		item := object.GetAccountItemByName("Affiliation", org)
		itemsChanged = append(itemsChanged, item)
	}
	if oldUser.Title != newUser.Title {
		item := object.GetAccountItemByName("Title", org)
		itemsChanged = append(itemsChanged, item)
	}
	if oldUser.Homepage != newUser.Homepage {
		item := object.GetAccountItemByName("Homepage", org)
		itemsChanged = append(itemsChanged, item)
	}
	if oldUser.Bio != newUser.Bio {
		item := object.GetAccountItemByName("Bio", org)
		itemsChanged = append(itemsChanged, item)
	}
	if oldUser.Tag != newUser.Tag {
		item := object.GetAccountItemByName("Tag", org)
		itemsChanged = append(itemsChanged, item)
	}
	if oldUser.SignupApplication != newUser.SignupApplication {
		item := object.GetAccountItemByName("Signup application", org)
		itemsChanged = append(itemsChanged, item)
	}

	oldUserRolesJson, _ := json.Marshal(oldUser.Roles)
	newUserRolesJson, _ := json.Marshal(newUser.Roles)
	if string(oldUserRolesJson) != string(newUserRolesJson) {
		item := object.GetAccountItemByName("Roles", org)
		itemsChanged = append(itemsChanged, item)
	}

	oldUserPermissionJson, _ := json.Marshal(oldUser.Permissions)
	newUserPermissionJson, _ := json.Marshal(newUser.Permissions)
	if string(oldUserPermissionJson) != string(newUserPermissionJson) {
		item := object.GetAccountItemByName("Permissions", org)
		itemsChanged = append(itemsChanged, item)
	}

	oldUserPropertiesJson, _ := json.Marshal(oldUser.Properties)
	newUserPropertiesJson, _ := json.Marshal(newUser.Properties)
	if string(oldUserPropertiesJson) != string(newUserPropertiesJson) {
		item := object.GetAccountItemByName("Properties", org)
		itemsChanged = append(itemsChanged, item)
	}

	if oldUser.IsAdmin != newUser.IsAdmin {
		item := object.GetAccountItemByName("Is admin", org)
		itemsChanged = append(itemsChanged, item)
	}
	if oldUser.IsGlobalAdmin != newUser.IsGlobalAdmin {
		item := object.GetAccountItemByName("Is global admin", org)
		itemsChanged = append(itemsChanged, item)
	}
	if oldUser.IsForbidden != newUser.IsForbidden {
		item := object.GetAccountItemByName("Is forbidden", org)
		itemsChanged = append(itemsChanged, item)
	}
	if oldUser.IsDeleted != newUser.IsDeleted {
		item := object.GetAccountItemByName("Is deleted", org)
		itemsChanged = append(itemsChanged, item)
	}

	for i := range itemsChanged {
		if pass, err := object.CheckAccountItemModifyRule(itemsChanged[i], c.getCurrentUser(), c.GetAcceptLanguage()); !pass {
			return pass, err
		}
	}
	return true, ""
}

// UpdateUser
// @Title UpdateUser
// @Tag User API
// @Description update user
// @Param   id     query    string  true        "The id of the user"
// @Param   body    body   object.User  true        "The details of the user"
// @Success 200 {object} controllers.Response The Response object
// @router /update-user [post]
func (c *ApiController) UpdateUser() {
	id := c.Input().Get("id")
	columnsStr := c.Input().Get("columns")

	if id == "" {
		id = c.GetSessionUsername()
	}

	var user object.User
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &user)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if user.DisplayName == "" {
		c.ResponseError(c.T("user:Display name cannot be empty"))
		return
	}

	columns := []string{}
	if columnsStr != "" {
		columns = strings.Split(columnsStr, ",")
	}

	isGlobalAdmin := c.IsGlobalAdmin()

	if pass, err := checkPermissionForUpdateUser(id, user, c); !pass {
		c.ResponseError(err)
		return
	}

	affected := object.UpdateUser(id, &user, columns, isGlobalAdmin)
	if affected {
		object.UpdateUserToOriginalDatabase(&user)
	}

	c.Data["json"] = wrapActionResponse(affected)
	c.ServeJSON()
}

// AddUser
// @Title AddUser
// @Tag User API
// @Description add user
// @Param   body    body   object.User  true        "The details of the user"
// @Success 200 {object} controllers.Response The Response object
// @router /add-user [post]
func (c *ApiController) AddUser() {
	var user object.User
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &user)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	count := object.GetUserCount("", "", "")
	if err := checkQuotaForUser(count); err != nil {
		c.ResponseError(err.Error())
		return
	}

	msg := object.CheckUsername(user.Name, c.GetAcceptLanguage())
	if msg != "" {
		c.ResponseError(msg)
		return
	}

	c.Data["json"] = wrapActionResponse(object.AddUser(&user))
	c.ServeJSON()
}

// DeleteUser
// @Title DeleteUser
// @Tag User API
// @Description delete user
// @Param   body    body   object.User  true        "The details of the user"
// @Success 200 {object} controllers.Response The Response object
// @router /delete-user [post]
func (c *ApiController) DeleteUser() {
	var user object.User
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &user)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = wrapActionResponse(object.DeleteUser(&user))
	c.ServeJSON()
}

// GetEmailAndPhone
// @Title GetEmailAndPhone
// @Tag User API
// @Description get email and phone by username
// @Param   username    formData   string  true        "The username of the user"
// @Param   organization    formData   string  true        "The organization of the user"
// @Success 200 {object} controllers.Response The Response object
// @router /get-email-and-phone [post]
func (c *ApiController) GetEmailAndPhone() {
	var form RequestForm
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &form)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	user := object.GetUserByFields(form.Organization, form.Username)
	if user == nil {
		c.ResponseError(fmt.Sprintf(c.T("general:The user: %s/%s doesn't exist"), form.Organization, form.Username))
		return
	}

	respUser := object.User{Name: user.Name}
	var contentType string
	switch form.Username {
	case user.Email:
		contentType = "email"
		respUser.Email = user.Email
	case user.Phone:
		contentType = "phone"
		respUser.Phone = user.Phone
	case user.Name:
		contentType = "username"
		respUser.Email = util.GetMaskedEmail(user.Email)
		respUser.Phone = util.GetMaskedPhone(user.Phone)
	}

	c.ResponseOk(respUser, contentType)
}

// SetPassword
// @Title SetPassword
// @Tag Account API
// @Description set password
// @Param   userOwner   formData    string  true        "The owner of the user"
// @Param   userName   formData    string  true        "The name of the user"
// @Param   oldPassword   formData    string  true        "The old password of the user"
// @Param   newPassword   formData    string  true        "The new password of the user"
// @Success 200 {object} controllers.Response The Response object
// @router /set-password [post]
func (c *ApiController) SetPassword() {
	userOwner := c.Ctx.Request.Form.Get("userOwner")
	userName := c.Ctx.Request.Form.Get("userName")
	oldPassword := c.Ctx.Request.Form.Get("oldPassword")
	newPassword := c.Ctx.Request.Form.Get("newPassword")

	requestUserId := c.GetSessionUsername()
	userId := fmt.Sprintf("%s/%s", userOwner, userName)

	hasPermission, err := object.CheckUserPermission(requestUserId, userId, userOwner, true, c.GetAcceptLanguage())
	if !hasPermission {
		c.ResponseError(err.Error())
		return
	}

	targetUser := object.GetUser(userId)

	if oldPassword != "" {
		msg := object.CheckPassword(targetUser, oldPassword, c.GetAcceptLanguage())
		if msg != "" {
			c.ResponseError(msg)
			return
		}
	}

	if strings.Contains(newPassword, " ") {
		c.ResponseError(c.T("user:New password cannot contain blank space."))
		return
	}

	if len(newPassword) <= 5 {
		c.ResponseError(c.T("user:New password must have at least 6 characters"))
		return
	}

	targetUser.Password = newPassword
	object.SetUserField(targetUser, "password", targetUser.Password)
	c.Data["json"] = Response{Status: "ok"}
	c.ServeJSON()
}

// CheckUserPassword
// @Title CheckUserPassword
// @router /check-user-password [post]
// @Tag User API
func (c *ApiController) CheckUserPassword() {
	var user object.User
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &user)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	_, msg := object.CheckUserPassword(user.Owner, user.Name, user.Password, c.GetAcceptLanguage())
	if msg == "" {
		c.ResponseOk()
	} else {
		c.ResponseError(msg)
	}
}

// GetSortedUsers
// @Title GetSortedUsers
// @Tag User API
// @Description
// @Param   owner     query    string  true        "The owner of users"
// @Param   sorter     query    string  true        "The DB column name to sort by, e.g., created_time"
// @Param   limit     query    string  true        "The count of users to return, e.g., 25"
// @Success 200 {array} object.User The Response object
// @router /get-sorted-users [get]
func (c *ApiController) GetSortedUsers() {
	owner := c.Input().Get("owner")
	sorter := c.Input().Get("sorter")
	limit := util.ParseInt(c.Input().Get("limit"))

	c.Data["json"] = object.GetMaskedUsers(object.GetSortedUsers(owner, sorter, limit))
	c.ServeJSON()
}

// GetUserCount
// @Title GetUserCount
// @Tag User API
// @Description
// @Param   owner     query    string  true        "The owner of users"
// @Param   isOnline     query    string  true        "The filter for query, 1 for online, 0 for offline, empty string for all users"
// @Success 200 {int} int The count of filtered users for an organization
// @router /get-user-count [get]
func (c *ApiController) GetUserCount() {
	owner := c.Input().Get("owner")
	isOnline := c.Input().Get("isOnline")

	count := 0
	if isOnline == "" {
		count = object.GetUserCount(owner, "", "")
	} else {
		count = object.GetOnlineUserCount(owner, util.ParseInt(isOnline))
	}

	c.Data["json"] = count
	c.ServeJSON()
}
