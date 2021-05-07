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
	"encoding/json"

	"github.com/casdoor/casdoor/object"
)

// @Title GetGlobalUsers
// @Description get global users
// @Success 200 {array} object.User The Response object
// @router /get-global-users [get]
func (c *ApiController) GetGlobalUsers() {
	c.Data["json"] = object.GetMaskedUsers(object.GetGlobalUsers())
	c.ServeJSON()
}

// @Title GetUsers
// @Description
// @Param   owner     query    string  true        "The owner of users"
// @Success 200 {array} object.User The Response object
// @router /get-users [get]
func (c *ApiController) GetUsers() {
	owner := c.Input().Get("owner")

	c.Data["json"] = object.GetMaskedUsers(object.GetUsers(owner))
	c.ServeJSON()
}

// @Title GetUser
// @Description get user
// @Param   id     query    string  true        "The id of the user"
// @Success 200 {object} object.User The Response object
// @router /get-user [get]
func (c *ApiController) GetUser() {
	id := c.Input().Get("id")

	c.Data["json"] = object.GetMaskedUser(object.GetUser(id))
	c.ServeJSON()
}

// @Title UpdateUser
// @Description update user
// @Param   id     query    string  true        "The id of the user"
// @Param   body    body   object.User  true        "The details of the user"
// @Success 200 {object} controllers.Response The Response object
// @router /update-user [post]
func (c *ApiController) UpdateUser() {
	id := c.Input().Get("id")

	var user object.User
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &user)
	if err != nil {
		panic(err)
	}

	c.Data["json"] = wrapActionResponse(object.UpdateUser(id, &user))
	c.ServeJSON()
}

// @Title AddUser
// @Description add user
// @Param   body    body   object.User  true        "The details of the user"
// @Success 200 {object} controllers.Response The Response object
// @router /add-user [post]
func (c *ApiController) AddUser() {
	var user object.User
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &user)
	if err != nil {
		panic(err)
	}

	c.Data["json"] = wrapActionResponse(object.AddUser(&user))
	c.ServeJSON()
}

// @Title DeleteUser
// @Description delete user
// @Param   body    body   object.User  true        "The details of the user"
// @Success 200 {object} controllers.Response The Response object
// @router /delete-user [post]
func (c *ApiController) DeleteUser() {
	var user object.User
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &user)
	if err != nil {
		panic(err)
	}

	c.Data["json"] = wrapActionResponse(object.DeleteUser(&user))
	c.ServeJSON()
}
