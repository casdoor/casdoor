// Copyright 2020 The casbin Authors. All Rights Reserved.
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

func (c *ApiController) GetUsers() {
	owner := c.Input().Get("owner")

	c.Data["json"] = object.GetUsers(owner)
	c.ServeJSON()
}

func (c *ApiController) GetUser() {
	id := c.Input().Get("id")

	c.Data["json"] = object.GetUser(id)
	c.ServeJSON()
}

func (c *ApiController) UpdateUser() {
	id := c.Input().Get("id")

	var user object.User
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &user)
	if err != nil {
		panic(err)
	}

	c.Data["json"] = object.UpdateUser(id, &user)
	c.ServeJSON()
}

func (c *ApiController) AddUser() {
	var user object.User
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &user)
	if err != nil {
		panic(err)
	}

	c.Data["json"] = object.AddUser(&user)
	c.ServeJSON()
}

func (c *ApiController) DeleteUser() {
	var user object.User
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &user)
	if err != nil {
		panic(err)
	}

	c.Data["json"] = object.DeleteUser(&user)
	c.ServeJSON()
}
