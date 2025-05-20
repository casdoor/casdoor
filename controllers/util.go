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
	"strings"

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
	enableErrorMask2 := conf.GetConfigBool("enableErrorMask2")
	if enableErrorMask2 {
		error = c.T("subscription:Error")

		resp := &Response{Status: "error", Msg: error}
		c.ResponseJsonData(resp, data...)
		return
	}

	enableErrorMask := conf.GetConfigBool("enableErrorMask")
	if enableErrorMask {
		if strings.HasPrefix(error, "The user: ") && strings.HasSuffix(error, " doesn't exist") || strings.HasPrefix(error, "用户: ") && strings.HasSuffix(error, "不存在") {
			error = c.T("check:password or code is incorrect")
		}
	}

	resp := &Response{Status: "error", Msg: error}
	c.ResponseJsonData(resp, data...)
}

func (c *ApiController) T(error string) string {
	return i18n.Translate(c.GetAcceptLanguage(), error)
}

// GetAcceptLanguage ...
func (c *ApiController) GetAcceptLanguage() string {
	language := c.Ctx.Request.Header.Get("Accept-Language")
	if len(language) > 2 {
		language = language[0:2]
	}
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

	if object.IsAppUser(userId) {
		tmpUserId := c.Input().Get("userId")
		if tmpUserId != "" {
			userId = tmpUserId
		}
	}

	user, err := object.GetUser(userId)
	if err != nil {
		c.ResponseError(err.Error())
		return nil, false
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

	if !user.IsAdmin {
		c.ResponseError(c.T("general:this operation requires administrator to perform"))
		return "", false
	}

	return user.Owner, true
}

func (c *ApiController) IsOrgAdmin() (bool, bool) {
	userId, ok := c.RequireSignedIn()
	if !ok {
		return false, true
	}

	if object.IsAppUser(userId) {
		return true, true
	}

	user, err := object.GetUser(userId)
	if err != nil {
		c.ResponseError(err.Error())
		return false, false
	}
	if user == nil {
		c.ClearUserSession()
		c.ResponseError(fmt.Sprintf(c.T("general:The user: %s doesn't exist"), userId))
		return false, false
	}

	return user.IsAdmin, true
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

func refineFullFilePath(fullFilePath string) (string, string) {
	tokens := strings.Split(fullFilePath, "/")
	if len(tokens) >= 2 && tokens[0] == "Direct" && tokens[1] != "" {
		providerName := tokens[1]
		res := strings.Join(tokens[2:], "/")
		return providerName, "/" + res
	} else {
		return "", fullFilePath
	}
}

func (c *ApiController) GetProviderFromContext(category string) (*object.Provider, error) {
	providerName := c.Input().Get("provider")
	if providerName == "" {
		field := c.Input().Get("field")
		value := c.Input().Get("value")
		if field == "provider" && value != "" {
			providerName = value
		} else {
			fullFilePath := c.Input().Get("fullFilePath")
			providerName, _ = refineFullFilePath(fullFilePath)
		}
	}

	if providerName != "" {
		provider, err := object.GetProvider(util.GetId("admin", providerName))
		if err != nil {
			return nil, err
		}

		if provider == nil {
			err = fmt.Errorf(c.T("util:The provider: %s is not found"), providerName)
			return nil, err
		}

		return provider, nil
	}

	userId, ok := c.RequireSignedIn()
	if !ok {
		return nil, fmt.Errorf(c.T("general:Please login first"))
	}

	application, err := object.GetApplicationByUserId(userId)
	if err != nil {
		return nil, err
	}

	if application == nil {
		return nil, fmt.Errorf(c.T("util:No application is found for userId: %s"), userId)
	}

	provider, err := application.GetProviderByCategory(category)
	if err != nil {
		return nil, err
	}

	if provider == nil {
		return nil, fmt.Errorf(c.T("util:No provider for category: %s is found for application: %s"), category, application.Name)
	}

	return provider, nil
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

func checkQuotaForUser() error {
	quota := conf.GetConfigQuota().User
	if quota == -1 {
		return nil
	}

	count, err := object.GetUserCount("", "", "", "")
	if err != nil {
		return err
	}

	if int(count) >= quota {
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
