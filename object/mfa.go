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
	"fmt"

	"github.com/casdoor/casdoor/util"

	"github.com/beego/beego/context"
)

const MfaRecoveryCodesSession = "mfa_recovery_codes"

type MfaSessionData struct {
	UserId string
}

type MfaProps struct {
	Enabled       bool     `json:"enabled"`
	IsPreferred   bool     `json:"isPreferred"`
	MfaType       string   `json:"mfaType" form:"mfaType"`
	Secret        string   `json:"secret,omitempty"`
	CountryCode   string   `json:"countryCode,omitempty"`
	URL           string   `json:"url,omitempty"`
	RecoveryCodes []string `json:"recoveryCodes,omitempty"`
}

type MfaInterface interface {
	Initiate(ctx *context.Context, userId string) (*MfaProps, error)
	SetupVerify(ctx *context.Context, passcode string) error
	Enable(ctx *context.Context, user *User) error
	Verify(passcode string) error
}

const (
	EmailType = "email"
	SmsType   = "sms"
	TotpType  = "app"
)

const (
	MfaSessionUserId = "MfaSessionUserId"
	NextMfa          = "NextMfa"
	RequiredMfa      = "RequiredMfa"
)

func GetMfaUtil(mfaType string, config *MfaProps) MfaInterface {
	switch mfaType {
	case SmsType:
		return NewSmsMfaUtil(config)
	case EmailType:
		return NewEmailMfaUtil(config)
	case TotpType:
		return NewTotpMfaUtil(config)
	}

	return nil
}

func MfaRecover(user *User, recoveryCode string) error {
	hit := false

	if len(user.RecoveryCodes) == 0 {
		return fmt.Errorf("do not have recovery codes")
	}

	for _, code := range user.RecoveryCodes {
		if code == recoveryCode {
			hit = true
			user.RecoveryCodes = util.DeleteVal(user.RecoveryCodes, code)
			break
		}
	}
	if !hit {
		return fmt.Errorf("recovery code not found")
	}

	_, err := UpdateUser(user.GetId(), user, []string{"recovery_codes"}, user.IsAdminUser())
	if err != nil {
		return err
	}

	return nil
}

func GetAllMfaProps(user *User, masked bool) []*MfaProps {
	mfaProps := []*MfaProps{}

	for _, mfaType := range []string{SmsType, EmailType, TotpType} {
		mfaProps = append(mfaProps, user.GetMfaProps(mfaType, masked))
	}
	return mfaProps
}

func (user *User) GetMfaProps(mfaType string, masked bool) *MfaProps {
	mfaProps := &MfaProps{}

	if mfaType == SmsType {
		if !user.MfaPhoneEnabled {
			return &MfaProps{
				Enabled: false,
				MfaType: mfaType,
			}
		}

		mfaProps = &MfaProps{
			Enabled:     user.MfaPhoneEnabled,
			MfaType:     mfaType,
			CountryCode: user.CountryCode,
		}
		if masked {
			mfaProps.Secret = util.GetMaskedPhone(user.Phone)
		} else {
			mfaProps.Secret = user.Phone
		}
	} else if mfaType == EmailType {
		if !user.MfaEmailEnabled {
			return &MfaProps{
				Enabled: false,
				MfaType: mfaType,
			}
		}

		mfaProps = &MfaProps{
			Enabled: user.MfaEmailEnabled,
			MfaType: mfaType,
		}
		if masked {
			mfaProps.Secret = util.GetMaskedEmail(user.Email)
		} else {
			mfaProps.Secret = user.Email
		}
	} else if mfaType == TotpType {
		if user.TotpSecret == "" {
			return &MfaProps{
				Enabled: false,
				MfaType: mfaType,
			}
		}

		mfaProps = &MfaProps{
			Enabled: true,
			MfaType: mfaType,
		}
		if masked {
			mfaProps.Secret = ""
		} else {
			mfaProps.Secret = user.TotpSecret
		}
	}

	if user.PreferredMfaType == mfaType {
		mfaProps.IsPreferred = true
	}
	return mfaProps
}

func DisabledMultiFactorAuth(user *User) error {
	user.PreferredMfaType = ""
	user.RecoveryCodes = []string{}
	user.MfaPhoneEnabled = false
	user.MfaEmailEnabled = false
	user.TotpSecret = ""

	_, err := updateUser(user.GetId(), user, []string{"preferred_mfa_type", "recovery_codes", "mfa_phone_enabled", "mfa_email_enabled", "totp_secret"})
	if err != nil {
		return err
	}
	return nil
}

func SetPreferredMultiFactorAuth(user *User, mfaType string) error {
	user.PreferredMfaType = mfaType

	_, err := UpdateUser(user.GetId(), user, []string{"preferred_mfa_type"}, user.IsAdminUser())
	if err != nil {
		return err
	}
	return nil
}
