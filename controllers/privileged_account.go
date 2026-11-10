// Copyright 2026 The Casdoor Authors. All Rights Reserved.
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

	"github.com/beego/beego/v2/core/utils/pagination"
	"github.com/casdoor/casdoor/edition"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

func (c *ApiController) requirePamLicense() bool {
	if err := edition.Require(edition.Enterprise, edition.FeaturePam); err != nil {
		c.ResponseError(err.Error())
		return false
	}
	return true
}

// requirePamAdmin checks that the current user administers the organization, an empty one stands for all of them in
// the lists of a global admin
func (c *ApiController) requirePamAdmin(owner string) bool {
	if (owner == "" && !c.IsGlobalAdmin()) || (owner != "" && !c.IsAdminOfOrganization(owner)) {
		c.ResponseError(c.T("auth:Unauthorized operation"))
		return false
	}
	return true
}

// GetPrivilegedAccounts
// @Title GetPrivilegedAccounts
// @Tag Privileged Account API
// @Description get the privileged accounts of an organization, without their passwords
// @Param   owner     query    string  true        "The organization"
// @Success 200 {array} object.PrivilegedAccount The Response object
// @router /get-privileged-accounts [get]
func (c *ApiController) GetPrivilegedAccounts() {
	owner := c.Ctx.Input.Query("owner")
	if !c.requirePamAdmin(owner) {
		return
	}
	limit := c.Ctx.Input.Query("pageSize")
	page := c.Ctx.Input.Query("p")
	field := c.Ctx.Input.Query("field")
	value := c.Ctx.Input.Query("value")
	sortField := c.Ctx.Input.Query("sortField")
	sortOrder := c.Ctx.Input.Query("sortOrder")

	if limit == "" || page == "" {
		accounts, err := object.GetPrivilegedAccounts(owner)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}
		c.ResponseOk(object.GetMaskedPrivilegedAccounts(accounts))
		return
	}

	pageSize := util.ParseInt(limit)
	count, err := object.GetPrivilegedAccountCount(owner, field, value)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	paginator := pagination.NewPaginator(c.Ctx.Request, pageSize, count)
	accounts, err := object.GetPaginationPrivilegedAccounts(owner, paginator.Offset(), pageSize, field, value, sortField, sortOrder)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	c.ResponseOk(object.GetMaskedPrivilegedAccounts(accounts), paginator.Nums())
}

// GetPrivilegedAccount
// @Title GetPrivilegedAccount
// @Tag Privileged Account API
// @Description get a privileged account, without its passwords
// @Param   id     query    string  true        "The id ( owner/name ) of the privileged account"
// @Success 200 {object} object.PrivilegedAccount The Response object
// @router /get-privileged-account [get]
func (c *ApiController) GetPrivilegedAccount() {
	id := c.Ctx.Input.Query("id")
	owner, _ := util.GetOwnerAndNameFromIdNoCheck(id)
	if !c.requirePamAdmin(owner) {
		return
	}
	account, err := object.GetPrivilegedAccount(id)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	c.ResponseOk(object.GetMaskedPrivilegedAccount(account))
}

// AddPrivilegedAccount
// @Title AddPrivilegedAccount
// @Tag Privileged Account API
// @Description add a privileged account to the vault
// @Param   body    body   object.PrivilegedAccount  true        "The privileged account"
// @Success 200 {object} controllers.Response The Response object
// @router /add-privileged-account [post]
func (c *ApiController) AddPrivilegedAccount() {
	if !c.requirePamLicense() {
		return
	}
	var account object.PrivilegedAccount
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &account); err != nil {
		c.ResponseError(err.Error())
		return
	}
	if !c.requirePamAdmin(account.Owner) {
		return
	}
	if err := object.ValidatePrivilegedAccount(&account, c.GetAcceptLanguage()); err != nil {
		c.ResponseError(err.Error())
		return
	}
	account.CreatedTime = util.GetCurrentTime()
	account.UpdatedTime = account.CreatedTime

	c.Data["json"] = wrapActionResponse(object.AddPrivilegedAccount(&account))
	c.ServeJSON()
}

// UpdatePrivilegedAccount
// @Title UpdatePrivilegedAccount
// @Tag Privileged Account API
// @Description update a privileged account, a password sent as *** is kept
// @Param   id     query    string  true        "The id ( owner/name ) of the privileged account"
// @Param   body    body   object.PrivilegedAccount  true        "The privileged account"
// @Success 200 {object} controllers.Response The Response object
// @router /update-privileged-account [post]
func (c *ApiController) UpdatePrivilegedAccount() {
	if !c.requirePamLicense() {
		return
	}
	id := c.Ctx.Input.Query("id")
	owner, _ := util.GetOwnerAndNameFromIdNoCheck(id)
	if !c.requirePamAdmin(owner) {
		return
	}
	var account object.PrivilegedAccount
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &account); err != nil {
		c.ResponseError(err.Error())
		return
	}
	if account.Owner != owner {
		c.ResponseError(c.T("auth:Unauthorized operation"))
		return
	}
	if err := object.ValidatePrivilegedAccount(&account, c.GetAcceptLanguage()); err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = wrapActionResponse(object.UpdatePrivilegedAccount(id, &account))
	c.ServeJSON()
}

// DeletePrivilegedAccount
// @Title DeletePrivilegedAccount
// @Tag Privileged Account API
// @Description delete a privileged account that nobody has checked out
// @Param   body    body   object.PrivilegedAccount  true        "The privileged account"
// @Success 200 {object} controllers.Response The Response object
// @router /delete-privileged-account [post]
func (c *ApiController) DeletePrivilegedAccount() {
	if !c.requirePamLicense() {
		return
	}
	var account object.PrivilegedAccount
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &account); err != nil {
		c.ResponseError(err.Error())
		return
	}
	if !c.requirePamAdmin(account.Owner) {
		return
	}

	c.Data["json"] = wrapActionResponse(object.DeletePrivilegedAccount(&account))
	c.ServeJSON()
}

// RotatePrivilegedAccount
// @Title RotatePrivilegedAccount
// @Tag Privileged Account API
// @Description set a new random password of a privileged account on its target system now
// @Param   id     query    string  true        "The id ( owner/name ) of the privileged account"
// @Success 200 {object} object.PrivilegedAccount The Response object
// @router /rotate-privileged-account [post]
func (c *ApiController) RotatePrivilegedAccount() {
	if !c.requirePamLicense() {
		return
	}
	id := c.Ctx.Input.Query("id")
	owner, _ := util.GetOwnerAndNameFromIdNoCheck(id)
	if !c.requirePamAdmin(owner) {
		return
	}
	account, err := object.GetPrivilegedAccount(id)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	if account == nil {
		c.ResponseError(c.T("general:Missing parameter"))
		return
	}

	if err = object.RotatePrivilegedAccount(account, "manual"); err != nil {
		c.ResponseError(err.Error())
		return
	}
	account, err = object.GetPrivilegedAccount(id)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	c.ResponseOk(object.GetMaskedPrivilegedAccount(account))
}

// GetPrivilegedRequests
// @Title GetPrivilegedRequests
// @Tag Privileged Account API
// @Description get all privileged access requests of an organization, for its admins
// @Param   owner     query    string  true        "The organization"
// @Success 200 {array} object.PrivilegedRequest The Response object
// @router /get-privileged-requests [get]
func (c *ApiController) GetPrivilegedRequests() {
	owner := c.Ctx.Input.Query("owner")
	if !c.requirePamAdmin(owner) {
		return
	}
	limit := c.Ctx.Input.Query("pageSize")
	page := c.Ctx.Input.Query("p")
	field := c.Ctx.Input.Query("field")
	value := c.Ctx.Input.Query("value")
	sortField := c.Ctx.Input.Query("sortField")
	sortOrder := c.Ctx.Input.Query("sortOrder")

	if limit == "" || page == "" {
		requests, err := object.GetPrivilegedRequests(owner)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}
		c.ResponseOk(requests)
		return
	}

	pageSize := util.ParseInt(limit)
	count, err := object.GetPrivilegedRequestCount(owner, field, value)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	paginator := pagination.NewPaginator(c.Ctx.Request, pageSize, count)
	requests, err := object.GetPaginationPrivilegedRequests(owner, paginator.Offset(), pageSize, field, value, sortField, sortOrder)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	c.ResponseOk(requests, paginator.Nums())
}

// requirePamUser returns the signed-in user for the privileged access self-service, an application credential can't
// act as a user here
func (c *ApiController) requirePamUser(needLicense bool) (*object.User, bool) {
	if needLicense && !c.requirePamLicense() {
		return nil, false
	}
	userId, ok := c.RequireSignedIn()
	if !ok {
		return nil, false
	}
	if object.IsAppUser(userId) {
		c.ResponseError(c.T("auth:Unauthorized operation"))
		return nil, false
	}
	user, err := object.GetUser(userId)
	if err != nil {
		c.ResponseError(err.Error())
		return nil, false
	}
	if user == nil || user.IsForbidden || user.IsDeleted {
		c.ResponseError(c.T("auth:Unauthorized operation"))
		return nil, false
	}
	return user, true
}

// GetPrivilegedAccess
// @Title GetPrivilegedAccess
// @Tag Privileged Access API
// @Description get the privileged accounts the current user may request, the user's requests and the requests the user approves
// @Success 200 {object} object.PamAccess The Response object
// @router /get-privileged-access [get]
func (c *ApiController) GetPrivilegedAccess() {
	user, ok := c.requirePamUser(false)
	if !ok {
		return
	}
	access, err := object.GetPamAccess(user)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	c.ResponseOk(access)
}

type privilegedAccessForm struct {
	Account         string `json:"account"`
	Reason          string `json:"reason"`
	DurationMinutes int    `json:"durationMinutes"`
}

// RequestPrivilegedAccess
// @Title RequestPrivilegedAccess
// @Tag Privileged Access API
// @Description request a privileged account or role of the current user's organization for a while
// @Param   body    body   controllers.privilegedAccessForm  true        "The account, the reason and the duration in minutes"
// @Success 200 {object} object.PrivilegedRequest The Response object
// @router /request-privileged-access [post]
func (c *ApiController) RequestPrivilegedAccess() {
	user, ok := c.requirePamUser(true)
	if !ok {
		return
	}
	var form privilegedAccessForm
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &form); err != nil {
		c.ResponseError(err.Error())
		return
	}
	request, err := object.CreatePrivilegedRequest(user, form.Account, form.Reason, form.DurationMinutes, object.GetPamOrigin(c.Ctx.Request.Host), c.GetAcceptLanguage())
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	c.ResponseOk(request)
}

type privilegedDecisionForm struct {
	Comment string `json:"comment"`
}

func (c *ApiController) decidePrivilegedRequest(approve bool) {
	user, ok := c.requirePamUser(true)
	if !ok {
		return
	}
	var form privilegedDecisionForm
	if len(c.Ctx.Input.RequestBody) > 0 {
		if err := json.Unmarshal(c.Ctx.Input.RequestBody, &form); err != nil {
			c.ResponseError(err.Error())
			return
		}
	}
	request, err := object.DecidePrivilegedRequest(user, c.Ctx.Input.Query("id"), approve, form.Comment, object.GetPamOrigin(c.Ctx.Request.Host), c.GetAcceptLanguage())
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	c.ResponseOk(request)
}

// ApprovePrivilegedRequest
// @Title ApprovePrivilegedRequest
// @Tag Privileged Access API
// @Description approve a pending request, by one of the account's approvers other than the requester
// @Param   id     query    string  true        "The id ( owner/name ) of the request"
// @Param   body    body   controllers.privilegedDecisionForm  false        "The comment"
// @Success 200 {object} object.PrivilegedRequest The Response object
// @router /approve-privileged-request [post]
func (c *ApiController) ApprovePrivilegedRequest() {
	c.decidePrivilegedRequest(true)
}

// RejectPrivilegedRequest
// @Title RejectPrivilegedRequest
// @Tag Privileged Access API
// @Description reject a pending request, by one of the account's approvers other than the requester
// @Param   id     query    string  true        "The id ( owner/name ) of the request"
// @Param   body    body   controllers.privilegedDecisionForm  false        "The comment"
// @Success 200 {object} object.PrivilegedRequest The Response object
// @router /reject-privileged-request [post]
func (c *ApiController) RejectPrivilegedRequest() {
	c.decidePrivilegedRequest(false)
}

// CancelPrivilegedRequest
// @Title CancelPrivilegedRequest
// @Tag Privileged Access API
// @Description withdraw the current user's pending request
// @Param   id     query    string  true        "The id ( owner/name ) of the request"
// @Success 200 {object} object.PrivilegedRequest The Response object
// @router /cancel-privileged-request [post]
func (c *ApiController) CancelPrivilegedRequest() {
	user, ok := c.requirePamUser(true)
	if !ok {
		return
	}
	request, err := object.CancelPrivilegedRequest(user, c.Ctx.Input.Query("id"), c.GetAcceptLanguage())
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	c.ResponseOk(request)
}

// ReturnPrivilegedRequest
// @Title ReturnPrivilegedRequest
// @Tag Privileged Access API
// @Description end the current user's access before it expires
// @Param   id     query    string  true        "The id ( owner/name ) of the request"
// @Success 200 {object} object.PrivilegedRequest The Response object
// @router /return-privileged-request [post]
func (c *ApiController) ReturnPrivilegedRequest() {
	user, ok := c.requirePamUser(true)
	if !ok {
		return
	}
	request, err := object.ReturnPrivilegedRequest(user, c.Ctx.Input.Query("id"), c.GetAcceptLanguage())
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	c.ResponseOk(request)
}

type revealedPassword struct {
	Username string                    `json:"username"`
	Password string                    `json:"password"`
	Request  *object.PrivilegedRequest `json:"request"`
}

// RevealPrivilegedPassword
// @Title RevealPrivilegedPassword
// @Tag Privileged Access API
// @Description get the password of the account of the current user's approved or active request, every call is recorded
// @Param   id     query    string  true        "The id ( owner/name ) of the request"
// @Success 200 {object} controllers.revealedPassword The Response object
// @router /reveal-privileged-password [post]
func (c *ApiController) RevealPrivilegedPassword() {
	user, ok := c.requirePamUser(true)
	if !ok {
		return
	}
	password, request, err := object.RevealPrivilegedPassword(user, c.Ctx.Input.Query("id"), c.GetAcceptLanguage())
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	account, err := object.GetPrivilegedAccount(util.GetId(request.Owner, request.Account))
	if err != nil || account == nil {
		c.ResponseError(c.T("auth:Unauthorized operation"))
		return
	}
	c.ResponseOk(revealedPassword{Username: account.Username, Password: password, Request: request})
}
