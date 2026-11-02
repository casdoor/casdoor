// Copyright 2026 The Casdoor Authors. All Rights Reserved.
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
	"strings"
	"time"

	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

const pendingUserMergeSession = "pendingUserMerge"

// pendingUserMerge is the account that a provider being linked already belongs to, the
// provider sign-in proved it, so the merge only has to be confirmed within a few minutes
type pendingUserMerge struct {
	UserId   string `json:"userId"`
	Provider string `json:"provider"`
	Expire   int64  `json:"expire"`
}

func (c *ApiController) setPendingUserMerge(user *object.User, provider string) {
	pending := pendingUserMerge{UserId: user.GetId(), Provider: provider, Expire: time.Now().Add(10 * time.Minute).Unix()}
	c.SetSession(pendingUserMergeSession, util.StructToJson(pending))
}

func (c *ApiController) popPendingUserMerge() *pendingUserMerge {
	text, _ := c.GetSession(pendingUserMergeSession).(string)
	c.DelSession(pendingUserMergeSession)
	if text == "" {
		return nil
	}

	pending := &pendingUserMerge{}
	if err := json.Unmarshal([]byte(text), pending); err != nil || time.Now().Unix() > pending.Expire {
		return nil
	}
	return pending
}

// getUserByDest returns the account other than the signed-in one that owns the email or phone
func getUserByDest(user *object.User, destType string, dest string, countryCode string) (*object.User, string, error) {
	switch destType {
	case object.VerifyTypeEmail:
		other, err := object.GetUserByField(user.Owner, "email", strings.ToLower(dest))
		return other, dest, err
	case object.VerifyTypePhone:
		phone, normalizedCountryCode, ok := util.GetNormalizedPhone(dest, user.GetCountryCode(countryCode))
		if !ok {
			return nil, "", nil
		}
		other, err := object.GetUserByPhoneAndCountryCode(user.Owner, phone, normalizedCountryCode)
		if err != nil {
			return nil, "", err
		}
		checkDest, _ := util.GetE164Number(phone, normalizedCountryCode)
		return other, checkDest, nil
	}
	return nil, "", fmt.Errorf("unknown type: %s", destType)
}

// isUserDestMergeable tells whether the code for a new email or phone may be sent: the address
// is free, or the account that owns it can be merged with the signed-in one either way
func isUserDestMergeable(user *object.User, destType string, dest string, countryCode string) bool {
	if user == nil {
		return true
	}

	other, _, err := getUserByDest(user, destType, dest, countryCode)
	if err != nil || other == nil || other.Name == user.Name {
		return err == nil
	}
	return object.GetUserMergeError(user, other, "en") == "" || object.GetUserMergeError(other, user, "en") == ""
}

// getMergePair orders the two accounts as "merged from" and "kept", keep is "current" or "other"
func getMergePair(user *object.User, other *object.User, keep string) (*object.User, *object.User) {
	if keep == "other" {
		return user, other
	}
	return other, user
}

// CheckUserMerge
// @Tag User API
// @Title CheckUserMerge
// @Description tell whether a new email or phone of the signed-in user belongs to another account that can be merged
// @Param   type        query string true  "email or phone"
// @Param   dest        query string true  "the new email or phone"
// @Param   countryCode query string false "the country code of the phone"
// @Param   keep        query string false "the account kept by the merge: current or other"
// @Success 200 {object} controllers.Response The Response object, data is "merge" when the two accounts can be merged
// @router /check-user-merge [get]
func (c *ApiController) CheckUserMerge() {
	user, ok := c.RequireSignedInUser()
	if !ok {
		return
	}

	destType := c.Ctx.Input.Query("type")
	other, _, err := getUserByDest(user, destType, c.Ctx.Input.Query("dest"), c.Ctx.Input.Query("countryCode"))
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	if other == nil || other.Name == user.Name {
		c.ResponseOk("")
		return
	}

	from, into := getMergePair(user, other, c.Ctx.Input.Query("keep"))
	if msg := object.GetUserMergeError(from, into, c.GetAcceptLanguage()); msg != "" {
		if destType == object.VerifyTypeEmail {
			c.ResponseError(c.T("check:Email already exists"))
		} else {
			c.ResponseError(c.T("check:Phone already exists"))
		}
		return
	}

	c.ResponseOk("merge")
}

// MergeUser
// @Tag User API
// @Title MergeUser
// @Description merge the signed-in user with the account that owns a verified email, phone or third-party account
// @Param   type        formData string true  "email, phone or provider"
// @Param   dest        formData string false "the email or phone"
// @Param   countryCode formData string false "the country code of the phone"
// @Param   code        formData string false "the verification code sent to the email or phone"
// @Param   keep        formData string false "the account kept by the merge: current or other"
// @Param   oauthCode   formData string false "the authorization code the signed-in user got, it is reissued to the kept account"
// @Param   redirectUri formData string false "the redirect URI of the authorization code"
// @Param   state       formData string false "the state of the authorization code"
// @Success 200 {object} controllers.Response The Response object, data is the reissued authorization code
// @router /merge-user [post]
func (c *ApiController) MergeUser() {
	user, ok := c.RequireSignedInUser()
	if !ok {
		return
	}

	destType := c.Ctx.Request.Form.Get("type")
	keep := c.Ctx.Request.Form.Get("keep")

	var other *object.User
	var err error
	if destType == "provider" {
		// the provider sign-in of the link flow already proved the other account, it stays the current one
		pending := c.popPendingUserMerge()
		if pending == nil {
			c.ResponseError(c.T("auth:Unauthorized operation"))
			return
		}
		other, err = object.GetUser(pending.UserId)
		keep = "current"
	} else {
		code := c.Ctx.Request.Form.Get("code")
		if code == "" {
			c.ResponseError(c.T("general:Missing parameter"))
			return
		}

		var checkDest string
		other, checkDest, err = getUserByDest(user, destType, c.Ctx.Request.Form.Get("dest"), c.Ctx.Request.Form.Get("countryCode"))
		if err == nil && other != nil {
			err = object.CheckVerifyCodeWithLimitAndIp(user, util.GetClientIpFromRequest(c.Ctx.Request), checkDest, code, c.GetAcceptLanguage())
			if err == nil {
				err = object.DisableVerificationCode(checkDest)
			}
		}
	}
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	if other == nil || other.Name == user.Name {
		c.ResponseError(c.T("auth:Unauthorized operation"))
		return
	}

	from, into := getMergePair(user, other, keep)
	if msg := object.GetUserMergeError(from, into, c.GetAcceptLanguage()); msg != "" {
		c.ResponseError(msg)
		return
	}

	// deleting the merged account expires its tokens, so the pending authorization code is read first
	var codeRequest *object.OAuthCodeRequest
	if oauthCode := c.Ctx.Request.Form.Get("oauthCode"); oauthCode != "" && from.Name == user.Name {
		codeRequest, err = object.TakeOAuthCodeRequest(oauthCode, from)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}
		if codeRequest == nil {
			c.ResponseError(c.T("auth:Unauthorized operation"))
			return
		}
	}

	err = object.MergeUser(from, into)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	util.LogInfo(c.Ctx, "API: [%s] merged into [%s]", from.GetId(), into.GetId())

	if into.Name == user.Name {
		c.ResponseOk("", into.GetId())
		return
	}

	c.renewSessionIdForUser(into.GetId())
	c.SetSessionUsername(into.GetId())
	sessionId := c.Ctx.Input.CruSession.SessionID(context.Background())

	applicationName := "app-built-in"
	if codeRequest != nil {
		applicationName = codeRequest.Application.Name
	}
	_, err = object.AddSession(&object.Session{
		Owner:       into.Owner,
		Name:        into.Name,
		Application: applicationName,
		SessionId:   []string{sessionId},
		SessionInfos: []*object.SessionInfo{{
			SessionId:      sessionId,
			CreatedTime:    util.GetCurrentTime(),
			LastActiveTime: util.GetCurrentTime(),
			Ip:             util.GetClientIpFromRequest(c.Ctx.Request),
			UserAgent:      c.Ctx.Request.UserAgent(),
		}},
	})
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	newCode := ""
	if codeRequest != nil {
		code, err := object.GetOAuthCode(into.GetId(), codeRequest.Application.ClientId, codeRequest.Provider, codeRequest.SigninMethod, "code", c.Ctx.Request.Form.Get("redirectUri"), codeRequest.Scope, c.Ctx.Request.Form.Get("state"), codeRequest.Nonce, codeRequest.Challenge, codeRequest.Resource, sessionId, time.Now().Unix(), c.Ctx.Request.Host, c.GetAcceptLanguage())
		if err != nil {
			c.ResponseError(err.Error())
			return
		}
		if code.Message != "" {
			c.ResponseError(code.Message)
			return
		}
		newCode = code.Code
	}

	c.ResponseOk(newCode, into.GetId())
}
