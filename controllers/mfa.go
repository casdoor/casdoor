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

// MfaSetupInitiate
// @Title MfaSetupInitiate
// @Tag MFA API
// @Description setup MFA
// @param owner	form	string	true	"owner of user"
// @param name	form	string	true	"name of user"
// @param type	form	string	true	"MFA auth type"
// @Success 200 {object}   The Response object
// @router /mfa/setup/initiate [post]
func (c *ApiController) MfaSetupInitiate() {
	owner := c.Ctx.Request.Form.Get("owner")
	name := c.Ctx.Request.Form.Get("name")
	authType := c.Ctx.Request.Form.Get("type")
	userId := util.GetId(owner, name)

	if len(userId) == 0 {
		c.ResponseError(http.StatusText(http.StatusBadRequest))
		return
	}

	MfaUtil := object.GetMfaUtil(authType, nil)
	if MfaUtil == nil {
		c.ResponseError("Invalid auth type")
	}
	user := object.GetUser(userId)
	if user == nil {
		c.ResponseError("User doesn't exist")
		return
	}

	issuer := beego.AppConfig.String("appname")
	accountName := user.GetId()

	mfaProps, err := MfaUtil.Initiate(c.Ctx, issuer, accountName)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	resp := mfaProps
	c.ResponseOk(resp)
}

// MfaSetupVerify
// @Title MfaSetupVerify
// @Tag MFA API
// @Description setup verify totp
// @param	secret		form	string	true	"MFA secret"
// @param	passcode	form 	string 	true	"MFA passcode"
// @Success 200 {object}  Response object
// @router /mfa/setup/verify [post]
func (c *ApiController) MfaSetupVerify() {
	authType := c.Ctx.Request.Form.Get("type")
	passcode := c.Ctx.Request.Form.Get("passcode")

	if authType == "" || passcode == "" {
		c.ResponseError("missing auth type or passcode")
		return
	}
	MfaUtil := object.GetMfaUtil(authType, nil)

	err := MfaUtil.SetupVerify(c.Ctx, passcode)
	if err != nil {
		c.ResponseError(err.Error())
	} else {
		c.ResponseOk(http.StatusText(http.StatusOK))
	}
}

// MfaSetupEnable
// @Title MfaSetupEnable
// @Tag MFA API
// @Description enable totp
// @param owner	form	string	true	"owner of user"
// @param name	form	string	true	"name of user"
// @param type	form	string	true	"MFA auth type"
// @Success 200 {object}  Response object
// @router /mfa/setup/enable [post]
func (c *ApiController) MfaSetupEnable() {
	owner := c.Ctx.Request.Form.Get("owner")
	name := c.Ctx.Request.Form.Get("name")
	authType := c.Ctx.Request.Form.Get("type")

	user := object.GetUser(util.GetId(owner, name))
	if user == nil {
		c.ResponseError("User doesn't exist")
		return
	}

	twoFactor := object.GetMfaUtil(authType, nil)
	err := twoFactor.Enable(c.Ctx, user)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(http.StatusText(http.StatusOK))
}

// DeleteMfa
// @Title DeleteMfa
// @Tag MFA API
// @Description: Delete MFA
// @param owner	form	string	true	"owner of user"
// @param name	form	string	true	"name of user"
// @param id	form	string	true	"id of user's MFA props"
// @Success 200 {object}  Response object
// @router /delete-mfa/ [post]
func (c *ApiController) DeleteMfa() {
	id := c.Ctx.Request.Form.Get("id")
	owner := c.Ctx.Request.Form.Get("owner")
	name := c.Ctx.Request.Form.Get("name")
	userId := util.GetId(owner, name)

	user := object.GetUser(userId)
	if user == nil {
		c.ResponseError("User doesn't exist")
		return
	}

	mfaProps := user.MultiFactorAuths[:0]
	i := 0
	for _, mfaProp := range mfaProps {
		if mfaProp.Id != id {
			mfaProps[i] = mfaProp
			i++
		}
	}
	user.MultiFactorAuths = mfaProps
	object.UpdateUser(userId, user, []string{"multi_factor_auths"}, user.IsAdminUser())
	c.ResponseOk(user.MultiFactorAuths)
}

// SetPreferredMfa
// @Title SetPreferredMfa
// @Tag MFA API
// @Description: Set specific Mfa Preferred
// @param owner	form	string	true	"owner of user"
// @param name	form	string	true	"name of user"
// @param id	form	string	true	"id of user's MFA props"
// @Success 200 {object}  Response object
// @router /set-preferred-mfa [post]
func (c *ApiController) SetPreferredMfa() {
	id := c.Ctx.Request.Form.Get("id")
	owner := c.Ctx.Request.Form.Get("owner")
	name := c.Ctx.Request.Form.Get("name")
	userId := util.GetId(owner, name)

	user := object.GetUser(userId)
	if user == nil {
		c.ResponseError("User doesn't exist")
		return
	}

	mfaProps := user.MultiFactorAuths
	for i, mfaProp := range user.MultiFactorAuths {
		if mfaProp.Id == id {
			mfaProps[i].IsPreferred = true
		} else {
			mfaProps[i].IsPreferred = false
		}
	}

	object.UpdateUser(userId, user, []string{"multi_factor_auths"}, user.IsAdminUser())

	for i, mfaProp := range mfaProps {
		mfaProps[i] = object.GetMaskedProps(mfaProp)
	}
	c.ResponseOk(mfaProps)
}
