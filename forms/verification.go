package forms

import (
	"github.com/casdoor/casdoor/i18n"
	"strings"
)

type VerificationForm struct {
	Dest          string `form:"dest"`
	Type          string `form:"type"`
	CountryCode   string `form:"countryCode"`
	ApplicationId string `form:"applicationId"`
	Method        string `form:"method"`
	CheckUser     string `form:"checkUser"`

	CaptchaType  string `form:"captchaType"`
	ClientSecret string `form:"clientSecret"`
	CaptchaToken string `form:"captchaToken"`
}

const (
	SendVerifyCode = 0
	VerifyCaptcha  = 1
)

func (form *VerificationForm) CheckParameter(checkType int, lang string) string {
	if checkType == SendVerifyCode {
		if form.Type == "" {
			return i18n.Translate(lang, "general:Missing parameter") + ": type."
		}
		if form.Dest == "" {
			return i18n.Translate(lang, "general:Missing parameter") + ": dest."
		}
		if form.CaptchaType == "" {
			return i18n.Translate(lang, "general:Missing parameter") + ": checkType."
		}
		if !strings.Contains(form.ApplicationId, "/") {
			return i18n.Translate(lang, "verification:Wrong parameter") + ": applicationId."
		}
	}

	if form.CaptchaType != "none" {
		if form.CaptchaToken == "" {
			return i18n.Translate(lang, "general:Missing parameter") + ": captchaToken."
		}
		if form.ClientSecret == "" {
			return i18n.Translate(lang, "general:Missing parameter") + ": clientSecret."
		}
	}

	return ""
}
