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
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/casdoor/casdoor/util"
)

const GEETESTCaptchaVerifyUrl = "http://gcaptcha4.geetest.com/validate"

type GEETESTCaptchaProvider struct{}

func NewGEETESTCaptchaProvider() *GEETESTCaptchaProvider {
	captcha := &GEETESTCaptchaProvider{}
	return captcha
}

func (captcha *GEETESTCaptchaProvider) VerifyCaptcha(token, clientSecret string) (bool, error) {
	pathData, err := url.ParseQuery(token)
	if err != nil {
		return false, err
	}

	signToken := util.GetHmacSha256(clientSecret, pathData["lot_number"][0])

	formData := make(url.Values)
	formData["lot_number"] = []string{pathData["lot_number"][0]}
	formData["captcha_output"] = []string{pathData["captcha_output"][0]}
	formData["pass_token"] = []string{pathData["pass_token"][0]}
	formData["gen_time"] = []string{pathData["gen_time"][0]}
	formData["sign_token"] = []string{signToken}
	captchaId := pathData["captcha_id"][0]

	cli := http.Client{Timeout: time.Second * 5}
	resp, err := cli.PostForm(fmt.Sprintf("%s?captcha_id=%s", GEETESTCaptchaVerifyUrl, captchaId), formData)
	if err != nil || resp.StatusCode != 200 {
		return false, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	type captchaResponse struct {
		Result string `json:"result"`
		Reason string `json:"reason"`
	}
	captchaResp := &captchaResponse{}
	err = json.Unmarshal(body, captchaResp)
	if err != nil {
		return false, err
	}

	if captchaResp.Result == "success" {
		return true, nil
	}

	return false, errors.New(captchaResp.Reason)
}
