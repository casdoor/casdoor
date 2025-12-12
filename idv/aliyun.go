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

package idv

import (
	"fmt"

	cloudauth "github.com/alibabacloud-go/cloudauth-20190307/v3/client"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	"github.com/alibabacloud-go/tea/tea"
)

const (
	// DefaultAlibabaCloudEndpoint is the default endpoint for Alibaba Cloud ID verification service
	DefaultAlibabaCloudEndpoint = "cloudauth.cn-shanghai.aliyuncs.com"
)

type AlibabaCloudIdvProvider struct {
	ClientId     string
	ClientSecret string
	Endpoint     string
}

func NewAlibabaCloudIdvProvider(clientId string, clientSecret string, endpoint string) *AlibabaCloudIdvProvider {
	return &AlibabaCloudIdvProvider{
		ClientId:     clientId,
		ClientSecret: clientSecret,
		Endpoint:     endpoint,
	}
}

func (provider *AlibabaCloudIdvProvider) VerifyIdentity(idCardType string, idCard string, realName string) (bool, error) {
	if provider.ClientId == "" || provider.ClientSecret == "" {
		return false, fmt.Errorf("Alibaba Cloud credentials not configured")
	}

	if idCard == "" || realName == "" {
		return false, fmt.Errorf("ID card and real name are required")
	}

	// Default endpoint if not configured
	endpoint := provider.Endpoint
	if endpoint == "" {
		endpoint = DefaultAlibabaCloudEndpoint
	}

	// Create client configuration
	config := &openapi.Config{
		AccessKeyId:     tea.String(provider.ClientId),
		AccessKeySecret: tea.String(provider.ClientSecret),
		Endpoint:        tea.String(endpoint),
	}

	// Create Alibaba Cloud Auth client
	client, err := cloudauth.NewClient(config)
	if err != nil {
		return false, fmt.Errorf("failed to create Alibaba Cloud client: %v", err)
	}

	// Prepare verification request using Id2MetaVerify API
	// This API verifies Chinese ID card number and real name
	// Reference: https://help.aliyun.com/zh/id-verification/financial-grade-id-verification/server-side-integration-2
	request := &cloudauth.Id2MetaVerifyRequest{
		IdentifyNum: tea.String(idCard),
		UserName:    tea.String(realName),
		ParamType:   tea.String("normal"),
	}

	// Send verification request
	response, err := client.Id2MetaVerify(request)
	if err != nil {
		return false, fmt.Errorf("failed to verify identity with Alibaba Cloud: %v", err)
	}

	// Check response
	if response == nil || response.Body == nil {
		return false, fmt.Errorf("empty response from Alibaba Cloud")
	}

	// Check if the API call was successful
	if response.Body.Code == nil || *response.Body.Code != "200" {
		message := "unknown error"
		if response.Body.Message != nil {
			message = *response.Body.Message
		}
		return false, fmt.Errorf("Alibaba Cloud API error: %s", message)
	}

	// Check verification result
	// BizCode "1" means verification passed
	if response.Body.ResultObject != nil && response.Body.ResultObject.BizCode != nil {
		if *response.Body.ResultObject.BizCode == "1" {
			return true, nil
		}
		return false, fmt.Errorf("identity verification failed: BizCode=%s", *response.Body.ResultObject.BizCode)
	}

	return false, fmt.Errorf("identity verification failed: missing result")
}
