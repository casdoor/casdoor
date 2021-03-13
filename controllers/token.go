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

func (c *ApiController) GetTokens() {
	owner := c.Input().Get("owner")

	c.Data["json"] = object.GetTokens(owner)
	c.ServeJSON()
}

func (c *ApiController) GetToken() {
	id := c.Input().Get("id")

	c.Data["json"] = object.GetToken(id)
	c.ServeJSON()
}

func (c *ApiController) UpdateToken() {
	id := c.Input().Get("id")

	var token object.Token
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &token)
	if err != nil {
		panic(err)
	}

	c.Data["json"] = object.UpdateToken(id, &token)
	c.ServeJSON()
}

func (c *ApiController) AddToken() {
	var token object.Token
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &token)
	if err != nil {
		panic(err)
	}

	c.Data["json"] = object.AddToken(&token)
	c.ServeJSON()
}

func (c *ApiController) DeleteToken() {
	var token object.Token
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &token)
	if err != nil {
		panic(err)
	}

	c.Data["json"] = object.DeleteToken(&token)
	c.ServeJSON()
}
