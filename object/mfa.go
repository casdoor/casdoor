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

type TwoFactorSessionData struct {
	UserId        string
	EnableSession bool
	AutoSignIn    bool
}

type MfaProps struct {
	Id            string   `json:"id,omitempty"`
	IsPreferred   bool     `json:"isPreferred,omitempty"`
	AuthType      string   `json:"type,omitempty" form:"type"`
	Secret        string   `json:"secret,omitempty"`
	CountryCode   string   `json:"countryCode,omitempty"`
	URL           string   `json:"url,omitempty"`
	RecoveryCodes []string `json:"recoveryCodes,omitempty"`
}

type TwoFactorInterface interface {
	SetupVerify(ctx *context.Context, passCode string) error
	Verify(passCode string) error
	Initiate(ctx *context.Context, name1 string, name2 string) (*MfaProps, error)
	Enable(ctx *context.Context, user *User) error
}

const (
	SmsType  = "sms"
	TotpType = "app"
)

const (
	TwoFactorSessionUserId        = "TwoFactorSessionUserId"
	TwoFactorSessionEnableSession = "TwoFactorSessionEnableSession"
	TwoFactorSessionAutoSignIn    = "TwoFactorSessionAutoSignIn"
	NextTwoFactor                 = "nextTwoFactor"
)

func GetTwoFactorUtil(providerType string, config *MfaProps) TwoFactorInterface {
	switch providerType {
	case SmsType:
		return NewSmsTwoFactor(config)
	case TotpType:
		return nil
	}

	return nil
}

func RecoverTfs(user *User, recoveryCode string, authType string) (bool, error) {
	hit := false
	twoFactor := &MfaProps{}

	for _, twoFactorProp := range user.TwoFactorAuth {
		if twoFactorProp.AuthType == authType {
			twoFactor = twoFactorProp
		}
	}
	if len(twoFactor.RecoveryCodes) == 0 {
		return false, fmt.Errorf("")
	}

	for i, code := range twoFactor.RecoveryCodes {
		if code == recoveryCode {
			twoFactor.RecoveryCodes[i] = ""
			hit = true
			break
		}
	}
	if !hit {
		return false, fmt.Errorf("")
	}
	affected := UpdateUser(user.GetId(), user, []string{"two_factor_auth"}, user.IsAdminUser())
	if !affected {
		return false, fmt.Errorf("")
	}
	return true, nil
}

func GetMaskedProps(props *MfaProps) *MfaProps {
	maskedProps := &MfaProps{
		AuthType:    SmsType,
		Id:          props.Id,
		IsPreferred: props.IsPreferred,
	}

	if props.AuthType == SmsType {
		if props.CountryCode != "" {
			maskedProps.Secret = util.GetMaskedPhone(props.Secret)
		} else {
			maskedProps.Secret = util.GetMaskedEmail(props.Secret)
		}
	}
	return maskedProps
}
