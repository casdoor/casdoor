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

	"github.com/casdoor/casdoor/conf"
	"github.com/casdoor/casdoor/i18n"
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
	language := c.Ctx.Request.Header.Get("Accept-Language")
	return conf.GetLanguage(language)
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
		c.ResponseError(c.T("general:Please login first"), "Please login first")
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

	user, err := object.GetUser(userId)
	if err != nil {
		panic(err)
	}

	if user == nil {
		c.ClearUserSession()
		c.ResponseError(fmt.Sprintf(c.T("general:The user: %s doesn't exist"), userId))
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

// IsMaskedEnabled ...
func (c *ApiController) IsMaskedEnabled() (bool, bool) {
	isMaskEnabled := true
	withSecret := c.Input().Get("withSecret")
	if withSecret == "1" {
		isMaskEnabled = false

		if conf.IsDemoMode() {
			c.ResponseError(c.T("general:this operation is not allowed in demo mode"))
			return false, isMaskEnabled
		}

		_, ok := c.RequireAdmin()
		if !ok {
			return false, isMaskEnabled
		}
	}

	return true, isMaskEnabled
}

func (c *ApiController) GetProviderFromContext(category string) (*object.Provider, *object.User, bool) {
	providerName := c.Input().Get("provider")
	if providerName != "" {
		provider, err := object.GetProvider(util.GetId("admin", providerName))
		if err != nil {
			panic(err)
		}

		if provider == nil {
			c.ResponseError(fmt.Sprintf(c.T("util:The provider: %s is not found"), providerName))
			return nil, nil, false
		}
		return provider, nil, true
	}

	userId, ok := c.RequireSignedIn()
	if !ok {
		return nil, nil, false
	}

	application, user, err := object.GetApplicationByUserId(userId)
	if err != nil {
		panic(err)
	}

	if application == nil {
		c.ResponseError(fmt.Sprintf(c.T("util:No application is found for userId: %s"), userId))
		return nil, nil, false
	}

	provider, err := application.GetProviderByCategory(category)
	if err != nil {
		panic(err)
	}

	if provider == nil {
		c.ResponseError(fmt.Sprintf(c.T("util:No provider for category: %s is found for application: %s"), category, application.Name))
		return nil, nil, false
	}

	return provider, user, true
}

func checkQuotaForApplication(count int) error {
	quota := conf.GetConfigQuota().Application
	if quota == -1 {
		return nil
	}
	if count >= quota {
		return fmt.Errorf("application quota is exceeded")
	}
	return nil
}

func checkQuotaForOrganization(count int) error {
	quota := conf.GetConfigQuota().Organization
	if quota == -1 {
		return nil
	}
	if count >= quota {
		return fmt.Errorf("organization quota is exceeded")
	}
	return nil
}

func checkQuotaForProvider(count int) error {
	quota := conf.GetConfigQuota().Provider
	if quota == -1 {
		return nil
	}
	if count >= quota {
		return fmt.Errorf("provider quota is exceeded")
	}
	return nil
}

func checkQuotaForUser(count int) error {
	quota := conf.GetConfigQuota().User
	if quota == -1 {
		return nil
	}
	if count >= quota {
		return fmt.Errorf("user quota is exceeded")
	}
	return nil
}

func getInvalidSmsReceivers(smsForm SmsForm) []string {
	var invalidReceivers []string
	for _, receiver := range smsForm.Receivers {
		// The receiver phone format: E164 like +8613854673829 +441932567890
		if !util.IsPhoneValid(receiver, "") {
			invalidReceivers = append(invalidReceivers, receiver)
		}
	}
	return invalidReceivers
}
