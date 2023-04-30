package object

import (
	"github.com/beego/beego/context"
	"github.com/google/uuid"
)

type TwoFactorSessionData struct {
	UserId        string
	EnableSession bool
	AutoSignIn    bool
}

type TFAProps struct {
	AuthType      string      `json:"type" form:"type"`
	Secret        string      `json:"secret"`
	URL           string      `json:"url,omitempty"`
	RecoveryCodes []uuid.UUID `json:"recoveryCodes,omitempty"`
}

type TwoFactorInterface interface {
	Verify(ctx *context.Context, passCode string) bool
	Initiate(secret string, name string) (*TFAProps, error)
	GetProps() *TFAProps
	Enable(ctx *context.Context) error
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
