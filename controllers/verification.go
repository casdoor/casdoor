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
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/casdoor/casdoor/captcha"
	"github.com/casdoor/casdoor/form"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

const (
	SignupVerification   = "signup"
	ResetVerification    = "reset"
	LoginVerification    = "login"
	ForgetVerification   = "forget"
	MfaSetupVerification = "mfaSetup"
	MfaAuthVerification  = "mfaAuth"
)

// SendVerificationCode ...
// @Title SendVerificationCode
// @Tag Verification API
// @router /send-verification-code [post]
func (c *ApiController) SendVerificationCode() {
	var vform form.VerificationForm
	err := c.ParseForm(&vform)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	remoteAddr := util.GetIPFromRequest(c.Ctx.Request)

	if msg := vform.CheckParameter(form.SendVerifyCode, c.GetAcceptLanguage()); msg != "" {
		c.ResponseError(msg)
		return
	}

	if vform.CaptchaType != "none" {
		if captchaProvider := captcha.GetCaptchaProvider(vform.CaptchaType); captchaProvider == nil {
			c.ResponseError(c.T("general:don't support captchaProvider: ") + vform.CaptchaType)
			return
		} else if isHuman, err := captchaProvider.VerifyCaptcha(vform.CaptchaToken, vform.ClientSecret); err != nil {
			c.ResponseError(err.Error())
			return
		} else if !isHuman {
			c.ResponseError(c.T("verification:Turing test failed."))
			return
		}
	}

	application, err := object.GetApplication(vform.ApplicationId)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	organization, err := object.GetOrganization(util.GetId(application.Owner, application.Organization))
	if err != nil {
		c.ResponseError(c.T(err.Error()))
	}

	if organization == nil {
		c.ResponseError(c.T("check:Organization does not exist"))
		return
	}

	var user *object.User
	// checkUser != "", means method is ForgetVerification
	if vform.CheckUser != "" {
		owner := application.Organization
		user, err = object.GetUser(util.GetId(owner, vform.CheckUser))
		if err != nil {
			c.ResponseError(err.Error())
			return
		}
	}

	// mfaSessionData != nil, means method is MfaSetupVerification
	if mfaSessionData := c.getMfaSessionData(); mfaSessionData != nil {
		user, err = object.GetUser(mfaSessionData.UserId)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}
	}

	sendResp := errors.New("invalid dest type")

	switch vform.Type {
	case object.VerifyTypeEmail:
		if !util.IsEmailValid(vform.Dest) {
			c.ResponseError(c.T("check:Email is invalid"))
			return
		}

		if vform.Method == LoginVerification || vform.Method == ForgetVerification {
			if user != nil && util.GetMaskedEmail(user.Email) == vform.Dest {
				vform.Dest = user.Email
			}

			user, err = object.GetUserByEmail(organization.Name, vform.Dest)
			if err != nil {
				c.ResponseError(err.Error())
				return
			}

			if user == nil {
				c.ResponseError(c.T("verification:the user does not exist, please sign up first"))
				return
			}
		} else if vform.Method == ResetVerification {
			user = c.getCurrentUser()
		} else if vform.Method == MfaAuthVerification {
			mfaProps := user.GetPreferMfa(false)
			if user != nil && util.GetMaskedEmail(mfaProps.Secret) == vform.Dest {
				vform.Dest = mfaProps.Secret
			}
		}

		provider, err := application.GetEmailProvider()
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		sendResp = object.SendVerificationCodeToEmail(organization, user, provider, remoteAddr, vform.Dest)
	case object.VerifyTypePhone:
		if vform.Method == LoginVerification || vform.Method == ForgetVerification {
			if user != nil && util.GetMaskedPhone(user.Phone) == vform.Dest {
				vform.Dest = user.Phone
			}

			if user, err = object.GetUserByPhone(organization.Name, vform.Dest); err != nil {
				c.ResponseError(err.Error())
				return
			} else if user == nil {
				c.ResponseError(c.T("verification:the user does not exist, please sign up first"))
				return
			}

			vform.CountryCode = user.GetCountryCode(vform.CountryCode)
		} else if vform.Method == ResetVerification {
			if user = c.getCurrentUser(); user != nil {
				vform.CountryCode = user.GetCountryCode(vform.CountryCode)
			}
		} else if vform.Method == MfaAuthVerification {
			mfaProps := user.GetPreferMfa(false)
			if user != nil && util.GetMaskedPhone(mfaProps.Secret) == vform.Dest {
				vform.Dest = mfaProps.Secret
			}

			vform.CountryCode = mfaProps.CountryCode
		}

		provider, err := application.GetSmsProvider()
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		if phone, ok := util.GetE164Number(vform.Dest, vform.CountryCode); !ok {
			c.ResponseError(fmt.Sprintf(c.T("verification:Phone number is invalid in your region %s"), vform.CountryCode))
			return
		} else {
			sendResp = object.SendVerificationCodeToPhone(organization, user, provider, remoteAddr, phone)
		}
	}

	if vform.Method == MfaSetupVerification {
		c.SetSession(object.MfaSmsCountryCodeSession, vform.CountryCode)
		c.SetSession(object.MfaSmsDestSession, vform.Dest)
	}

	if sendResp != nil {
		c.ResponseError(sendResp.Error())
	} else {
		c.ResponseOk()
	}
}

// VerifyCaptcha ...
// @Title VerifyCaptcha
// @Tag Verification API
// @router /verify-captcha [post]
func (c *ApiController) VerifyCaptcha() {
	var vform form.VerificationForm
	err := c.ParseForm(&vform)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if msg := vform.CheckParameter(form.VerifyCaptcha, c.GetAcceptLanguage()); msg != "" {
		c.ResponseError(msg)
		return
	}

	provider := captcha.GetCaptchaProvider(vform.CaptchaType)
	if provider == nil {
		c.ResponseError(c.T("verification:Invalid captcha provider."))
		return
	}

	isValid, err := provider.VerifyCaptcha(vform.CaptchaToken, vform.ClientSecret)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(isValid)
}

// ResetEmailOrPhone ...
// @Tag Account API
// @Title ResetEmailOrPhone
// @router /api/reset-email-or-phone [post]
func (c *ApiController) ResetEmailOrPhone() {
	user, ok := c.RequireSignedInUser()
	if !ok {
		return
	}

	destType := c.Ctx.Request.Form.Get("type")
	dest := c.Ctx.Request.Form.Get("dest")
	code := c.Ctx.Request.Form.Get("code")

	if util.IsStringsEmpty(destType, dest, code) {
		c.ResponseError(c.T("general:Missing parameter"))
		return
	}

	checkDest := dest
	organization, err := object.GetOrganizationByUser(user)
	if err != nil {
		c.ResponseError(c.T(err.Error()))
		return
	}

	if destType == object.VerifyTypePhone {
		if object.HasUserByField(user.Owner, "phone", dest) {
			c.ResponseError(c.T("check:Phone already exists"))
			return
		}

		phoneItem := object.GetAccountItemByName("Phone", organization)
		if phoneItem == nil {
			c.ResponseError(c.T("verification:Unable to get the phone modify rule."))
			return
		}

		if pass, errMsg := object.CheckAccountItemModifyRule(phoneItem, user.IsAdminUser(), c.GetAcceptLanguage()); !pass {
			c.ResponseError(errMsg)
			return
		}
		if checkDest, ok = util.GetE164Number(dest, user.GetCountryCode("")); !ok {
			c.ResponseError(fmt.Sprintf(c.T("verification:Phone number is invalid in your region %s"), user.CountryCode))
			return
		}
	} else if destType == object.VerifyTypeEmail {
		if object.HasUserByField(user.Owner, "email", dest) {
			c.ResponseError(c.T("check:Email already exists"))
			return
		}

		emailItem := object.GetAccountItemByName("Email", organization)
		if emailItem == nil {
			c.ResponseError(c.T("verification:Unable to get the email modify rule."))
			return
		}

		if pass, errMsg := object.CheckAccountItemModifyRule(emailItem, user.IsAdminUser(), c.GetAcceptLanguage()); !pass {
			c.ResponseError(errMsg)
			return
		}
	}

	if result := object.CheckVerificationCode(checkDest, code, c.GetAcceptLanguage()); result.Code != object.VerificationSuccess {
		c.ResponseError(result.Msg)
		return
	}

	switch destType {
	case object.VerifyTypeEmail:
		user.Email = dest
		_, err = object.SetUserField(user, "email", user.Email)
	case object.VerifyTypePhone:
		user.Phone = dest
		_, err = object.SetUserField(user, "phone", user.Phone)
	default:
		c.ResponseError(c.T("verification:Unknown type"))
		return
	}
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	err = object.DisableVerificationCode(checkDest)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk()
}

// VerifyCode
// @Tag Verification API
// @Title VerifyCode
// @router /api/verify-code [post]
func (c *ApiController) VerifyCode() {
	var authForm form.AuthForm
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &authForm)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	var user *object.User
	if authForm.Name != "" {
		user, err = object.GetUserByFields(authForm.Organization, authForm.Name)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}
	}

	var checkDest string
	if strings.Contains(authForm.Username, "@") {
		if user != nil && util.GetMaskedEmail(user.Email) == authForm.Username {
			authForm.Username = user.Email
		}
		checkDest = authForm.Username
	} else {
		if user != nil && util.GetMaskedPhone(user.Phone) == authForm.Username {
			authForm.Username = user.Phone
		}
	}

	if user, err = object.GetUserByFields(authForm.Organization, authForm.Username); err != nil {
		c.ResponseError(err.Error())
		return
	} else if user == nil {
		c.ResponseError(fmt.Sprintf(c.T("general:The user: %s doesn't exist"), util.GetId(authForm.Organization, authForm.Username)))
		return
	}

	verificationCodeType := object.GetVerifyType(authForm.Username)
	if verificationCodeType == object.VerifyTypePhone {
		authForm.CountryCode = user.GetCountryCode(authForm.CountryCode)
		var ok bool
		if checkDest, ok = util.GetE164Number(authForm.Username, authForm.CountryCode); !ok {
			c.ResponseError(fmt.Sprintf(c.T("verification:Phone number is invalid in your region %s"), authForm.CountryCode))
			return
		}
	}

	if result := object.CheckVerificationCode(checkDest, authForm.Code, c.GetAcceptLanguage()); result.Code != object.VerificationSuccess {
		c.ResponseError(result.Msg)
		return
	}
	err = object.DisableVerificationCode(checkDest)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	c.SetSession("verifiedCode", authForm.Code)

	c.ResponseOk()
}
