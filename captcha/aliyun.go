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
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	openapiutil "github.com/alibabacloud-go/openapi-util/service"
	teaUtil "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
)

const AliyunCaptchaVerifyUrl = "captcha.cn-shanghai.aliyuncs.com"

type VerifyCaptchaRequest struct {
	CaptchaVerifyParam *string `json:"CaptchaVerifyParam,omitempty" xml:"CaptchaVerifyParam,omitempty"`
	SceneId            *string `json:"SceneId,omitempty" xml:"SceneId,omitempty"`
}

type VerifyCaptchaResponseBodyResult struct {
	VerifyResult *bool `json:"VerifyResult,omitempty" xml:"VerifyResult,omitempty"`
}

type VerifyCaptchaResponseBody struct {
	Code    *string `json:"Code,omitempty" xml:"Code,omitempty"`
	Message *string `json:"Message,omitempty" xml:"Message,omitempty"`
	// Id of the request
	RequestId *string                          `json:"RequestId,omitempty" xml:"RequestId,omitempty"`
	Result    *VerifyCaptchaResponseBodyResult `json:"Result,omitempty" xml:"Result,omitempty" type:"Struct"`
	Success   *bool                            `json:"Success,omitempty" xml:"Success,omitempty"`
}

type VerifyIntelligentCaptchaResponseBodyResult struct {
	VerifyCode   *string `json:"VerifyCode,omitempty" xml:"VerifyCode,omitempty"`
	VerifyResult *bool   `json:"VerifyResult,omitempty" xml:"VerifyResult,omitempty"`
}

type VerifyIntelligentCaptchaResponseBody struct {
	Code    *string `json:"Code,omitempty" xml:"Code,omitempty"`
	Message *string `json:"Message,omitempty" xml:"Message,omitempty"`
	// Id of the request
	RequestId *string                                     `json:"RequestId,omitempty" xml:"RequestId,omitempty"`
	Result    *VerifyIntelligentCaptchaResponseBodyResult `json:"Result,omitempty" xml:"Result,omitempty" type:"Struct"`
	Success   *bool                                       `json:"Success,omitempty" xml:"Success,omitempty"`
}

type VerifyIntelligentCaptchaResponse struct {
	Headers    map[string]*string                    `json:"headers,omitempty" xml:"headers,omitempty" require:"true"`
	StatusCode *int32                                `json:"statusCode,omitempty" xml:"statusCode,omitempty" require:"true"`
	Body       *VerifyIntelligentCaptchaResponseBody `json:"body,omitempty" xml:"body,omitempty" require:"true"`
}
type AliyunCaptchaProvider struct{}

func NewAliyunCaptchaProvider() *AliyunCaptchaProvider {
	captcha := &AliyunCaptchaProvider{}
	return captcha
}

func (captcha *AliyunCaptchaProvider) VerifyCaptcha(token, clientId, clientSecret, clientId2 string) (bool, error) {
	config := &openapi.Config{}

	config.Endpoint = tea.String(AliyunCaptchaVerifyUrl)
	config.ConnectTimeout = tea.Int(5000)
	config.ReadTimeout = tea.Int(5000)
	config.AccessKeyId = tea.String(clientId)
	config.AccessKeySecret = tea.String(clientSecret)

	client := new(openapi.Client)
	err := client.Init(config)
	if err != nil {
		return false, err
	}

	request := VerifyCaptchaRequest{CaptchaVerifyParam: tea.String(token), SceneId: tea.String(clientId2)}

	err = teaUtil.ValidateModel(&request)
	if err != nil {
		return false, err
	}

	runtime := &teaUtil.RuntimeOptions{}

	body := map[string]interface{}{}
	if !tea.BoolValue(teaUtil.IsUnset(request.CaptchaVerifyParam)) {
		body["CaptchaVerifyParam"] = request.CaptchaVerifyParam
	}

	if !tea.BoolValue(teaUtil.IsUnset(request.SceneId)) {
		body["SceneId"] = request.SceneId
	}

	req := &openapi.OpenApiRequest{
		Body: openapiutil.ParseToMap(body),
	}
	params := &openapi.Params{
		Action:      tea.String("VerifyIntelligentCaptcha"),
		Version:     tea.String("2023-03-05"),
		Protocol:    tea.String("HTTPS"),
		Pathname:    tea.String("/"),
		Method:      tea.String("POST"),
		AuthType:    tea.String("AK"),
		Style:       tea.String("RPC"),
		ReqBodyType: tea.String("formData"),
		BodyType:    tea.String("json"),
	}

	res := &VerifyIntelligentCaptchaResponse{}

	resBody, err := client.CallApi(params, req, runtime)
	if err != nil {
		return false, err
	}

	err = tea.Convert(resBody, &res)
	if err != nil {
		return false, err
	}

	if res.Body.Result.VerifyResult != nil && *res.Body.Result.VerifyResult {
		return true, nil
	}

	return false, nil
}
