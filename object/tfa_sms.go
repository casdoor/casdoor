package object

import (
	"errors"
	"fmt"

	"github.com/beego/beego/context"
	"github.com/google/uuid"
)

type SmsTFA struct {
	Config *TfaProps
}

func (tfa *SmsTFA) SetupVerify(ctx *context.Context, passCode string) error {
	dest := ctx.Input.CruSession.Get("tfa_dest").(string)
	if result := CheckVerificationCode(dest, passCode, "en"); result.Code != VerificationSuccess {
		return errors.New(result.Msg)
	}
	return nil
}

func (tfa *SmsTFA) Verify(passCode string) error {
	if result := CheckVerificationCode(tfa.Config.Secret, passCode, "en"); result.Code != VerificationSuccess {
		return errors.New(result.Msg)
	}
	return nil
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

	for i, twoFactorProp := range user.TwoFactorAuth {
		if twoFactorProp.AuthType == SmsType {
			user.TwoFactorAuth = append(user.TwoFactorAuth[:i], user.TwoFactorAuth[i+1:]...)
		}
	}
	user.TwoFactorAuth = append(user.TwoFactorAuth, tfa.Config)

	affected := UpdateUser(user.GetId(), user, []string{"two_factor_auth"}, user.IsAdminUser())
	if !affected {
		return fmt.Errorf("failed to enable two factor authentication")
	}

	return nil
}

func NewSmsTwoFactor(config *TfaProps) *SmsTFA {
	if config == nil {
		config = &TfaProps{
			AuthType: SmsType,
		}
	}
	return &SmsTFA{
		Config: config,
	}
}
