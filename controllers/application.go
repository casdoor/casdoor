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

// @Title GetApplications
// @Description get all applications
// @Param   owner     query    string  true        "The owner of applications."
// @Success 200 {array} object.Application The Response object
// @router /get-applications [get]
func (c *ApiController) GetApplications() {
	owner := c.Input().Get("owner")

	c.Data["json"] = object.GetApplications(owner)
	c.ServeJSON()
}

// @Title GetApplication
// @Description get the detail of an application
// @Param   id     query    string  true        "The id of the application."
// @Success 200 {object} object.Application The Response object
// @router /get-application [get]
func (c *ApiController) GetApplication() {
	id := c.Input().Get("id")

	c.Data["json"] = object.GetApplication(id)
	c.ServeJSON()
}

// @Title GetDefaultApplication
// @Description get the detail of the default application
// @Param   owner     query    string  true        "The owner of the application."
// @Success 200 {object} object.Application The Response object
// @router /get-default-application [get]
func (c *ApiController) GetDefaultApplication() {
	//owner := c.Input().Get("owner")

	if c.GetSessionUser() == "" {
		c.Data["json"] = nil
		c.ServeJSON()
		return
	}

	username := c.GetSessionUser()
	user := object.GetUser(username)

	c.Data["json"] = object.GetApplicationByUser(user)
	c.ServeJSON()
}

// @Title UpdateApplication
// @Description update an application
// @Param   id     query    string  true        "The id of the application"
// @Param   body    body   object.Application  true        "The details of the application"
// @Success 200 {object} controllers.Response The Response object
// @router /update-application [post]
func (c *ApiController) UpdateApplication() {
	id := c.Input().Get("id")

	var application object.Application
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &application)
	if err != nil {
		panic(err)
	}

	c.Data["json"] = wrapActionResponse(object.UpdateApplication(id, &application))
	c.ServeJSON()
}

// @Title AddApplication
// @Description add an application
// @Param   body    body   object.Application  true        "The details of the application"
// @Success 200 {object} controllers.Response The Response object
// @router /add-application [post]
func (c *ApiController) AddApplication() {
	var application object.Application
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &application)
	if err != nil {
		panic(err)
	}

	c.Data["json"] = wrapActionResponse(object.AddApplication(&application))
	c.ServeJSON()
}

// @Title DeleteApplication
// @Description delete an application
// @Param   body    body   object.Application  true        "The details of the application"
// @Success 200 {object} controllers.Response The Response object
// @router /delete-application [post]
func (c *ApiController) DeleteApplication() {
	var application object.Application
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &application)
	if err != nil {
		panic(err)
	}

	c.Data["json"] = wrapActionResponse(object.DeleteApplication(&application))
	c.ServeJSON()
}
