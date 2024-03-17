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

// Casdoor will expose its providers as services to SDK
// We are going to implement those services as APIs here

package controllers

import (
	"fmt"

	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

// FaceIDSigninBegin
// @Title FaceIDSigninBegin
// @Tag Login API
// @Description FaceId Login Flow 1st stage
// @Param   owner     query    string  true        "owner"
// @Param   name     query    string  true        "name"
// @Success 200 {object} controllers.Response The Response object
// @router /faceid-signin-begin [get]
func (c *ApiController) FaceIDSigninBegin() {
	userOwner := c.Input().Get("owner")
	userName := c.Input().Get("name")

	user, err := object.GetUserByFields(userOwner, userName)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	if user == nil {
		c.ResponseError(fmt.Sprintf(c.T("general:The user: %s doesn't exist"), util.GetId(userOwner, userName)))
		return
	}

	if len(user.FaceIds) == 0 {
		c.ResponseError(c.T("check:Face data does not exist, cannot log in"))
		return
	}

	c.ResponseOk()
}
