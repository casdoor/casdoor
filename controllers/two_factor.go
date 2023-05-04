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
	"github.com/casdoor/casdoor/util"
)

// TwoFactorSetupInitiate
// @Title TwoFactorSetupInitiate
// @Tag Two-Factor API
// @Description setup totp
// @param userId	form	string	true	" "<owner>/<name>" of user"
// @Success 200 {object}   The Response object
// @router /mfa/setup/initiate [post]
func (c *ApiController) TwoFactorSetupInitiate() {
	owner := c.Ctx.Request.Form.Get("owner")
	name := c.Ctx.Request.Form.Get("name")
	authType := c.Ctx.Request.Form.Get("type")
	userId := util.GetId(owner, name)

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
	owner := c.Ctx.Request.Form.Get("owner")
	name := c.Ctx.Request.Form.Get("name")
	authType := c.Ctx.Request.Form.Get("type")

	user := object.GetUser(util.GetId(owner, name))
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

// TwoFactorDelete
// @Title TwoFactorDelete
// @Tag Two-Factor API
// @Description: Remove Totp
// @param	userId	form	string	true	"Id of user"
// @Success 200 {object}  Response object
// @router /mfa/ [delete]
func (c *ApiController) TwoFactorDelete() {
	id := c.Ctx.Request.Form.Get("id")
	owner := c.Ctx.Request.Form.Get("owner")
	name := c.Ctx.Request.Form.Get("name")
	userId := util.GetId(owner, name)

	user := object.GetUser(userId)
	if user == nil {
		c.ResponseError("User doesn't exist")
		return
	}

	twoFactorProps := user.TwoFactorAuth[:0]
	i := 0
	for _, twoFactorProp := range twoFactorProps {
		if twoFactorProp.Id != id {
			twoFactorProps[i] = twoFactorProp
			i++
		}
	}
	user.TwoFactorAuth = twoFactorProps
	object.UpdateUser(userId, user, []string{"two_factor_auth"}, user.IsAdminUser())
	c.ResponseOk(user.TwoFactorAuth)
}
