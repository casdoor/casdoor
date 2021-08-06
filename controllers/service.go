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

	"github.com/casbin/casdoor/object"
	"github.com/casbin/casdoor/util"
	sender "github.com/casdoor/go-sms-sender"
)

// @Title SendEmail
// @Description This API is not for Casdoor frontend to call, it is for Casdoor SDKs.
// @Param   clientId    query    string  true        "The clientId of the application"
// @Param   clientSecret    query    string  true        "The clientSecret of the application"
// @Param   body    body   emailForm    true        "Details of the email request"
// @Success 200 {object}  Response object
// @router /api/send-email [post]
func (c *ApiController) SendEmail() {
	clientId := c.Input().Get("clientId")
	clientSecret := c.Input().Get("clientSecret")
	app := object.GetApplicationByClientIdAndSecret(clientId, clientSecret)
	if app == nil {
		c.ResponseError("Invalid clientId or clientSecret.")
		return
	}

	provider := app.GetEmailProvider()
	if provider == nil {
		c.ResponseError("No Email provider is found")
		return
	}

	var emailForm struct {
		Title     string   `json:"title"`
		Content   string   `json:"content"`
		Receivers []string `json:"receivers"`
		Sender    string   `json:"sender"`
	}

	err := json.Unmarshal(c.Ctx.Input.RequestBody, &emailForm)
	if err != nil {
		c.ResponseError("Request body error.")
		return
	}

	if util.IsStrsEmpty(emailForm.Title, emailForm.Content, emailForm.Sender) {
		c.ResponseError("Missing parameters.")
		return
	}

	var invalidEmails []string
	for _, receiver := range emailForm.Receivers {
		if !util.IsEmailValid(receiver) {
			invalidEmails = append(invalidEmails, receiver)
		}
	}

	if len(invalidEmails) != 0 {
		c.ResponseError("Invalid Email addresses", invalidEmails)
		return
	}

	ok := 0
	for _, receiver := range emailForm.Receivers {
		if msg := object.SendEmail(
			provider,
			emailForm.Title,
			emailForm.Content,
			receiver,
			emailForm.Sender);
			len(msg) == 0 {
			ok++
		}
	}

	c.Data["json"] = Response{Status: "ok", Data: ok}
	c.ServeJSON()
}

// @Title SendSms
// @Description This API is not for Casdoor frontend to call, it is for Casdoor SDKs.
// @Param   clientId    query    string  true        "The clientId of the application"
// @Param   clientSecret    query    string  true        "The clientSecret of the application"
// @Param   body    body   smsForm    true        "Details of the sms request"
// @Success 200 {object}  Response object
// @router /api/send-sms [post]
func (c *ApiController) SendSms() {
	clientId := c.Input().Get("clientId")
	clientSecret := c.Input().Get("clientSecret")
	app := object.GetApplicationByClientIdAndSecret(clientId, clientSecret)
	if app == nil {
		c.ResponseError("Invalid clientId or clientSecret.")
		return
	}

	provider := app.GetSmsProvider()
	if provider == nil {
		c.ResponseError("No SMS provider is found")
		return
	}

	client := sender.NewSmsClient(
		provider.Type,
		provider.ClientId,
		provider.ClientSecret,
		provider.SignName,
		provider.RegionId,
		provider.TemplateCode,
		provider.AppId,
	)
	if client == nil {
		c.ResponseError("Invalid provider info.")
		return
	}

	var smsForm struct {
		Receivers  []string          `json:"receivers"`
		Parameters map[string]string `json:"parameters"`
	}

	err := json.Unmarshal(c.Ctx.Input.RequestBody, &smsForm)
	if err != nil {
		c.ResponseError("Request body error.")
		return
	}

	var invalidReceivers []string
	for _, receiver := range smsForm.Receivers {
		if !util.IsPhoneCnValid(receiver) {
			invalidReceivers = append(invalidReceivers, receiver)
		}
	}

	if len(invalidReceivers) != 0{
		c.ResponseError("Invalid phone numbers", invalidReceivers)
		return
	}

	client.SendMessage(smsForm.Parameters, smsForm.Receivers...)
	c.Data["json"] = Response{Status: "ok"}
	c.ServeJSON()
}
