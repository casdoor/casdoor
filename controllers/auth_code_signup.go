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
	"fmt"
	"strings"

	"github.com/casdoor/casdoor/form"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

// answerCodeSigninSignup answers a verification-code sign-in of an email or phone that has
// no account. When the application takes signups, the code is checked but not used up: the
// signup page posts it again with the rest of the signup items, and /api/signup checks it
// the same way as a code the signup page sent itself
func (c *ApiController) answerCodeSigninSignup(application *object.Application, authForm *form.AuthForm) {
	verifyType := object.GetVerifyType(authForm.Username)
	if !application.IsCodeSigninSignupEnabled(verifyType) || !application.IsSignupAllowedFor(authForm.Organization) {
		c.ResponseError(fmt.Sprintf(c.T("general:The user: %s doesn't exist"), util.GetId(authForm.Organization, authForm.Username)))
		return
	}

	dest := strings.TrimSpace(authForm.Username)
	checkDest := dest
	countryCode := authForm.CountryCode
	if verifyType == object.VerifyTypePhone {
		// a national number is the one of the organization's first region, as
		// /api/send-verification-code resolves it for an address with no account
		if countryCode == "" && !strings.HasPrefix(dest, "+") {
			organization, err := object.GetOrganization(util.GetId("admin", application.Organization))
			if err != nil {
				c.ResponseError(err.Error())
				return
			}
			if organization != nil && len(organization.CountryCodes) > 0 {
				countryCode = organization.CountryCodes[0]
			}
		}

		var ok bool
		if checkDest, ok = util.GetE164Number(dest, countryCode); !ok {
			c.ResponseError(fmt.Sprintf(c.T("verification:Phone number is invalid in your region %s"), countryCode))
			return
		}
		dest, countryCode = util.ParseE164Phone(checkDest)
	}

	clientIp := util.GetClientIpFromRequest(c.Ctx.Request)
	err := object.CheckVerifyCodeWithLimitAndIp(nil, clientIp, checkDest, authForm.Code, c.GetAcceptLanguage())
	if err != nil {
		c.ResponseError(fmt.Sprintf("%s - %s", verifyType, err.Error()))
		return
	}

	c.ResponseOk(object.NextSignup, map[string]string{
		"type":        verifyType,
		"dest":        dest,
		"countryCode": countryCode,
	})
}
