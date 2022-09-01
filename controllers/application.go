// Copyright 2021 The Casdoor Authors. All Rights Reserved.
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

	"github.com/astaxie/beego/utils/pagination"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

// GetApplications
// @Title GetApplications
// @Tag Application API
// @Description get all applications
// @Param   owner     query    string  true        "The owner of applications."
// @Success 200 {array} object.Application The Response object
// @router /get-applications [get]
func (c *ApiController) GetApplications() {
	userId := c.GetSessionUsername()
	owner := c.Input().Get("owner")
	limit := c.Input().Get("pageSize")
	page := c.Input().Get("p")
	field := c.Input().Get("field")
	value := c.Input().Get("value")
	sortField := c.Input().Get("sortField")
	sortOrder := c.Input().Get("sortOrder")
	organization := c.Input().Get("organization")

	if limit == "" || page == "" {
		var applications []*object.Application
		if organization == "" {
			applications = object.GetApplications(owner)
		} else {
			applications = object.GetApplicationsByOrganizationName(owner, organization)
		}

		c.Data["json"] = object.GetMaskedApplications(applications, userId)
		c.ServeJSON()
	} else {
		limit := util.ParseInt(limit)
		paginator := pagination.SetPaginator(c.Ctx, limit, int64(object.GetApplicationCount(owner, field, value)))
		applications := object.GetMaskedApplications(object.GetPaginationApplications(owner, paginator.Offset(), limit, field, value, sortField, sortOrder), userId)
		c.ResponseOk(applications, paginator.Nums())
	}
}

// GetApplication
// @Title GetApplication
// @Tag Application API
// @Description get the detail of an application
// @Param   id     query    string  true        "The id of the application."
// @Success 200 {object} object.Application The Response object
// @router /get-application [get]
func (c *ApiController) GetApplication() {
	userId := c.GetSessionUsername()
	id := c.Input().Get("id")

	c.Data["json"] = object.GetMaskedApplication(object.GetApplication(id), userId)
	c.ServeJSON()
}

// GetUserApplication
// @Title GetUserApplication
// @Tag Application API
// @Description get the detail of the user's application
// @Param   id     query    string  true        "The id of the user"
// @Success 200 {object} object.Application The Response object
// @router /get-user-application [get]
func (c *ApiController) GetUserApplication() {
	userId := c.GetSessionUsername()
	id := c.Input().Get("id")
	user := object.GetUser(id)
	if user == nil {
		c.ResponseError(fmt.Sprintf("The user: %s doesn't exist", id))
		return
	}

	c.Data["json"] = object.GetMaskedApplication(object.GetApplicationByUser(user), userId)
	c.ServeJSON()
}

// GetOrganizationApplications
// @Title GetOrganizationApplications
// @Tag Application API
// @Description get the detail of the organization's application
// @Param   organization     query    string  true        "The organization name"
// @Success 200 {array} object.Application The Response object
// @router /get-organization-applications [get]
func (c *ApiController) GetOrganizationApplications() {
	userId := c.GetSessionUsername()
	owner := c.Input().Get("owner")
	organization := c.Input().Get("organization")

	if organization == "" {
		c.ResponseError("Parameter organization is missing")
		return
	}

	applications := object.GetApplicationsByOrganizationName(owner, organization)
	c.Data["json"] = object.GetMaskedApplications(applications, userId)
	c.ServeJSON()
}

// UpdateApplication
// @Title UpdateApplication
// @Tag Application API
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
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = wrapActionResponse(object.UpdateApplication(id, &application))
	c.ServeJSON()
}

// AddApplication
// @Title AddApplication
// @Tag Application API
// @Description add an application
// @Param   body    body   object.Application  true        "The details of the application"
// @Success 200 {object} controllers.Response The Response object
// @router /add-application [post]
func (c *ApiController) AddApplication() {
	var application object.Application
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &application)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = wrapActionResponse(object.AddApplication(&application))
	c.ServeJSON()
}

// DeleteApplication
// @Title DeleteApplication
// @Tag Application API
// @Description delete an application
// @Param   body    body   object.Application  true        "The details of the application"
// @Success 200 {object} controllers.Response The Response object
// @router /delete-application [post]
func (c *ApiController) DeleteApplication() {
	var application object.Application
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &application)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = wrapActionResponse(object.DeleteApplication(&application))
	c.ServeJSON()
}
