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
	"github.com/casdoor/casdoor/object"
	"github.com/google/uuid"
)

type TotpInit struct {
	Secret       string `json:"secret"`
	RecoveryCode string `json:"recoveryCode"`
	URL          string `json:"url"`
}

// TwoFactorSetupInitTotp
// @Title TwoFactorSetupInitTotp
// @Tag Two-Factor API
// @Description setup totp
// @param userId	form	string	true	"Id of user"
// @Success 200 {object}  controllers.TotpInit The Response object
// @router /two-factor/setup/totp/init [post]
func (c *ApiController) TwoFactorSetupInitTotp() {
	userId := c.Ctx.Request.Form.Get("userId")
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
		c.ResponseError("User has Totp two-factor authentication enabled")
		return
	}

	application := object.GetApplicationByUser(user)

	issuer := beego.AppConfig.String("appname")
	accountName := fmt.Sprintf("%s/%s", application.Name, user.Name)

	key, err := object.NewTotpKey(issuer, accountName)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	recoveryCode, err := uuid.NewRandom()
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	resp := TotpInit{
		Secret:       key.Secret(),
		RecoveryCode: strings.ReplaceAll(recoveryCode.String(), "-", ""),
		URL:          key.URL(),
	}
	c.ResponseOk(resp)
}

// TwoFactorSetupVerityTotp
// @Title TwoFactorSetupVerityTotp
// @Tag Two-Factor API
// @Description setup verity totp
// @param	secret		form	string	true	"totp secret"
// @param	passcode	form 	string 	true	"totp passcode"
// @Success 200 {object}  Response object
// @router /two-factor/setup/totp/verity [post]
func (c *ApiController) TwoFactorSetupVerityTotp() {
	secret := c.Ctx.Request.Form.Get("secret")
	passcode := c.Ctx.Request.Form.Get("passcode")
	ok := object.ValidateTotpPassCode(passcode, secret)
	if ok {
		c.ResponseOk(http.StatusText(http.StatusOK))
	} else {
		c.ResponseError(http.StatusText(http.StatusUnauthorized))
	}
}

// TwoFactorEnableTotp
// @Title TwoFactorEnableTotp
// @Tag Two-Factor API
// @Description enable totp
// @param	userId		form	string	true	"Id of user"
// @param  	secret		form	string	true	"totp secret"
// @Success 200 {object}  Response object
// @router /two-factor/totp [post]
func (c *ApiController) TwoFactorEnableTotp() {
	userId := c.Ctx.Request.Form.Get("userId")
	secret := c.Ctx.Request.Form.Get("secret")
	recoveryCode := c.Ctx.Request.Form.Get("recoveryCode")

	user := object.GetUser(userId)
	if user == nil {
		c.ResponseError("User doesn't exist")
		return
	}

	object.SetUserField(user, "totp_secret", secret)
	object.SetUserField(user, "two_factor_recovery_code", recoveryCode)

	c.ResponseOk(http.StatusText(http.StatusOK))
}

// TwoFactorRemoveTotp
// @Title TwoFactorRemoveTotp
// @Tag Two-Factor API
// @Description: Remove Totp
// @param	userId	form	string	true	"Id of user"
// @Success 200 {object}  Response object
// @router /two-factor/totp [delete]
func (c *ApiController) TwoFactorRemoveTotp() {
	userId := c.Ctx.Request.Form.Get("userId")
	user := object.GetUser(userId)
	if user == nil {
		c.ResponseError("User doesn't exist")
		return
	}

	object.SetUserField(user, "totp_secret", "")
	c.ResponseOk(http.StatusText(http.StatusOK))
}

// TwoFactorAuthTotp
// @Title TwoFactorAuthTotp
// @Tag Totp API
// @Description Auth Totp
// @param	passcode	form	string	true	"totp passcode"
// @Success 200 {object}  Response object
// @router /two-factor/auth/totp [post]
func (c *ApiController) TwoFactorAuthTotp() {
	totpSessionData := c.getTotpSessionData()
	if totpSessionData == nil {
		c.ResponseError(http.StatusText(http.StatusBadRequest))
		return
	}

	user := object.GetUser(totpSessionData.UserId)
	if user == nil {
		c.ResponseError("User does not exist")
		return
	}

	passcode := c.Ctx.Request.Form.Get("passcode")
	ok := object.ValidateTotpPassCode(passcode, user.TotpSecret)
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

// TwoFactorRecoverTotp
// @Title TwoFactorRecoverTotp
// @Tag Totp API
// @Description recover totp
// @param	recoveryCode	form	string	true	"totp recovery code"
// @Success 200 {object}  Response object
// @router /two-factor/auth/totp/recover [post]
func (c *ApiController) TwoFactorRecoverTotp() {
	totpSessionData := c.getTotpSessionData()
	if totpSessionData == nil {
		c.ResponseError(http.StatusText(http.StatusBadRequest))
		return
	}

	user := object.GetUser(totpSessionData.UserId)
	if user == nil {
		c.ResponseError("User does not exist")
		return
	}

	recoveryCode := c.Ctx.Request.Form.Get("recoveryCode")
	ok, err := object.RecoverTotp(user, recoveryCode)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
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
