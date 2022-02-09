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

	"github.com/astaxie/beego/utils/pagination"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

// GetCerts
// @Title GetCerts
// @Tag Cert API
// @Description get certs
// @Param   owner     query    string  true        "The owner of certs"
// @Success 200 {array} object.Cert The Response object
// @router /get-certs [get]
func (c *ApiController) GetCerts() {
	owner := c.Input().Get("owner")
	limit := c.Input().Get("pageSize")
	page := c.Input().Get("p")
	field := c.Input().Get("field")
	value := c.Input().Get("value")
	sortField := c.Input().Get("sortField")
	sortOrder := c.Input().Get("sortOrder")
	if limit == "" || page == "" {
		c.Data["json"] = object.GetMaskedCerts(object.GetCerts(owner))
		c.ServeJSON()
	} else {
		limit := util.ParseInt(limit)
		paginator := pagination.SetPaginator(c.Ctx, limit, int64(object.GetCertCount(owner, field, value)))
		certs := object.GetMaskedCerts(object.GetPaginationCerts(owner, paginator.Offset(), limit, field, value, sortField, sortOrder))
		c.ResponseOk(certs, paginator.Nums())
	}
}

// @Title GetCert
// @Tag Cert API
// @Description get cert
// @Param   id    query    string  true        "The id of the cert"
// @Success 200 {object} object.Cert The Response object
// @router /get-cert [get]
func (c *ApiController) GetCert() {
	id := c.Input().Get("id")

	c.Data["json"] = object.GetMaskedCert(object.GetCert(id))
	c.ServeJSON()
}

// @Title UpdateCert
// @Tag Cert API
// @Description update cert
// @Param   id    query    string  true        "The id of the cert"
// @Param   body    body   object.Cert  true        "The details of the cert"
// @Success 200 {object} controllers.Response The Response object
// @router /update-cert [post]
func (c *ApiController) UpdateCert() {
	id := c.Input().Get("id")

	var cert object.Cert
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &cert)
	if err != nil {
		panic(err)
	}

	c.Data["json"] = wrapActionResponse(object.UpdateCert(id, &cert))
	c.ServeJSON()
}

// @Title AddCert
// @Tag Cert API
// @Description add cert
// @Param   body    body   object.Cert  true        "The details of the cert"
// @Success 200 {object} controllers.Response The Response object
// @router /add-cert [post]
func (c *ApiController) AddCert() {
	var cert object.Cert
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &cert)
	if err != nil {
		panic(err)
	}

	c.Data["json"] = wrapActionResponse(object.AddCert(&cert))
	c.ServeJSON()
}

// @Title DeleteCert
// @Tag Cert API
// @Description delete cert
// @Param   body    body   object.Cert  true        "The details of the cert"
// @Success 200 {object} controllers.Response The Response object
// @router /delete-cert [post]
func (c *ApiController) DeleteCert() {
	var cert object.Cert
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &cert)
	if err != nil {
		panic(err)
	}

	c.Data["json"] = wrapActionResponse(object.DeleteCert(&cert))
	c.ServeJSON()
}
