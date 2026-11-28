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

package object

import (
	"crypto/subtle"
	"errors"
	"fmt"
	"time"

	"github.com/casdoor/casdoor/i18n"
	"github.com/casdoor/casdoor/util"
	"github.com/golang-jwt/jwt/v5"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

// RecoveryCodeType is the step-up method of a one-time recovery code
const RecoveryCodeType = "recovery"

const (
	StepUpTokenType               = "step-up"
	defaultStepUpTokenExpireInSec = 300
	maxStepUpTokenExpireInSec     = 3600
	stepUpIpErrorDest             = "mfa-step-up"
)

var StepUpMfaTypes = []string{TotpType, EmailType, SmsType, RadiusType, RecoveryCodeType}

// MfaTotpStep keeps the last accepted TOTP time step of a user, so a passcode is accepted only once
type MfaTotpStep struct {
	Owner string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name  string `xorm:"varchar(100) notnull pk" json:"name"`
	Step  int64  `xorm:"notnull" json:"step"`
}

// MfaStepUp is a passed step-up check, its name is the jti of the issued proof token
type MfaStepUp struct {
	Owner        string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name         string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime  string `xorm:"varchar(100)" json:"createdTime"`
	User         string `xorm:"varchar(100) index" json:"user"`
	Application  string `xorm:"varchar(100)" json:"application"`
	MfaType      string `xorm:"varchar(100)" json:"mfaType"`
	Action       string `xorm:"varchar(500)" json:"action"`
	ExpireTime   int64  `json:"expireTime"`
	RedeemedTime string `xorm:"varchar(100)" json:"redeemedTime"`
}

type StepUpClaims struct {
	TokenType string   `json:"tokenType"`
	Owner     string   `json:"owner"`
	Name      string   `json:"name"`
	Amr       []string `json:"amr"`
	AuthTime  int64    `json:"auth_time"`
	MfaType   string   `json:"mfa_type"`
	Action    string   `json:"action,omitempty"`
	jwt.RegisteredClaims
}

type StepUpMethod struct {
	MfaType     string `json:"mfaType"`
	Dest        string `json:"dest,omitempty"`
	CountryCode string `json:"countryCode,omitempty"`
	IsPreferred bool   `json:"isPreferred"`
}

func (application *Application) GetStepUpMfaTypes() []string {
	if len(application.StepUpMfaTypes) == 0 {
		return StepUpMfaTypes
	}

	res := []string{}
	for _, mfaType := range application.StepUpMfaTypes {
		if util.InSlice(StepUpMfaTypes, mfaType) {
			res = append(res, mfaType)
		}
	}
	return res
}

func (application *Application) GetStepUpExpireInSeconds() int {
	expireInSeconds := application.StepUpExpireInSeconds
	if expireInSeconds <= 0 {
		return defaultStepUpTokenExpireInSec
	}
	if expireInSeconds > maxStepUpTokenExpireInSec {
		return maxStepUpTokenExpireInSec
	}
	return expireInSeconds
}

// GetStepUpMethods returns the factors the user has enrolled that the application accepts for step-up,
// the preferred one first, with the destinations masked
func GetStepUpMethods(user *User, application *Application) []*StepUpMethod {
	methods := []*StepUpMethod{}
	for _, mfaType := range application.GetStepUpMfaTypes() {
		if mfaType == RecoveryCodeType {
			if len(user.RecoveryCodes) > 0 {
				methods = append(methods, &StepUpMethod{MfaType: mfaType})
			}
			continue
		}

		mfaProps := user.GetMfaProps(mfaType, true)
		if !mfaProps.Enabled {
			continue
		}
		method := &StepUpMethod{MfaType: mfaType, IsPreferred: mfaProps.IsPreferred}
		if mfaType == EmailType || mfaType == SmsType {
			method.Dest = mfaProps.Secret
			method.CountryCode = mfaProps.CountryCode
		}
		if method.IsPreferred {
			methods = append([]*StepUpMethod{method}, methods...)
		} else {
			methods = append(methods, method)
		}
	}
	return methods
}

// GetStepUpMfaProps returns the enrolled factor of the user for a step-up check, never a factor taken from the request
func GetStepUpMfaProps(user *User, application *Application, mfaType string, lang string) (*MfaProps, error) {
	if !util.InSlice(application.GetStepUpMfaTypes(), mfaType) {
		return nil, fmt.Errorf(i18n.Translate(lang, "mfa:The MFA type: %s is not allowed for step-up verification by the application"), mfaType)
	}

	if mfaType == RecoveryCodeType {
		if len(user.RecoveryCodes) == 0 {
			return nil, errors.New(i18n.Translate(lang, "mfa:The user has no recovery codes"))
		}
		return &MfaProps{Enabled: true, MfaType: RecoveryCodeType}, nil
	}

	mfaProps := user.GetMfaProps(mfaType, false)
	if !mfaProps.Enabled {
		return nil, fmt.Errorf(i18n.Translate(lang, "mfa:The user has not enabled the MFA type: %s"), mfaType)
	}
	return mfaProps, nil
}

// VerifyMfaPasscode checks a passcode against an enrolled factor, a TOTP passcode is accepted only once
func VerifyMfaPasscode(user *User, mfaProps *MfaProps, passcode string, lang string) error {
	if mfaProps.MfaType == TotpType {
		return verifyTotpPasscodeOnce(user, mfaProps.Secret, passcode, lang)
	}

	mfaProps.User = user
	mfaUtil := GetMfaUtil(mfaProps.MfaType, mfaProps)
	if mfaUtil == nil {
		return errors.New("Invalid multi-factor authentication type")
	}
	return mfaUtil.Verify(passcode, lang)
}

func verifyTotpPasscodeOnce(user *User, totpSecret string, passcode string, lang string) error {
	if totpSecret == "" {
		return errors.New("totp is not enabled")
	}

	secret, period, digits := parseTotpSecret(totpSecret)
	now := time.Now().UTC()
	currentStep := now.Unix() / int64(period)
	matchedStep := int64(-1)
	for _, step := range []int64{currentStep - 1, currentStep, currentStep + 1} {
		code, err := totp.GenerateCodeCustom(secret, time.Unix(step*int64(period), 0).UTC(), totp.ValidateOpts{
			Period:    period,
			Digits:    digits,
			Algorithm: otp.AlgorithmSHA1,
		})
		if err != nil {
			return err
		}
		if subtle.ConstantTimeCompare([]byte(code), []byte(passcode)) == 1 {
			matchedStep = step
			break
		}
	}
	if matchedStep < 0 {
		return errors.New("totp passcode error")
	}

	return useTotpStep(user, matchedStep, lang)
}

func useTotpStep(user *User, step int64, lang string) error {
	usedErr := errors.New(i18n.Translate(lang, "mfa:This code has already been used, please wait for the next one"))

	affected, err := ormer.Engine.Where("owner = ? and name = ? and step < ?", user.Owner, user.Name, step).Cols("step").Update(&MfaTotpStep{Step: step})
	if err != nil {
		return err
	}
	if affected == 1 {
		return nil
	}

	existing := MfaTotpStep{Owner: user.Owner, Name: user.Name}
	has, err := ormer.Engine.Get(&existing)
	if err != nil {
		return err
	}
	if has {
		return usedErr
	}

	_, err = ormer.Engine.Insert(&MfaTotpStep{Owner: user.Owner, Name: user.Name, Step: step})
	if err != nil {
		// another request inserted the row for the same passcode first
		return usedErr
	}
	return nil
}

// VerifyStepUpWithLimit counts a failure against the user, like the MFA step of a sign-in, and against
// the client IP when one is given, so neither a user nor an IP can guess codes without limit
func VerifyStepUpWithLimit(user *User, remoteAddr string, verify func() error, lang string) error {
	if err := checkVerifyCodeIpErrorTimes(remoteAddr, stepUpIpErrorDest, lang); err != nil {
		return err
	}

	err := VerifyMfaWithLimit(user, verify, lang)
	if err != nil {
		if remoteAddr != "" {
			_ = recordVerifyCodeIpErrorInfo(remoteAddr, stepUpIpErrorDest, lang)
		}
		return err
	}

	resetVerifyCodeIpErrorTimes(remoteAddr, stepUpIpErrorDest)
	return nil
}

func getStepUpAmr(mfaType string) []string {
	switch mfaType {
	case TotpType, RadiusType:
		return []string{"mfa", "otp"}
	case SmsType:
		return []string{"mfa", "sms", "otp"}
	case EmailType:
		return []string{"mfa", "email", "otp"}
	default:
		return []string{"mfa", RecoveryCodeType}
	}
}

// IssueStepUpToken signs a short-lived, single-use proof of a passed step-up check for the application
func IssueStepUpToken(host string, application *Application, user *User, mfaType string, action string) (string, int, error) {
	issuer, err := GetOidcIssuerByApplication(host, application)
	if err != nil {
		return "", 0, err
	}

	signingApplication := getSigningApplication(application, user.Owner)
	cert, err := GetCertByApplication(signingApplication)
	if err != nil {
		return "", 0, err
	}
	if cert == nil {
		return "", 0, fmt.Errorf("no cert found for application: %s", application.GetId())
	}

	now := time.Now()
	expireInSeconds := application.GetStepUpExpireInSeconds()
	expireTime := now.Add(time.Duration(expireInSeconds) * time.Second)
	jti := util.GenerateId()
	claims := StepUpClaims{
		TokenType: StepUpTokenType,
		Owner:     user.Owner,
		Name:      user.Name,
		Amr:       getStepUpAmr(mfaType),
		AuthTime:  now.Unix(),
		MfaType:   mfaType,
		Action:    action,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issuer,
			Subject:   user.Id,
			Audience:  []string{application.ClientId},
			ExpiresAt: jwt.NewNumericDate(expireTime),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        jti,
		},
	}

	token := jwt.NewWithClaims(GetJwtSigningMethod(signingApplication), claims)
	key, err := getJwtSigningKey(token.Method, cert)
	if err != nil {
		return "", 0, err
	}
	token.Header["typ"] = "step-up+jwt"
	token.Header["kid"] = cert.GetKeyId()
	tokenString, err := signJwtToken(token, key)
	if err != nil {
		return "", 0, err
	}

	stepUp := &MfaStepUp{
		Owner:       user.Owner,
		Name:        jti,
		CreatedTime: util.GetCurrentTime(),
		User:        user.Name,
		Application: application.GetId(),
		MfaType:     mfaType,
		Action:      action,
		ExpireTime:  expireTime.Unix(),
	}
	_, err = ormer.Engine.Insert(stepUp)
	if err != nil {
		return "", 0, err
	}

	return tokenString, expireInSeconds, nil
}

// RedeemStepUpToken checks a proof token for the application and marks it used, so it is accepted only once
func RedeemStepUpToken(application *Application, tokenString string, action string, lang string) (*StepUpClaims, error) {
	cert, err := getCertByApplicationOfToken(application, tokenString)
	if err != nil {
		return nil, err
	}
	if cert == nil || cert.Certificate == "" {
		return nil, fmt.Errorf("no cert found for application: %s", application.GetId())
	}

	claims := &StepUpClaims{}
	_, err = jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		kid, _ := token.Header["kid"].(string)
		return getJwtVerifyKey(token, cert.getCertificateByKeyId(kid))
	}, jwt.WithAudience(application.ClientId), jwt.WithExpirationRequired())
	if err != nil {
		return nil, fmt.Errorf(i18n.Translate(lang, "mfa:The step-up token is invalid: %s"), err.Error())
	}
	if claims.TokenType != StepUpTokenType || claims.ID == "" {
		return nil, errors.New(i18n.Translate(lang, "mfa:The token is not a step-up token"))
	}
	if action != "" && claims.Action != action {
		return nil, errors.New(i18n.Translate(lang, "mfa:The step-up token was issued for another action"))
	}

	affected, err := ormer.Engine.Where("owner = ? and name = ? and application = ? and (redeemed_time = '' or redeemed_time is null)", claims.Owner, claims.ID, application.GetId()).
		Cols("redeemed_time").Update(&MfaStepUp{RedeemedTime: util.GetCurrentTime()})
	if err != nil {
		return nil, err
	}
	if affected != 1 {
		return nil, errors.New(i18n.Translate(lang, "mfa:The step-up token has already been used"))
	}

	return claims, nil
}

func GetStepUpRecords(application *Application, limit int) ([]*MfaStepUp, error) {
	stepUps := []*MfaStepUp{}
	err := ormer.Engine.Where("application = ?", application.GetId()).Desc("created_time").Limit(limit).Find(&stepUps)
	return stepUps, err
}
