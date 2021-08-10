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
	"fmt"
	"strconv"
	"strings"

	"github.com/casbin/casdoor/object"
	"github.com/casbin/casdoor/original"
	"github.com/casbin/casdoor/util"
)

const (
	ResponseTypeLogin = "login"
	ResponseTypeCode  = "code"
)

type RequestForm struct {
	Type string `json:"type"`

	Organization string `json:"organization"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	Name         string `json:"name"`
	Email        string `json:"email"`
	Phone        string `json:"phone"`
	Affiliation  string `json:"affiliation"`
	Region       string `json:"region"`

	Application string `json:"application"`
	Provider    string `json:"provider"`
	Code        string `json:"code"`
	State       string `json:"state"`
	RedirectUri string `json:"redirectUri"`
	Method      string `json:"method"`

	EmailCode   string `json:"emailCode"`
	PhoneCode   string `json:"phoneCode"`
	PhonePrefix string `json:"phonePrefix"`

	AutoSignin bool `json:"autoSignin"`
}

type Response struct {
	Status string      `json:"status"`
	Msg    string      `json:"msg"`
	Data   interface{} `json:"data"`
	Data2  interface{} `json:"data2"`
}

type HumanCheck struct {
	Type         string      `json:"type"`
	AppKey       string      `json:"appKey"`
	Scene        string      `json:"scene"`
	CaptchaId    string      `json:"captchaId"`
	CaptchaImage interface{} `json:"captchaImage"`
}

// Signup
// @Title Signup
// @Description sign up a new user
// @Param   username     formData    string  true        "The username to sign up"
// @Param   password     formData    string  true        "The password"
// @Success 200 {object} controllers.Response The Response object
// @router /signup [post]
func (c *ApiController) Signup() {
	if c.GetSessionUsername() != "" {
		c.ResponseError("Please sign out first before signing up", c.GetSessionUsername())
		return
	}

	var form RequestForm
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &form)
	if err != nil {
		panic(err)
	}

	application := object.GetApplication(fmt.Sprintf("admin/%s", form.Application))
	if !application.EnableSignUp {
		c.ResponseError("The application does not allow to sign up new account")
		return
	}

	if application.IsSignupItemEnabled("Email") {
		checkResult := object.CheckVerificationCode(form.Email, form.EmailCode)
		if len(checkResult) != 0 {
			c.ResponseError(fmt.Sprintf("Email%s", checkResult))
			return
		}
	}

	var checkPhone string
	if application.IsSignupItemEnabled("Phone") {
		checkPhone = fmt.Sprintf("+%s%s", form.PhonePrefix, form.Phone)
		checkResult := object.CheckVerificationCode(checkPhone, form.PhoneCode)
		if len(checkResult) != 0 {
			c.ResponseError(fmt.Sprintf("Phone%s", checkResult))
			return
		}
	}

	userId := fmt.Sprintf("%s/%s", form.Organization, form.Username)

	organization := object.GetOrganization(fmt.Sprintf("%s/%s", "admin", form.Organization))
	msg := object.CheckUserSignup(application, organization, form.Username, form.Password, form.Name, form.Email, form.Phone, form.Affiliation)
	if msg != "" {
		c.ResponseError(msg)
		return
	}

	id := util.GenerateId()
	if application.GetSignupItemRule("ID") == "Incremental" {
		lastUser := object.GetLastUser(form.Organization)
		lastIdInt := util.ParseInt(lastUser.Id)
		id = strconv.Itoa(lastIdInt + 1)
	}

	username := form.Username
	if !application.IsSignupItemVisible("Username") {
		username = id
	}

	user := &object.User{
		Owner:             form.Organization,
		Name:              username,
		CreatedTime:       util.GetCurrentTime(),
		Id:                id,
		Type:              "normal-user",
		Password:          form.Password,
		DisplayName:       form.Name,
		Avatar:            organization.DefaultAvatar,
		Email:             form.Email,
		Phone:             form.Phone,
		Address:           []string{},
		Affiliation:       form.Affiliation,
		Region:            form.Region,
		IsAdmin:           false,
		IsGlobalAdmin:     false,
		IsForbidden:       false,
		SignupApplication: application.Name,
		Properties:        map[string]string{},
	}

	affected := object.AddUser(user)
	if affected {
		original.AddUserToOriginalDatabase(user)
	}

	if application.HasPromptPage() {
		// The prompt page needs the user to be signed in
		c.SetSessionUsername(user.GetId())
	}

	object.DisableVerificationCode(form.Email)
	object.DisableVerificationCode(checkPhone)

	util.LogInfo(c.Ctx, "API: [%s] is signed up as new user", userId)

	c.ResponseOk(userId)
}

// Logout
// @Title Logout
// @Description logout the current user
// @Success 200 {object} controllers.Response The Response object
// @router /logout [post]
func (c *ApiController) Logout() {
	user := c.GetSessionUsername()
	util.LogInfo(c.Ctx, "API: [%s] logged out", user)

	c.SetSessionUsername("")
	c.SetSessionData(nil)

	c.ResponseOk(user)
}

// GetAccount
// @Title GetAccount
// @Description get the details of the current account
// @Success 200 {object} controllers.Response The Response object
// @router /get-account [get]
func (c *ApiController) GetAccount() {
	userId, ok := c.RequireSignedIn()
	if !ok {
		return
	}

	user := object.GetUser(userId)
	if user == nil {
		c.ResponseError(fmt.Sprintf("The user: %s doesn't exist", userId))
		return
	}

	organization := object.GetOrganizationByUser(user)

	c.ResponseOk(user, organization)
}

// UploadFile
// @Title UploadFile
// @Description upload file
// @Param   folder      query       string  true    "The folder"
// @Param   subFolder   query       string  true    "The sub folder"
// @Param   file        formData    string  true    "The file"
// @Success 200 {object} controllers.Response The Response object
// @router /upload-file [post]
func (c *ApiController) UploadFile() {
	userId, ok := c.RequireSignedIn()
	if !ok {
		return
	}

	folder := c.Input().Get("folder")
	subFolder := c.Input().Get("subFolder")

	user := object.GetUser(userId)
	application := object.GetApplicationByUser(user)
	provider := application.GetStorageProvider()
	if provider == nil {
		c.ResponseError("No storage provider is found")
		return
	}
	file, header, err := c.GetFile("file")
	defer file.Close()
	if err != nil {
		c.ResponseError("Missing parameter")
		return
	}

	fileType := header.Header.Get("Content-Type")

	fileSuffix := ""
	switch fileType {
	case "image/png":
		fileSuffix = "png"
	case "text/html":
		fileSuffix = "html"
	}

	fileUrl, err := object.UploadFile(provider, folder, subFolder, file, fileSuffix)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	switch folder {
	case "avatar":
		user.Avatar = fileUrl
		object.UpdateUser(user.GetId(), user)
	case "termsofuse":
		appId := fmt.Sprintf("admin/%s", strings.Split(subFolder, "/")[0])
		app := object.GetApplication(appId)
		app.TermsOfUse = fileUrl
		object.UpdateApplication(appId, app)
	}

	c.ResponseOk(fileUrl)
}

// GetHumanCheck ...
func (c *ApiController) GetHumanCheck() {
	c.Data["json"] = HumanCheck{Type: "none"}

	provider := object.GetDefaultHumanCheckProvider()
	if provider == nil {
		id, img := object.GetCaptcha()
		c.Data["json"] = HumanCheck{Type: "captcha", CaptchaId: id, CaptchaImage: img}
		c.ServeJSON()
		return
	}

	c.ServeJSON()
}
