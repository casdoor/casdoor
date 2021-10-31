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

// Casdoor will expose its providers as services to SDK
// We are going to implement those services as APIs here

package controllers

import (
	"encoding/json"
	"fmt"

	"github.com/casbin/casdoor/object"
	"github.com/casbin/casdoor/util"
)

// SendEmail
// @Title SendEmail
// @Description This API is not for Casdoor frontend to call, it is for Casdoor SDKs.
// @Param   clientId    query    string  true        "The clientId of the application"
// @Param   clientSecret    query    string  true    "The clientSecret of the application"
// @Param   body    body   emailForm    true         "Details of the email request"
// @Success 200 {object}  Response object
// @router /api/send-email [post]
func (c *ApiController) SendEmail() {
	provider, _, ok := c.GetProviderFromContext("Email")
	if !ok {
		return
	}

	var emailForm struct {
		Title     string   `json:"title"`
		Content   string   `json:"content"`
		Sender    string   `json:"sender"`
		Receivers []string `json:"receivers"`
	}
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &emailForm)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if util.IsStrsEmpty(emailForm.Title, emailForm.Content, emailForm.Sender) {
		c.ResponseError(fmt.Sprintf("Empty parameters for emailForm: %v", emailForm))
		return
	}

	invalidReceivers := []string{}
	for _, receiver := range emailForm.Receivers {
		if !util.IsEmailValid(receiver) {
			invalidReceivers = append(invalidReceivers, receiver)
		}
	}

	if len(invalidReceivers) != 0 {
		c.ResponseError(fmt.Sprintf("Invalid Email receivers: %s", invalidReceivers))
		return
	}

	for _, receiver := range emailForm.Receivers {
		err = object.SendEmail(provider, emailForm.Title, emailForm.Content, receiver, emailForm.Sender)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}
	}

	c.ResponseOk()
}

// SendSms
// @Title SendSms
// @Description This API is not for Casdoor frontend to call, it is for Casdoor SDKs.
// @Param   clientId    query    string  true        "The clientId of the application"
// @Param   clientSecret    query    string  true    "The clientSecret of the application"
// @Param   body    body   smsForm    true           "Details of the sms request"
// @Success 200 {object}  Response object
// @router /api/send-sms [post]
func (c *ApiController) SendSms() {
	provider, _, ok := c.GetProviderFromContext("SMS")
	if !ok {
		return
	}

	var smsForm struct {
		Content   string   `json:"content"`
		Receivers []string `json:"receivers"`
		OrgId     string   `json:"organizationId"` // e.g. "admin/built-in"
	}
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &smsForm)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	org := object.GetOrganization(smsForm.OrgId)
	var invalidReceivers []string
	for idx, receiver := range smsForm.Receivers {
		if !util.IsPhoneCnValid(receiver) {
			invalidReceivers = append(invalidReceivers, receiver)
		} else {
			smsForm.Receivers[idx] = fmt.Sprintf("+%s%s", org.PhonePrefix, receiver)
		}
	}

	if len(invalidReceivers) != 0 {
		c.ResponseError(fmt.Sprintf("Invalid phone receivers: %s", invalidReceivers))
		return
	}

	err = object.SendSms(provider, smsForm.Content, smsForm.Receivers...)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk()
}
