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

	"github.com/casdoor/casdoor/edition"
	"github.com/casdoor/casdoor/object"
)

// getFormFillTarget reads where the client fills the form: the page URL for the browser extension, the program and
// its window for the desktop client
func (c *ApiController) getFormFillTarget() *object.FormFillTarget {
	return &object.FormFillTarget{
		Url:       c.Ctx.Input.Query("url"),
		Program:   c.Ctx.Input.Query("program"),
		Title:     c.Ctx.Input.Query("title"),
		Publisher: c.Ctx.Input.Query("publisher"),
		FileHash:  c.Ctx.Input.Query("fileHash"),
		Launch:    c.Ctx.Input.Query("launch") == "true",
	}
}

// requireFormFillUser returns the signed-in user for form autofill. Only users themselves get their managed accounts,
// an application credential can't act as a user here.
func (c *ApiController) requireFormFillUser() (*object.User, bool) {
	if err := edition.Require(edition.Enterprise, edition.FeatureFormFill); err != nil {
		c.ResponseError(err.Error())
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
	if user == nil || user.IsForbidden {
		c.ResponseError(c.T("auth:Unauthorized operation"))
		return nil, false
	}
	return user, true
}

// GetFormFillSites
// @Title GetFormFillSites
// @Tag Form Autofill API
// @Description get the legacy systems whose sign-in forms are filled for the current user, without the passwords
// @Success 200 {array} object.FormFillSite The Response object
// @router /get-form-fill-sites [get]
func (c *ApiController) GetFormFillSites() {
	user, ok := c.requireFormFillUser()
	if !ok {
		return
	}
	sites, err := object.GetFormFillSites(user)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	c.ResponseOk(sites)
}

// GetFormFillAccount
// @Title GetFormFillAccount
// @Tag Form Autofill API
// @Description get the current user's managed account of an application to fill its sign-in form, every call is recorded
// @Param   application query string true  "The application name"
// @Param   username    query string false "The username of the managed account, the first one when empty"
// @Param   url         query string false "The sign-in page the form is on, for the browser extension"
// @Param   program     query string false "The full path of the desktop program, for the desktop client"
// @Param   title       query string false "The title of the program's sign-in window"
// @Param   publisher   query string false "The subject CN of the program's Authenticode certificate"
// @Param   fileHash    query string false "The SHA-256 of the program's executable"
// @Param   launch      query string false "true when the client starts the program with the account in its command line"
// @Success 200 {object} object.FormFillAccount The Response object
// @router /get-form-fill-account [get]
func (c *ApiController) GetFormFillAccount() {
	user, ok := c.requireFormFillUser()
	if !ok {
		return
	}
	application := c.Ctx.Input.Query("application")
	username := c.Ctx.Input.Query("username")
	account, err := object.GetFormFillAccount(user, application, username, c.getFormFillTarget(), c.GetAcceptLanguage())
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	c.ResponseOk(account)
}

type saveFormFillAccountForm struct {
	Application string `json:"application"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	Url         string `json:"url"`
	Program     string `json:"program"`
	Title       string `json:"title"`
	Publisher   string `json:"publisher"`
	FileHash    string `json:"fileHash"`
}

// SaveFormFillAccount
// @Title SaveFormFillAccount
// @Tag Form Autofill API
// @Description save the account that the current user signed into an application with by hand as a managed account
// @Param   body body controllers.saveFormFillAccountForm true "The application, username, password and the sign-in page"
// @Success 200 {object} controllers.Response The Response object
// @router /save-form-fill-account [post]
func (c *ApiController) SaveFormFillAccount() {
	user, ok := c.requireFormFillUser()
	if !ok {
		return
	}
	var form saveFormFillAccountForm
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &form); err != nil {
		c.ResponseError(err.Error())
		return
	}
	target := &object.FormFillTarget{Url: form.Url, Program: form.Program, Title: form.Title, Publisher: form.Publisher, FileHash: form.FileHash}
	err := object.SaveFormFillAccount(user, form.Application, form.Username, form.Password, target, c.GetAcceptLanguage())
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	c.ResponseOk()
}
