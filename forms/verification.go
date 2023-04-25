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

package forms

import (
	"strings"

	"github.com/casdoor/casdoor/i18n"
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
