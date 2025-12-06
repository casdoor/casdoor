// Copyright 2025 The Casdoor Authors. All Rights Reserved.
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
"encoding/json"

"github.com/casdoor/casdoor/object"
"github.com/casdoor/casdoor/util"
)

type VerifyIdentificationRequest struct {
Owner    string `json:"owner"`
Name     string `json:"name"`
RealName string `json:"realName"`
}

// VerifyIdentification
// @Title VerifyIdentification
// @Tag User API
// @Description verify user's real identification
// @Param   body    body   VerifyIdentificationRequest  true        "The details of the verification request"
// @Success 200 {object} controllers.Response The Response object
// @router /verify-identification [post]
func (c *ApiController) VerifyIdentification() {
var req VerifyIdentificationRequest
err := json.Unmarshal(c.Ctx.Input.RequestBody, &req)
if err != nil {
c.ResponseError(err.Error())
return
}

userId := util.GetId(req.Owner, req.Name)
user, err := object.GetUser(userId)
if err != nil {
c.ResponseError(err.Error())
return
}

if user == nil {
c.ResponseError(c.T("user:The user does not exist"))
return
}

requestUserId := c.GetSessionUsername()
if requestUserId != userId && !c.IsAdmin() {
c.ResponseError(c.T("auth:Unauthorized operation"))
return
}

application, err := object.GetApplicationByUser(user)
if err != nil {
c.ResponseError(err.Error())
return
}

if application == nil {
c.ResponseError(c.T("application:The application does not exist"))
return
}

provider, err := object.GetIdvProviderByApplication(util.GetId(application.Owner, application.Name), "false", c.GetAcceptLanguage())
if err != nil {
c.ResponseError(err.Error())
return
}

if provider == nil {
c.ResponseError(c.T("provider:No ID verification provider configured for this application"))
return
}

verified, err := object.VerifyIdentification(user, provider, req.RealName, c.GetAcceptLanguage())
if err != nil {
c.ResponseError(err.Error())
return
}

if verified {
c.ResponseOk("Verification successful", user)
} else {
c.ResponseError(c.T("user:Verification failed"))
}
}

type TestIdvProviderRequest struct {
Provider   object.Provider `json:"provider"`
IdCardType string          `json:"idCardType"`
IdCard     string          `json:"idCard"`
RealName   string          `json:"realName"`
}

// TestIdvProvider
// @Title TestIdvProvider
// @Tag Provider API
// @Description test ID verification provider
// @Param   body    body   TestIdvProviderRequest  true        "The details of the test request"
// @Success 200 {object} controllers.Response The Response object
// @router /test-idv-provider [post]
func (c *ApiController) TestIdvProvider() {
var req TestIdvProviderRequest
err := json.Unmarshal(c.Ctx.Input.RequestBody, &req)
if err != nil {
c.ResponseError(err.Error())
return
}

if !c.IsAdmin() {
c.ResponseError(c.T("auth:Unauthorized operation"))
return
}

if req.IdCardType == "" {
req.IdCardType = "ID_CARD"
}

if req.IdCard == "" {
req.IdCard = "123456789"
}

if req.RealName == "" {
req.RealName = "Test User"
}

testUser := &object.User{
IdCardType: req.IdCardType,
IdCard:     req.IdCard,
}

verified, err := object.VerifyIdentification(testUser, &req.Provider, req.RealName, c.GetAcceptLanguage())
if err != nil {
c.ResponseError(err.Error())
return
}

if verified {
c.ResponseOk("Test successful - Provider is working correctly")
} else {
c.ResponseError(c.T("provider:Test failed - Provider verification returned false"))
}
}
