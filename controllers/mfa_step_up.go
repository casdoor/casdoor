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

	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

type stepUpContext struct {
	user         *object.User
	application  *object.Application
	organization *object.Organization
	isAppCaller  bool
}

// getStepUpContext resolves who is verified and for which application: the signed-in user for the
// application of their access token or the given clientId, or, for an application calling with its own
// credentials, the user it names in its organization
func (c *ApiController) getStepUpContext() (*stepUpContext, bool) {
	username, ok := c.RequireSignedIn()
	if !ok {
		return nil, false
	}

	clientId := strings.TrimSpace(c.GetString("clientId"))
	ctx := &stepUpContext{}
	var err error

	if object.IsAppUser(username) {
		ctx.isAppCaller = true
		ctx.application, err = object.GetApplicationByUserId(username)
		if err != nil {
			c.ResponseError(err.Error())
			return nil, false
		}
		if ctx.application == nil {
			c.ResponseError(fmt.Sprintf(c.T("auth:The application: %s does not exist"), username))
			return nil, false
		}

		userId := c.GetString("userId")
		if userId == "" {
			userId = util.GetId(c.GetString("owner"), c.GetString("name"))
		}
		owner, name, err := util.GetOwnerAndNameFromIdWithError(userId)
		if err != nil || owner == "" || name == "" {
			c.ResponseError(c.T("general:Missing parameter") + ": owner, name")
			return nil, false
		}
		if owner != ctx.application.Organization {
			c.ResponseError(c.T("auth:Unauthorized operation"))
			return nil, false
		}
		ctx.user, err = object.GetUser(userId)
		if err != nil {
			c.ResponseError(err.Error())
			return nil, false
		}
		if ctx.user == nil {
			c.ResponseError(fmt.Sprintf(c.T("general:The user: %s doesn't exist"), userId))
			return nil, false
		}
	} else {
		ctx.user, ok = c.RequireSignedInUser()
		if !ok {
			return nil, false
		}

		_, aud := c.GetSessionOidc()
		if aud != "" {
			// an access token is bound to its application, it cannot ask for a proof for another one
			if clientId != "" && clientId != aud {
				c.ResponseError(c.T("auth:Unauthorized operation"))
				return nil, false
			}
			clientId = aud
		} else if !c.checkCredentialedOrigin() {
			return nil, false
		}

		if clientId == "" {
			c.ResponseError(c.T("general:Missing parameter") + ": clientId")
			return nil, false
		}
		ctx.application, err = object.GetApplicationByClientId(clientId)
		if err != nil {
			c.ResponseError(err.Error())
			return nil, false
		}
		if ctx.application == nil {
			c.ResponseError(fmt.Sprintf(c.T("auth:The application: %s does not exist"), clientId))
			return nil, false
		}
		// an administrator may try the step-up check of an application of the organization on its own account
		if ctx.user.Owner != ctx.application.Organization && !c.IsAdminOfOrganization(ctx.application.Organization) {
			c.ResponseError(c.T("auth:Unauthorized operation"))
			return nil, false
		}
	}

	if !ctx.application.EnableStepUpMfa {
		c.ResponseError(fmt.Sprintf(c.T("mfa:Step-up verification is not enabled for the application: %s"), ctx.application.Name))
		return nil, false
	}
	if ctx.user.IsForbidden || ctx.user.IsDeleted {
		c.ResponseError(c.T("check:The user is forbidden to sign in, please contact the administrator"))
		return nil, false
	}

	ctx.organization, err = object.GetOrganizationByUser(ctx.user)
	if err != nil {
		c.ResponseError(err.Error())
		return nil, false
	}
	if ctx.organization == nil {
		c.ResponseError(c.T("check:Organization does not exist"))
		return nil, false
	}

	c.Ctx.Input.SetParam("recordUserId", ctx.user.GetId())
	return ctx, true
}

// GetStepUpMethods
// @Title GetStepUpMethods
// @Tag MFA API
// @Description list the enrolled MFA factors of a user that an application accepts for a step-up check
// @Param   clientId    query    string  false  "The client ID of the application, not needed with an access token or the application's own credentials"
// @Param   owner       query    string  false  "The organization of the user, only for an application calling with its own credentials"
// @Param   name        query    string  false  "The name of the user, only for an application calling with its own credentials"
// @Success 200 {object} controllers.Response The Response object
// @router /mfa/step-up/methods [get]
func (c *ApiController) GetStepUpMethods() {
	ctx, ok := c.getStepUpContext()
	if !ok {
		return
	}

	c.ResponseOk(object.GetStepUpMethods(ctx.user, ctx.application))
}

// SendStepUpCode
// @Title SendStepUpCode
// @Tag MFA API
// @Description send a step-up code to the email address or phone number the user enrolled for MFA, never to one from the request
// @Param   mfaType     formData string  true   "email or sms"
// @Param   clientId    formData string  false  "The client ID of the application"
// @Param   owner       formData string  false  "The organization of the user, only for an application calling with its own credentials"
// @Param   name        formData string  false  "The name of the user, only for an application calling with its own credentials"
// @Success 200 {object} controllers.Response The Response object
// @router /mfa/step-up/send-code [post]
func (c *ApiController) SendStepUpCode() {
	ctx, ok := c.getStepUpContext()
	if !ok {
		return
	}

	mfaType := c.GetString("mfaType")
	if mfaType != object.EmailType && mfaType != object.SmsType {
		c.ResponseError(c.T("general:Missing parameter") + ": mfaType")
		return
	}

	lang := c.GetAcceptLanguage()
	mfaProps, err := object.GetStepUpMfaProps(ctx.user, ctx.application, mfaType, lang)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	clientIp := util.GetClientIpFromRequest(c.Ctx.Request)
	if mfaType == object.EmailType {
		provider, err := ctx.application.GetEmailProvider(MfaAuthVerification)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}
		if provider == nil {
			c.ResponseError(fmt.Sprintf(c.T("verification:please add an Email provider to the \"Providers\" list for the application: %s"), ctx.application.Name))
			return
		}
		err = object.SendVerificationCodeToEmail(ctx.organization, ctx.user, provider, clientIp, mfaProps.Secret, MfaAuthVerification, c.Ctx.Request.Host, ctx.application.Name, ctx.application, lang)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}
		c.ResponseOk(util.GetMaskedEmail(mfaProps.Secret))
		return
	}

	countryCode := ctx.user.GetCountryCode(mfaProps.CountryCode)
	provider, err := ctx.application.GetSmsProvider(MfaAuthVerification, countryCode)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	if provider == nil {
		c.ResponseError(fmt.Sprintf(c.T("verification:please add a SMS provider to the \"Providers\" list for the application: %s"), ctx.application.Name))
		return
	}
	phone, ok := util.GetE164Number(mfaProps.Secret, countryCode)
	if !ok {
		c.ResponseError(fmt.Sprintf(c.T("verification:Phone number is invalid in your region %s"), countryCode))
		return
	}
	err = object.SendVerificationCodeToPhone(ctx.organization, ctx.user, provider, clientIp, phone, ctx.application, lang)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	c.ResponseOk(util.GetMaskedPhone(mfaProps.Secret))
}

// VerifyStepUp
// @Title VerifyStepUp
// @Tag MFA API
// @Description verify a fresh code of an MFA factor the user has enrolled, without a new sign-in, and return a short-lived single-use proof token for the application
// @Param   mfaType      formData string  true   "app, email, sms, radius or recovery"
// @Param   passcode     formData string  false  "The code of the factor"
// @Param   recoveryCode formData string  false  "A recovery code, when mfaType is recovery"
// @Param   action       formData string  false  "A value naming the action the check approves, copied into the proof token"
// @Param   clientId     formData string  false  "The client ID of the application"
// @Param   owner        formData string  false  "The organization of the user, only for an application calling with its own credentials"
// @Param   name         formData string  false  "The name of the user, only for an application calling with its own credentials"
// @Success 200 {object} controllers.Response The Response object
// @router /mfa/step-up/verify [post]
func (c *ApiController) VerifyStepUp() {
	ctx, ok := c.getStepUpContext()
	if !ok {
		return
	}

	mfaType := c.GetString("mfaType")
	passcode := strings.TrimSpace(c.GetString("passcode"))
	recoveryCode := strings.TrimSpace(c.GetString("recoveryCode"))
	action := c.GetString("action")
	if mfaType == "" {
		c.ResponseError(c.T("general:Missing parameter") + ": mfaType")
		return
	}
	if len(action) > 500 {
		c.ResponseError(c.T("general:Missing parameter") + ": action")
		return
	}
	if mfaType == object.RecoveryCodeType && recoveryCode == "" {
		recoveryCode = passcode
	}
	if (mfaType == object.RecoveryCodeType && recoveryCode == "") || (mfaType != object.RecoveryCodeType && passcode == "") {
		c.ResponseError("missing passcode or recovery code")
		return
	}

	lang := c.GetAcceptLanguage()
	mfaProps, err := object.GetStepUpMfaProps(ctx.user, ctx.application, mfaType, lang)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	// an application calling for its users from one server would be locked out by the IP limit, its
	// calls count against each user's limit only
	clientIp := ""
	if !ctx.isAppCaller {
		clientIp = util.GetClientIpFromRequest(c.Ctx.Request)
	}

	// no master verification code and no remembered MFA: every check needs a fresh code of the enrolled factor
	err = object.VerifyStepUpWithLimit(ctx.user, clientIp, func() error {
		if mfaType == object.RecoveryCodeType {
			return object.MfaRecover(ctx.user, recoveryCode)
		}
		return object.VerifyMfaPasscode(ctx.user, mfaProps, passcode, lang)
	}, lang)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	token, expiresIn, err := object.IssueStepUpToken(c.Ctx.Request.Host, ctx.application, ctx.user, mfaType, action)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(map[string]interface{}{
		"token":     token,
		"tokenType": object.StepUpTokenType,
		"expiresIn": expiresIn,
		"mfaType":   mfaType,
	})
}

// RedeemStepUpToken
// @Title RedeemStepUpToken
// @Tag MFA API
// @Description check a step-up proof token with the application's own credentials and mark it used, a token is accepted only once
// @Param   token    formData string  true   "The proof token returned by /api/mfa/step-up/verify"
// @Param   action   formData string  false  "When given, the token must have been issued for this action"
// @Success 200 {object} controllers.Response The Response object
// @router /mfa/step-up/redeem [post]
func (c *ApiController) RedeemStepUpToken() {
	username, ok := c.RequireSignedIn()
	if !ok {
		return
	}
	if !object.IsAppUser(username) {
		c.ResponseError(c.T("auth:Unauthorized operation"))
		return
	}

	application, err := object.GetApplicationByUserId(username)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	if application == nil {
		c.ResponseError(fmt.Sprintf(c.T("auth:The application: %s does not exist"), username))
		return
	}

	token := strings.TrimSpace(c.GetString("token"))
	if token == "" {
		c.ResponseError(c.T("general:Missing parameter") + ": token")
		return
	}

	claims, err := object.RedeemStepUpToken(application, token, c.GetString("action"), c.GetAcceptLanguage())
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Ctx.Input.SetParam("recordUserId", util.GetId(claims.Owner, claims.Name))
	c.ResponseOk(claims)
}

// GetStepUpRecords
// @Title GetStepUpRecords
// @Tag MFA API
// @Description list the latest passed step-up checks of an application
// @Param   id    query    string  true   "The id ( owner/name ) of the application"
// @Success 200 {object} controllers.Response The Response object
// @router /get-step-up-records [get]
func (c *ApiController) GetStepUpRecords() {
	application, err := object.GetApplication(c.GetString("id"))
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	if application == nil {
		c.ResponseError(fmt.Sprintf(c.T("auth:The application: %s does not exist"), c.GetString("id")))
		return
	}
	if !c.IsAdminOfOrganization(application.Organization) {
		c.ResponseError(c.T("auth:Unauthorized operation"))
		return
	}

	records, err := object.GetStepUpRecords(application, 20)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	c.ResponseOk(records)
}
