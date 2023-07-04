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
		maskedUsers, err := object.GetMaskedUsers(object.GetGlobalUsers())
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		c.Data["json"] = maskedUsers
		c.ServeJSON()
	} else {
		limit := util.ParseInt(limit)
		count, err := object.GetGlobalUserCount(field, value)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		paginator := pagination.SetPaginator(c.Ctx, limit, count)
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
	owner := c.Input().Get("owner")
	groupName := c.Input().Get("groupName")
	limit := c.Input().Get("pageSize")
	page := c.Input().Get("p")
	field := c.Input().Get("field")
	value := c.Input().Get("value")
	sortField := c.Input().Get("sortField")
	sortOrder := c.Input().Get("sortOrder")

	if limit == "" || page == "" {
		if groupName != "" {
			maskedUsers, err := object.GetMaskedUsers(object.GetGroupUsers(groupName))
			if err != nil {
				c.ResponseError(err.Error())
				return
			}
			c.ResponseOk(maskedUsers)
			return
		}

		maskedUsers, err := object.GetMaskedUsers(object.GetUsers(owner))
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		c.Data["json"] = maskedUsers
		c.ServeJSON()
	} else {
		limit := util.ParseInt(limit)
		count, err := object.GetUserCount(owner, field, value, groupName)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		paginator := pagination.SetPaginator(c.Ctx, limit, count)
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
	id := c.Input().Get("id")
	email := c.Input().Get("email")
	phone := c.Input().Get("phone")
	userId := c.Input().Get("userId")
	owner := c.Input().Get("owner")
	var err error
	var userFromUserId *object.User
	if userId != "" && owner != "" {
		userFromUserId, err = object.GetUserByUserId(owner, userId)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		id = util.GetId(userFromUserId.Owner, userFromUserId.Name)
	}

	if owner == "" {
		owner = util.GetOwnerFromId(id)
	}

	organization, err := object.GetOrganization(util.GetId("admin", owner))
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if !organization.IsProfilePublic {
		requestUserId := c.GetSessionUsername()
		hasPermission, err := object.CheckUserPermission(requestUserId, id, false, c.GetAcceptLanguage())
		if !hasPermission {
			c.ResponseError(err.Error())
			return
		}
	}

	var user *object.User
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

	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	user.MultiFactorAuths = object.GetAllMfaProps(user, true)
	err = object.ExtendUserWithRolesAndPermissions(user)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	maskedUser, err := object.GetMaskedUser(user)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = maskedUser
	c.ServeJSON()
}

// UpdateUser
// @Title UpdateUser
// @Tag User API
// @Description update user
// @Param   id     query    string  true        "The id ( owner/name ) of the user"
// @Param   body    body   object.User  true        "The details of the user"
// @Success 200 {object} controllers.Response The Response object
// @router /update-user [post]
func (c *ApiController) UpdateUser() {
	id := c.Input().Get("id")
	columnsStr := c.Input().Get("columns")

	var user object.User
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &user)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if id == "" {
		id = c.GetSessionUsername()
		if id == "" {
			c.ResponseError(c.T("general:Missing parameter"))
			return
		}
	}
	oldUser, err := object.GetUser(id)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if oldUser == nil {
		c.ResponseError(fmt.Sprintf(c.T("general:The user: %s doesn't exist"), id))
		return
	}

	if oldUser.Owner == "built-in" && oldUser.Name == "admin" && (user.Owner != "built-in" || user.Name != "admin") {
		c.ResponseError(c.T("auth:Unauthorized operation"))
		return
	}

	if msg := object.CheckUpdateUser(oldUser, &user, c.GetAcceptLanguage()); msg != "" {
		c.ResponseError(msg)
		return
	}

	isAdmin := c.IsAdmin()
	if pass, err := object.CheckPermissionForUpdateUser(oldUser, &user, isAdmin, c.GetAcceptLanguage()); !pass {
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

	count, err := object.GetUserCount("", "", "", "")
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if err := checkQuotaForUser(int(count)); err != nil {
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

	//if userOwner == "built-in" && userName == "admin" {
	//	c.ResponseError(c.T("auth:Unauthorized operation"))
	//	return
	//}

	if strings.Contains(newPassword, " ") {
		c.ResponseError(c.T("user:New password cannot contain blank space."))
		return
	}

	userId := ""
	requestUserId := c.GetSessionUsername()
	if userOwner != "" && userName != "" {
		userId = util.GetId(userOwner, userName)
	} else {
		userId = requestUserId
	}

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
		c.SetSession("verifiedCode", "")
	}

	targetUser, err := object.GetUser(userId)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if oldPassword != "" {
		msg := object.CheckPassword(targetUser, oldPassword, c.GetAcceptLanguage())
		if msg != "" {
			c.ResponseError(msg)
			return
		}
	}

	msg := object.CheckPasswordComplexity(targetUser, newPassword)
	if msg != "" {
		c.ResponseError(msg)
		return
	}

	if requestUserId == userId {
		c.ClearUserSession()
		util.LogInfo(c.Ctx, "API: current user '[%s]' logged out", targetUser.Name)
	} else {
		util.LogInfo(c.Ctx, "API: [%s] logged out", targetUser.Name)
	}

	object.DeleteSessionId(util.GetSessionId(targetUser.Owner, targetUser.Name, targetUser.SignupApplication), c.Ctx.Input.CruSession.SessionID())

	targetUser.Password = newPassword
	_, err = object.SetUserField(targetUser, "password", targetUser.Password)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	c.ResponseOk()
}

// CompleteUserSessions
// @Title CompleteUserSessions
// @Tag User API
// @Description completes all user sessions
// @Param   body    body   object.User  true        "The details of the user"
// @Success 200 {object} controllers.Response The Response object
// @router /api/complete-user-sessions [post]
func (c *ApiController) CompleteUserSessions() {
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
	// TODO: implement this
	c.ResponseOk()
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

	maskedUsers, err := object.GetMaskedUsers(object.GetSortedUsers(owner, sorter, limit))
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = maskedUsers
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

	c.Data["json"] = count
	c.ServeJSON()
}

// AddUserkeys
// @Title AddUserkeys
// @router /add-user-keys [post]
// @Tag User API
func (c *ApiController) AddUserkeys() {
	var user object.User
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &user)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	isAdmin := c.IsAdmin()
	affected, err := object.AddUserkeys(&user, isAdmin)
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

	c.Data["json"] = wrapActionResponse(object.RemoveUserFromGroup(owner, name, groupName))
	c.ServeJSON()
}
