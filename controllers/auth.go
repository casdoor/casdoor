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
	"github.com/casdoor/casdoor/idp"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

func codeToResponse(code *object.Code) *Response {
	if code.Code == "" {
		return &Response{Status: "error", Msg: code.Message, Data: code.Code}
	} else {
		return &Response{Status: "ok", Msg: "", Data: code.Code}
	}
}

func (c *ApiController) HandleLoggedIn(application *object.Application, user *object.User, form *RequestForm) *Response {
	userId := user.GetId()
	resp := &Response{}
	if form.Type == ResponseTypeLogin {
		if application.EnableMfa {
			addr := object.GetUserField(user, mfaMethods[application.MfaMethod])
			if addr == "" {
				resp = &Response{Status: "error", Msg: fmt.Sprintf("MFAMethod error:%s", mfaMethods[application.MfaMethod])}
				return resp
			}
			util.LogInfo(c.Ctx, "API: [%s] need to verify email/phone/others", userId)
			resp = &Response{Status: "error", Msg: fmt.Sprint("need to verify email/phone"), Data: user.Name, Data2: addr}
			return resp
		}
		c.SetSessionUsername(userId)
		util.LogInfo(c.Ctx, "API: [%s] signed in", userId)
		resp = &Response{Status: "ok", Msg: "", Data: userId}
	} else if form.Type == ResponseTypeCode {
		clientId := c.Input().Get("clientId")
		responseType := c.Input().Get("responseType")
		redirectUri := c.Input().Get("redirectUri")
		scope := c.Input().Get("scope")
		state := c.Input().Get("state")

		code := object.GetOAuthCode(userId, clientId, responseType, redirectUri, scope, state)
		resp = codeToResponse(code)

		if application.HasPromptPage() {
			// The prompt page needs the user to be signed in
			c.SetSessionUsername(userId)
		}
	} else {
		resp = &Response{Status: "error", Msg: fmt.Sprintf("Unknown response type: %s", form.Type)}
	}

	// if user did not check auto signin
	if resp.Status == "ok" && !form.AutoSignin {
		timestamp := time.Now().Unix()
		timestamp += 3600 * 24
		c.SetSessionData(&SessionData{
			ExpireTime: timestamp,
		})
	}

	return resp
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
	var resp Response

	clientId := c.Input().Get("clientId")
	responseType := c.Input().Get("responseType")
	redirectUri := c.Input().Get("redirectUri")
	scope := c.Input().Get("scope")
	state := c.Input().Get("state")

	msg, application := object.CheckOAuthLogin(clientId, responseType, redirectUri, scope, state)
	if msg != "" {
		resp = Response{Status: "error", Msg: msg, Data: application}
	} else {
		resp = Response{Status: "ok", Msg: "", Data: application}
	}
	c.Data["json"] = resp
	c.ServeJSON()
}

// @Title Login
// @Description login
// @Param   oAuthParams     query    string  true        "oAuth parameters"
// @Param   body    body   RequestForm  true        "Login information"
// @Success 200 {object} controllers.api_controller.Response The Response object
// @router /login [post]
func (c *ApiController) Login() {
	resp := &Response{Status: "null", Msg: ""}
	var form RequestForm
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &form)
	if err != nil {
		resp = &Response{Status: "error", Msg: err.Error()}
		c.Data["json"] = resp
		c.ServeJSON()
		return
	}

	// Two-step verification required
	if form.MfaMethod != "" || (form.Password == "" && form.Username != "") {
		if form.Type == ResponseTypeLogin {
			if c.GetSessionUsername() != "" {
				resp = &Response{Status: "error", Msg: "Please log out first before signing in", Data: c.GetSessionUsername()}
				c.Data["json"] = resp
				c.ServeJSON()
				return
			}
		}
		var verificationCodeType string
		
		var user *object.User
		
		// check result through Email/Phone/others
		if strings.Contains(form.Email, "@") {
			checkResult := object.CheckVerificationCode(form.Email, form.EmailCode)
			if checkResult != "" {
				c.ResponseError(fmt.Sprintf("Email:%s", checkResult))
				return
			}
		} else if form.Phone != "" {
			checkResult := object.CheckVerificationCode(fmt.Sprintf("%s%s", form.PhonePrefix, form.Phone), form.PhoneCode)
			if checkResult != "" {
				c.ResponseError(fmt.Sprintf("Phone:%s", checkResult))
				return
			}
		} else {
			// Adapt to ForgetPage, use email instead of email/phone
			checkResult := object.CheckVerificationCode(fmt.Sprintf("%s%s", form.PhonePrefix, form.Email), form.EmailCode)
			if checkResult != "" {
				c.ResponseError(fmt.Sprintf("Phone:%s", checkResult))
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
			}
			object.DisableVerificationCode(form.Email)
			break
		case "phone":
			if user.Phone != form.Phone {
				c.ResponseError("wrong phone!")
			}
			object.DisableVerificationCode(form.Phone)
			break
		}
		
		c.SetSessionUsername(userId)
		util.LogInfo(c.Ctx, "API: [%s] signed in", userId)
		resp = &Response{Status: "ok", Msg: "", Data: userId}
	} else if form.Username != "" {
		if form.Type == ResponseTypeLogin {
			if c.GetSessionUsername() != "" {
				resp = &Response{Status: "error", Msg: "Please log out first before signing in", Data: c.GetSessionUsername()}
				c.Data["json"] = resp
				c.ServeJSON()
				return
			}
		}

		var user *object.User
		var msg string
		
		if form.Password != "" {
			password := form.Password
			user, msg = object.CheckUserLogin(form.Organization, form.Username, password)
		}

		if msg != "" || user == nil {
			resp = &Response{Status: "error", Msg: msg, Data: ""}
		} else {
			application := object.GetApplication(fmt.Sprintf("admin/%s", form.Application))

			// patch User.Phone to prefix/phoneNumber, to facilitate front-end processing
			organization := object.GetOrganization(fmt.Sprintf("%s/%s", "admin", application.Organization))
			user.Phone = fmt.Sprintf("+%s/%s", organization.PhonePrefix, user.Phone)

			resp = c.HandleLoggedIn(application, user, &form)

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
			resp = &Response{Status: "error", Msg: fmt.Sprintf("provider: %s does not exist", provider.Type)}
			c.Data["json"] = resp
			c.ServeJSON()
			return
		}

		idProvider.SetHttpClient(httpClient)

		if form.State != beego.AppConfig.String("authState") && form.State != application.Name {
			resp = &Response{Status: "error", Msg: fmt.Sprintf("state expected: \"%s\", but got: \"%s\"", beego.AppConfig.String("authState"), form.State)}
			c.Data["json"] = resp
			c.ServeJSON()
			return
		}

		// https://github.com/golang/oauth2/issues/123#issuecomment-103715338
		token, err := idProvider.GetToken(form.Code)
		if err != nil {
			resp = &Response{Status: "error", Msg: err.Error()}
			c.Data["json"] = resp
			c.ServeJSON()
			return
		}

		if !token.Valid() {
			resp = &Response{Status: "error", Msg: "Invalid token"}
			c.Data["json"] = resp
			c.ServeJSON()
			return
		}

		userInfo, err := idProvider.GetUserInfo(token)
		if err != nil {
			resp = &Response{Status: "error", Msg: fmt.Sprintf("Failed to login in: %s", err.Error())}
			c.Data["json"] = resp
			c.ServeJSON()
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
				//	avatar := UploadAvatarToOSS(res.Avatar, userId)
				//	object.LinkMemberAccount(userId, "avatar", avatar)
				//}

				resp = c.HandleLoggedIn(application, user, &form)

				record := util.Records(c.Ctx)
				record.Organization = application.Organization
				record.Username = user.Name

				object.AddRecord(record)
			} else {
				// Sign up via OAuth
				if !application.EnableSignUp {
					resp = &Response{Status: "error", Msg: fmt.Sprintf("The account for provider: %s and username: %s (%s) does not exist and is not allowed to sign up as new account, please contact your IT support", provider.Type, userInfo.Username, userInfo.DisplayName)}
					c.Data["json"] = resp
					c.ServeJSON()
					return
				}

				if !providerItem.CanSignUp {
					resp = &Response{Status: "error", Msg: fmt.Sprintf("The account for provider: %s and username: %s (%s) does not exist and is not allowed to sign up as new account via %s, please use another way to sign up", provider.Type, userInfo.Username, userInfo.DisplayName, provider.Type)}
					c.Data["json"] = resp
					c.ServeJSON()
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

				resp = c.HandleLoggedIn(application, user, &form)

				record := util.Records(c.Ctx)
				record.Organization = application.Organization
				record.Username = user.Name

				object.AddRecord(record)
			}
			//resp = &Response{Status: "ok", Msg: "", Data: res}
		} else { // form.Method != "signup"
			userId := c.GetSessionUsername()
			if userId == "" {
				resp = &Response{Status: "error", Msg: "The account does not exist", Data: userInfo}
				c.Data["json"] = resp
				c.ServeJSON()
				return
			}

			oldUser := object.GetUserByField(application.Organization, provider.Type, userInfo.Id)
			if oldUser == nil {
				oldUser = object.GetUserByField(application.Organization, provider.Type, userInfo.Username)
			}
			if oldUser != nil {
				resp = &Response{Status: "error", Msg: fmt.Sprintf("The account for provider: %s and username: %s (%s) is already linked to another account: %s (%s)", provider.Type, userInfo.Username, userInfo.DisplayName, oldUser.Name, oldUser.DisplayName)}
				c.Data["json"] = resp
				c.ServeJSON()
				return
			}

			user := object.GetUser(userId)

			// sync info from 3rd-party if possible
			object.SetUserOAuthProperties(organization, user, provider.Type, userInfo)

			isLinked := object.LinkUserAccount(user, provider.Type, userInfo.Id)
			if isLinked {
				resp = &Response{Status: "ok", Msg: "", Data: isLinked}
			} else {
				resp = &Response{Status: "error", Msg: "Failed to link user account", Data: isLinked}
			}
			//if len(object.GetMemberAvatar(userId)) == 0 {
			//	avatar := UploadAvatarToOSS(tempUserAccount.AvatarUrl, userId)
			//	object.LinkUserAccount(userId, "avatar", avatar)
			//}
		}
	} else {
		panic("unknown authentication type (not password or provider), form = " + util.StructToJson(form))
	}

	c.Data["json"] = resp
	c.ServeJSON()
}
