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
	"strings"

	"github.com/casdoor/casdoor/object"
)

func (c *ApiController) SendVerificationCode() {
	destType := c.Ctx.Request.Form.Get("type")
	dest := c.Ctx.Request.Form.Get("dest")
	remoteAddr := c.Ctx.Request.RemoteAddr
	remoteAddr = remoteAddr[:strings.LastIndex(remoteAddr, ":")]

	if len(destType) == 0 || len(dest) == 0 {
		c.Data["json"] = Response{Status: "error", Msg: "Missing parameter."}
		c.ServeJSON()
		return
	}

	ret := "Invalid dest type."
	switch destType {
	case "email":
		ret = object.SendVerificationCodeToEmail(remoteAddr, dest)
	case "phone":
		ret = object.SendVerificationCodeToPhone(remoteAddr, dest)
	}

	var status string
	if len(ret) == 0 {
		status = "ok"
	} else {
		status = "error"
	}

	c.Data["json"] = Response{Status: status, Msg: ret}
	c.ServeJSON()
}

func (c *ApiController) ResetEmailOrPhone() {
	userId := c.GetSessionUser()
	if len(userId) == 0 {
		c.ResponseError("Please sign in first")
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

	if ret := object.CheckVerificationCode(dest, code); len(ret) != 0 {
		c.ResponseError(ret)
		return
	}

	switch destType {
	case "email":
		user.Email = dest
		object.SetUserField(user, "email", user.Email)
	case "phone":
		if strings.Index(dest, "+86") == 0 {
			user.PhonePrefix = "86"
			user.Phone = dest[3:]
		} else if strings.Index(dest, "+1") == 0 {
			user.PhonePrefix = "1"
			user.Phone = dest[2:]
		}
		object.SetUserField(user, "phone", user.Phone)
		object.SetUserField(user, "phone_prefix", user.PhonePrefix)
	default:
		c.ResponseError("Unknown type.")
		return
	}

	c.Data["json"] = Response{Status: "ok"}
	c.ServeJSON()
}
