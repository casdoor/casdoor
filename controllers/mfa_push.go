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
	"sync"
	"time"

	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

var lastPushTimes sync.Map

// SendMfaPush
// @Title SendMfaPush
// @Tag MFA API
// @Description send a sign-in confirmation card to the user's phone, for the sign-in waiting for the second factor or the push MFA being set up
// @Param provider   formData string false "the provider, only when setting up"
// @Param receiver   formData string false "the receiver, only when setting up"
// @Param application formData string false "the application name shown on the card"
// @Success 200 {object} controllers.Response The Response object
// @router /mfa/push/send [post]
func (c *ApiController) SendMfaPush() {
	providerId := c.Ctx.Request.Form.Get("provider")
	receiver := c.Ctx.Request.Form.Get("receiver")
	application := c.Ctx.Request.Form.Get("application")

	// the sign-in waiting for the second factor uses the factor the user enabled, setting it up uses the form
	userId := c.getMfaUserSession()
	isSetup := userId == ""
	if isSetup {
		userId = c.GetSessionUsername()
	}
	if userId == "" {
		c.ResponseError(c.T("general:Please login first"))
		return
	}

	user, err := object.GetUser(userId)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	if user == nil {
		c.ResponseError(c.T("general:The user doesn't exist"))
		return
	}

	if !isSetup {
		providerId = user.MfaPushProvider
		receiver = user.MfaPushReceiver
	} else {
		provider, err := object.GetProvider(providerId)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}
		// the messages go to the account linked to the user when the action says which one, e.g. the WeCom user ID
		if field := object.GetPushReceiverField(provider); field != "" {
			if value := object.GetUserFieldValue(user, field); value != "" {
				receiver = value
			}
		}
	}

	if value, ok := lastPushTimes.Load(userId); ok && time.Since(value.(time.Time)) < 20*time.Second {
		c.ResponseError("please wait for 20 seconds before sending again")
		return
	}
	lastPushTimes.Store(userId, time.Now())

	origin := object.GetFrontendOrigin(c.Ctx.Request.Host)
	challenge, err := object.SendPushChallenge(userId, providerId, receiver, application, util.GetClientIpFromRequest(c.Ctx.Request), c.Ctx.Request.UserAgent(), origin)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(challenge.Id, receiver)
}

// GetMfaPushChallenge
// @Title GetMfaPushChallenge
// @Tag MFA API
// @Description get a sign-in confirmation request, for the confirmation page and the sign-in page waiting for it
// @Param id query string true "the request id"
// @Success 200 {object} object.PushChallenge The Response object
// @router /mfa/push/challenge [get]
func (c *ApiController) GetMfaPushChallenge() {
	challenge := object.GetPushChallenge(c.Ctx.Input.Query("id"))
	if challenge == nil {
		c.ResponseError("the sign-in request doesn't exist or has expired")
		return
	}
	c.ResponseOk(challenge)
}

// AnswerMfaPushChallenge
// @Title AnswerMfaPushChallenge
// @Tag MFA API
// @Description approve or deny a sign-in confirmation request on the phone
// @Param id      query string true "the request id"
// @Param approve query string true "true or false"
// @Success 200 {object} controllers.Response The Response object
// @router /mfa/push/answer [post]
func (c *ApiController) AnswerMfaPushChallenge() {
	challenge, err := object.AnswerPushChallenge(c.Ctx.Input.Query("id"), c.Ctx.Input.Query("approve") == "true")
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	c.ResponseOk(challenge.Status)
}
