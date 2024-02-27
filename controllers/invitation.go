// Copyright 2023 The Casdoor Authors. All Rights Reserved.
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

	"github.com/beego/beego/utils/pagination"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

// GetInvitations
// @Title GetInvitations
// @Tag Invitation API
// @Description get invitations
// @Param   owner     query    string  true        "The owner of invitations"
// @Success 200 {array} object.Invitation The Response object
// @router /get-invitations [get]
func (c *ApiController) GetInvitations() {
	owner := c.Input().Get("owner")
	limit := c.Input().Get("pageSize")
	page := c.Input().Get("p")
	field := c.Input().Get("field")
	value := c.Input().Get("value")
	sortField := c.Input().Get("sortField")
	sortOrder := c.Input().Get("sortOrder")

	if limit == "" || page == "" {
		invitations, err := object.GetInvitations(owner)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		c.ResponseOk(invitations)
	} else {
		limit := util.ParseInt(limit)
		count, err := object.GetInvitationCount(owner, field, value)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		paginator := pagination.SetPaginator(c.Ctx, limit, count)
		invitations, err := object.GetPaginationInvitations(owner, paginator.Offset(), limit, field, value, sortField, sortOrder)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		c.ResponseOk(invitations, paginator.Nums())
	}
}

// GetInvitation
// @Title GetInvitation
// @Tag Invitation API
// @Description get invitation
// @Param   id     query    string  true        "The id ( owner/name ) of the invitation"
// @Success 200 {object} object.Invitation The Response object
// @router /get-invitation [get]
func (c *ApiController) GetInvitation() {
	id := c.Input().Get("id")

	invitation, err := object.GetInvitation(id)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(invitation)
}

// GetInvitationCodeInfo
// @Title GetInvitationCodeInfo
// @Tag Invitation API
// @Description get invitation code information
// @Param   code     query    string  true        "Invitation code"
// @Success 200 {object} object.Invitation The Response object
// @router /get-invitation-info [get]
func (c *ApiController) GetInvitationCodeInfo() {
	code := c.Input().Get("code")
	applicationId := c.Input().Get("applicationId")

	application, err := object.GetApplication(applicationId)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	invitation, msg := object.GetInvitationByCode(code, application.Organization, c.GetAcceptLanguage())
	if msg != "" {
		c.ResponseError(msg)
		return
	}

	c.ResponseOk(object.GetMaskedInvitation(invitation))
}

// UpdateInvitation
// @Title UpdateInvitation
// @Tag Invitation API
// @Description update invitation
// @Param   id     query    string  true        "The id ( owner/name ) of the invitation"
// @Param   body    body   object.Invitation  true        "The details of the invitation"
// @Success 200 {object} controllers.Response The Response object
// @router /update-invitation [post]
func (c *ApiController) UpdateInvitation() {
	id := c.Input().Get("id")

	var invitation object.Invitation
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &invitation)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = wrapActionResponse(object.UpdateInvitation(id, &invitation, c.GetAcceptLanguage()))
	c.ServeJSON()
}

// AddInvitation
// @Title AddInvitation
// @Tag Invitation API
// @Description add invitation
// @Param   body    body   object.Invitation  true        "The details of the invitation"
// @Success 200 {object} controllers.Response The Response object
// @router /add-invitation [post]
func (c *ApiController) AddInvitation() {
	var invitation object.Invitation
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &invitation)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = wrapActionResponse(object.AddInvitation(&invitation, c.GetAcceptLanguage()))
	c.ServeJSON()
}

// DeleteInvitation
// @Title DeleteInvitation
// @Tag Invitation API
// @Description delete invitation
// @Param   body    body   object.Invitation  true        "The details of the invitation"
// @Success 200 {object} controllers.Response The Response object
// @router /delete-invitation [post]
func (c *ApiController) DeleteInvitation() {
	var invitation object.Invitation
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &invitation)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = wrapActionResponse(object.DeleteInvitation(&invitation))
	c.ServeJSON()
}

// VerifyInvitation
// @Title VerifyInvitation
// @Tag Invitation API
// @Description verify invitation
// @Param   id     query    string  true        "The id ( owner/name ) of the invitation"
// @Success 200 {object} controllers.Response The Response object
// @router /verify-invitation [get]
func (c *ApiController) VerifyInvitation() {
	id := c.Input().Get("id")

	payment, attachInfo, err := object.VerifyInvitation(id)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(payment, attachInfo)
}
