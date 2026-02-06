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
	"fmt"
	"strings"

	"github.com/beego/beego/v2/core/utils/pagination"
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
	owner := c.Ctx.Input.Query("owner")
	limit := c.Ctx.Input.Query("pageSize")
	page := c.Ctx.Input.Query("p")
	field := c.Ctx.Input.Query("field")
	value := c.Ctx.Input.Query("value")
	sortField := c.Ctx.Input.Query("sortField")
	sortOrder := c.Ctx.Input.Query("sortOrder")

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

		paginator := pagination.NewPaginator(c.Ctx.Request, limit, count)
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
	id := c.Ctx.Input.Query("id")

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
	code := c.Ctx.Input.Query("code")
	applicationId := c.Ctx.Input.Query("applicationId")

	application, err := object.GetApplication(applicationId)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	if application == nil {
		c.ResponseError(fmt.Sprintf(c.T("auth:The application: %s does not exist"), applicationId))
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
	id := c.Ctx.Input.Query("id")

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
	id := c.Ctx.Input.Query("id")

	payment, attachInfo, err := object.VerifyInvitation(id)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(payment, attachInfo)
}

// SendInvitation
// @Title VerifyInvitation
// @Tag Invitation API
// @Description verify invitation
// @Param   id     query    string	true        "The id ( owner/name ) of the invitation"
// @Param   body    body	[]string  true        "The details of the invitation"
// @Success 200 {object} controllers.Response The Response object
// @router /send-invitation [post]
func (c *ApiController) SendInvitation() {
	id := c.Ctx.Input.Query("id")

	var destinations []string
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &destinations)

	if !c.IsAdmin() {
		c.ResponseError(c.T("auth:Unauthorized operation"))
		return
	}

	invitation, err := object.GetInvitation(id)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	if invitation == nil {
		c.ResponseError(fmt.Sprintf(c.T("invitation:Invitation %s does not exist"), id))
		return
	}

	organization, err := object.GetOrganization(fmt.Sprintf("admin/%s", invitation.Owner))
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	if organization == nil {
		c.ResponseError(fmt.Sprintf(c.T("auth:The organization: %s does not exist"), invitation.Owner))
		return
	}

	var application *object.Application
	if invitation.Application != "" {
		application, err = object.GetApplication(fmt.Sprintf("admin/%s-org-%s", invitation.Application, invitation.Owner))
		if err != nil {
			c.ResponseError(err.Error())
			return
		}
	} else {
		application, err = object.GetApplicationByOrganizationName(invitation.Owner)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}
	}

	if application == nil {
		c.ResponseError(fmt.Sprintf(c.T("general:The organization: %s should have one application at least"), invitation.Owner))
		return
	}

	if application.IsShared {
		application.Name = fmt.Sprintf("%s-org-%s", application.Name, invitation.Owner)
	}

	provider, err := application.GetEmailProvider("Invitation")
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	if provider == nil {
		c.ResponseError(fmt.Sprintf(c.T("verification:please add an Email provider to the \"Providers\" list for the application: %s"), invitation.Owner))
		return
	}

	content := provider.Metadata

	content = strings.ReplaceAll(content, "%code", invitation.Code)
	content = strings.ReplaceAll(content, "%link", invitation.GetInvitationLink(c.Ctx.Request.Host, application.Name))

	err = object.SendEmail(provider, provider.Title, content, destinations, organization.DisplayName)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk()
}
