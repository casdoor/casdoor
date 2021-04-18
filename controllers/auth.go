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

func (c *ApiController) HandleLoggedIn(userId string, form *RequestForm) *Response {
	resp := &Response{}
	if form.Type == ResponseTypeLogin {
		c.SetSessionUser(userId)
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
	} else {
		resp = &Response{Status: "error", Msg: fmt.Sprintf("Unknown response type: %s", form.Type)}
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
		panic(err)
	}

	if form.Username != "" {
		if form.Type == ResponseTypeLogin {
			if c.GetSessionUser() != "" {
				resp = &Response{Status: "error", Msg: "Please log out first before signing in", Data: c.GetSessionUser()}
				c.Data["json"] = resp
				c.ServeJSON()
				return
			}
		}

		userId := fmt.Sprintf("%s/%s", form.Organization, form.Username)
		password := form.Password
		msg := object.CheckUserLogin(userId, password)

		if msg != "" {
			resp = &Response{Status: "error", Msg: msg, Data: ""}
		} else {
			resp = c.HandleLoggedIn(userId, &form)
		}
	} else if form.Provider != "" {
		application := object.GetApplication(fmt.Sprintf("admin/%s", form.Application))
		provider := object.GetProvider(fmt.Sprintf("admin/%s", form.Provider))

		idProvider := idp.GetIdProvider(provider.Type, provider.ClientId, provider.ClientSecret, form.RedirectUri)
		idProvider.SetHttpClient(httpClient)

		if form.State != beego.AppConfig.String("AuthState") && form.State != application.Name {
			resp = &Response{Status: "error", Msg: fmt.Sprintf("state expected: \"%s\", but got: \"%s\"", beego.AppConfig.String("AuthState"), form.State)}
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
			userId := object.GetUserIdByField(application, provider.Type, userInfo.Username)
			if userId != "" {
				//if object.IsForbidden(userId) {
				//	c.forbiddenAccountResp(userId)
				//	return
				//}

				//if len(object.GetMemberAvatar(userId)) == 0 {
				//	avatar := UploadAvatarToOSS(res.Avatar, userId)
				//	object.LinkMemberAccount(userId, "avatar", avatar)
				//}

				resp = c.HandleLoggedIn(userId, &form)
			} else {
				//if userId := object.GetUserIdByField(application, "email", userInfo.Email); userId != "" {
				//	resp = c.HandleLoggedIn(userId, &form)
				//
				//	object.LinkUserAccount(userId, provider.Type, userInfo.Username)
				//}

				if !application.EnableSignUp {
					resp = &Response{Status: "error", Msg: fmt.Sprintf("The account for provider: %s and username: %s does not exist and is not allowed to register as new account, please contact your IT support", provider.Type, userInfo.Username)}
					c.Data["json"] = resp
					c.ServeJSON()
					return
				} else {
					resp = &Response{Status: "error", Msg: fmt.Sprintf("The account for provider: %s and username: %s does not exist, please register an account first", provider.Type, userInfo.Username)}
					c.Data["json"] = resp
					c.ServeJSON()
					return
				}
			}
			//resp = &Response{Status: "ok", Msg: "", Data: res}
		} else {
			userId := c.GetSessionUser()
			if userId == "" {
				resp = &Response{Status: "error", Msg: "The account does not exist", Data: userInfo}
				c.Data["json"] = resp
				c.ServeJSON()
				return
			}

			user := object.GetUser(userId)
			isLinked := object.LinkUserAccount(user, provider.Type, userInfo.Username)
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
		panic("unknown authentication type (not password or provider)")
	}

	c.Data["json"] = resp
	c.ServeJSON()
}
