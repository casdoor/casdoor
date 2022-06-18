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
	"errors"
	"fmt"
	"strings"

	"github.com/casdoor/casdoor/captcha"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

func (c *ApiController) getCurrentUser() *object.User {
	var user *object.User
	userId := c.GetSessionUsername()
	if userId == "" {
		user = nil
	} else {
		user = object.GetUser(userId)
	}
	return user
}

// SendVerificationCode ...
// @Title SendVerificationCode
// @Tag Verification API
// @router /send-verification-code [post]
func (c *ApiController) SendVerificationCode() {
	destType := c.Ctx.Request.Form.Get("type")
	dest := c.Ctx.Request.Form.Get("dest")
	orgId := c.Ctx.Request.Form.Get("organizationId")
	checkType := c.Ctx.Request.Form.Get("checkType")
	checkId := c.Ctx.Request.Form.Get("checkId")
	checkKey := c.Ctx.Request.Form.Get("checkKey")
	checkUser := c.Ctx.Request.Form.Get("checkUser")
	remoteAddr := util.GetIPFromRequest(c.Ctx.Request)

	if len(destType) == 0 || len(dest) == 0 || len(orgId) == 0 || !strings.Contains(orgId, "/") || len(checkType) == 0 {
		c.ResponseError("Missing parameter.")
		return
	}

	provider := captcha.GetCaptchaProvider(checkType)
	if provider == nil {
		c.ResponseError("Invalid captcha provider.")
		return
	}

	if checkKey == "" {
		c.ResponseError("Missing parameter: checkKey.")
		return
	}
	isHuman, err := provider.VerifyCaptcha(checkKey, checkId)
	if err != nil {
		c.ResponseError("Failed to verify captcha: %v", err)
		return
	}

	if !isHuman {
		c.ResponseError("Turing test failed.")
		return
	}

	user := c.getCurrentUser()
	organization := object.GetOrganization(orgId)
	application := object.GetApplicationByOrganizationName(organization.Name)

	if checkUser == "true" && user == nil && object.GetUserByFields(organization.Name, dest) == nil {
		c.ResponseError("Please login first")
		return
	}

	sendResp := errors.New("Invalid dest type")

	if user == nil && checkUser != "" && checkUser != "true" {
		_, name := util.GetOwnerAndNameFromId(orgId)
		user = object.GetUser(fmt.Sprintf("%s/%s", name, checkUser))
	}
	switch destType {
	case "email":
		if user != nil && util.GetMaskedEmail(user.Email) == dest {
			dest = user.Email
		}
		if !util.IsEmailValid(dest) {
			c.ResponseError("Invalid Email address")
			return
		}

		provider := application.GetEmailProvider()
		sendResp = object.SendVerificationCodeToEmail(organization, user, provider, remoteAddr, dest)
	case "phone":
		if user != nil && util.GetMaskedPhone(user.Phone) == dest {
			dest = user.Phone
		}
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
		provider := application.GetSmsProvider()
		sendResp = object.SendVerificationCodeToPhone(organization, user, provider, remoteAddr, dest)
	}

	if sendResp != nil {
		c.Data["json"] = Response{Status: "error", Msg: sendResp.Error()}
	} else {
		c.Data["json"] = Response{Status: "ok"}
	}

	c.ServeJSON()
}

// ResetEmailOrPhone ...
// @Tag Account API
// @Title ResetEmailOrPhone
// @router /api/reset-email-or-phone [post]
func (c *ApiController) ResetEmailOrPhone() {
	userId, ok := c.RequireSignedIn()
	if !ok {
		return
	}

	user := object.GetUser(userId)
	if user == nil {
		c.ResponseError(fmt.Sprintf("The user: %s doesn't exist", userId))
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

// VerifyCaptcha ...
// @Title VerifyCaptcha
// @Tag Verification API
// @router /verify-captcha [post]
func (c *ApiController) VerifyCaptcha() {
	captchaType := c.Ctx.Request.Form.Get("captchaType")

	captchaToken := c.Ctx.Request.Form.Get("captchaToken")
	clientSecret := c.Ctx.Request.Form.Get("clientSecret")
	if captchaToken == "" {
		c.ResponseError("Missing parameter: captchaToken.")
		return
	}
	if clientSecret == "" {
		c.ResponseError("Missing parameter: clientSecret.")
		return
	}

	provider := captcha.GetCaptchaProvider(captchaType)
	if provider == nil {
		c.ResponseError("Invalid captcha provider.")
		return
	}

	isValid, err := provider.VerifyCaptcha(captchaToken, clientSecret)
	if err != nil {
		c.ResponseError("Failed to verify captcha: %v", err)
		return
	}

	c.ResponseOk(isValid)
}
