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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"

	"github.com/astaxie/beego"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
	"golang.org/x/oauth2"
)

var githubEndpoint = oauth2.Endpoint{
	AuthURL:  "https://github.com/login/oauth/authorize",
	TokenURL: "https://github.com/login/oauth/access_token",
}

var githubOauthConfig = &oauth2.Config{
	ClientID:     beego.AppConfig.String("GithubAuthClientID"),
	ClientSecret: beego.AppConfig.String("GithubAuthClientSecret"),
	RedirectURL:  "",
	Scopes:       []string{"user:email", "read:user"},
	Endpoint:     githubEndpoint,
}

func (c *ApiController) AuthLogin() {
	applicationName := c.Input().Get("application")
	providerName := c.Input().Get("provider")
	code := c.Input().Get("code")
	state := c.Input().Get("state")
	method := c.Input().Get("method")
	RedirectURL := c.Input().Get("redirect_url")

	application := object.GetApplication(fmt.Sprintf("admin/%s", applicationName))
	provider := object.GetProvider(fmt.Sprintf("admin/%s", providerName))
	githubOauthConfig.ClientID = provider.ClientId
	githubOauthConfig.ClientSecret = provider.ClientSecret

	var resp Response
	var res authResponse
	res.IsAuthenticated = true

	if state != beego.AppConfig.String("AuthState") {
		res.IsAuthenticated = false
		resp = Response{Status: "fail", Msg: "unauthorized", Data: res}
		c.ServeJSON()
		return
	}

	githubOauthConfig.RedirectURL = RedirectURL

	// https://github.com/golang/oauth2/issues/123#issuecomment-103715338
	ctx := context.WithValue(oauth2.NoContext, oauth2.HTTPClient, httpClient)
	token, err := githubOauthConfig.Exchange(ctx, code)
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
	var tempUserEmail []userEmailFromGithub
	var tempUserAccount userInfoFromGithub
	wg.Add(2)
	go func() {
		req, err := http.NewRequest("GET", "https://api.github.com/user/emails", nil)
		if err != nil {
			panic(err)
		}
		req.Header.Add("Authorization", "token "+token.AccessToken)
		response, err := httpClient.Do(req)
		if err != nil {
			panic(err)
		}
		defer response.Body.Close()
		contents, err := ioutil.ReadAll(response.Body)

		err = json.Unmarshal(contents, &tempUserEmail)
		if err != nil {
			res.IsAuthenticated = false
			panic(err)
		}
		for _, v := range tempUserEmail {
			if v.Primary == true {
				res.Email = v.Email
				break
			}
		}
		wg.Done()
	}()
	go func() {
		req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
		if err != nil {
			panic(err)
		}
		req.Header.Add("Authorization", "token "+token.AccessToken)
		response2, err := httpClient.Do(req)
		if err != nil {
			panic(err)
		}
		defer response2.Body.Close()
		contents2, err := ioutil.ReadAll(response2.Body)
		err = json.Unmarshal(contents2, &tempUserAccount)
		if err != nil {
			res.IsAuthenticated = false
			panic(err)
		}
		wg.Done()
	}()
	wg.Wait()

	if method == "signup" {
		userId := object.HasGithub(application, tempUserAccount.Login)
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
				_ = object.LinkUserAccount(userId, "github", tempUserAccount.Login)
			} else {
				res.IsSignedUp = false
			}
		}
		res.Method = tempUserAccount.Login
		res.Avatar = tempUserAccount.AvatarUrl
		resp = Response{Status: "ok", Msg: "success", Data: res}
	} else {
		memberId := c.GetSessionUser()
		if memberId == "" {
			resp = Response{Status: "fail", Msg: "no account exist", Data: res}
			c.Data["json"] = resp
			c.ServeJSON()
			return
		}
		linkRes := object.LinkUserAccount(memberId, "github_account", tempUserAccount.Login)
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
