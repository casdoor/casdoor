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
	phonePrefix := c.Ctx.Request.Form.Get("phonePrefix")
	checkType := c.Ctx.Request.Form.Get("checkType")
	checkId := c.Ctx.Request.Form.Get("checkId")
	checkKey := c.Ctx.Request.Form.Get("checkKey")
	checkUser := c.Ctx.Request.Form.Get("checkUser")
	applicationId := c.Ctx.Request.Form.Get("applicationId")
	method := c.Ctx.Request.Form.Get("method")
	remoteAddr := util.GetIPFromRequest(c.Ctx.Request)

	if destType == "" {
		c.ResponseError(c.T("general:Missing parameter") + ": type.")
		return
	}
	if dest == "" {
		c.ResponseError(c.T("general:Missing parameter") + ": dest.")
		return
	}
	if applicationId == "" {
		c.ResponseError(c.T("general:Missing parameter") + ": applicationId.")
		return
	}
	if !strings.Contains(applicationId, "/") {
		c.ResponseError(c.T("verification:Wrong parameter") + ": applicationId.")
		return
	}
	if checkType == "" {
		c.ResponseError(c.T("general:Missing parameter") + ": checkType.")
		return
	}

	captchaProvider := captcha.GetCaptchaProvider(checkType)

	if captchaProvider != nil {
		if checkKey == "" {
			c.ResponseError(c.T("general:Missing parameter") + ": checkKey.")
			return
		}
		isHuman, err := captchaProvider.VerifyCaptcha(checkKey, checkId)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		if !isHuman {
			c.ResponseError(c.T("verification:Turing test failed."))
			return
		}
	}

	user := c.getCurrentUser()
	application := object.GetApplication(applicationId)
	organization := object.GetOrganization(fmt.Sprintf("%s/%s", application.Owner, application.Organization))
	if organization == nil {
		c.ResponseError(c.T("verification:Organization does not exist"))
		return
	}

	if checkUser == "true" && user == nil && object.GetUserByFields(organization.Name, dest) == nil {
		c.ResponseError(c.T("general:Please login first"))
		return
	}

	sendResp := errors.New("invalid dest type")

	if user == nil && checkUser != "" && checkUser != "true" {
		name := application.Organization
		user = object.GetUser(fmt.Sprintf("%s/%s", name, checkUser))
	}
	switch destType {
	case "email":
		if user != nil && util.GetMaskedEmail(user.Email) == dest {
			dest = user.Email
		}
		if !util.IsEmailValid(dest) {
			c.ResponseError(c.T("verification:Email is invalid"))
			return
		}

		userByEmail := object.GetUserByEmail(organization.Name, dest)
		if userByEmail == nil && method != "signup" && method != "reset" {
			c.ResponseError(c.T("verification:the user does not exist, please sign up first"))
			return
		}

		provider := application.GetEmailProvider()
		sendResp = object.SendVerificationCodeToEmail(organization, user, provider, remoteAddr, dest)
	case "phone":
		if user != nil && util.GetMaskedPhone(user.Phone) == dest {
			dest = user.Phone
		}
		if !util.IsPhoneCnValid(dest) {
			c.ResponseError(c.T("verification:Phone number is invalid"))
			return
		}

		userByPhone := object.GetUserByPhone(organization.Name, dest)
		if userByPhone == nil && method != "signup" && method != "reset" {
			c.ResponseError(c.T("verification:the user does not exist, please sign up first"))
			return
		}

		if phonePrefix == "" || phonePrefix == "undefined" {
			if user == nil || user.PhonePrefix == "" {
				//phonePrefix = organization.PhonePrefix
				fmt.Println("this")
			} else {
				phonePrefix = user.PhonePrefix
			}
		}

		dest = fmt.Sprintf("%s%s", phonePrefix, dest)
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
	user, ok := c.RequireSignedInUser()
	if !ok {
		return
	}

	destType := c.Ctx.Request.Form.Get("type")
	dest := c.Ctx.Request.Form.Get("dest")
	code := c.Ctx.Request.Form.Get("code")
	phonePrefix := c.Ctx.Request.Form.Get("phonePrefix")

	if len(dest) == 0 || len(code) == 0 || len(destType) == 0 {
		c.ResponseError(c.T("general:Missing parameter"))
		return
	}

	checkDest := dest
	organization := object.GetOrganizationByUser(user)
	if destType == "phone" {
		if object.HasUserByField(user.Owner, "phone", dest) {
			c.ResponseError(c.T("check:Phone already exists"))
			return
		}

		phoneItem := object.GetAccountItemByName("Phone", organization)
		if phoneItem == nil {
			c.ResponseError(c.T("verification:Unable to get the phone modify rule."))
			return
		}

		if pass, errMsg := object.CheckAccountItemModifyRule(phoneItem, user, c.GetAcceptLanguage()); !pass {
			c.ResponseError(errMsg)
			return
		}

		checkDest = fmt.Sprintf("%s%s", phonePrefix, dest)
	} else if destType == "email" {
		if object.HasUserByField(user.Owner, "email", user.Email) {
			c.ResponseError(c.T("check:Email already exists"))
			return
		}

		emailItem := object.GetAccountItemByName("Email", organization)
		if emailItem == nil {
			c.ResponseError(c.T("verification:Unable to get the email modify rule."))
			return
		}

		if pass, errMsg := object.CheckAccountItemModifyRule(emailItem, user, c.GetAcceptLanguage()); !pass {
			c.ResponseError(errMsg)
			return
		}
	}
	if ret := object.CheckVerificationCode(checkDest, code, c.GetAcceptLanguage()); len(ret) != 0 {
		c.ResponseError(ret)
		return
	}

	switch destType {
	case "email":
		user.Email = dest
		object.SetUserField(user, "email", user.Email)
	case "phone":
		user.Phone = dest
		user.PhonePrefix = phonePrefix
		object.SetUserField(user, "phone_prefix", user.PhonePrefix)
		object.SetUserField(user, "phone", user.Phone)
	default:
		c.ResponseError(c.T("verification:Unknown type"))
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
		c.ResponseError(c.T("general:Missing parameter") + ": captchaToken.")
		return
	}
	if clientSecret == "" {
		c.ResponseError(c.T("general:Missing parameter") + ": clientSecret.")
		return
	}

	provider := captcha.GetCaptchaProvider(captchaType)
	if provider == nil {
		c.ResponseError(c.T("verification:Invalid captcha provider."))
		return
	}

	isValid, err := provider.VerifyCaptcha(captchaToken, clientSecret)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(isValid)
}
