// Copyright 2020 The casbin Authors. All Rights Reserved.
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
	"context"
	"fmt"
	"sync"

	"github.com/astaxie/beego"
	"github.com/casdoor/casdoor/idp"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
	"golang.org/x/oauth2"
)

func (c *ApiController) AuthLogin() {
	applicationName := c.Input().Get("application")
	providerName := c.Input().Get("provider")
	code := c.Input().Get("code")
	state := c.Input().Get("state")
	method := c.Input().Get("method")
	redirectUrl := c.Input().Get("redirect_url")

	application := object.GetApplication(fmt.Sprintf("admin/%s", applicationName))
	provider := object.GetProvider(fmt.Sprintf("admin/%s", providerName))

	idProvider := idp.GetIdProvider(provider.Type)
	oauthConfig := idProvider.GetConfig()
	oauthConfig.ClientID = provider.ClientId
	oauthConfig.ClientSecret = provider.ClientSecret
	oauthConfig.RedirectURL = redirectUrl

	var resp Response
	var res authResponse
	res.IsAuthenticated = true

	if state != beego.AppConfig.String("AuthState") {
		res.IsAuthenticated = false
		resp = Response{Status: "fail", Msg: "unauthorized", Data: res}
		c.ServeJSON()
		return
	}

	// https://github.com/golang/oauth2/issues/123#issuecomment-103715338
	ctx := context.WithValue(oauth2.NoContext, oauth2.HTTPClient, httpClient)
	token, err := oauthConfig.Exchange(ctx, code)
	if err != nil {
		res.IsAuthenticated = false
		panic(err)
	}

	if !token.Valid() {
		resp = Response{Status: "fail", Msg: "unauthorized", Data: res}
		c.Data["json"] = resp
		c.ServeJSON()
		return
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		res.Email = idProvider.GetEmail(httpClient, token)
		wg.Done()
	}()
	go func() {
		res.Method, res.Avatar = idProvider.GetLoginAndAvatar(httpClient, token)
		wg.Done()
	}()
	wg.Wait()

	if method == "signup" {
		userId := object.HasGithub(application, res.Method)
		if userId != "" {
			//if len(object.GetMemberAvatar(userId)) == 0 {
			//	avatar := UploadAvatarToOSS(tempUserAccount.AvatarUrl, userId)
			//	object.LinkMemberAccount(userId, "avatar", avatar)
			//}
			c.SetSessionUser(userId)
			util.LogInfo(c.Ctx, "API: [%s] signed in", userId)
			res.IsSignedUp = true
		} else {
			if userId := object.HasMail(application, res.Email); userId != "" {
				c.SetSessionUser(userId)
				util.LogInfo(c.Ctx, "API: [%s] signed in", userId)
				res.IsSignedUp = true
				_ = object.LinkUserAccount(userId, "github", res.Method)
			} else {
				res.IsSignedUp = false
			}
		}
		resp = Response{Status: "ok", Msg: "success", Data: res}
	} else {
		memberId := c.GetSessionUser()
		if memberId == "" {
			resp = Response{Status: "fail", Msg: "no account exist", Data: res}
			c.Data["json"] = resp
			c.ServeJSON()
			return
		}
		linkRes := object.LinkUserAccount(memberId, "github_account", res.Method)
		if linkRes {
			resp = Response{Status: "ok", Msg: "success", Data: linkRes}
		} else {
			resp = Response{Status: "fail", Msg: "link account failed", Data: linkRes}
		}
		//if len(object.GetMemberAvatar(memberId)) == 0 {
		//	avatar := UploadAvatarToOSS(tempUserAccount.AvatarUrl, memberId)
		//	object.LinkUserAccount(memberId, "avatar", avatar)
		//}
	}

	c.Data["json"] = resp
	c.ServeJSON()
}
