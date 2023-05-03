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
	"strings"

	"github.com/casdoor/casdoor/util"

	"github.com/beego/beego/context"
	"github.com/google/uuid"
)

const (
	MfaSmsDestSession          = "mfa_dest"
	MfaSmsRecoveryCodesSession = "mfa_recovery_codes"
)

type SmsMfa struct {
	Config *MfaProps
}

func (mfa *SmsMfa) SetupVerify(ctx *context.Context, passCode string) error {
	dest := ctx.Input.CruSession.Get(MfaSmsDestSession).(string)
	if result := CheckVerificationCode(dest, passCode, "en"); result.Code != VerificationSuccess {
		return errors.New(result.Msg)
	}
	return nil
}

func (mfa *SmsMfa) Verify(passCode string) error {
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

	twoFactorProps := MfaProps{
		RecoveryCodes: []string{recoveryCode.String()},
	}
	return &twoFactorProps, nil
}

func (mfa *SmsMfa) Enable(ctx *context.Context, user *User) error {
	dest := ctx.Input.CruSession.Get(MfaSmsDestSession)
	recoveryCodes := ctx.Input.CruSession.Get(MfaSmsRecoveryCodesSession)

	if dest == nil || recoveryCodes == nil {
		return fmt.Errorf("mfa authentication dest or recovery codes is nil")
	}

	if strings.HasPrefix(dest.(string), "+") {
		countryCode, err := util.GetCountryCodeFromE164Number(dest.(string))
		if err != nil {
			return err
		}
		mfa.Config.CountryCode = countryCode
	}

	mfa.Config.AuthType = SmsType
	mfa.Config.Id = uuid.NewString()
	mfa.Config.Secret = dest.(string)
	mfa.Config.RecoveryCodes = recoveryCodes.([]string)

	for i, twoFactorProp := range user.TwoFactorAuth {
		if twoFactorProp.AuthType == SmsType {
			user.TwoFactorAuth = append(user.TwoFactorAuth[:i], user.TwoFactorAuth[i+1:]...)
		}
	}
	user.TwoFactorAuth = append(user.TwoFactorAuth, mfa.Config)

	affected := UpdateUser(user.GetId(), user, []string{"two_factor_auth"}, user.IsAdminUser())
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
