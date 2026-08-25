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
	"strings"

	"github.com/beego/beego/v2/core/logs"
	"github.com/casdoor/casdoor/conf"
	"github.com/casdoor/casdoor/form"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

const (
	maskedRecoveryActiveSessionKey       = "maskedRecoveryActive"
	maskedRecoveryOrganizationSessionKey = "maskedRecoveryOrganization"
	maskedRecoveryIdentifierSessionKey   = "maskedRecoveryIdentifier"
	maskedRecoveryApplicationSessionKey  = "maskedRecoveryApplication"
	maskedRecoveryUserSessionKey         = "maskedRecoveryUser"
	maskedRecoveryDestSessionKey         = "maskedRecoveryDest"

	maskedRecoveryEmail = "a***@example.com"
	maskedRecoveryPhone = "*******"
)

func isMaskedPasswordRecovery(method string) bool {
	return method == ForgetVerification && conf.GetConfigBool("enableErrorMask")
}

func getMaskedPasswordRecoveryPublicUser(identifier string) object.User {
	return object.User{
		Name:  identifier,
		Email: maskedRecoveryEmail,
		Phone: maskedRecoveryPhone,
	}
}

func (c *ApiController) getSessionString(key string) string {
	value := c.GetSession(key)
	if value == nil {
		return ""
	}

	result, _ := value.(string)
	return result
}

func (c *ApiController) prepareMaskedPasswordRecovery(organization string, identifier string) {
	c.SetSession(maskedRecoveryActiveSessionKey, "true")
	c.SetSession(maskedRecoveryOrganizationSessionKey, organization)
	c.SetSession(maskedRecoveryIdentifierSessionKey, identifier)
	c.SetSession(maskedRecoveryApplicationSessionKey, "")
	c.SetSession(maskedRecoveryUserSessionKey, "")
	c.SetSession(maskedRecoveryDestSessionKey, "")

	user, err := object.GetUserByFields(organization, identifier)
	if err != nil {
		logs.Error("Failed to prepare masked password recovery: %v", err)
		return
	}
	if user == nil || user.IsDeleted || user.IsForbidden || object.CheckLdapPasswordForget(user) != nil {
		return
	}

	c.SetSession(maskedRecoveryUserSessionKey, user.GetId())
}

func (c *ApiController) getMaskedPasswordRecoveryUser() (*object.User, error) {
	userId := c.getSessionString(maskedRecoveryUserSessionKey)
	if userId == "" {
		return nil, nil
	}

	user, err := object.GetUser(userId)
	if err != nil || user == nil {
		return user, err
	}
	if user.IsDeleted || user.IsForbidden || object.CheckLdapPasswordForget(user) != nil {
		return nil, nil
	}

	return user, nil
}

func selectMaskedPasswordRecoveryType(user *object.User, preferredType string, canEmail bool, canPhone bool) string {
	if preferredType == object.VerifyTypeEmail && canEmail && user.Email != "" {
		return object.VerifyTypeEmail
	}
	if preferredType == object.VerifyTypePhone && canPhone && user.Phone != "" {
		return object.VerifyTypePhone
	}
	if canPhone && user.Phone != "" {
		return object.VerifyTypePhone
	}
	if canEmail && user.Email != "" {
		return object.VerifyTypeEmail
	}

	return ""
}

func (c *ApiController) sendMaskedPasswordRecoveryCode(application *object.Application, organization *object.Organization, preferredType string, clientIp string) {
	c.SetSession(maskedRecoveryApplicationSessionKey, application.Name)
	c.SetSession(maskedRecoveryDestSessionKey, "")

	user, err := c.getMaskedPasswordRecoveryUser()
	if err != nil {
		logs.Error("Failed to resolve masked password recovery user: %v", err)
		return
	}
	if user == nil || user.Owner != application.Organization {
		return
	}

	emailProvider, emailErr := application.GetEmailProvider(ForgetVerification)
	if emailErr != nil {
		logs.Error("Failed to resolve password recovery email provider: %v", emailErr)
	}

	countryCode := user.GetCountryCode("")
	smsProvider, smsErr := application.GetSmsProvider(ForgetVerification, countryCode)
	if smsErr != nil {
		logs.Error("Failed to resolve password recovery SMS provider: %v", smsErr)
	}

	verificationType := selectMaskedPasswordRecoveryType(user, preferredType, emailProvider != nil, smsProvider != nil)
	var dest string
	switch verificationType {
	case object.VerifyTypeEmail:
		dest = user.Email
		if emailProvider.HttpHeaders == nil {
			emailProvider.HttpHeaders = map[string]string{}
		}
		if _, ok := emailProvider.HttpHeaders["Accept-Language"]; !ok {
			emailProvider.HttpHeaders["Accept-Language"] = c.GetAcceptLanguage()
		}
		err = object.SendVerificationCodeToEmail(organization, user, emailProvider, clientIp, dest, ForgetVerification, c.Ctx.Request.Host, application.Name, application)
	case object.VerifyTypePhone:
		var ok bool
		dest, ok = util.GetE164Number(user.Phone, countryCode)
		if !ok {
			return
		}
		err = object.SendVerificationCodeToPhone(organization, user, smsProvider, clientIp, dest, application)
	default:
		return
	}
	if err != nil {
		logs.Error("Failed to send masked password recovery code: %v", err)
		return
	}

	c.SetSession(maskedRecoveryDestSessionKey, dest)
}

func shouldEnableMaskedPasswordRecoveryCaptcha(application *object.Application, clientIp string) bool {
	for _, providerItem := range application.Providers {
		if providerItem.Provider == nil || providerItem.Provider.Category != "Captcha" {
			continue
		}

		switch providerItem.Rule {
		case "Always", "Dynamic":
			return true
		case "Internet-Only":
			return util.IsInternetIp(clientIp)
		default:
			return false
		}
	}

	return false
}

func (c *ApiController) isMaskedPasswordRecoveryCaptchaRequest(organization string, identifier string) bool {
	return conf.GetConfigBool("enableErrorMask") &&
		c.getSessionString(maskedRecoveryActiveSessionKey) == "true" &&
		c.getSessionString(maskedRecoveryOrganizationSessionKey) == organization &&
		c.getSessionString(maskedRecoveryIdentifierSessionKey) == identifier
}

func (c *ApiController) isMaskedPasswordRecoveryVerification(authForm *form.AuthForm) bool {
	return conf.GetConfigBool("enableErrorMask") &&
		c.getSessionString(maskedRecoveryActiveSessionKey) == "true" &&
		c.getSessionString(maskedRecoveryOrganizationSessionKey) == authForm.Organization &&
		strings.EqualFold(c.getSessionString(maskedRecoveryIdentifierSessionKey), authForm.Name) &&
		c.getSessionString(maskedRecoveryApplicationSessionKey) == authForm.Application
}

func (c *ApiController) verifyMaskedPasswordRecoveryCode(authForm *form.AuthForm) {
	responseError := func() {
		c.ResponseError(c.T("verification:Wrong verification code!"))
	}

	user, err := c.getMaskedPasswordRecoveryUser()
	dest := c.getSessionString(maskedRecoveryDestSessionKey)
	if err != nil || user == nil || dest == "" {
		responseError()
		return
	}

	passed, err := c.checkOrgMasterVerificationCode(user, authForm.Code)
	if err != nil {
		responseError()
		return
	}
	if !passed {
		clientIp := util.GetClientIpFromRequest(c.Ctx.Request)
		if err = object.CheckVerifyCodeWithLimitAndIp(user, clientIp, dest, authForm.Code, c.GetAcceptLanguage()); err != nil {
			responseError()
			return
		}
		if err = object.DisableVerificationCode(dest); err != nil {
			responseError()
			return
		}
	}

	c.SetSession("verifiedCode", authForm.Code)
	c.SetSession("verifiedUserId", user.GetId())
	c.clearMaskedPasswordRecovery()
	c.ResponseOk()
}

func (c *ApiController) clearMaskedPasswordRecovery() {
	c.SetSession(maskedRecoveryActiveSessionKey, "")
	c.SetSession(maskedRecoveryOrganizationSessionKey, "")
	c.SetSession(maskedRecoveryIdentifierSessionKey, "")
	c.SetSession(maskedRecoveryApplicationSessionKey, "")
	c.SetSession(maskedRecoveryUserSessionKey, "")
	c.SetSession(maskedRecoveryDestSessionKey, "")
}
