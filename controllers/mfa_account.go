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

// GetMfaAccounts
// @Tag Mfa Account Api
// @Title GetMfaAccounts
// @Description get MFA accounts
// @Param       id          query       string  true        "The id ( owner/name ) of the user"
// @Success     200         {array}     object.MfaAccount    The MfaAccounts object
// @router /get-mfa-accounts [get]
func (c *ApiController) GetMfaAccounts() {
	id := c.Input().Get("id")
	user, err := object.GetUser(id)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	if user == nil {
		c.ResponseError(fmt.Sprintf(c.T("general:The user: %s doesn't exist"), id))
		return
	}

	c.ResponseOk(user.MfaAccounts, len(user.MfaAccounts))
}

// AddMfaAccount
// @Tag Mfa Account Api
// @Title AddMfaAccount
// @Param       id          query       string               true       "The id ( owner/name ) of the user"
// @Param       MfaAccount  body        object.MfaAccount    true       "MfaAccount object"
// @Success     200         {object}    controllers.Response Success or error
// @router /add-mfa-account [post]
func (c *ApiController) AddMfaAccount() {
	id := c.Input().Get("id")

	user, err := object.GetUser(id)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	if user == nil {
		c.ResponseError(fmt.Sprintf(c.T("general:The user: %s doesn't exist"), id))
		return
	}

	var mfaAccount object.MfaAccount
	err = json.Unmarshal(c.Ctx.Input.RequestBody, &mfaAccount)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = wrapActionResponse(object.AddMfaAccount(user, &mfaAccount))
	c.ServeJSON()
}

// DeleteMfaAccount
// @Tag Mfa Account Api
// @Title DeleteMfaAccount
// @Param       id          query       string               true       "The id ( owner/name ) of the user"
// @Param       MfaAccount  body        object.MfaAccount    true       "MfaAccount object"
// @Success     200         {object}    controllers.Response Success or error
// @router /delete-mfa-account [post]
func (c *ApiController) DeleteMfaAccount() {
	id := c.Input().Get("id")

	user, err := object.GetUser(id)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	if user == nil {
		c.ResponseError(fmt.Sprintf(c.T("general:The user: %s doesn't exist"), id))
		return
	}

	var mfaAccount object.MfaAccount
	err = json.Unmarshal(c.Ctx.Input.RequestBody, &mfaAccount)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = wrapActionResponse(object.DeleteMfaAccount(user, &mfaAccount))
	c.ServeJSON()
}

// UpdateMfaAccount
// @Tag Mfa Account Api
// @Title UpdateMfaAccount
// @Param       id          query       string              true       "The id ( owner/name ) of the user"
// @Param       MfaAccount  body        object.MfaAccount   true       "MfaAccount object"
// @Success     200         {object}    controllers.Response Success or error
// @router /update-mfa-account [post]
func (c *ApiController) UpdateMfaAccount() {
	id := c.Input().Get("id")

	user, err := object.GetUser(id)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	if user == nil {
		c.ResponseError(fmt.Sprintf(c.T("general:The user: %s doesn't exist"), id))
		return
	}

	var mfaAccount object.MfaAccount
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
		c.ResponseError("Mfa account updated failed")
		return
	}

	c.ResponseOk("Mfa account updated successfully")
}
