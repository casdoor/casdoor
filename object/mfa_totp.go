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
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

const (
	MfaTotpPeriodInSeconds = 30
)

type TotpMfa struct {
	*MfaProps
	period     uint
	secretSize uint
	digits     otp.Digits
}

func (mfa *TotpMfa) Initiate(userId string) (*MfaProps, error) {
	//issuer := beego.AppConfig.String("appname")
	//if issuer == "" {
	//	issuer = "casdoor"
	//}
	issuer := "Casdoor"

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

	mfaProps := MfaProps{
		MfaType: mfa.MfaType,
		Secret:  key.Secret(),
		URL:     key.URL(),
	}
	return &mfaProps, nil
}

func (mfa *TotpMfa) SetupVerify(passcode string) error {
	result, err := totp.ValidateCustom(passcode, mfa.Secret, time.Now().UTC(), totp.ValidateOpts{
		Period:    MfaTotpPeriodInSeconds,
		Skew:      1,
		Digits:    otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA1,
	})
	if err != nil {
		return err
	}

	if result {
		return nil
	} else {
		return errors.New("totp passcode error")
	}
}

func (mfa *TotpMfa) Enable(user *User) error {
	columns := []string{"recovery_codes", "preferred_mfa_type", "totp_secret"}

	user.RecoveryCodes = append(user.RecoveryCodes, mfa.RecoveryCodes...)
	user.TotpSecret = mfa.Secret
	if user.PreferredMfaType == "" {
		user.PreferredMfaType = mfa.MfaType
	}

	_, err := updateUser(user.GetId(), user, columns)
	if err != nil {
		return err
	}

	return nil
}

func (mfa *TotpMfa) Verify(passcode string) error {
	result, err := totp.ValidateCustom(passcode, mfa.Secret, time.Now().UTC(), totp.ValidateOpts{
		Period:    MfaTotpPeriodInSeconds,
		Skew:      1,
		Digits:    otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA1,
	})
	if err != nil {
		return err
	}

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
		MfaProps:   config,
		period:     MfaTotpPeriodInSeconds,
		secretSize: 20,
		digits:     otp.DigitsSix,
	}
}
