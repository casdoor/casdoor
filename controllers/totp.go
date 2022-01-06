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

	"github.com/casbin/casdoor/object"
)

type TOTPInit struct {
	Secret        string `json:"secret"`
	RecoveryCodes string `json:"recovery-code"`
	URL           string `json:"qrcode-url"`
}

// InitTOTP
// @Title: InitTOTP
// @Tag: TOTP API
// @Description: Initialize the user's TOTP information and return recovery_code and secret
// @Success: 200 {object} TOTPInit The Response object
// @router: /totp [get]
func (c *ApiController) InitTOTP() {
	userId, ok := c.RequireSignedIn()
	if !ok {
		return
	}
	user := object.GetUser(userId)
	if user == nil {
		c.ResponseError(fmt.Sprintf("The user: %s doesn't exist", userId))
		return
	}
	if user.Is2fa {
		c.ResponseError(fmt.Sprintf("The user: %s has turned on 2fa already", userId))
		return
	}
	application := object.GetApplicationByUser(user)
	key, recoveryCode := object.NewTOTPKey(application.Name, user.Name, application.TotpPeriod, application.TotpSecretSize)

	resp := TOTPInit{
		Secret:        key.Secret(),
		RecoveryCodes: recoveryCode,
		URL:           key.URL(),
	}
	object.SetUserField(user, "totp_secret", key.Secret())
	object.SetUserField(user, "recovery_code", recoveryCode)
	c.Data["json"] = resp
	c.ServeJSON()
}

// SetTOTP
// @Title: SetTOTP
// @Tag: TOTP API
// @Description: Verify that the user has set the correct TOTP
// @Param: secret  formData    string  true        "The totp secret returned by casdoor"
// @Param: code  formData    string  true        "Code generated using totp secret"
// @Success: controllers.Response The Response object
// @router: /totp [post]
func (c *ApiController) SetTOTP() {
	userId, ok := c.RequireSignedIn()
	if !ok {
		return
	}
	user := object.GetUser(userId)
	if user == nil {
		c.ResponseError(fmt.Sprintf("The user: %s doesn't exist", userId))
		return
	}
	if user.Is2fa {
		c.ResponseError(fmt.Sprintf("The user: %s has turned on 2fa already", userId))
		return
	}
	secret := c.Input().Get("secret")
	code := c.Input().Get("code")
	if secret != user.TotpSecret {
		c.ResponseError("secret is not match")
		return
	}

	if object.ValidatePassCode(code, secret) {
		object.SetUserField(user, "is2fa", "1")
		c.ResponseOk()
	} else {
		c.ResponseError("get wrong code")
	}

}

// DeleteTOTP
// @Title: DeleteTOTP
// @Tag: TOTP API
// @Description: Delete TOTP set by the user
// @Param: recovery-code  formData    string  true        "The recovery-code returned by casdoor"
// @Success: controllers.Response The Response object
// @router: /totp [delete]
func (c *ApiController) DeleteTOTP() {
	userId, ok := c.RequireSignedIn()
	if !ok {
		return
	}
	user := object.GetUser(userId)
	if user == nil {
		c.ResponseError(fmt.Sprintf("The user: %s doesn't exist", userId))
		return
	}
	recoveryCode := c.Input().Get("recovery-code")
	if recoveryCode != user.RecoveryCode {
		c.ResponseError("get wrong recovery code")
	}
	object.SetUserField(user, "totp_secret", "")
	object.SetUserField(user, "is2fa", "0")
	object.SetUserField(user, "recovery_code", "")
	c.ResponseOk()
}
