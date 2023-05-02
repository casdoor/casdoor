package object

import (
	"fmt"

	"github.com/beego/beego/context"
)

type TwoFactorSessionData struct {
	UserId        string
	EnableSession bool
	AutoSignIn    bool
}

type TfaProps struct {
	Id            string   `json:"id,omitempty"`
	IsPreferred   bool     `json:"isPreferred,omitempty"`
	AuthType      string   `json:"type,omitempty" form:"type"`
	Secret        string   `json:"secret,omitempty"`
	URL           string   `json:"url,omitempty"`
	RecoveryCodes []string `json:"recoveryCodes,omitempty"`
}

type TwoFactorInterface interface {
	SetupVerify(ctx *context.Context, passCode string) error
	Verify(passCode string) error
	Initiate(ctx *context.Context, name1 string, name2 string) (*TfaProps, error)
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

func GetTwoFactorUtil(providerType string, config *TfaProps) TwoFactorInterface {
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
	twoFactor := &TfaProps{}

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

func GetMaskedProps(props *TfaProps) *TfaProps {
	maskedProps := &TfaProps{
		AuthType:    SmsType,
		Id:          props.Id,
		IsPreferred: props.IsPreferred,
	}
	return maskedProps
}
