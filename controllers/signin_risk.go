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
	"strconv"
	"strings"
	"time"

	"github.com/casdoor/casdoor/ee"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

const (
	signinAssessmentKey  = "signinAssessment"
	signinRiskSessionKey = "signinRisk"
	riskMfaSessionKey    = "riskMfa"
	riskDeniedCode       = "risk-denied"
)

// assessSigninRisk scores the sign-in once per request under the adaptive risk and the device trust of the user's
// organization, and runs its pre-sign-in flows. It returns true when it has refused the sign-in
func (c *ApiController) assessSigninRisk(user *object.User) (*ee.SigninAssessment, bool) {
	if value, ok := c.Ctx.Input.GetData(signinAssessmentKey).(*ee.SigninAssessment); ok {
		return value, false
	}
	if user == nil {
		return nil, false
	}

	clientIp := util.GetClientIpFromRequest(c.Ctx.Request)
	applicationName, _ := c.Ctx.Input.GetData("signinApplication").(string)

	device := object.CheckSigninDevice(user, applicationName, c.Ctx.Request)
	if device.Denied != "" {
		c.ResponseErrorWithCode(riskDeniedCode, c.T(device.Denied))
		return nil, true
	}
	c.Ctx.Input.SetData("deviceSatisfiesMfa", device.SatisfiesMfa)

	assessment := ee.AssessSignin(&ee.RiskInput{
		User:            user,
		Application:     applicationName,
		Ip:              clientIp,
		DeviceId:        c.Ctx.Request.Header.Get("X-Casdoor-Device"),
		UserAgent:       c.Ctx.Request.UserAgent(),
		Header:          c.Ctx.Request.Header,
		TrustedDevice:   device.Trusted,
		UntrustedDevice: device.Untrusted,
	})

	// the pre-sign-in flows see the risk, and can refuse the sign-in or ask for MFA on top of it
	flowResult := object.RunSigninFlows(user, applicationName, clientIp, c.Ctx.Request, object.FlowTriggerPreSignin, assessment.ToRisk())
	if flowResult != nil && flowResult.Deny != "" {
		ee.MarkSigninRiskByFlow(user.Owner, assessment, ee.RiskDecisionDeny)
		c.ResponseErrorWithCode(riskDeniedCode, flowResult.Deny)
		return nil, true
	}
	if flowResult != nil && flowResult.RequireMfa && (assessment == nil || assessment.Decision == ee.RiskDecisionAllow) {
		if assessment == nil {
			assessment = &ee.SigninAssessment{Signals: []string{}}
		}
		assessment.Decision = ee.RiskDecisionMfa
		ee.MarkSigninRiskByFlow(user.Owner, assessment, ee.RiskDecisionMfa)
	}
	// a trusted device is the possession factor when the policy counts it, except for what a flow asks for
	if assessment != nil && assessment.Decision == ee.RiskDecisionMfa && device.SatisfiesMfa && (flowResult == nil || !flowResult.RequireMfa) {
		assessment.Decision = ee.RiskDecisionAllow
	}
	c.Ctx.Input.SetData(signinAssessmentKey, assessment)

	if assessment == nil {
		return nil, false
	}
	c.SetSession(signinRiskSessionKey, fmt.Sprintf("%s|%s|%d", user.GetId(), assessment.EventName, time.Now().Unix()))
	if assessment.Decision == ee.RiskDecisionDeny {
		c.ResponseErrorWithCode(riskDeniedCode, c.T("auth:This sign-in looks risky and has been blocked, please contact your administrator"))
		return assessment, true
	}
	return assessment, false
}

// popSigninRiskSession returns the risk event of the sign-in that the earlier request of this sign-in assessed, such
// as the one before its MFA
func (c *ApiController) popSigninRiskSession(user *object.User) (string, bool) {
	value, _ := c.GetSession(signinRiskSessionKey).(string)
	if value == "" {
		return "", false
	}
	c.SetSession(signinRiskSessionKey, "")
	c.SetSession(riskMfaSessionKey, "")
	parts := strings.Split(value, "|")
	if len(parts) != 3 || parts[0] != user.GetId() {
		return "", false
	}
	unix, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil || time.Since(time.Unix(unix, 0)) > 15*time.Minute {
		return "", false
	}
	return parts[1], true
}

// checkSigninRiskOnLoggedIn completes the risk event of a sign-in, or assesses a sign-in method that doesn't pass
// through the MFA check (a magic link, a certificate, a device authorization), where only a refusal applies.
// It returns true when it has refused the sign-in
func (c *ApiController) checkSigninRiskOnLoggedIn(application *object.Application, user *object.User) bool {
	if c.Ctx.Input.Param("sessionSignin") == "true" {
		return false
	}

	assessment, ok := c.Ctx.Input.GetData(signinAssessmentKey).(*ee.SigninAssessment)
	if !ok {
		if eventName, found := c.popSigninRiskSession(user); found {
			ee.CompleteSigninRiskEvent(user.Owner, eventName, application.Name)
			return false
		}
		c.Ctx.Input.SetData("signinApplication", application.Name)
		var denied bool
		assessment, denied = c.assessSigninRisk(user)
		if denied {
			return true
		}
	}
	c.SetSession(signinRiskSessionKey, "")
	c.SetSession(riskMfaSessionKey, "")
	if assessment != nil {
		ee.CompleteSigninRiskEvent(user.Owner, assessment.EventName, application.Name)
	}
	return false
}

// getRiskMfaProps offers the email and the phone of the user for the MFA that a risky sign-in asks for, when the
// user hasn't set up the MFA that the method of the sign-in doesn't already prove
func getRiskMfaProps(user *object.User, verificationType string) []*object.MfaProps {
	props := []*object.MfaProps{}
	if user.Email != "" && verificationType != object.EmailType {
		props = append(props, &object.MfaProps{Enabled: true, MfaType: object.EmailType, Secret: util.GetMaskedEmail(user.Email)})
	}
	if user.Phone != "" && verificationType != object.SmsType {
		props = append(props, &object.MfaProps{Enabled: true, MfaType: object.SmsType, Secret: util.GetMaskedPhone(user.Phone), CountryCode: user.CountryCode})
	}
	return props
}

// getMfaAuthProps returns the MFA of the user that a sign-in verifies, including the email and the phone of a risky
// sign-in's MFA
func (c *ApiController) getMfaAuthProps(user *object.User, mfaType string) *object.MfaProps {
	props := user.GetMfaProps(mfaType, false)
	if props.Enabled {
		return props
	}
	if riskMfa, _ := c.GetSession(riskMfaSessionKey).(string); riskMfa != user.GetId() {
		return props
	}
	if mfaType == object.EmailType && user.Email != "" {
		return &object.MfaProps{Enabled: true, MfaType: mfaType, Secret: user.Email}
	}
	if mfaType == object.SmsType && user.Phone != "" {
		return &object.MfaProps{Enabled: true, MfaType: mfaType, Secret: user.Phone, CountryCode: user.CountryCode}
	}
	return props
}
