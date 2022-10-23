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
	"fmt"
	"github.com/casdoor/casdoor/i18n"
	"strconv"

	"github.com/casdoor/casdoor/conf"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

// ResponseJsonData ...
func (c *ApiController) ResponseJsonData(resp *Response, data ...interface{}) {
	switch len(data) {
	case 2:
		resp.Data2 = data[1]
		fallthrough
	case 1:
		resp.Data = data[0]
	}
	c.Data["json"] = resp
	c.ServeJSON()
}

// ResponseOk ...
func (c *ApiController) ResponseOk(data ...interface{}) {
	resp := &Response{Status: "ok"}
	c.ResponseJsonData(resp, data...)
}

// ResponseError ...
func (c *ApiController) ResponseError(error string, data ...interface{}) {
	resp := &Response{Status: "error", Msg: error}
	c.ResponseJsonData(resp, data...)
}

func (c *ApiController) T(error string) string {
	return i18n.Translate(c.GetAcceptLanguage(), error)
}

// GetAcceptLanguage ...
func (c *ApiController) GetAcceptLanguage() string {
	lang := c.Ctx.Request.Header.Get("Accept-Language")
	return lang[0:2]
}

// SetTokenErrorHttpStatus ...
func (c *ApiController) SetTokenErrorHttpStatus() {
	_, ok := c.Data["json"].(*object.TokenError)
	if ok {
		if c.Data["json"].(*object.TokenError).Error == object.InvalidClient {
			c.Ctx.Output.SetStatus(401)
			c.Ctx.Output.Header("WWW-Authenticate", "Basic realm=\"OAuth2\"")
		} else {
			c.Ctx.Output.SetStatus(400)
		}
	}
	_, ok = c.Data["json"].(*object.TokenWrapper)
	if ok {
		c.Ctx.Output.SetStatus(200)
	}
}

// RequireSignedIn ...
func (c *ApiController) RequireSignedIn() (string, bool) {
	userId := c.GetSessionUsername()
	if userId == "" {
		c.ResponseError(c.T("LoginErr.SignInFirst"))
		return "", false
	}
	return userId, true
}

// RequireSignedInUser ...
func (c *ApiController) RequireSignedInUser() (*object.User, bool) {
	userId, ok := c.RequireSignedIn()
	if !ok {
		return nil, false
	}

	user := object.GetUser(userId)
	if user == nil {
		c.ResponseError(fmt.Sprintf(c.T("UserErr.DoNotExist"), userId))
		return nil, false
	}
	return user, true
}

// RequireAdmin ...
func (c *ApiController) RequireAdmin() (string, bool) {
	user, ok := c.RequireSignedInUser()
	if !ok {
		return "", false
	}

	if user.Owner == "built-in" {
		return "", true
	}
	return user.Owner, true
}

func getInitScore() (int, error) {
	return strconv.Atoi(conf.GetConfigString("initScore"))
}

func (c *ApiController) GetProviderFromContext(category string) (*object.Provider, *object.User, bool) {
	providerName := c.Input().Get("provider")
	if providerName != "" {
		provider := object.GetProvider(util.GetId(providerName))
		if provider == nil {
			c.ResponseError(c.T("ProviderErr.ProviderNotFound"), providerName)
			return nil, nil, false
		}
		return provider, nil, true
	}

	userId, ok := c.RequireSignedIn()
	if !ok {
		return nil, nil, false
	}

	application, user := object.GetApplicationByUserId(userId)
	if application == nil {
		c.ResponseError(fmt.Sprintf(c.T("ApplicationErr.AppNotFoundForUserID"), userId))
		return nil, nil, false
	}

	provider := application.GetProviderByCategory(category)
	if provider == nil {
		c.ResponseError(fmt.Sprintf(c.T("ProviderErr.ProviderNotFoundForCategory"), category, application.Name))
		return nil, nil, false
	}

	return provider, user, true
}
