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

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/astaxie/beego/logs"
)

const ReCaptchaVerifyURL = "https://recaptcha.net/recaptcha/api/siteverify"

type ReCaptchaProvider struct {
	ClientSecret string
}

func NewReCaptchaProvider(clientSecret string) *ReCaptchaProvider {
	captcha := &ReCaptchaProvider{}

	captcha.ClientSecret = clientSecret
	return captcha
}

func (captcha *ReCaptchaProvider) VerifyCaptcha(token string) (bool, error) {
	reqData := url.Values{
		"secret":   {captcha.ClientSecret},
		"response": {token},
	}
	resp, err := http.PostForm(ReCaptchaVerifyURL, reqData)
	if err != nil {
		logs.Error("Failed post to verify captcha: %v", err)
		return false, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		logs.Error("Failed to read from verify captcha: %v", err)
		return false, err
	}
	type captchaResponse struct {
		Success bool `json:"success"`
	}
	captchaResp := &captchaResponse{}
	err = json.Unmarshal(body, captchaResp)
	if err != nil || captchaResp == nil {
		logs.Error("Failed to unmarshal from verify captcha: %v", err)
		return false, err
	}

	return captchaResp.Success, nil
}
