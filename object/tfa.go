package object

import (
	"github.com/beego/beego/context"
)

type TwoFactorSessionData struct {
	UserId        string
	EnableSession bool
	AutoSignIn    bool
}

type TfaProps struct {
	AuthType      string   `json:"type,omitempty" form:"type"`
	Secret        string   `json:"secret,omitempty"`
	URL           string   `json:"url,omitempty"`
	RecoveryCodes []string `json:"recoveryCodes,omitempty"`
}

type TwoFactorInterface interface {
	Verify(ctx *context.Context, passCode string) bool
	Initiate(ctx *context.Context, name1 string, name2 string) (*TfaProps, error)
	GetProps() *TfaProps
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

func GetTwoFactorUtil(providerType string) TwoFactorInterface {
	switch providerType {
	case SmsType:
		return NewSmsTwoFactor(providerType)
	case TotpType:
		return nil
	}

	return nil
}
