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

	"github.com/casdoor/casdoor/object"
)

type TOTPInit struct {
	Secret        string `json:"secret"`
	RecoveryCodes string `json:"recoveryCode"`
	URL           string `json:"url"`
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
	if user.TwoFactor {
		c.ResponseError(fmt.Sprintf("The user: %s has two-factor authentication enabled", userId))
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
