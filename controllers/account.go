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
	"strconv"
	"strings"

	"github.com/astaxie/beego"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
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
	IdCard       string `json:"idCard"`
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

	RelayState   string `json:"relayState"`
	SamlResponse string `json:"samlResponse"`
}

type Response struct {
	Status string      `json:"status"`
	Msg    string      `json:"msg"`
	Sub    string      `json:"sub"`
	Name   string      `json:"name"`
	Data   interface{} `json:"data"`
	Data2  interface{} `json:"data2"`
}

type Userinfo struct {
	Sub         string `json:"sub"`
	Iss         string `json:"iss"`
	Aud         string `json:"aud"`
	Name        string `json:"name,omitempty"`
	DisplayName string `json:"preferred_username,omitempty"`
	Email       string `json:"email,omitempty"`
	Avatar      string `json:"picture,omitempty"`
	Address     string `json:"address,omitempty"`
	Phone       string `json:"phone,omitempty"`
}

type HumanCheck struct {
	Type         string      `json:"type"`
	AppKey       string      `json:"appKey"`
	Scene        string      `json:"scene"`
	CaptchaId    string      `json:"captchaId"`
	CaptchaImage interface{} `json:"captchaImage"`
}

// Signup
// @Tag Login API
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

	organization := object.GetOrganization(fmt.Sprintf("%s/%s", "admin", form.Organization))
	msg := object.CheckUserSignup(application, organization, form.Username, form.Password, form.Name, form.Email, form.Phone, form.Affiliation)
	if msg != "" {
		c.ResponseError(msg)
		return
	}

	if application.IsSignupItemVisible("Email") && form.Email != "" {
		checkResult := object.CheckVerificationCode(form.Email, form.EmailCode)
		if len(checkResult) != 0 {
			c.ResponseError(fmt.Sprintf("Email: %s", checkResult))
			return
		}
	}

	var checkPhone string
	if application.IsSignupItemVisible("Phone") && form.Phone != "" {
		checkPhone = fmt.Sprintf("+%s%s", form.PhonePrefix, form.Phone)
		checkResult := object.CheckVerificationCode(checkPhone, form.PhoneCode)
		if len(checkResult) != 0 {
			c.ResponseError(fmt.Sprintf("Phone: %s", checkResult))
			return
		}
	}

	userId := fmt.Sprintf("%s/%s", form.Organization, form.Username)

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
		IdCard:            form.IdCard,
		Region:            form.Region,
		Score:             getInitScore(),
		IsAdmin:           false,
		IsGlobalAdmin:     false,
		IsForbidden:       false,
		IsDeleted:         false,
		SignupApplication: application.Name,
		Properties:        map[string]string{},
	}

	affected := object.AddUser(user)
	if !affected {
		c.ResponseError(fmt.Sprintf("Failed to create user, user information is invalid: %s", util.StructToJson(user)))
		return
	}

	object.AddUserToOriginalDatabase(user)

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
// @Tag Login API
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
// @Tag Account API
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

	organization := object.GetMaskedOrganization(object.GetOrganizationByUser(user))
	resp := Response{
		Status: "ok",
		Sub:    user.Id,
		Name:   user.Name,
		Data:   user,
		Data2:  organization,
	}
	c.Data["json"] = resp
	c.ServeJSON()
}

// UserInfo
// @Title UserInfo
// @Tag Account API
// @Description return user information according to OIDC standards
// @Success 200 {object} controllers.Userinfo The Response object
// @router /userinfo [get]
func (c *ApiController) GetUserinfo() {
	userId, ok := c.RequireSignedIn()
	if !ok {
		return
	}
	user := object.GetUser(userId)
	if user == nil {
		c.ResponseError(fmt.Sprintf("The user: %s doesn't exist", userId))
		return
	}
	scope, aud := c.GetSessionOidc()
	iss := beego.AppConfig.String("origin")
	resp := Userinfo{
		Sub: user.Id,
		Iss: iss,
		Aud: aud,
	}
	if strings.Contains(scope, "profile") {
		resp.Name = user.Name
		resp.DisplayName = user.DisplayName
		resp.Avatar = user.Avatar
	}
	if strings.Contains(scope, "email") {
		resp.Email = user.Email
	}
	if strings.Contains(scope, "address") {
		resp.Address = user.Location
	}
	if strings.Contains(scope, "phone") {
		resp.Phone = user.Phone
	}
	c.Data["json"] = resp
	c.ServeJSON()
}

// GetHumanCheck ...
// @Tag Login API
// @Title GetHumancheck
// @router /api/get-human-check [get]
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
