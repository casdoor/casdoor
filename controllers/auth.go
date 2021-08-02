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
	"time"

	"github.com/astaxie/beego"
	"github.com/casbin/casdoor/idp"
	"github.com/casbin/casdoor/object"
	"github.com/casbin/casdoor/util"
)

func (c *ApiController) HandleLoggedIn(application *object.Application, user *object.User, form *RequestForm) {
	userId := user.GetId()
	if form.Type == ResponseTypeLogin {
		c.SetSessionUsername(userId)
		util.LogInfo(c.Ctx, "API: [%s] signed in", userId)
		c.Data["json"] = Response{Status: "ok", Msg: "", Data: userId}
		c.ServeJSON()
	} else if form.Type == ResponseTypeCode {
		clientId := c.Input().Get("clientId")
		responseType := c.Input().Get("responseType")
		redirectUri := c.Input().Get("redirectUri")
		scope := c.Input().Get("scope")
		state := c.Input().Get("state")

		code := object.GetOAuthCode(userId, clientId, responseType, redirectUri, scope, state)
		if code.Code == "" {
			c.ResponseError(code.Message, code.Code)
		} else {
			c.Data["json"] = Response{Status: "ok", Msg: "", Data: code.Code}
			c.ServeJSON()
		}

		if application.HasPromptPage() {
			// The prompt page needs the user to be signed in
			c.SetSessionUsername(userId)
		}
	} else {
		c.ResponseError(fmt.Sprintf("Unknown response type: %s", form.Type))
		return
	}

	// if user did not check auto signin
	if !form.AutoSignin {
		timestamp := time.Now().Unix()
		timestamp += 3600 * 24
		c.SetSessionData(&SessionData{
			ExpireTime: timestamp,
		})
	}
}

// @Title GetApplicationLogin
// @Description get application login
// @Param   clientId    query    string  true        "client id"
// @Param   responseType    query    string  true        "response type"
// @Param   redirectUri    query    string  true        "redirect uri"
// @Param   scope    query    string  true        "scope"
// @Param   state    query    string  true        "state"
// @Success 200 {object} controllers.api_controller.Response The Response object
// @router /update-application [get]
func (c *ApiController) GetApplicationLogin() {
	clientId := c.Input().Get("clientId")
	responseType := c.Input().Get("responseType")
	redirectUri := c.Input().Get("redirectUri")
	scope := c.Input().Get("scope")
	state := c.Input().Get("state")

	msg, application := object.CheckOAuthLogin(clientId, responseType, redirectUri, scope, state)
	if msg != "" {
		c.ResponseError(msg, application)
		return
	}

	c.Data["json"] = Response{Status: "ok", Msg: "", Data: application}
	c.ServeJSON()
}

func setHttpClient(idProvider idp.IdProvider, providerType string) {
	if providerType == "GitHub" || providerType == "Google" || providerType == "Facebook" || providerType == "LinkedIn" {
		idProvider.SetHttpClient(proxyHttpClient)
	} else {
		idProvider.SetHttpClient(defaultHttpClient)
	}
}

// @Title Login
// @Description login
// @Param   oAuthParams     query    string  true        "oAuth parameters"
// @Param   body    body   RequestForm  true        "Login information"
// @Success 200 {object} controllers.api_controller.Response The Response object
// @router /login [post]
func (c *ApiController) Login() {
	var form RequestForm
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &form)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if form.Username != "" {
		if form.Type == ResponseTypeLogin {
			if c.GetSessionUsername() != "" {
				c.ResponseError("Please log out first before signing in", c.GetSessionUsername())
				return
			}
		}

		var user *object.User
		var msg string

		if form.Password == "" {
			var verificationCodeType string

			// check result through Email or Phone
			if strings.Contains(form.Email, "@") {
				verificationCodeType = "email"
				checkResult := object.CheckVerificationCode(form.Email, form.EmailCode)
				if len(checkResult) != 0 {
					responseText := fmt.Sprintf("Email%s", checkResult)
					c.ResponseError(responseText)
					return
				}
			} else {
				verificationCodeType = "phone"
				checkPhone := fmt.Sprintf("+%s%s", form.PhonePrefix, form.Email)
				checkResult := object.CheckVerificationCode(checkPhone, form.EmailCode)
				if len(checkResult) != 0 {
					responseText := fmt.Sprintf("Phone%s", checkResult)
					c.ResponseError(responseText)
					return
				}
			}

			// get user
			var userId string
			if form.Username == "" {
				userId, _ = c.RequireSignedIn()
			} else {
				userId = fmt.Sprintf("%s/%s", form.Organization, form.Username)
			}

			user = object.GetUser(userId)
			if user == nil {
				c.ResponseError("No such user.")
				return
			}

			// disable the verification code
			switch verificationCodeType {
			case "email":
				if user.Email != form.Email {
					c.ResponseError("wrong email!")
					return
				}
				object.DisableVerificationCode(form.Email)
				break
			case "phone":
				if user.Phone != form.Email {
					c.ResponseError("wrong phone!")
					return
				}
				object.DisableVerificationCode(form.Email)
				break
			}
		} else {
			password := form.Password
			user, msg = object.CheckUserLogin(form.Organization, form.Username, password)
		}

		if msg != "" {
			c.ResponseError(msg)
			return
		} else {
			application := object.GetApplication(fmt.Sprintf("admin/%s", form.Application))
			c.HandleLoggedIn(application, user, &form)

			record := util.Records(c.Ctx)
			record.Organization = application.Organization
			record.Username = user.Name

			object.AddRecord(record)
		}
	} else if form.Provider != "" {
		application := object.GetApplication(fmt.Sprintf("admin/%s", form.Application))
		organization := object.GetOrganization(fmt.Sprintf("%s/%s", "admin", application.Organization))
		provider := object.GetProvider(fmt.Sprintf("admin/%s", form.Provider))
		providerItem := application.GetProviderItem(provider.Name)

		idProvider := idp.GetIdProvider(provider.Type, provider.ClientId, provider.ClientSecret, form.RedirectUri)
		if idProvider == nil {
			c.ResponseError(fmt.Sprintf("provider: %s does not exist", provider.Type))
			return
		}

		setHttpClient(idProvider, provider.Type)

		if form.State != beego.AppConfig.String("authState") && form.State != application.Name {
			c.ResponseError(fmt.Sprintf("state expected: \"%s\", but got: \"%s\"", beego.AppConfig.String("authState"), form.State))
			return
		}

		// https://github.com/golang/oauth2/issues/123#issuecomment-103715338
		token, err := idProvider.GetToken(form.Code)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		if !token.Valid() {
			c.ResponseError("Invalid token")
			return
		}

		userInfo, err := idProvider.GetUserInfo(token)
		if err != nil {
			c.ResponseError(fmt.Sprintf("Failed to login in: %s", err.Error()))
			return
		}

		if form.Method == "signup" {
			user := object.GetUserByField(application.Organization, provider.Type, userInfo.Id)
			if user == nil {
				user = object.GetUserByField(application.Organization, provider.Type, userInfo.Username)
			}
			if user == nil {
				user = object.GetUserByField(application.Organization, "name", userInfo.Username)
			}

			if user != nil {
				// Sign in via OAuth

				//if object.IsForbidden(userId) {
				//	c.forbiddenAccountResp(userId)
				//	return
				//}

				//if len(object.GetMemberAvatar(userId)) == 0 {
				//	avatar := UploadAvatarToStorage(res.Avatar, userId)
				//	object.LinkMemberAccount(userId, "avatar", avatar)
				//}

				c.HandleLoggedIn(application, user, &form)

				record := util.Records(c.Ctx)
				record.Organization = application.Organization
				record.Username = user.Name

				 object.AddRecord(record)
			} else {
				// Sign up via OAuth
				if !application.EnableSignUp {
					c.ResponseError(fmt.Sprintf("The account for provider: %s and username: %s (%s) does not exist and is not allowed to sign up as new account, please contact your IT support", provider.Type, userInfo.Username, userInfo.DisplayName))
					return
				}

				if !providerItem.CanSignUp {
					c.ResponseError(fmt.Sprintf("The account for provider: %s and username: %s (%s) does not exist and is not allowed to sign up as new account via %s, please use another way to sign up", provider.Type, userInfo.Username, userInfo.DisplayName, provider.Type))
					return
				}

				properties := map[string]string{}
				properties["no"] = strconv.Itoa(len(object.GetUsers(application.Organization)) + 2)
				user := &object.User{
					Owner:             application.Organization,
					Name:              userInfo.Username,
					CreatedTime:       util.GetCurrentTime(),
					Id:                util.GenerateId(),
					Type:              "normal-user",
					DisplayName:       userInfo.DisplayName,
					Avatar:            userInfo.AvatarUrl,
					Email:             userInfo.Email,
					Score:             200,
					IsAdmin:           false,
					IsGlobalAdmin:     false,
					IsForbidden:       false,
					SignupApplication: application.Name,
					Properties:        properties,
				}
				object.AddUser(user)

				// sync info from 3rd-party if possible
				object.SetUserOAuthProperties(organization, user, provider.Type, userInfo)

				object.LinkUserAccount(user, provider.Type, userInfo.Id)

				c.HandleLoggedIn(application, user, &form)

				record := util.Records(c.Ctx)
				record.Organization = application.Organization
				record.Username = user.Name

				object.AddRecord(record)
			}
			//resp = &Response{Status: "ok", Msg: "", Data: res}
		} else { // form.Method != "signup"
			userId := c.GetSessionUsername()
			if userId == "" {
				c.ResponseError("The account does not exist", userInfo)
				return
			}

			oldUser := object.GetUserByField(application.Organization, provider.Type, userInfo.Id)
			if oldUser == nil {
				oldUser = object.GetUserByField(application.Organization, provider.Type, userInfo.Username)
			}
			if oldUser != nil {
				c.ResponseError(fmt.Sprintf("The account for provider: %s and username: %s (%s) is already linked to another account: %s (%s)", provider.Type, userInfo.Username, userInfo.DisplayName, oldUser.Name, oldUser.DisplayName))
				return
			}

			user := object.GetUser(userId)

			// sync info from 3rd-party if possible
			object.SetUserOAuthProperties(organization, user, provider.Type, userInfo)

			isLinked := object.LinkUserAccount(user, provider.Type, userInfo.Id)
			if isLinked {
				c.Data["json"] = Response{Status: "ok", Msg: "", Data: isLinked}
				c.ServeJSON()
				return
			} else {
				c.ResponseError("Failed to link user account", isLinked)
				return
			}
			//if len(object.GetMemberAvatar(userId)) == 0 {
			//	avatar := UploadAvatarToStorage(tempUserAccount.AvatarUrl, userId)
			//	object.LinkUserAccount(userId, "avatar", avatar)
			//}
		}
	} else {
		panic("unknown authentication type (not password or provider), form = " + util.StructToJson(form))
	}

	c.Data["json"] = Response{Status: "null"}
	c.ServeJSON()
}
