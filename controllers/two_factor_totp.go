// Copyright 2022 The casbin Authors. All Rights Reserved.
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
	"net/http"
	"strings"

	"github.com/astaxie/beego"
	jsoniter "github.com/json-iterator/go"

	"github.com/casdoor/casdoor/object"
	"github.com/google/uuid"
)

type TOTPInit struct {
	Secret       string `json:"secret"`
	RecoveryCode string `json:"recoveryCode"`
	URL          string `json:"url"`
}

// TwoFactorSetupInitTOTP
// @Title: Setup init TOTP
// @Tag: Two-Factor API
// @router: /two-factor/setup/totp/init [post]
func (c *ApiController) TwoFactorSetupInitTOTP() {
	userId := jsoniter.Get(c.Ctx.Input.RequestBody, "userId").ToString()
	if len(userId) == 0 {
		c.ResponseError(http.StatusText(http.StatusBadRequest))
		return
	}

	user := object.GetUser(userId)
	if user == nil {
		c.ResponseError("User doesn't exist")
		return
	}
	if len(user.TotpSecret) != 0 {
		c.ResponseError("User has TOTP two-factor authentication enabled")
		return
	}

	application := object.GetApplicationByUser(user)

	issuer := beego.AppConfig.String("appname")
	accountName := fmt.Sprintf("%s/%s", application.Name, user.Name)

	key, err := object.NewTOTPKey(issuer, accountName)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	recoveryCode, err := uuid.NewRandom()
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	resp := TOTPInit{
		Secret:       key.Secret(),
		RecoveryCode: strings.ReplaceAll(recoveryCode.String(), "-", ""),
		URL:          key.URL(),
	}
	c.ResponseOk(resp)
}

// TwoFactorSetupVerityTOTP
// @Title: Setup verity TOTP
// @Tag: Two-Factor API
// @router: /two-factor/setup/totp/verity [post]
func (c *ApiController) TwoFactorSetupVerityTOTP() {
	secret := jsoniter.Get(c.Ctx.Input.RequestBody, "secret").ToString()
	passcode := jsoniter.Get(c.Ctx.Input.RequestBody, "passcode").ToString()
	ok := object.ValidateTOTPPassCode(passcode, secret)
	if ok {
		c.ResponseOk(http.StatusText(http.StatusOK))
	} else {
		c.ResponseError(http.StatusText(http.StatusUnauthorized))
	}
}

// TwoFactorEnableTOTP
// @Title: Enable TOTP
// @Tag: Two-Factor API
// @router: /two-factor/totp [post]
func (c *ApiController) TwoFactorEnableTOTP() {
	userId := jsoniter.Get(c.Ctx.Input.RequestBody, "userId").ToString()
	secret := jsoniter.Get(c.Ctx.Input.RequestBody, "secret").ToString()
	recoveryCode := jsoniter.Get(c.Ctx.Input.RequestBody, "recoveryCode").ToString()

	user := object.GetUser(userId)
	if user == nil {
		c.ResponseError("User doesn't exist")
		return
	}

	object.SetUserField(user, "totp_secret", secret)
	object.SetUserField(user, "two_factor_recovery_code", recoveryCode)

	c.ResponseOk(http.StatusText(http.StatusOK))
}

// TwoFactorRemoveTOTP
// @Title: Remove TOTP
// @Tag: Two-Factor API
// @router: /two-factor/totp [delete]
func (c *ApiController) TwoFactorRemoveTOTP() {
	userId := jsoniter.Get(c.Ctx.Input.RequestBody, "userId").ToString()

	user := object.GetUser(userId)
	if user == nil {
		c.ResponseError("User doesn't exist")
		return
	}

	object.SetUserField(user, "totp_secret", "")
	c.ResponseOk(http.StatusText(http.StatusOK))
}

// TwoFactorAuthTOTP
// @Title: Auth TOTP
// @Tag: TOTP API
// @router: /two-factor/auth/totp [post]
func (c *ApiController) TwoFactorAuthTOTP() {
	totpSessionData := c.GetTOTPSessionData()
	if totpSessionData == nil {
		c.ResponseError(http.StatusText(http.StatusBadRequest))
		return
	}

	user := object.GetUser(totpSessionData.UserId)
	if user == nil {
		c.ResponseError("User does not exist")
		return
	}

	passcode := jsoniter.Get(c.Ctx.Input.RequestBody, "passcode").ToString()
	ok := object.ValidateTOTPPassCode(passcode, user.TotpSecret)
	if ok {
		if totpSessionData.EnableSession {
			c.SetSessionUsername(totpSessionData.UserId)
		}
		if !totpSessionData.AutoSignIn {
			c.setExpireForSession()
		}
		c.ResponseOk(http.StatusText(http.StatusOK))
	} else {
		c.ResponseError(http.StatusText(http.StatusUnauthorized))
	}
}
