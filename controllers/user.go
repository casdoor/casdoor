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

package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/casdoor/casdoor/util"
	"strings"

	"github.com/casdoor/casdoor/object"
)

// @Title GetGlobalUsers
// @Description get global users
// @Success 200 {array} object.User The Response object
// @router /get-global-users [get]
func (c *ApiController) GetGlobalUsers() {
	c.Data["json"] = object.GetMaskedUsers(object.GetGlobalUsers())
	c.ServeJSON()
}

// @Title GetUsers
// @Description
// @Param   owner     query    string  true        "The owner of users"
// @Success 200 {array} object.User The Response object
// @router /get-users [get]
func (c *ApiController) GetUsers() {
	owner := c.Input().Get("owner")

	c.Data["json"] = object.GetMaskedUsers(object.GetUsers(owner))
	c.ServeJSON()
}

// @Title GetUser
// @Description get user
// @Param   id     query    string  true        "The id of the user"
// @Success 200 {object} object.User The Response object
// @router /get-user [get]
func (c *ApiController) GetUser() {
	id := c.Input().Get("id")

	c.Data["json"] = object.GetMaskedUser(object.GetUser(id))
	c.ServeJSON()
}

// @Title UpdateUser
// @Description update user
// @Param   id     query    string  true        "The id of the user"
// @Param   body    body   object.User  true        "The details of the user"
// @Success 200 {object} controllers.Response The Response object
// @router /update-user [post]
func (c *ApiController) UpdateUser() {
	id := c.Input().Get("id")

	var user object.User
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &user)
	if err != nil {
		panic(err)
	}

	if user.DisplayName == "" {
		c.ResponseError("Display name cannot be empty")
		return
	}

	c.Data["json"] = wrapActionResponse(object.UpdateUser(id, &user))
	c.ServeJSON()
}

// @Title AddUser
// @Description add user
// @Param   body    body   object.User  true        "The details of the user"
// @Success 200 {object} controllers.Response The Response object
// @router /add-user [post]
func (c *ApiController) AddUser() {
	var user object.User
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &user)
	if err != nil {
		panic(err)
	}

	c.Data["json"] = wrapActionResponse(object.AddUser(&user))
	c.ServeJSON()
}

// @Title DeleteUser
// @Description delete user
// @Param   body    body   object.User  true        "The details of the user"
// @Success 200 {object} controllers.Response The Response object
// @router /delete-user [post]
func (c *ApiController) DeleteUser() {
	var user object.User
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &user)
	if err != nil {
		panic(err)
	}

	c.Data["json"] = wrapActionResponse(object.DeleteUser(&user))
	c.ServeJSON()
}

// @Title ForgetPassword
// @Description verify email or phoneId and change password
// @Param   username    formData   string  true        "The username of the user"
// @Param   password    formData   string  true        "The password of the user"
// @Param   organization    formData   string  true        "The organization of the user"
// @Param   email    formData   string  true        "The email of the user"
// @Param   emailCode    formData   string  true        "The emailCode of the user"
// @Success 200 {object} controllers.Response The Response object
// @router /forget-password [post]
func (c *ApiController) ForgetPassword() {
	var resp Response

	var form RequestForm
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &form)
	if err != nil {
		panic(err)
	}

	var verificationCodeType string

	// check result through Email or Phone
	if strings.Contains(form.Email, "@") {
		verificationCodeType = "email"
		checkResult := object.CheckVerificationCode(form.Email, form.EmailCode)
		if len(checkResult) != 0 {
			responseText := fmt.Sprintf("Email%s", checkResult)
			c.ResponseError(responseText)
			return
		}
	} else {
		verificationCodeType = "phone"
		checkPhone := fmt.Sprintf("+%s%s", form.PhonePrefix, form.Email)
		checkResult := object.CheckVerificationCode(checkPhone, form.EmailCode)
		if len(checkResult) != 0 {
			responseText := fmt.Sprintf("Phone%s", checkResult)
			c.ResponseError(responseText)
			return
		}
	}

	// get user
	var userId string
	if form.Username == "" {
		userId, _ = c.RequireSignedIn()
	} else {
		userId = fmt.Sprintf("%s/%s", form.Organization, form.Username)
	}

	user := object.GetUser(userId)
	if user == nil {
		c.ResponseError("No such user.")
		return
	}

	// reset Password
	user.Password = form.Password
	object.SetUserField(user, "password", user.Password)

	msg := object.CheckPassword(user, user.Password)
	if msg != "" {
		resp = Response{Status: "error", Msg: msg, Data: ""}
	} else {
		util.LogInfo(c.Ctx, "API: [%s] has changed password", userId)
		resp = Response{Status: "ok", Msg: "", Data: userId}

		// disable the verification code
		switch verificationCodeType {
		case "email":
			object.DisableVerificationCode(form.Email)
			break
		case "phone":
			object.DisableVerificationCode(form.Phone)
			break
		}
	}

	c.Data["json"] = resp
	c.ServeJSON()
}

// @Title SetPassword
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

	requestUserId := c.GetSessionUser()
	if requestUserId == "" {
		c.ResponseError("Please login first.")
		return
	}
	requestUser := object.GetUser(requestUserId)
	if requestUser == nil {
		c.ResponseError("Session outdated. Please login again.")
		return
	}

	userId := fmt.Sprintf("%s/%s", userOwner, userName)
	targetUser := object.GetUser(userId)
	if targetUser == nil {
		c.ResponseError("Invalid user id.")
		return
	}

	hasPermission := false

	if requestUser.IsGlobalAdmin {
		hasPermission = true
	} else if requestUserId == userId {
		hasPermission = true
	} else if targetUser.Owner == requestUser.Owner && requestUser.IsAdmin {
		hasPermission = true
	}

	if !hasPermission {
		c.ResponseError("You don't have the permission to do this.")
		return
	}

	msg := object.CheckPassword(targetUser, oldPassword)
	if msg != "" {
		c.ResponseError(msg)
		return
	}

	if strings.Index(newPassword, " ") >= 0 {
		c.ResponseError("New password cannot contain blank space.")
		return
	}

	if len(newPassword) <= 5 {
		c.ResponseError("New password must have at least 6 characters")
		return
	}

	targetUser.Password = newPassword
	object.SetUserField(targetUser, "password", targetUser.Password)
	c.Data["json"] = Response{Status: "ok"}
	c.ServeJSON()
}
