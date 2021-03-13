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

func (c *ApiController) GetProviders() {
	owner := c.Input().Get("owner")

	c.Data["json"] = object.GetProviders(owner)
	c.ServeJSON()
}

func (c *ApiController) GetProvider() {
	id := c.Input().Get("id")

	c.Data["json"] = object.GetProvider(id)
	c.ServeJSON()
}

func (c *ApiController) UpdateProvider() {
	id := c.Input().Get("id")

	var provider object.Provider
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &provider)
	if err != nil {
		panic(err)
	}

	c.Data["json"] = object.UpdateProvider(id, &provider)
	c.ServeJSON()
}

func (c *ApiController) AddProvider() {
	var provider object.Provider
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &provider)
	if err != nil {
		panic(err)
	}

	c.Data["json"] = object.AddProvider(&provider)
	c.ServeJSON()
}

func (c *ApiController) DeleteProvider() {
	var provider object.Provider
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &provider)
	if err != nil {
		panic(err)
	}

	c.Data["json"] = object.DeleteProvider(&provider)
	c.ServeJSON()
}
