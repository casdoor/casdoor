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

	"github.com/beego/beego/utils/pagination"
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

// GetVerifications
// @Title GetVerifications
// @Tag Verification API
// @Description get payments
// @Param   owner     query    string  true        "The owner of payments"
// @Success 200 {array} object.Verification The Response object
// @router /get-payments [get]
func (c *ApiController) GetVerifications() {
	owner := c.Input().Get("owner")
	limit := c.Input().Get("pageSize")
	page := c.Input().Get("p")
	field := c.Input().Get("field")
	value := c.Input().Get("value")
	sortField := c.Input().Get("sortField")
	sortOrder := c.Input().Get("sortOrder")

	if limit == "" || page == "" {
		payments, err := object.GetVerifications(owner)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		c.ResponseOk(payments)
	} else {
		limit := util.ParseInt(limit)
		count, err := object.GetVerificationCount(owner, field, value)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		paginator := pagination.SetPaginator(c.Ctx, limit, count)
		payments, err := object.GetPaginationVerifications(owner, paginator.Offset(), limit, field, value, sortField, sortOrder)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}

		c.ResponseOk(payments, paginator.Nums())
	}
}

// GetUserVerifications
// @Title GetUserVerifications
// @Tag Verification API
// @Description get payments for a user
// @Param   owner     query    string  true        "The owner of payments"
// @Param   organization    query   string  true   "The organization of the user"
// @Param   user    query   string  true           "The username of the user"
// @Success 200 {array} object.Verification The Response object
// @router /get-user-payments [get]
func (c *ApiController) GetUserVerifications() {
	owner := c.Input().Get("owner")
	user := c.Input().Get("user")

	payments, err := object.GetUserVerifications(owner, user)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(payments)
}

// GetVerification
// @Title GetVerification
// @Tag Verification API
// @Description get payment
// @Param   id     query    string  true        "The id ( owner/name ) of the payment"
// @Success 200 {object} object.Verification The Response object
// @router /get-payment [get]
func (c *ApiController) GetVerification() {
	id := c.Input().Get("id")

	payment, err := object.GetVerification(id)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(payment)
}

// SendVerificationCode ...
// @Title SendVerificationCode
// @Tag Verification API
// @router /send-verification-code [post]
// @Success 200 {object} object.Userinfo The Response object
func (c *ApiController) SendVerificationCode() {
	var vform form.VerificationForm
	err := c.ParseForm(&vform)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	clientIp := util.GetClientIpFromRequest(c.Ctx.Request)

	if msg := vform.CheckParameter(form.SendVerifyCode, c.GetAcceptLanguage()); msg != "" {
		c.ResponseError(msg)
		return
	}

	provider, err := object.GetCaptchaProviderByApplication(vform.ApplicationId, "false", c.GetAcceptLanguage())
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if provider != nil {
		if vform.CaptchaType != provider.Type {
			c.ResponseError(c.T("verification:Turing test failed."))
			return
		}

		if provider.Type != "Default" {
			vform.ClientSecret = provider.ClientSecret
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
		if user == nil || user.IsDeleted {
			c.ResponseError(c.T("verification:the user does not exist, please sign up first"))
			return
		}

		if user.IsForbidden {
			c.ResponseError(c.T("check:The user is forbidden to sign in, please contact the administrator"))
			return
		}
	}

	// mfaUserSession != "", means method is MfaAuthVerification
	if mfaUserSession := c.getMfaUserSession(); mfaUserSession != "" {
		user, err = object.GetUser(mfaUserSession)
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
			mfaProps := user.GetPreferredMfaProps(false)
			if user != nil && util.GetMaskedEmail(mfaProps.Secret) == vform.Dest {
				vform.Dest = mfaProps.Secret
			}
		}

		provider, err = application.GetEmailProvider(vform.Method)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}
		if provider == nil {
			c.ResponseError(fmt.Sprintf(c.T("verification:please add an Email provider to the \"Providers\" list for the application: %s"), application.Name))
			return
		}

		sendResp = object.SendVerificationCodeToEmail(organization, user, provider, clientIp, vform.Dest)
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
		} else if vform.Method == ResetVerification || vform.Method == MfaSetupVerification {
			if vform.CountryCode == "" {
				if user = c.getCurrentUser(); user != nil {
					vform.CountryCode = user.GetCountryCode(vform.CountryCode)
				}
			}
		} else if vform.Method == MfaAuthVerification {
			mfaProps := user.GetPreferredMfaProps(false)
			if user != nil && util.GetMaskedPhone(mfaProps.Secret) == vform.Dest {
				vform.Dest = mfaProps.Secret
			}

			vform.CountryCode = mfaProps.CountryCode
			vform.CountryCode = user.GetCountryCode(vform.CountryCode)
		}

		provider, err = application.GetSmsProvider(vform.Method, vform.CountryCode)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}
		if provider == nil {
			c.ResponseError(fmt.Sprintf(c.T("verification:please add a SMS provider to the \"Providers\" list for the application: %s"), application.Name))
			return
		}

		if phone, ok := util.GetE164Number(vform.Dest, vform.CountryCode); !ok {
			c.ResponseError(fmt.Sprintf(c.T("verification:Phone number is invalid in your region %s"), vform.CountryCode))
			return
		} else {
			sendResp = object.SendVerificationCodeToPhone(organization, user, provider, clientIp, phone)
		}
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
// @Success 200 {object} object.Userinfo The Response object
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

	captchaProvider, err := object.GetCaptchaProviderByOwnerName(vform.ApplicationId, c.GetAcceptLanguage())
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	if captchaProvider.Type != "Default" {
		vform.ClientSecret = captchaProvider.ClientSecret
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
// @router /reset-email-or-phone [post]
// @Success 200 {object} object.Userinfo The Response object
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

	result, err := object.CheckVerificationCode(checkDest, code, c.GetAcceptLanguage())
	if err != nil {
		c.ResponseError(c.T(err.Error()))
		return
	}
	if result.Code != object.VerificationSuccess {
		c.ResponseError(result.Msg)
		return
	}

	switch destType {
	case object.VerifyTypeEmail:
		user.Email = dest
		user.EmailVerified = true
		_, err = object.UpdateUser(user.GetId(), user, []string{"email", "email_verified"}, false)
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
// @router /verify-code [post]
// @Success 200 {object} object.Userinfo The Response object
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

	passed, err := c.checkOrgMasterVerificationCode(user, authForm.Code)
	if err != nil {
		c.ResponseError(c.T(err.Error()))
		return
	}

	if !passed {
		result, err := object.CheckVerificationCode(checkDest, authForm.Code, c.GetAcceptLanguage())
		if err != nil {
			c.ResponseError(err.Error())
			return
		}
		if result.Code != object.VerificationSuccess {
			c.ResponseError(result.Msg)
			return
		}

		err = object.DisableVerificationCode(checkDest)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}
	}

	c.SetSession("verifiedCode", authForm.Code)
	c.SetSession("verifiedUserId", user.GetId())
	c.ResponseOk()
}
