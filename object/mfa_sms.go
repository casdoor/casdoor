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

	"github.com/casdoor/casdoor/util"
)

type SmsMfa struct {
	*MfaProps
}

func (mfa *SmsMfa) Initiate(userId string) (*MfaProps, error) {
	mfaProps := MfaProps{
		MfaType: mfa.MfaType,
	}
	return &mfaProps, nil
}

func (mfa *SmsMfa) SetupVerify(passCode string) error {
	if !util.IsEmailValid(mfa.Secret) {
		mfa.Secret, _ = util.GetE164Number(mfa.Secret, mfa.CountryCode)
	}

	result, err := CheckVerificationCode(mfa.Secret, passCode, "en")
	if err != nil {
		return err
	}
	if result.Code != VerificationSuccess {
		return errors.New(result.Msg)
	}

	return nil
}

func (mfa *SmsMfa) Enable(user *User) error {
	columns := []string{"recovery_codes", "preferred_mfa_type"}

	user.RecoveryCodes = append(user.RecoveryCodes, mfa.RecoveryCodes...)
	if user.PreferredMfaType == "" {
		user.PreferredMfaType = mfa.MfaType
	}

	if mfa.MfaType == SmsType {
		user.MfaPhoneEnabled = true
		columns = append(columns, "mfa_phone_enabled", "phone", "country_code")
	} else if mfa.MfaType == EmailType {
		user.MfaEmailEnabled = true
		user.EmailVerified = true
		columns = append(columns, "mfa_email_enabled", "email", "email_verified")
	}

	_, err := UpdateUser(user.GetId(), user, columns, false)
	if err != nil {
		return err
	}

	return nil
}

func (mfa *SmsMfa) Verify(passCode string) error {
	if !util.IsEmailValid(mfa.Secret) {
		mfa.Secret, _ = util.GetE164Number(mfa.Secret, mfa.CountryCode)
	}

	result, err := CheckVerificationCode(mfa.Secret, passCode, "en")
	if err != nil {
		return err
	}
	if result.Code != VerificationSuccess {
		return errors.New(result.Msg)
	}

	return nil
}

func NewSmsMfaUtil(config *MfaProps) *SmsMfa {
	if config == nil {
		config = &MfaProps{
			MfaType: SmsType,
		}
	}
	return &SmsMfa{
		config,
	}
}

func NewEmailMfaUtil(config *MfaProps) *SmsMfa {
	if config == nil {
		config = &MfaProps{
			MfaType: EmailType,
		}
	}
	return &SmsMfa{
		config,
	}
}
