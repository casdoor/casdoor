// Copyright 2021 The Casdoor Authors. All Rights Reserved.
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

package storage

import (
	"os"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/aliyun/credentials-go/credentials"
	casdoorOss "github.com/casdoor/oss"
	"github.com/casdoor/oss/aliyun"
)

func NewAliyunOssStorageProvider(clientId string, clientSecret string, region string, bucket string, endpoint string) casdoorOss.StorageInterface {
	// Check if RRSA is available (empty credentials + environment variables set)
	if (clientId == "" || clientId == "rrsa") &&
		(clientSecret == "" || clientSecret == "rrsa") &&
		os.Getenv("ALIBABA_CLOUD_ROLE_ARN") != "" {
		// Use RRSA to get temporary credentials
		config := &credentials.Config{}
		config.SetType("oidc_role_arn")
		config.SetRoleArn(os.Getenv("ALIBABA_CLOUD_ROLE_ARN"))
		config.SetOIDCProviderArn(os.Getenv("ALIBABA_CLOUD_OIDC_PROVIDER_ARN"))
		config.SetOIDCTokenFilePath(os.Getenv("ALIBABA_CLOUD_OIDC_TOKEN_FILE"))
		config.SetRoleSessionName("casdoor-oss")

		// Set STS endpoint if provided
		if stsEndpoint := os.Getenv("ALIBABA_CLOUD_STS_ENDPOINT"); stsEndpoint != "" {
			config.SetSTSEndpoint(stsEndpoint)
		}

		credential, err := credentials.NewCredential(config)
		if err == nil {
			accessKeyId, errId := credential.GetAccessKeyId()
			accessKeySecret, errSecret := credential.GetAccessKeySecret()
			securityToken, errToken := credential.GetSecurityToken()

			if errId == nil && errSecret == nil && errToken == nil &&
				accessKeyId != nil && accessKeySecret != nil && securityToken != nil {
				// Successfully obtained RRSA credentials
				sp := aliyun.New(&aliyun.Config{
					AccessID:      *accessKeyId,
					AccessKey:     *accessKeySecret,
					Bucket:        bucket,
					Endpoint:      endpoint,
					ClientOptions: []oss.ClientOption{oss.SecurityToken(*securityToken)},
				})
				return sp
			}
		}
		// If RRSA fails, fall through to static credentials (which will fail if empty)
	}

	// Use static credentials (existing behavior)
	sp := aliyun.New(&aliyun.Config{
		AccessID:  clientId,
		AccessKey: clientSecret,
		Bucket:    bucket,
		Endpoint:  endpoint,
	})

	return sp
}
