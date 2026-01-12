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

	"github.com/beego/beego/v2/core/utils/pagination"
	"github.com/casdoor/casdoor/conf"
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
	limit := c.Ctx.Input.Query("pageSize")
	page := c.Ctx.Input.Query("p")
	field := c.Ctx.Input.Query("field")
	value := c.Ctx.Input.Query("value")
	sortField := c.Ctx.Input.Query("sortField")
	sortOrder := c.Ctx.Input.Query("sortOrder")

	if limit == "" || page == "" {
		users, err := object.GetMaskedUsers(object.GetGlobalUsers())
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		c.ResponseOk(users)
	} else {
		limit := util.ParseInt(limit)
		count, err := object.GetGlobalUserCount(field, value)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		paginator := pagination.NewPaginator(c.Ctx.Request, limit, count)
		users, err := object.GetPaginationGlobalUsers(paginator.Offset(), limit, field, value, sortField, sortOrder)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		users, err = object.GetMaskedUsers(users)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

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
	owner := c.Ctx.Input.Query("owner")
	groupName := c.Ctx.Input.Query("groupName")
	limit := c.Ctx.Input.Query("pageSize")
	page := c.Ctx.Input.Query("p")
	field := c.Ctx.Input.Query("field")
	value := c.Ctx.Input.Query("value")
	sortField := c.Ctx.Input.Query("sortField")
	sortOrder := c.Ctx.Input.Query("sortOrder")

	if limit == "" || page == "" {
		if groupName != "" {
			users, err := object.GetMaskedUsers(object.GetGroupUsers(util.GetId(owner, groupName)))
			if err != nil {
				c.ResponseError(err.Error())
				return
			}
			c.ResponseOk(users)
			return
		}

		users, err := object.GetMaskedUsers(object.GetUsers(owner))
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		c.ResponseOk(users)
	} else {
		limit := util.ParseInt(limit)
		count, err := object.GetUserCount(owner, field, value, groupName)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		paginator := pagination.NewPaginator(c.Ctx.Request, limit, count)
		users, err := object.GetPaginationUsers(owner, paginator.Offset(), limit, field, value, sortField, sortOrder, groupName)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		users, err = object.GetMaskedUsers(users)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		c.ResponseOk(users, paginator.Nums())
	}
}

// GetUser
// @Title GetUser
// @Tag User API
// @Description get user
// @Param   id     query    string  false        "The id ( owner/name ) of the user"
// @Param   owner  query    string  false        "The owner of the user"
// @Param   email  query    string  false 	     "The email of the user"
// @Param   phone  query    string  false 	     "The phone of the user"
// @Param   userId query    string  false 	     "The userId of the user"
// @Success 200 {object} object.User The Response object
// @router /get-user [get]
func (c *ApiController) GetUser() {
	id := c.Ctx.Input.Query("id")
	email := c.Ctx.Input.Query("email")
	phone := c.Ctx.Input.Query("phone")
	userId := c.Ctx.Input.Query("userId")
	owner := c.Ctx.Input.Query("owner")
	var err error
	var userFromUserId *object.User
	if userId != "" && owner != "" {
		userFromUserId, err = object.GetUserByUserId(owner, userId)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}
		if userFromUserId == nil {
			c.ResponseOk(nil)
			return
		}

		id = util.GetId(userFromUserId.Owner, userFromUserId.Name)
	}

	var user *object.User
	if id == "" && owner == "" {
		switch {
		case email != "":
			user, err = object.GetUserByEmailOnly(email)
		case phone != "":
			user, err = object.GetUserByPhoneOnly(phone)
		case userId != "":
			user, err = object.GetUserByUserIdOnly(userId)
		}
	} else {
		if owner == "" {
			owner = util.GetOwnerFromId(id)
		}

		switch {
		case email != "":
			user, err = object.GetUserByEmail(owner, email)
		case phone != "":
			user, err = object.GetUserByPhone(owner, phone)
		case userId != "":
			user = userFromUserId
		default:
			user, err = object.GetUser(id)
		}
	}

	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	var organization *object.Organization
	if user != nil {
		organization, err = object.GetOrganizationByUser(user)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}
		if organization == nil {
			c.ResponseError(fmt.Sprintf(c.T("auth:The organization: %s does not exist"), owner))
			return
		}

		if !organization.IsProfilePublic {
			requestUserId := c.GetSessionUsername()
			var hasPermission bool
			hasPermission, err = object.CheckUserPermission(requestUserId, user.GetId(), false, c.GetAcceptLanguage())
			if !hasPermission {
				c.ResponseError(err.Error())
				return
			}
		}
	}

	if user != nil {
		user.MultiFactorAuths = object.GetAllMfaProps(user, true)
	}

	err = object.ExtendUserWithRolesAndPermissions(user)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	isAdminOrSelf := c.IsAdminOrSelf(user)
	user, err = object.GetMaskedUser(user, isAdminOrSelf)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if organization != nil && user != nil {
		user, err = object.GetFilteredUser(user, c.IsAdmin(), c.IsAdminOrSelf(user), organization.AccountItems)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}
	}

	c.ResponseOk(user)
}

// UpdateUser
// @Title UpdateUser
// @Tag User API
// @Description update user
// @Param   id     query    string  false        "The id ( owner/name ) of the user"
// @Param   userId query    string  false        "The userId (UUID) of the user"
// @Param   owner  query    string  false        "The owner of the user (required when using userId)"
// @Param   body    body   object.User  true        "The details of the user"
// @Success 200 {object} controllers.Response The Response object
// @router /update-user [post]
func (c *ApiController) UpdateUser() {
	id := c.Ctx.Input.Query("id")
	userId := c.Ctx.Input.Query("userId")
	owner := c.Ctx.Input.Query("owner")
	columnsStr := c.Ctx.Input.Query("columns")

	var user object.User
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &user)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if id == "" && userId == "" {
		id = c.GetSessionUsername()
		if id == "" {
			c.ResponseError(c.T("general:Missing parameter"))
			return
		}
	}

	var userFromUserId *object.User
	if userId != "" && owner != "" {
		userFromUserId, err = object.GetUserByUserId(owner, userId)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}
		if userFromUserId == nil {
			c.ResponseError(fmt.Sprintf(c.T("general:The user: %s doesn't exist"), userId))
			return
		}

		id = util.GetId(userFromUserId.Owner, userFromUserId.Name)
	}

	var oldUser *object.User
	if userId != "" {
		oldUser = userFromUserId
	} else {
		oldUser, err = object.GetUser(id)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}
	}

	if oldUser == nil {
		c.ResponseError(fmt.Sprintf(c.T("general:The user: %s doesn't exist"), id))
		return
	}

	if oldUser.Owner == "built-in" && oldUser.Name == "admin" && (user.Owner != "built-in" || user.Name != "admin") {
		c.ResponseError(c.T("auth:Unauthorized operation"))
		return
	}

	if user.MfaEmailEnabled && user.Email == "" {
		c.ResponseError(c.T("user:MFA email is enabled but email is empty"))
		return
	}

	if user.MfaPhoneEnabled && user.Phone == "" {
		c.ResponseError(c.T("user:MFA phone is enabled but phone number is empty"))
		return
	}

	if msg := object.CheckUpdateUser(oldUser, &user, c.GetAcceptLanguage()); msg != "" {
		c.ResponseError(msg)
		return
	}

	isUsernameLowered := conf.GetConfigBool("isUsernameLowered")
	if isUsernameLowered {
		user.Name = strings.ToLower(user.Name)
	}

	isAdmin := c.IsAdmin()
	allowDisplayNameEmpty := c.Ctx.Input.Query("allowEmpty") != ""
	if pass, err := object.CheckPermissionForUpdateUser(oldUser, &user, isAdmin, allowDisplayNameEmpty, c.GetAcceptLanguage()); !pass {
		c.ResponseError(err)
		return
	}

	columns := []string{}
	if columnsStr != "" {
		columns = strings.Split(columnsStr, ",")
	}

	affected, err := object.UpdateUser(id, &user, columns, isAdmin)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if affected {
		err = object.UpdateUserToOriginalDatabase(&user)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}
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

	if err := checkQuotaForUser(); err != nil {
		c.ResponseError(err.Error())
		return
	}

	emptyUser := object.User{}
	msg := object.CheckUpdateUser(&emptyUser, &user, c.GetAcceptLanguage())
	if msg != "" {
		c.ResponseError(msg)
		return
	}

	// Set RegisterSource based on the current user if not already set
	if user.RegisterType == "" {
		user.RegisterType = "Add User"
	}
	if user.RegisterSource == "" {
		currentUser := c.getCurrentUser()
		if currentUser != nil {
			user.RegisterSource = currentUser.GetId()
		}
	}

	c.Data["json"] = wrapActionResponse(object.AddUser(&user, c.GetAcceptLanguage()))
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

	if user.Owner == "built-in" && user.Name == "admin" {
		c.ResponseError(c.T("auth:Unauthorized operation"))
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
// @router /get-email-and-phone [get]
func (c *ApiController) GetEmailAndPhone() {
	organization := c.Ctx.Request.Form.Get("organization")
	username := c.Ctx.Request.Form.Get("username")

	enableErrorMask2 := conf.GetConfigBool("enableErrorMask2")
	if enableErrorMask2 {
		c.ResponseError("Error")
		return
	}

	user, err := object.GetUserByFields(organization, username)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if user == nil {
		c.ResponseError(fmt.Sprintf(c.T("general:The user: %s doesn't exist"), util.GetId(organization, username)))
		return
	}

	respUser := object.User{Name: user.Name}
	var contentType string
	switch username {
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
	code := c.Ctx.Request.Form.Get("code")

	// if userOwner == "built-in" && userName == "admin" {
	//	c.ResponseError(c.T("auth:Unauthorized operation"))
	//	return
	// }

	userId := util.GetId(userOwner, userName)

	user, err := object.GetUser(userId)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	if user == nil {
		c.ResponseError(fmt.Sprintf(c.T("general:The user: %s doesn't exist"), userId))
		return
	}

	// Get organization to check for password obfuscation settings
	organization, err := object.GetOrganizationByUser(user)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	if organization == nil {
		c.ResponseError(fmt.Sprintf(c.T("auth:the organization: %s is not found"), user.Owner))
		return
	}

	// Deobfuscate passwords if organization has password obfuscator configured
	// Note: Deobfuscation is optional - if it fails, we treat the password as plain text
	// This allows SDKs and raw HTTP API calls to work without obfuscation support
	if organization.PasswordObfuscatorType != "" && organization.PasswordObfuscatorType != "Plain" {
		if oldPassword != "" {
			deobfuscatedOldPassword, deobfuscateErr := util.GetUnobfuscatedPassword(organization.PasswordObfuscatorType, organization.PasswordObfuscatorKey, oldPassword)
			if deobfuscateErr == nil {
				oldPassword = deobfuscatedOldPassword
			}
		}

		if newPassword != "" {
			deobfuscatedNewPassword, deobfuscateErr := util.GetUnobfuscatedPassword(organization.PasswordObfuscatorType, organization.PasswordObfuscatorKey, newPassword)
			if deobfuscateErr == nil {
				newPassword = deobfuscatedNewPassword
			}
		}
	}

	if strings.Contains(newPassword, " ") {
		c.ResponseError(c.T("user:New password cannot contain blank space."))
		return
	}

	requestUserId := c.GetSessionUsername()
	if requestUserId == "" && code == "" {
		c.ResponseError(c.T("general:Please login first"), "Please login first")
		return
	} else if code == "" {
		hasPermission, err := object.CheckUserPermission(requestUserId, userId, true, c.GetAcceptLanguage())
		if !hasPermission {
			c.ResponseError(err.Error())
			return
		}
	} else {
		if code != c.GetSession("verifiedCode") {
			c.ResponseError(c.T("general:Missing parameter"))
			return
		}
		if userId != c.GetSession("verifiedUserId") {
			c.ResponseError(c.T("general:Wrong userId"))
			return
		}
		c.SetSession("verifiedCode", "")
		c.SetSession("verifiedUserId", "")
	}

	targetUser, err := object.GetUser(userId)
	if targetUser == nil {
		c.ResponseError(fmt.Sprintf(c.T("general:The user: %s doesn't exist"), userId))
		return
	}
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	isAdmin := c.IsAdmin()
	if isAdmin {
		if oldPassword != "" {
			err = object.CheckPassword(targetUser, oldPassword, c.GetAcceptLanguage())
			if err != nil {
				c.ResponseError(err.Error())
				return
			}
		}
	} else if code == "" {
		if targetUser.Password != "" || user.Ldap != "" {
			if user.Ldap == "" {
				err = object.CheckPassword(targetUser, oldPassword, c.GetAcceptLanguage())
			} else {
				err = object.CheckLdapUserPassword(targetUser, oldPassword, c.GetAcceptLanguage())
			}
			if err != nil {
				c.ResponseError(err.Error())
				return
			}
		}
	}

	msg := object.CheckPasswordComplexity(targetUser, newPassword, c.GetAcceptLanguage())
	if msg != "" {
		c.ResponseError(msg)
		return
	}

	// Check if the new password is the same as the current password
	if !object.CheckPasswordNotSameAsCurrent(targetUser, newPassword, organization) {
		c.ResponseError(c.T("user:The new password must be different from your current password"))
		return
	}

	application, err := object.GetApplicationByUser(targetUser)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	if application == nil {
		c.ResponseError(fmt.Sprintf(c.T("auth:the application for user %s is not found"), userId))
		return
	}

	clientIp := util.GetClientIpFromRequest(c.Ctx.Request)
	err = object.CheckEntryIp(clientIp, targetUser, application, organization, c.GetAcceptLanguage())
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	targetUser.Password = newPassword
	targetUser.UpdateUserPassword(organization)
	targetUser.NeedUpdatePassword = false
	targetUser.LastChangePasswordTime = util.GetCurrentTime()

	if user.Ldap == "" {
		_, err = object.UpdateUser(userId, targetUser, []string{"password", "password_salt", "need_update_password", "password_type", "last_change_password_time"}, false)
	} else {
		if isAdmin {
			err = object.ResetLdapPassword(targetUser, "", newPassword, c.GetAcceptLanguage())
		} else {
			err = object.ResetLdapPassword(targetUser, oldPassword, newPassword, c.GetAcceptLanguage())
		}
	}

	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk()
}

// CheckUserPassword
// @Title CheckUserPassword
// @router /check-user-password [post]
// @Tag User API
// @Success 200 {object} object.Userinfo The Response object
func (c *ApiController) CheckUserPassword() {
	var user object.User
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &user)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	/*
	 * Verified password with user as subject, if field ldap not empty,
	 * then `isPasswordWithLdapEnabled` is true
	 */
	_, err = object.CheckUserPassword(user.Owner, user.Name, user.Password, c.GetAcceptLanguage(), false, false, user.Ldap != "")
	if err != nil {
		c.ResponseError(err.Error())
	} else {
		c.ResponseOk()
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
	owner := c.Ctx.Input.Query("owner")
	sorter := c.Ctx.Input.Query("sorter")
	limit := util.ParseInt(c.Ctx.Input.Query("limit"))

	users, err := object.GetMaskedUsers(object.GetSortedUsers(owner, sorter, limit))
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(users)
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
	owner := c.Ctx.Input.Query("owner")
	isOnline := c.Ctx.Input.Query("isOnline")

	var count int64
	var err error
	if isOnline == "" {
		count, err = object.GetUserCount(owner, "", "", "")
	} else {
		count, err = object.GetOnlineUserCount(owner, util.ParseInt(isOnline))
	}
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(count)
}

// AddUserKeys
// @Title AddUserKeys
// @router /add-user-keys [post]
// @Tag User API
// @Success 200 {object} object.Userinfo The Response object
func (c *ApiController) AddUserKeys() {
	var user object.User
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &user)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	isAdmin := c.IsAdmin()
	affected, err := object.AddUserKeys(&user, isAdmin)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(affected)
}

func (c *ApiController) RemoveUserFromGroup() {
	owner := c.Ctx.Request.Form.Get("owner")
	name := c.Ctx.Request.Form.Get("name")
	groupName := c.Ctx.Request.Form.Get("groupName")

	organization, err := object.GetOrganization(util.GetId("admin", owner))
	if err != nil {
		return
	}
	item := object.GetAccountItemByName("Groups", organization)
	res, msg := object.CheckAccountItemModifyRule(item, c.IsAdmin(), c.GetAcceptLanguage())
	if !res {
		c.ResponseError(msg)
		return
	}

	affected, err := object.DeleteGroupForUser(util.GetId(owner, name), util.GetId(owner, groupName))
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(affected)
}

// ImpersonateUser
// @Title ImpersonateUser
// @Tag User API
// @Description set impersonation user for current admin session
// @Param   username    formData   string  true        "The username to impersonate (owner/name)"
// @Success 200 {object} controllers.Response The Response object
// @router /impersonation-user [post]
func (c *ApiController) ImpersonateUser() {
	org, ok := c.RequireAdmin()
	if !ok {
		return
	}

	username := c.Ctx.Request.Form.Get("username")
	if username == "" {
		c.ResponseError(c.T("general:Missing parameter"))
		return
	}

	owner, _, err := util.GetOwnerAndNameFromIdWithError(username)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if !(owner == org || org == "") {
		c.ResponseError(c.T("auth:Unauthorized operation"))
		return
	}

	targetUser, err := object.GetUser(username)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if targetUser == nil {
		c.ResponseError(fmt.Sprintf(c.T("general:The user: %s doesn't exist"), username))
		return
	}

	err = c.SetSession("impersonateUser", username)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	c.Ctx.SetCookie("impersonateUser", username)
	c.ResponseOk()
}

// ExitImpersonateUser
// @Title ExitImpersonateUser
// @Tag User API
// @Description clear impersonation info for current session
// @Success 200 {object} controllers.Response The Response object
// @router /exit-impersonation-user [post]
func (c *ApiController) ExitImpersonateUser() {
	_, ok := c.Ctx.Input.GetData("impersonating").(bool)
	if !ok {
		c.ResponseError(c.T("auth:Unauthorized operation"))
		return
	}

	err := c.SetSession("impersonateUser", "")
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	c.Ctx.SetCookie("impersonateUser", "", -1, "/")
	c.ResponseOk()
}

// VerifyIdentification
// @Title VerifyIdentification
// @Tag User API
// @Description verify user's real identity using ID Verification provider
// @Param   owner     query    string  false  "The owner of the user (optional, defaults to logged-in user)"
// @Param   name      query    string  false  "The name of the user (optional, defaults to logged-in user)"
// @Param   provider  query    string  false  "The name of the ID Verification provider (optional, auto-selected if not provided)"
// @Success 200 {object} controllers.Response The Response object
// @router /verify-identification [post]
func (c *ApiController) VerifyIdentification() {
	owner := c.Ctx.Input.Query("owner")
	name := c.Ctx.Input.Query("name")
	providerName := c.Ctx.Input.Query("provider")

	// If user not specified, use logged-in user
	if owner == "" || name == "" {
		loggedInUser := c.GetSessionUsername()
		if loggedInUser == "" {
			c.ResponseError(c.T("general:Please login first"))
			return
		}
		var err error
		owner, name, err = util.GetOwnerAndNameFromIdWithError(loggedInUser)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}
	} else {
		// If user is specified, check if current user has permission to verify other users
		// Only admins can verify other users
		loggedInUser := c.GetSessionUsername()
		if loggedInUser != util.GetId(owner, name) && !c.IsAdmin() {
			c.ResponseError(c.T("auth:Unauthorized operation"))
			return
		}
	}

	user, err := object.GetUser(util.GetId(owner, name))
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if user == nil {
		c.ResponseError(fmt.Sprintf(c.T("general:The user: %s doesn't exist"), util.GetId(owner, name)))
		return
	}

	if user.IdCard == "" || user.IdCardType == "" || user.RealName == "" {
		c.ResponseError(c.T("user:ID card information and real name are required"))
		return
	}

	if user.IsVerified {
		c.ResponseError(c.T("user:User is already verified"))
		return
	}

	var provider *object.Provider
	// If provider not specified, find suitable IDV provider from user's application
	if providerName == "" {
		application, err := object.GetApplicationByUser(user)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		if application == nil {
			c.ResponseError(c.T("user:No application found for user"))
			return
		}

		// Find IDV provider from application
		idvProvider, err := object.GetIdvProviderByApplication(util.GetId(application.Owner, application.Name), "false", c.GetAcceptLanguage())
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		if idvProvider == nil {
			c.ResponseError(c.T("provider:No ID Verification provider configured"))
			return
		}
		provider = idvProvider
	} else {
		provider, err = object.GetProvider(providerName)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		if provider == nil {
			c.ResponseError(fmt.Sprintf(c.T("provider:The provider: %s does not exist"), providerName))
			return
		}

		if provider.Category != "ID Verification" {
			c.ResponseError(c.T("provider:Provider is not an ID Verification provider"))
			return
		}
	}

	idvProvider := object.GetIdvProviderFromProvider(provider)
	if idvProvider == nil {
		c.ResponseError(c.T("provider:Failed to initialize ID Verification provider"))
		return
	}

	verified, err := idvProvider.VerifyIdentity(user.IdCardType, user.IdCard, user.RealName)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if !verified {
		c.ResponseError(c.T("user:Identity verification failed"))
		return
	}

	// Set IsVerified to true upon successful verification
	user.IsVerified = true
	_, err = object.UpdateUser(user.GetId(), user, []string{"is_verified"}, false)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(user.RealName)
}
