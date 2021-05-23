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
	"fmt"
	"strings"

	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

func (c *ApiController) SendVerificationCode() {
	destType := c.Ctx.Request.Form.Get("type")
	dest := c.Ctx.Request.Form.Get("dest")
	orgId := c.Ctx.Request.Form.Get("organizationId")
	checkType := c.Ctx.Request.Form.Get("checkType")
	checkId := c.Ctx.Request.Form.Get("checkId")
	checkKey := c.Ctx.Request.Form.Get("checkKey")
	remoteAddr := c.Ctx.Request.RemoteAddr
	remoteAddr = remoteAddr[:strings.LastIndex(remoteAddr, ":")]

	if len(destType) == 0 || len(dest) == 0 || len(orgId) == 0 || strings.Index(orgId, "/") < 0 || len(checkType) == 0 || len(checkId) == 0 || len(checkKey) == 0 {
		c.ResponseError("Missing parameter.")
		return
	}

	isHuman := false
	provider := object.GetDefaultHumanCheckProvider()
	if provider == nil {
		isHuman = object.VerifyCaptcha(checkId, checkKey)
	}

	if !isHuman {
		c.ResponseError("Turing test failed.")
		return
	}

	msg := "Invalid dest type."
	switch destType {
	case "email":
		if !util.IsEmailValid(dest) {
			c.ResponseError("Invalid Email address")
			return
		}
		msg = object.SendVerificationCodeToEmail(remoteAddr, dest)
	case "phone":
		if !util.IsPhoneCnValid(dest) {
			c.ResponseError("Invalid phone number")
			return
		}
		org := object.GetOrganization(orgId)
		if org == nil {
			c.ResponseError("Missing parameter.")
			return
		}
		dest = fmt.Sprintf("+%s%s", org.PhonePrefix, dest)
		msg = object.SendVerificationCodeToPhone(remoteAddr, dest)
	}

	status := "ok"
	if msg != "" {
		status = "error"
	}

	c.Data["json"] = Response{Status: status, Msg: msg}
	c.ServeJSON()
}

func (c *ApiController) ResetEmailOrPhone() {
	userId, ok := c.RequireSignedIn()
	if !ok {
		return
	}

	user := object.GetUser(userId)
	if user == nil {
		c.ResponseError("No such user.")
		return
	}

	destType := c.Ctx.Request.Form.Get("type")
	dest := c.Ctx.Request.Form.Get("dest")
	code := c.Ctx.Request.Form.Get("code")
	if len(dest) == 0 || len(code) == 0 || len(destType) == 0 {
		c.ResponseError("Missing parameter.")
		return
	}

	checkDest := dest
	if destType == "phone" {
		org := object.GetOrganizationByUser(user)
		phonePrefix := "86"
		if org != nil && org.PhonePrefix != "" {
			phonePrefix = org.PhonePrefix
		}
		checkDest = fmt.Sprintf("+%s%s", phonePrefix, dest)
	}
	if ret := object.CheckVerificationCode(checkDest, code); len(ret) != 0 {
		c.ResponseError(ret)
		return
	}

	switch destType {
	case "email":
		user.Email = dest
		object.SetUserField(user, "email", user.Email)
	case "phone":
		user.Phone = dest
		object.SetUserField(user, "phone", user.Phone)
	default:
		c.ResponseError("Unknown type.")
		return
	}

	object.DisableVerificationCode(checkDest)
	c.Data["json"] = Response{Status: "ok"}
	c.ServeJSON()
}
