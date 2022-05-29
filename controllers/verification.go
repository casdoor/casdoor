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
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/astaxie/beego/logs"
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

func (c *ApiController) verifyReCaptcha(token, secret, verifySite string) bool {
	reqData := url.Values{
		"secret":   {secret},
		"response": {token},
	}
	// resp, err := http.Post(verifySite, "application/x-www-form-urlencoded", strings.NewReader(reqData.Encode()))
	resp, err := http.PostForm(verifySite, reqData)
	if err != nil {
		logs.Error("Failed post to verify captcha: %v", err)
		return false
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		logs.Error("Failed to read from verify captcha: %v", err)
		return false
	}
	type captchaResponse struct {
		Success    bool     `json:"success"`
	}
	captchaResp := &captchaResponse{}
	err = json.Unmarshal(body, captchaResp)
	if err != nil || captchaResp == nil {
		logs.Error("Failed to unmarshal from verify captcha: %v", err)
		return false
	}

	return captchaResp.Success
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
	captchaToken := c.Ctx.Request.Form.Get("captchaToken")

	if len(destType) == 0 || len(dest) == 0 || len(orgId) == 0 || !strings.Contains(orgId, "/") || len(checkType) == 0 {
		c.ResponseError("Missing parameter.")
		return
	}

	isHuman := false
	captchaProvider := object.GetDefaultHumanCheckProvider()
	if captchaProvider == nil {
		if len(checkId) == 0 || len(checkKey) == 0 {
			c.ResponseError("Missing parameter.")
			return
		}
		isHuman = object.VerifyCaptcha(checkId, checkKey)
	} else {
		if len(captchaToken) == 0 {
			c.ResponseError("Missing parameter.")
			return
		}

		if captchaProvider.Type == "reCaptcha" {
			isHuman = c.verifyReCaptcha(captchaToken, captchaProvider.ClientSecret, ReCaptchaVerifySite)
		} else if captchaProvider.Type == "hCaptcha" {
			isHuman = c.verifyReCaptcha(captchaToken, captchaProvider.ClientSecret, HCaptchaVerifySite)
		}
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
