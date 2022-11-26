// Copyright 2022 The Casdoor Authors. All Rights Reserved.
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

package captcha

import "fmt"

type CaptchaProvider interface {
	VerifyCaptcha(token, clientSecret string) (bool, error)
}

func GetCaptchaProvider(captchaType string) CaptchaProvider {
	if captchaType == "Default" {
		return NewDefaultCaptchaProvider()
	} else if captchaType == "reCAPTCHA" {
		return NewReCaptchaProvider()
	} else if captchaType == "hCaptcha" {
		return NewHCaptchaProvider()
	} else if captchaType == "Aliyun Captcha" {
		return NewAliyunCaptchaProvider()
	} else if captchaType == "GEETEST" {
		return NewGEETESTCaptchaProvider()
	} else if captchaType == "Turnstile" {
		return NewTurnstileProvider()
	}
	return nil
}

func VerifyCaptchaByCaptchaType(captchaType, token, clientSecret string) (bool, error) {
	provider := GetCaptchaProvider(captchaType)
	if provider == nil {
		return false, fmt.Errorf("invalid captcha provider: %s", captchaType)
	}

	return provider.VerifyCaptcha(token, clientSecret)
}
