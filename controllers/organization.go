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

	"github.com/casbin/casdoor/object"
)

// @Title GetOrganizations
// @Description get organizations
// @Param   owner     query    string  true        "owner"
// @Success 200 {array} object.Organization The Response object
// @router /get-organizations [get]
func (c *ApiController) GetOrganizations() {
	owner := c.Input().Get("owner")

	c.Data["json"] = object.GetOrganizations(owner)
	c.ServeJSON()
}

// @Title GetOrganization
// @Description get organization
// @Param   id     query    string  true        "organization id"
// @Success 200 {object} object.Organization The Response object
// @router /get-organization [get]
func (c *ApiController) GetOrganization() {
	id := c.Input().Get("id")

	c.Data["json"] = object.GetOrganization(id)
	c.ServeJSON()
}

// @Title UpdateOrganization
// @Description update organization
// @Param   id     query    string  true        "The id of the organization"
// @Param   body    body   object.Organization  true        "The details of the organization"
// @Success 200 {object} controllers.Response The Response object
// @router /update-organization [post]
func (c *ApiController) UpdateOrganization() {
	id := c.Input().Get("id")

	var organization object.Organization
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &organization)
	if err != nil {
		panic(err)
	}

	c.Data["json"] = wrapActionResponse(object.UpdateOrganization(id, &organization))
	c.ServeJSON()
}

// @Title AddOrganization
// @Description add organization
// @Param   body    body   object.Organization  true        "The details of the organization"
// @Success 200 {object} controllers.Response The Response object
// @router /add-organization [post]
func (c *ApiController) AddOrganization() {
	var organization object.Organization
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &organization)
	if err != nil {
		panic(err)
	}

	c.Data["json"] = wrapActionResponse(object.AddOrganization(&organization))
	c.ServeJSON()
}

// @Title DeleteOrganization
// @Description delete organization
// @Param   body    body   object.Organization  true        "The details of the organization"
// @Success 200 {object} controllers.Response The Response object
// @router /delete-organization [post]
func (c *ApiController) DeleteOrganization() {
	var organization object.Organization
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &organization)
	if err != nil {
		panic(err)
	}

	c.Data["json"] = wrapActionResponse(object.DeleteOrganization(&organization))
	c.ServeJSON()
}
