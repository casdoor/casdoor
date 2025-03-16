// Copyright 2025 The Casdoor Authors. All Rights Reserved.
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

package faceId

import (
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	facebody20191230 "github.com/alibabacloud-go/facebody-20191230/v5/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
	"strings"
)

type AliyunFaceIdProvider struct {
	AccessKey    string
	AccessSecret string

	Endpoint              string
	QualityScoreThreshold float32
}

func NewAliyunFaceIdProvider(accessKey string, accessSecret string, endPoint string) *AliyunFaceIdProvider {
	return &AliyunFaceIdProvider{
		AccessKey:             accessKey,
		AccessSecret:          accessSecret,
		Endpoint:              endPoint,
		QualityScoreThreshold: 0.65,
	}
}

func (provider *AliyunFaceIdProvider) Check(base64ImageA string, base64ImageB string) (bool, error) {
	config := openapi.Config{
		AccessKeyId:     tea.String(provider.AccessKey),
		AccessKeySecret: tea.String(provider.AccessSecret),
	}
	config.Endpoint = tea.String(provider.Endpoint)
	client, err := facebody20191230.NewClient(&config)

	if err != nil {
		return false, err
	}

	compareFaceRequest := &facebody20191230.CompareFaceRequest{
		QualityScoreThreshold: tea.Float32(provider.QualityScoreThreshold),
		ImageDataA:            tea.String(strings.Replace(base64ImageA, "data:image/png;base64,", "", -1)),
		ImageDataB:            tea.String(strings.Replace(base64ImageB, "data:image/png;base64,", "", -1)),
	}

	runtime := &util.RuntimeOptions{}
	res, tryErr := func() (result *facebody20191230.CompareFaceResponse, _e error) {
		defer func() {
			if r := tea.Recover(recover()); r != nil {
				_e = r
			}
		}()
		// 复制代码运行请自行打印 API 的返回值
		result, err := client.CompareFaceWithOptions(compareFaceRequest, runtime)
		if err != nil {
			return nil, err
		}

		return result, nil
	}()

	if tryErr != nil {
		return false, tryErr
	}

	if res == nil {
		return false, nil
	}

	if *res.Body.Data.Thresholds[0] < *res.Body.Data.Confidence {
		return true, nil
	}

	return false, nil
}
