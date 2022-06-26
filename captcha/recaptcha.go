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
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

const ReCaptchaVerifyUrl = "https://recaptcha.net/recaptcha/api/siteverify"

type ReCaptchaProvider struct {
}

func NewReCaptchaProvider() *ReCaptchaProvider {
	captcha := &ReCaptchaProvider{}
	return captcha
}

func (captcha *ReCaptchaProvider) VerifyCaptcha(token, clientSecret string) (bool, error) {
	reqData := url.Values{
		"secret":   {clientSecret},
		"response": {token},
	}
	resp, err := http.PostForm(ReCaptchaVerifyUrl, reqData)
	if err != nil {
		return false, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	type captchaResponse struct {
		Success    bool     `json:"success"`
		ErrorCodes []string `json:"error-codes"`
	}
	captchaResp := &captchaResponse{}
	err = json.Unmarshal(body, captchaResp)
	if err != nil {
		return false, err
	}

	if len(captchaResp.ErrorCodes) > 0 {
		return false, errors.New(strings.Join(captchaResp.ErrorCodes, ","))
	}

	return captchaResp.Success, nil
}
