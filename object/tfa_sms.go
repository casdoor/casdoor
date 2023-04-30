package object

import (
	"fmt"

	"github.com/beego/beego/context"
	"github.com/google/uuid"
)

type SmsTFA struct {
	Config TFAProps
}

func (tfa SmsTFA) Verify(ctx *context.Context, passCode string) bool {
	dest := ctx.Input.CruSession.Get("tfa_dest").(string)
	if result := CheckVerificationCode(dest, passCode, "en"); result.Code != VerificationSuccess {
		return false
	}
	return true
}

func (tfa SmsTFA) GetProps() *TFAProps {
	return &tfa.Config
}

func (tfa SmsTFA) Initiate(name string, secret string) (*TFAProps, error) {
	recoveryCode, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	twoFactorProps := TFAProps{
		RecoveryCodes: []uuid.UUID{recoveryCode},
	}
	return &twoFactorProps, nil
}

func (tfa SmsTFA) Enable(ctx *context.Context) error {
	secret := ctx.Input.CruSession.Get("tfa_secret")
	recoveryCodes := ctx.Input.CruSession.Get("tfa_recovery_codes")
	if secret == nil || recoveryCodes == nil {
		return fmt.Errorf("two-factor authentication secret or recovery codes is nil")

	}

	tfa.Config.AuthType = SmsType
	tfa.Config.Secret = secret.(string)
	tfa.Config.RecoveryCodes = recoveryCodes.([]uuid.UUID)

	return nil
}

func NewSmsTwoFactor(authType string) *SmsTFA {
	return &SmsTFA{
		Config: TFAProps{
			AuthType: authType,
		},
	}
}
