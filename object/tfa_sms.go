package object

import (
	"fmt"

	"github.com/beego/beego/context"
	"github.com/google/uuid"
)

type SmsTFA struct {
	Config *TfaProps
}

func (tfa *SmsTFA) Verify(ctx *context.Context, passCode string) bool {
	dest := ctx.Input.CruSession.Get("tfa_dest").(string)
	if result := CheckVerificationCode(dest, passCode, "en"); result.Code != VerificationSuccess {
		return false
	}
	return true
}

func (tfa *SmsTFA) GetProps() *TfaProps {
	return tfa.Config
}

func (tfa *SmsTFA) Initiate(ctx *context.Context, name string, secret string) (*TfaProps, error) {
	recoveryCode, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	err = ctx.Input.CruSession.Set("tfa_recovery_codes", []string{recoveryCode.String()})
	if err != nil {
		return nil, err
	}

	twoFactorProps := TfaProps{
		RecoveryCodes: []string{recoveryCode.String()},
	}
	return &twoFactorProps, nil
}

func (tfa *SmsTFA) Enable(ctx *context.Context, user *User) error {
	secret := ctx.Input.CruSession.Get("tfa_dest")
	recoveryCodes := ctx.Input.CruSession.Get("tfa_recovery_codes")
	if secret == nil || recoveryCodes == nil {
		return fmt.Errorf("two-factor authentication secret or recovery codes is nil")

	}

	tfa.Config.AuthType = SmsType
	tfa.Config.Secret = secret.(string)
	tfa.Config.RecoveryCodes = recoveryCodes.([]string)

	for i, twoFactorProps := range user.TwoFactorAuth {
		if twoFactorProps.AuthType == tfa.GetProps().AuthType {
			user.TwoFactorAuth = append(user.TwoFactorAuth[:i], user.TwoFactorAuth[i+1:]...)
		}
	}
	user.TwoFactorAuth = append(user.TwoFactorAuth, tfa.GetProps())

	affected := UpdateUser(user.GetId(), user, []string{"two_factor_auth"}, user.IsAdminUser())
	if !affected {
		return fmt.Errorf("failed to enable two factor authentication")
	}

	return nil
}

func NewSmsTwoFactor(authType string) *SmsTFA {
	return &SmsTFA{
		Config: &TfaProps{
			AuthType: authType,
		},
	}
}
