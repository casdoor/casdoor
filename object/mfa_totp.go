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

	"github.com/beego/beego"
	"github.com/beego/beego/context"
	"github.com/google/uuid"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

const MfaTotpSecretSession = "mfa_totp_secret"

type TotpMfa struct {
	Config     *MfaProps
	period     uint
	secretSize uint
	digits     otp.Digits
}

func (mfa *TotpMfa) Initiate(ctx *context.Context, userId string) (*MfaProps, error) {
	issuer := beego.AppConfig.String("appname")
	if issuer == "" {
		issuer = "casdoor"
	}

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      issuer,
		AccountName: userId,
		Period:      mfa.period,
		SecretSize:  mfa.secretSize,
		Digits:      mfa.digits,
	})
	if err != nil {
		return nil, err
	}

	err = ctx.Input.CruSession.Set(MfaTotpSecretSession, key.Secret())
	if err != nil {
		return nil, err
	}

	recoveryCode := uuid.NewString()
	err = ctx.Input.CruSession.Set(MfaRecoveryCodesSession, []string{recoveryCode})
	if err != nil {
		return nil, err
	}

	mfaProps := MfaProps{
		MfaType:       mfa.Config.MfaType,
		RecoveryCodes: []string{recoveryCode},
		Secret:        key.Secret(),
		URL:           key.URL(),
	}
	return &mfaProps, nil
}

func (mfa *TotpMfa) SetupVerify(ctx *context.Context, passcode string) error {
	secret := ctx.Input.CruSession.Get(MfaTotpSecretSession).(string)
	result := totp.Validate(passcode, secret)

	if result {
		return nil
	} else {
		return errors.New("totp passcode error")
	}
}

func (mfa *TotpMfa) Enable(ctx *context.Context, user *User) error {
	recoveryCodes := ctx.Input.CruSession.Get(MfaRecoveryCodesSession).([]string)
	if len(recoveryCodes) == 0 {
		return fmt.Errorf("recovery codes is missing")
	}
	secret := ctx.Input.CruSession.Get(MfaTotpSecretSession).(string)
	if secret == "" {
		return fmt.Errorf("totp secret is missing")
	}

	columns := []string{"recovery_codes", "preferred_mfa_type", "totp_secret"}

	user.RecoveryCodes = append(user.RecoveryCodes, recoveryCodes...)
	user.TotpSecret = secret
	if user.PreferredMfaType == "" {
		user.PreferredMfaType = mfa.Config.MfaType
	}

	_, err := updateUser(user.GetId(), user, columns)
	if err != nil {
		return err
	}
	return nil
}

func (mfa *TotpMfa) Verify(passcode string) error {
	result := totp.Validate(passcode, mfa.Config.Secret)

	if result {
		return nil
	} else {
		return errors.New("totp passcode error")
	}
}

func NewTotpMfaUtil(config *MfaProps) *TotpMfa {
	if config == nil {
		config = &MfaProps{
			MfaType: TotpType,
		}
	}

	return &TotpMfa{
		Config:     config,
		period:     30,
		secretSize: 20,
		digits:     otp.DigitsSix,
	}
}
