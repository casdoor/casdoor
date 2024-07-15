// Copyright 2024 The Casdoor Authors. All Rights Reserved.
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
	"fmt"
	"github.com/casdoor/casdoor/object"
)

// GetMFAAccount
// @Tag MFA Account Api
// @Title GetMFAAccount
// @Description get MFA accounts
// @Param       id          query       string  true        "The id ( owner/name ) of the user"
// @Success     200         {array}     object.MFAAccount    The MFAAccounts object
// @router /get-mfa-accounts [get]
func (c *ApiController) GetMFAAccount() {
	id := c.Input().Get("id")
	user, err := object.GetUserByUserIdOnly(id)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	if user == nil {
		c.ResponseError(fmt.Sprintf(c.T("general:The user: %s doesn't exist"), id))
		return
	}

	c.ResponseOk(user.MFAAccounts, len(user.MFAAccounts))
}

// AddMFAAccounts
// @Tag MFA Account Api
// @Title AddMFAAccounts
// @Param       id          query       string               true       "The id ( owner/name ) of the user"
// @Param       MFAAccount  body        object.MFAAccount    true       "MFAAccount object"
// @Success     200         {object}    controllers.Response Success or error
// @router /add-mfa-account [post]
func (c *ApiController) AddMFAAccounts() {
	id := c.Input().Get("id")

	user, err := object.GetUserByUserIdOnly(id)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	if user == nil {
		c.ResponseError(fmt.Sprintf(c.T("general:The user: %s doesn't exist"), id))
		return
	}

	var mfaAccount object.MFAAccount
	err = json.Unmarshal(c.Ctx.Input.RequestBody, &mfaAccount)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = wrapActionResponse(object.AddMfaAccount(user, &mfaAccount))
	c.ServeJSON()
}

// DeleteMFAAccount
// @Tag MFA Account Api
// @Title DeleteMFAAccount
// @Param       id          query       string               true       "The id ( owner/name ) of the user"
// @Param       MFAAccount  body        object.MFAAccount    true       "MFAAccount object"
// @Success     200         {object}    controllers.Response Success or error
// @router /delete-mfa-account [post]
func (c *ApiController) DeleteMFAAccount() {
	id := c.Input().Get("id")

	user, err := object.GetUserByUserIdOnly(id)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	if user == nil {
		c.ResponseError(fmt.Sprintf(c.T("general:The user: %s doesn't exist"), id))
		return
	}

	var mfaAccount object.MFAAccount
	err = json.Unmarshal(c.Ctx.Input.RequestBody, &mfaAccount)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = wrapActionResponse(object.DeleteMfaAccount(user, &mfaAccount))
	c.ServeJSON()
}

// UpdateMFAAccount
// @Tag MFA Account Api
// @Title UpdateMFAAccount
// @Param       id          query       string              true       "The id ( owner/name ) of the user"
// @Param       MFAAccount  body        object.MFAAccount   true       "MFAAccount object"
// @Success     200         {object}    controllers.Response Success or error
// @router /update-mfa-account [post]
func (c *ApiController) UpdateMFAAccount() {
	id := c.Input().Get("id")

	user, err := object.GetUserByUserIdOnly(id)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	if user == nil {
		c.ResponseError(fmt.Sprintf(c.T("general:The user: %s doesn't exist"), id))
		return
	}

	var mfaAccount object.MFAAccount
	err = json.Unmarshal(c.Ctx.Input.RequestBody, &mfaAccount)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	success, err := object.UpdateMfaAccount(user, &mfaAccount)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	if !success {
		c.ResponseError("MFAAccount updated failed")
		return
	}

	c.ResponseOk("MFAAccount updated successfully")
}
