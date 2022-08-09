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
	"strconv"

	"github.com/casdoor/casdoor/conf"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

// ResponseOk ...
func (c *ApiController) ResponseOk(data ...interface{}) {
	resp := Response{Status: "ok"}
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

// ResponseError ...
func (c *ApiController) ResponseError(error string, data ...interface{}) {
	resp := Response{Status: "error", Msg: error}
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
		c.ResponseError("Please sign in first")
		return "", false
	}
	return userId, true
}

func getInitScore() int {
	score, err := strconv.Atoi(conf.GetConfigString("initScore"))
	if err != nil {
		panic(err)
	}

	return score
}

func (c *ApiController) GetProviderFromContext(category string) (*object.Provider, *object.User, bool) {
	providerName := c.Input().Get("provider")
	if providerName != "" {
		provider := object.GetProvider(util.GetId(providerName))
		if provider == nil {
			c.ResponseError(fmt.Sprintf("The provider: %s is not found", providerName))
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
		c.ResponseError(fmt.Sprintf("No application is found for userId: \"%s\"", userId))
		return nil, nil, false
	}

	provider := application.GetProviderByCategory(category)
	if provider == nil {
		c.ResponseError(fmt.Sprintf("No provider for category: \"%s\" is found for application: %s", category, application.Name))
		return nil, nil, false
	}

	return provider, user, true
}
