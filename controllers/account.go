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
	"net/http"
	"strconv"
	"strings"

	"github.com/casdoor/casdoor/form"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

const (
	ResponseTypeLogin   = "login"
	ResponseTypeCode    = "code"
	ResponseTypeToken   = "token"
	ResponseTypeIdToken = "id_token"
	ResponseTypeSaml    = "saml"
	ResponseTypeCas     = "cas"
)

type Response struct {
	Status string      `json:"status"`
	Msg    string      `json:"msg"`
	Sub    string      `json:"sub"`
	Name   string      `json:"name"`
	Data   interface{} `json:"data"`
	Data2  interface{} `json:"data2"`
}

type Captcha struct {
	Type          string `json:"type"`
	AppKey        string `json:"appKey"`
	Scene         string `json:"scene"`
	CaptchaId     string `json:"captchaId"`
	CaptchaImage  []byte `json:"captchaImage"`
	ClientId      string `json:"clientId"`
	ClientSecret  string `json:"clientSecret"`
	ClientId2     string `json:"clientId2"`
	ClientSecret2 string `json:"clientSecret2"`
	SubType       string `json:"subType"`
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
		c.ResponseError(c.T("account:Please sign out first"), c.GetSessionUsername())
		return
	}

	var authForm form.AuthForm
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &authForm)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	application := object.GetApplication(fmt.Sprintf("admin/%s", authForm.Application))
	if !application.EnableSignUp {
		c.ResponseError(c.T("account:The application does not allow to sign up new account"))
		return
	}

	organization := object.GetOrganization(util.GetId("admin", authForm.Organization))
	msg := object.CheckUserSignup(application, organization, &authForm, c.GetAcceptLanguage())
	if msg != "" {
		c.ResponseError(msg)
		return
	}

	if application.IsSignupItemVisible("Email") && application.GetSignupItemRule("Email") != "No verification" && authForm.Email != "" {
		checkResult := object.CheckVerificationCode(authForm.Email, authForm.EmailCode, c.GetAcceptLanguage())
		if checkResult.Code != object.VerificationSuccess {
			c.ResponseError(checkResult.Msg)
			return
		}
	}

	var checkPhone string
	if application.IsSignupItemVisible("Phone") && application.GetSignupItemRule("Phone") != "No verification" && authForm.Phone != "" {
		checkPhone, _ = util.GetE164Number(authForm.Phone, authForm.CountryCode)
		checkResult := object.CheckVerificationCode(checkPhone, authForm.PhoneCode, c.GetAcceptLanguage())
		if checkResult.Code != object.VerificationSuccess {
			c.ResponseError(checkResult.Msg)
			return
		}
	}

	id := util.GenerateId()
	if application.GetSignupItemRule("ID") == "Incremental" {
		lastUser := object.GetLastUser(authForm.Organization)

		lastIdInt := -1
		if lastUser != nil {
			lastIdInt = util.ParseInt(lastUser.Id)
		}

		id = strconv.Itoa(lastIdInt + 1)
	}

	username := authForm.Username
	if !application.IsSignupItemVisible("Username") {
		username = id
	}

	initScore, err := organization.GetInitScore()
	if err != nil {
		c.ResponseError(fmt.Errorf(c.T("account:Get init score failed, error: %w"), err).Error())
		return
	}

	user := &object.User{
		Owner:             authForm.Organization,
		Name:              username,
		CreatedTime:       util.GetCurrentTime(),
		Id:                id,
		Type:              "normal-user",
		Password:          authForm.Password,
		DisplayName:       authForm.Name,
		Avatar:            organization.DefaultAvatar,
		Email:             authForm.Email,
		Phone:             authForm.Phone,
		CountryCode:       authForm.CountryCode,
		Address:           []string{},
		Affiliation:       authForm.Affiliation,
		IdCard:            authForm.IdCard,
		Region:            authForm.Region,
		Score:             initScore,
		IsAdmin:           false,
		IsGlobalAdmin:     false,
		IsForbidden:       false,
		IsDeleted:         false,
		SignupApplication: application.Name,
		Properties:        map[string]string{},
		Karma:             0,
	}

	if len(organization.Tags) > 0 {
		tokens := strings.Split(organization.Tags[0], "|")
		if len(tokens) > 0 {
			user.Tag = tokens[0]
		}
	}

	if application.GetSignupItemRule("Display name") == "First, last" {
		if authForm.FirstName != "" || authForm.LastName != "" {
			user.DisplayName = fmt.Sprintf("%s %s", authForm.FirstName, authForm.LastName)
			user.FirstName = authForm.FirstName
			user.LastName = authForm.LastName
		}
	}

	affected := object.AddUser(user)
	if !affected {
		c.ResponseError(c.T("account:Failed to add user"), util.StructToJson(user))
		return
	}

	object.AddUserToOriginalDatabase(user)

	if application.HasPromptPage() {
		// The prompt page needs the user to be signed in
		c.SetSessionUsername(user.GetId())
	}

	object.DisableVerificationCode(authForm.Email)
	object.DisableVerificationCode(checkPhone)

	isSignupFromPricing := authForm.Plan != "" && authForm.Pricing != ""
	if isSignupFromPricing {
		object.Subscribe(organization.Name, user.Name, authForm.Plan, authForm.Pricing)
	}

	record := object.NewRecord(c.Ctx)
	record.Organization = application.Organization
	record.User = user.Name
	util.SafeGoroutine(func() { object.AddRecord(record) })

	userId := user.GetId()
	util.LogInfo(c.Ctx, "API: [%s] is signed up as new user", userId)

	c.ResponseOk(userId)
}

// Logout
// @Title Logout
// @Tag Login API
// @Description logout the current user
// @Param   id_token_hint   query        string  false        "id_token_hint"
// @Param   post_logout_redirect_uri    query    string  false     "post_logout_redirect_uri"
// @Param   state     query    string  false     "state"
// @Success 200 {object} controllers.Response The Response object
// @router /logout [get,post]
func (c *ApiController) Logout() {
	// https://openid.net/specs/openid-connect-rpinitiated-1_0-final.html
	accessToken := c.Input().Get("id_token_hint")
	redirectUri := c.Input().Get("post_logout_redirect_uri")
	state := c.Input().Get("state")

	user := c.GetSessionUsername()

	if accessToken == "" && redirectUri == "" {
		// TODO https://github.com/casdoor/casdoor/pull/1494#discussion_r1095675265
		if user == "" {
			c.ResponseOk()
			return
		}

		c.ClearUserSession()
		owner, username := util.GetOwnerAndNameFromId(user)
		object.DeleteSessionId(util.GetSessionId(owner, username, object.CasdoorApplication), c.Ctx.Input.CruSession.SessionID())

		util.LogInfo(c.Ctx, "API: [%s] logged out", user)

		application := c.GetSessionApplication()
		if application == nil || application.Name == "app-built-in" || application.HomepageUrl == "" {
			c.ResponseOk(user)
			return
		}
		c.ResponseOk(user, application.HomepageUrl)
		return
	} else {
		if redirectUri == "" {
			c.ResponseError(c.T("general:Missing parameter") + ": post_logout_redirect_uri")
			return
		}
		if accessToken == "" {
			c.ResponseError(c.T("general:Missing parameter") + ": id_token_hint")
			return
		}

		affected, application, token := object.ExpireTokenByAccessToken(accessToken)
		if !affected {
			c.ResponseError(c.T("token:Token not found, invalid accessToken"))
			return
		}

		if application == nil {
			c.ResponseError(fmt.Sprintf(c.T("auth:The application: %s does not exist")), token.Application)
			return
		}

		if application.IsRedirectUriValid(redirectUri) {
			if user == "" {
				user = util.GetId(token.Organization, token.User)
			}

			c.ClearUserSession()
			// TODO https://github.com/casdoor/casdoor/pull/1494#discussion_r1095675265
			owner, username := util.GetOwnerAndNameFromId(user)

			object.DeleteSessionId(util.GetSessionId(owner, username, object.CasdoorApplication), c.Ctx.Input.CruSession.SessionID())
			util.LogInfo(c.Ctx, "API: [%s] logged out", user)

			c.Ctx.Redirect(http.StatusFound, fmt.Sprintf("%s?state=%s", strings.TrimRight(redirectUri, "/"), state))
		} else {
			c.ResponseError(fmt.Sprintf(c.T("token:Redirect URI: %s doesn't exist in the allowed Redirect URI list"), redirectUri))
			return
		}
	}
}

// GetAccount
// @Title GetAccount
// @Tag Account API
// @Description get the details of the current account
// @Success 200 {object} controllers.Response The Response object
// @router /get-account [get]
func (c *ApiController) GetAccount() {
	user, ok := c.RequireSignedInUser()
	if !ok {
		return
	}

	managedAccounts := c.Input().Get("managedAccounts")
	if managedAccounts == "1" {
		user = object.ExtendManagedAccountsWithUser(user)
	}

	object.ExtendUserWithRolesAndPermissions(user)

	user.Permissions = object.GetMaskedPermissions(user.Permissions)
	user.Roles = object.GetMaskedRoles(user.Roles)

	organization := object.GetMaskedOrganization(object.GetOrganizationByUser(user))
	resp := Response{
		Status: "ok",
		Sub:    user.Id,
		Name:   user.Name,
		Data:   object.GetMaskedUser(user),
		Data2:  organization,
	}
	c.Data["json"] = resp
	c.ServeJSON()
}

// GetUserinfo
// UserInfo
// @Title UserInfo
// @Tag Account API
// @Description return user information according to OIDC standards
// @Success 200 {object} object.Userinfo The Response object
// @router /userinfo [get]
func (c *ApiController) GetUserinfo() {
	user, ok := c.RequireSignedInUser()
	if !ok {
		return
	}

	scope, aud := c.GetSessionOidc()
	host := c.Ctx.Request.Host
	userInfo := object.GetUserInfo(user, scope, aud, host)

	c.Data["json"] = userInfo
	c.ServeJSON()
}

// GetUserinfo2
// LaravelResponse
// @Title UserInfo2
// @Tag Account API
// @Description return Laravel compatible user information according to OAuth 2.0
// @Success 200 {object} LaravelResponse The Response object
// @router /user [get]
func (c *ApiController) GetUserinfo2() {
	user, ok := c.RequireSignedInUser()
	if !ok {
		return
	}

	// this API is used by "Api URL" of Flarum's FoF Passport plugin
	// https://github.com/FriendsOfFlarum/passport
	type LaravelResponse struct {
		Id              string `json:"id"`
		Name            string `json:"name"`
		Email           string `json:"email"`
		EmailVerifiedAt string `json:"email_verified_at"`
		CreatedAt       string `json:"created_at"`
		UpdatedAt       string `json:"updated_at"`
	}

	response := LaravelResponse{
		Id:              user.Id,
		Name:            user.Name,
		Email:           user.Email,
		EmailVerifiedAt: user.CreatedTime,
		CreatedAt:       user.CreatedTime,
		UpdatedAt:       user.UpdatedTime,
	}

	c.Data["json"] = response
	c.ServeJSON()
}

// GetCaptcha ...
// @Tag Login API
// @Title GetCaptcha
// @router /api/get-captcha [get]
func (c *ApiController) GetCaptcha() {
	applicationId := c.Input().Get("applicationId")
	isCurrentProvider := c.Input().Get("isCurrentProvider")

	captchaProvider, err := object.GetCaptchaProviderByApplication(applicationId, isCurrentProvider, c.GetAcceptLanguage())
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if captchaProvider != nil {
		if captchaProvider.Type == "Default" {
			id, img := object.GetCaptcha()
			c.ResponseOk(Captcha{Type: captchaProvider.Type, CaptchaId: id, CaptchaImage: img})
			return
		} else if captchaProvider.Type != "" {
			c.ResponseOk(Captcha{
				Type:          captchaProvider.Type,
				SubType:       captchaProvider.SubType,
				ClientId:      captchaProvider.ClientId,
				ClientSecret:  captchaProvider.ClientSecret,
				ClientId2:     captchaProvider.ClientId2,
				ClientSecret2: captchaProvider.ClientSecret2,
			})
			return
		}
	}

	c.ResponseOk(Captcha{Type: "none"})
}
