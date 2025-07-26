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

	"github.com/beego/beego/logs"
	"github.com/beego/beego/utils/pagination"
	"github.com/casdoor/casdoor/errorx"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
	"github.com/google/uuid"
	"github.com/mcuadros/go-defaults"
)

// GetApplications
// @Title GetApplications
// @Tag Application API
// @Description get all applications
// @Param 		pageSize 	query 		integer 			false 				"Page Size"
// @Param 		p 			query 		integer 				false 				"Page Number"
// @Param 		query 		query 		string 				false 				"Field"
// @Param 		sortField 	query 		string 				false 				"Sort Field"
// @Param 		sortOrder 	query 		string 				false 				"Sort Order: asc,desc"
// @Success 200 {array} object.ApplicationInfo The Response object
// @router /applications [get]
func (c *ApiController) GetApplications() {
	params := c.GetQueryParams()
	userId := c.GetSessionUsername()

	matchFilters, queryFilters := make(object.And), make(object.Or)
	matchFilters["organization"] = params.Organization
	matchFilters["is_shared"] = true

	queryFilters["displayName"] = params.Query
	applications, count, err := Query[object.Application](
		c,
		matchFilters,
		queryFilters,
		params,
	)
	if err != nil {
		c.ResponseErr(err)
		return
	}
	applications = object.GetMaskedApplications(applications, userId)
	c.ResponseOk(QueryResult(object.GetApplicationInfos(applications), count))
}

// GetApplication
// @Title GetApplication
// @Tag Application API old
// @Description get the detail of an application， 不知道为啥一定要有的一个applications开头的url
// @Param   id     query    string  true        "The id ( owner/name ) of the application."
// @Success 200 {object} object.Application The Response object
// @router applications/:appId [get]
func (c *ApiController) GetApplication2() {

}

// GetApplication
// @Title GetApplication
// @Tag Application API
// @Description get the detail of an application
// @Param   name     path    string  true        "The id ( name ) of the application."
// @Success 200 {object} object.ApplicationDetail The Response object
// @router /applications/:name [get]
func (c *ApiController) GetApplication() {
	userId := c.GetSessionUsername()
	id := c.Ctx.Input.Param(":name")
	organization := c.getOrganization()

	application, err := object.GetApplicationByOrganization(organization, id)
	if err != nil {
		c.ResponseErr(err)
		return
	}

	if c.Input().Get("withKey") != "" && application != nil && application.Cert != "" {
		cert, err := object.GetCert(util.GetId(application.Owner, application.Cert))
		if err != nil {
			c.ResponseErr(err)
			return
		}

		if cert == nil {
			cert, err = object.GetCert(util.GetId(application.Organization, application.Cert))
			if err != nil {
				c.ResponseErr(err)
				return
			}
		}

		if cert != nil {
			application.CertPublicKey = cert.Certificate
		}
	}

	clientIp := util.GetClientIpFromRequest(c.Ctx.Request)
	object.CheckEntryIp(clientIp, nil, application, nil, c.GetAcceptLanguage())

	c.ResponseOk(object.GetApplicationInfo(object.GetMaskedApplication(application, userId)))
}

// GetUserApplication
// @Title GetUserApplication
// @Tag Application API2
// @Description get the detail of the user's application
// @Param   id     query    string  true        "The id ( owner/name ) of the user"
// @Success 200 {object} object.Application The Response object
// @router /get-user-application [get]
func (c *ApiController) GetUserApplication() {
	userId := c.GetSessionUsername()
	id := c.Input().Get("id")

	user, err := object.GetUser(id)
	if err != nil {
		c.ResponseErr(err)
		return
	}
	if user == nil {
		c.ResponseError(fmt.Sprintf(c.T("general:The user: %s doesn't exist"), id))
		return
	}

	application, err := object.GetApplicationByUser(user)
	if err != nil {
		c.ResponseErr(err)
		return
	}
	if application == nil {
		c.ResponseError(fmt.Sprintf(c.T("general:The organization: %s should have one application at least"), user.Owner))
		return
	}

	c.ResponseOk(object.GetMaskedApplication(application, userId))
}

// GetOrganizationApplications
// @Title GetOrganizationApplications
// @Tag Application API2
// @Description get the detail of the organization's application
// @Param   organization     query    string  true        "The organization name"
// @Success 200 {array} object.Application The Response object
// @router /get-organization-applications [get]
func (c *ApiController) GetOrganizationApplications() {
	userId := c.GetSessionUsername()
	organization := c.Input().Get("organization")
	owner := c.Input().Get("owner")
	limit := c.Input().Get("pageSize")
	page := c.Input().Get("p")
	field := c.Input().Get("field")
	value := c.Input().Get("value")
	sortField := c.Input().Get("sortField")
	sortOrder := c.Input().Get("sortOrder")

	if organization == "" {
		c.ResponseError(c.T("general:Missing parameter") + ": organization")
		return
	}

	if limit == "" || page == "" {
		applications, err := object.GetOrganizationApplications(owner, organization)
		if err != nil {
			c.ResponseErr(err)
			return
		}

		applications, err = object.GetAllowedApplications(applications, userId, c.GetAcceptLanguage())
		if err != nil {
			c.ResponseErr(err)
			return
		}

		c.ResponseOk(object.GetMaskedApplications(applications, userId))
	} else {
		limit := util.ParseInt(limit)

		count, err := object.GetOrganizationApplicationCount(owner, organization, field, value)
		if err != nil {
			c.ResponseErr(err)
			return
		}

		paginator := pagination.SetPaginator(c.Ctx, limit, count)
		applications, err := object.GetPaginationOrganizationApplications(owner, organization, paginator.Offset(), limit, field, value, sortField, sortOrder)
		if err != nil {
			c.ResponseErr(err)
			return
		}

		applications, err = object.GetAllowedApplications(applications, userId, c.GetAcceptLanguage())
		if err != nil {
			c.ResponseErr(err)
			return
		}

		applications = object.GetMaskedApplications(applications, userId)
		c.ResponseOk(applications, paginator.Nums())
	}
}

// UpdateApplication
// @Title UpdateApplication
// @Tag Application API
// @Description update an application
// @Param   name     path    string  true        "The id ( owner/name ) of the application"
// @Param   body    body   object.AddApplicationInfo  true        "The details of the application"
// @Success 200 {object} controllers.Response The Response object
// @router /applications/:name [put]
func (c *ApiController) UpdateApplication() {
	appId := c.Ctx.Input.Param(":name")

	application, err := object.GetApplicationByOrganization(c.getOrganization(), appId)
	if err != nil {
		c.ResponseErr(err)
		return
	}
	if application == nil {
		c.ResponseErr(errorx.NotFoundAppErr)
		return
	}
	id := util.GetId(application.Owner, appId)


	organization, owner := application.Organization, application.Owner
	if err != nil {
		c.ResponseErr(err)
		return
	}

	err = json.Unmarshal(c.Ctx.Input.RequestBody, application)
	if err != nil {
		c.ResponseErr(err)
		return
	}
	userId := c.GetSessionUsername()
	userOwner, _ := util.GetOwnerAndNameFromId(userId)

	// 禁止修改
	if  userOwner != "built-in" {
		application.Organization = organization
		application.Owner = owner
	}

	if err = object.CheckIpWhitelist(application.IpWhitelist, c.GetAcceptLanguage()); err != nil {
		c.ResponseErr(err)
		return
	}
	c.Data["json"] = c.WrapResponse(object.UpdateApplication(id, application))
	c.ServeJSON()
}

// AddApplication
// @Title AddApplication
// @Tag Application API
// @Description add an application
// @Param   body    body   object.AddApplicationInfo  true        "The details of the application"
// @Success 200 {object} controllers.Response The Response object
// @router /applications [post]
func (c *ApiController) AddApplication() {
	var application object.Application
	defaults.SetDefaults(&application)
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &application)
	if err != nil {
		c.ResponseErr(errorx.InvalidParamErr)
		return
	}

	if err := checkQuotaForApplication(); err != nil {
		logs.Error("checkQuotaForApplication, err=%s", err)
		c.ResponseErr(err)
		return
	}

	if err = object.CheckIpWhitelist(application.IpWhitelist, c.GetAcceptLanguage()); err != nil {
		logs.Error("CheckIpWhitelist, err=%s", err)
		c.ResponseErr(err)
		return
	}
	organization := c.getOrganization()
	userId := c.GetSessionUsername()
	owner, _ := util.GetOwnerAndNameFromId(userId)
	if application.Organization == "" || owner != "built-in" {
		application.Organization = organization
	}
	if application.Owner == "" || owner != "built-in" {
		application.Owner = organization
	}

	if application.Name == "" {
		application.Name = uuid.NewString()
	}

	c.Data["json"] = c.WrapResponse(object.AddApplication(&application))
	c.ServeJSON()
}

// DeleteApplication
// @Title DeleteApplication
// @Tag Application API
// @Description delete an application
// @Param   body    body   object.DeleteApplicationParams   true        "application.name 列表"
// @Success 200 {object} controllers.Response The Response object
// @router /applications [delete]
func (c *ApiController) DeleteApplication() {
	var application object.DeleteApplicationParams
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &application)
	if err != nil {
		c.ResponseErr(errorx.InvalidParamErr)
		return
	}
	applicationIds := application.Applications
	if len(applicationIds) == 0 {
		c.ResponseOk(true)
		return
	}

	organization := c.getOrganization()
	var apps []*object.Application
	for _, applicationId := range applicationIds {
		// 内置管理可以删除任何应用，其他管理员只允许删除组织下的应用
		if organization != "built-in" {
			apps = append(apps, &object.Application{
				Owner: organization,
				Name:  applicationId,
			})
		}
	}

	c.Data["json"] = c.WrapResponse(object.DeleteApplication(apps))
	c.ServeJSON()
}
