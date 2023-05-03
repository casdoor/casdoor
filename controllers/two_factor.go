// Copyright 2023 The Casdoor Authors. All Rights Reserved.
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
	"net/http"

	"github.com/beego/beego"
	"github.com/casdoor/casdoor/object"
)

// TwoFactorSetupInitiate
// @Title TwoFactorSetupInitiate
// @Tag Two-Factor API
// @Description setup totp
// @param userId	form	string	true	" "<owner>/<name>" of user"
// @Success 200 {object}   The Response object
// @router /mfa/setup/initiate [post]
func (c *ApiController) TwoFactorSetupInitiate() {
	userId := c.Ctx.Request.Form.Get("userId")
	authType := c.Ctx.Request.Form.Get("type")

	if len(userId) == 0 {
		c.ResponseError(http.StatusText(http.StatusBadRequest))
		return
	}

	twoFactorUtil := object.GetTwoFactorUtil(authType, nil)
	if twoFactorUtil == nil {
		c.ResponseError("Invalid auth type")
	}
	user := object.GetUser(userId)
	if user == nil {
		c.ResponseError("User doesn't exist")
		return
	}

	issuer := beego.AppConfig.String("appname")
	accountName := user.GetId()

	twoFactorProps, err := twoFactorUtil.Initiate(c.Ctx, issuer, accountName)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	resp := twoFactorProps
	c.ResponseOk(resp)
}

// TwoFactorSetupVerify
// @Title TwoFactorSetupVerify
// @Tag Two-Factor API
// @Description setup verify totp
// @param	secret		form	string	true	"totp secret"
// @param	passcode	form 	string 	true	"totp passcode"
// @Success 200 {object}  Response object
// @router /mfa/setup/totp/verify [post]
func (c *ApiController) TwoFactorSetupVerify() {
	authType := c.Ctx.Request.Form.Get("type")
	passcode := c.Ctx.Request.Form.Get("passcode")

	if authType == "" || passcode == "" {
		c.ResponseError("missing auth type or passcode")
		return
	}
	twoFactorUtil := object.GetTwoFactorUtil(authType, nil)

	err := twoFactorUtil.SetupVerify(c.Ctx, passcode)
	if err != nil {
		c.ResponseError(err.Error())
	} else {
		c.ResponseOk(http.StatusText(http.StatusOK))
	}
}

// TwoFactorSetupEnable
// @Title TwoFactorSetupEnable
// @Tag Two-Factor API
// @Description enable totp
// @param	userId		form	string	true	"Id of user"
// @param  	secret		form	string	true	"totp secret"
// @Success 200 {object}  Response object
// @router /mfa/setup/enable [post]
func (c *ApiController) TwoFactorSetupEnable() {
	userId := c.Ctx.Request.Form.Get("userId")
	authType := c.Ctx.Request.Form.Get("type")
	user := object.GetUser(userId)
	if user == nil {
		c.ResponseError("User doesn't exist")
		return
	}

	twoFactor := object.GetTwoFactorUtil(authType, nil)
	err := twoFactor.Enable(c.Ctx, user)

	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(http.StatusText(http.StatusOK))
}

// TwoFactorAuthVerify
// @Title TwoFactorAuthVerify
// @Tag Totp API
// @Description Auth Totp
// @param	passcode	form	string	true	"totp passcode"
// @Success 200 {object}  Response object
// @router /mfa/auth/verify [post]
func (c *ApiController) TwoFactorAuthVerify() {
	authType := c.Ctx.Request.Form.Get("type")
	passcode := c.Ctx.Request.Form.Get("passcode")
	totpSessionData := c.getMfaSessionData()
	if totpSessionData == nil {
		c.ResponseError(http.StatusText(http.StatusBadRequest))
		return
	}

	user := object.GetUser(totpSessionData.UserId)
	if user == nil {
		c.ResponseError("User does not exist")
		return
	}

	twoFactorUtil := object.GetTwoFactorUtil(authType, user.GetPreferTwoFactor(false))
	err := twoFactorUtil.Verify(passcode)
	if err != nil {
		c.ResponseError(http.StatusText(http.StatusUnauthorized))
	} else {
		if totpSessionData.EnableSession {
			c.SetSessionUsername(totpSessionData.UserId)
		}
		if !totpSessionData.AutoSignIn {
			c.setExpireForSession()
		}
		c.SetSession(object.TwoFactorSessionUserId, "")

		c.ResponseOk(http.StatusText(http.StatusOK))
	}
}

// TwoFactorAuthRecover
// @Title TwoFactorAuthRecover
// @Tag Totp API
// @Description recover mfa authentication
// @param	recoveryCode	form	string	true	"recovery code"
// @Success 200 {object}  Response object
// @router /mfa/auth/recover [post]
func (c *ApiController) TwoFactorAuthRecover() {
	authType := c.Ctx.Request.Form.Get("type")
	recoveryCode := c.Ctx.Request.Form.Get("recoveryCode")

	tfaSessionData := c.getMfaSessionData()
	if tfaSessionData == nil {
		c.ResponseError(http.StatusText(http.StatusBadRequest))
		return
	}

	user := object.GetUser(tfaSessionData.UserId)
	if user == nil {
		c.ResponseError("User does not exist")
		return
	}

	ok, err := object.RecoverTfs(user, recoveryCode, authType)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	if ok {
		if tfaSessionData.EnableSession {
			c.SetSessionUsername(tfaSessionData.UserId)
		}
		if !tfaSessionData.AutoSignIn {
			c.setExpireForSession()
		}
		c.SetSession(object.TwoFactorSessionUserId, "")

		c.ResponseOk(http.StatusText(http.StatusOK))
	} else {
		c.ResponseError(http.StatusText(http.StatusUnauthorized))
	}
}

// TwoFactorDelete
// @Title TwoFactorDelete
// @Tag Two-Factor API
// @Description: Remove Totp
// @param	userId	form	string	true	"Id of user"
// @Success 200 {object}  Response object
// @router /mfa/ [delete]
func (c *ApiController) TwoFactorDelete() {
	authType := c.Ctx.Request.Form.Get("type")
	userId := c.Ctx.Request.Form.Get("userId")
	user := object.GetUser(userId)
	if user == nil {
		c.ResponseError("User doesn't exist")
		return
	}

	twoFactorProps := user.TwoFactorAuth[:0]
	i := 0
	for _, twoFactorProp := range twoFactorProps {
		if twoFactorProp.AuthType != authType {
			twoFactorProps[i] = twoFactorProp
			i++
		}
	}
	user.TwoFactorAuth = twoFactorProps
	object.UpdateUser(userId, user, []string{"two_factor_auth"}, user.IsAdminUser())
	c.ResponseOk(user.TwoFactorAuth)
}
