// Copyright 2023 The Casdoor Authors. All Rights Reserved.
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
	"errors"
	"fmt"
	"github.com/casdoor/casdoor/util"

	"github.com/beego/beego/context"
	"github.com/google/uuid"
)

const (
	MfaSmsCountryCodeSession   = "mfa_country_code"
	MfaSmsDestSession          = "mfa_dest"
	MfaSmsRecoveryCodesSession = "mfa_recovery_codes"
)

type SmsMfa struct {
	Config *MfaProps
}

func (mfa *SmsMfa) SetupVerify(ctx *context.Context, passCode string) error {
	dest := ctx.Input.CruSession.Get(MfaSmsDestSession).(string)
	countryCode := ctx.Input.CruSession.Get(MfaSmsCountryCodeSession).(string)
	if countryCode != "" {
		dest, _ = util.GetE164Number(dest, countryCode)
	}

	if result := CheckVerificationCode(dest, passCode, "en"); result.Code != VerificationSuccess {
		return errors.New(result.Msg)
	}
	return nil
}

func (mfa *SmsMfa) Verify(passCode string) error {
	if !util.IsEmailValid(mfa.Config.Secret) {
		mfa.Config.Secret, _ = util.GetE164Number(mfa.Config.Secret, mfa.Config.CountryCode)
	}
	if result := CheckVerificationCode(mfa.Config.Secret, passCode, "en"); result.Code != VerificationSuccess {
		return errors.New(result.Msg)
	}
	return nil
}

func (mfa *SmsMfa) Initiate(ctx *context.Context, name string, secret string) (*MfaProps, error) {
	recoveryCode, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	err = ctx.Input.CruSession.Set(MfaSmsRecoveryCodesSession, []string{recoveryCode.String()})
	if err != nil {
		return nil, err
	}

	mfaProps := MfaProps{
		AuthType:      SmsType,
		RecoveryCodes: []string{recoveryCode.String()},
	}
	return &mfaProps, nil
}

func (mfa *SmsMfa) Enable(ctx *context.Context, user *User) error {
	dest := ctx.Input.CruSession.Get(MfaSmsDestSession).(string)
	recoveryCodes := ctx.Input.CruSession.Get(MfaSmsRecoveryCodesSession).([]string)
	countryCode := ctx.Input.CruSession.Get(MfaSmsCountryCodeSession).(string)

	if dest == "" || len(recoveryCodes) == 0 {
		return fmt.Errorf("MFA dest or recovery codes is empty")
	}

	if !util.IsEmailValid(dest) {
		mfa.Config.CountryCode = countryCode
	}

	mfa.Config.AuthType = SmsType
	mfa.Config.Id = uuid.NewString()
	mfa.Config.Secret = dest
	mfa.Config.RecoveryCodes = recoveryCodes

	for i, mfaProp := range user.MultiFactorAuths {
		if mfaProp.Secret == mfa.Config.Secret {
			user.MultiFactorAuths = append(user.MultiFactorAuths[:i], user.MultiFactorAuths[i+1:]...)
		}
	}
	user.MultiFactorAuths = append(user.MultiFactorAuths, mfa.Config)

	affected := UpdateUser(user.GetId(), user, []string{"multi_factor_auths"}, user.IsAdminUser())
	if !affected {
		return fmt.Errorf("failed to enable two factor authentication")
	}

	return nil
}

func NewSmsTwoFactor(config *MfaProps) *SmsMfa {
	if config == nil {
		config = &MfaProps{
			AuthType: SmsType,
		}
	}
	return &SmsMfa{
		Config: config,
	}
}
