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

func (c *ApiController) GetApplications() {
	owner := c.Input().Get("owner")

	c.Data["json"] = object.GetApplications(owner)
	c.ServeJSON()
}

func (c *ApiController) GetApplication() {
	id := c.Input().Get("id")

	c.Data["json"] = object.GetApplication(id)
	c.ServeJSON()
}

func (c *ApiController) UpdateApplication() {
	id := c.Input().Get("id")

	var application object.Application
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &application)
	if err != nil {
		panic(err)
	}

	c.Data["json"] = object.UpdateApplication(id, &application)
	c.ServeJSON()
}

func (c *ApiController) AddApplication() {
	var application object.Application
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &application)
	if err != nil {
		panic(err)
	}

	c.Data["json"] = object.AddApplication(&application)
	c.ServeJSON()
}

func (c *ApiController) DeleteApplication() {
	var application object.Application
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &application)
	if err != nil {
		panic(err)
	}

	c.Data["json"] = object.DeleteApplication(&application)
	c.ServeJSON()
}
