package object

import (
	"github.com/beego/beego/context"
	"github.com/google/uuid"
)

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

type TwoFactorSessionData struct {
	UserId        string
	EnableSession bool
	AutoSignIn    bool
}

type TwoFactorProps struct {
	AuthType      string      `json:"type" form:"type"`
	Secret        string      `json:"secret"`
	URL           string      `json:"url,omitempty"`
	RecoveryCodes []uuid.UUID `json:"recoveryCodes,omitempty"`
}

func GetTwoFactorUtil(providerType string) TwoFactorInterface {
	switch providerType {
	case SmsType:
		return NewSmsTwoFactor(providerType)
	case TotpType:
		return nil
	}

	return nil
}

type TwoFactorInterface interface {
	Verify(ctx *context.Context, passCode string) bool
	Initiate(secret string, name string) (*TwoFactorProps, error)
	GetProps() *TwoFactorProps
	Enable(ctx *context.Context) error
}
