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
	"testing"
)

func TestNewAliyunOssStorageProvider_StaticCredentials(t *testing.T) {
	// Test with static credentials (existing behavior)
	provider := NewAliyunOssStorageProvider("testAccessId", "testAccessKey", "oss-cn-hangzhou", "test-bucket", "https://oss-cn-hangzhou.aliyuncs.com")
	if provider == nil {
		t.Error("Expected provider to be created with static credentials")
	}
}

func TestNewAliyunOssStorageProvider_RRSADetection(t *testing.T) {
	// Save original environment
	originalRoleArn := os.Getenv("ALIBABA_CLOUD_ROLE_ARN")
	originalProviderArn := os.Getenv("ALIBABA_CLOUD_OIDC_PROVIDER_ARN")
	originalTokenFile := os.Getenv("ALIBABA_CLOUD_OIDC_TOKEN_FILE")

	// Restore environment after test
	defer func() {
		if originalRoleArn == "" {
			os.Unsetenv("ALIBABA_CLOUD_ROLE_ARN")
		} else {
			os.Setenv("ALIBABA_CLOUD_ROLE_ARN", originalRoleArn)
		}
		if originalProviderArn == "" {
			os.Unsetenv("ALIBABA_CLOUD_OIDC_PROVIDER_ARN")
		} else {
			os.Setenv("ALIBABA_CLOUD_OIDC_PROVIDER_ARN", originalProviderArn)
		}
		if originalTokenFile == "" {
			os.Unsetenv("ALIBABA_CLOUD_OIDC_TOKEN_FILE")
		} else {
			os.Setenv("ALIBABA_CLOUD_OIDC_TOKEN_FILE", originalTokenFile)
		}
	}()

	// Test RRSA detection with environment variables set
	os.Setenv("ALIBABA_CLOUD_ROLE_ARN", "acs:ram::123456789:role/test-role")
	os.Setenv("ALIBABA_CLOUD_OIDC_PROVIDER_ARN", "acs:ram::123456789:oidc-provider/test-provider")
	os.Setenv("ALIBABA_CLOUD_OIDC_TOKEN_FILE", "/var/run/secrets/token")

	// This should attempt to use RRSA (will fall back to static credentials if RRSA fails due to missing token file)
	provider := NewAliyunOssStorageProvider("", "", "oss-cn-hangzhou", "test-bucket", "https://oss-cn-hangzhou.aliyuncs.com")
	if provider == nil {
		t.Error("Expected provider to be created even if RRSA fails")
	}

	// Test with "rrsa" as placeholder value
	provider = NewAliyunOssStorageProvider("rrsa", "rrsa", "oss-cn-hangzhou", "test-bucket", "https://oss-cn-hangzhou.aliyuncs.com")
	if provider == nil {
		t.Error("Expected provider to be created with 'rrsa' placeholder")
	}
}

func TestNewAliyunOssStorageProvider_NoRRSA(t *testing.T) {
	// Ensure RRSA environment variables are not set
	originalRoleArn := os.Getenv("ALIBABA_CLOUD_ROLE_ARN")
	defer func() {
		if originalRoleArn == "" {
			os.Unsetenv("ALIBABA_CLOUD_ROLE_ARN")
		} else {
			os.Setenv("ALIBABA_CLOUD_ROLE_ARN", originalRoleArn)
		}
	}()
	os.Unsetenv("ALIBABA_CLOUD_ROLE_ARN")

	// Should use static credentials (which will be empty in this case)
	provider := NewAliyunOssStorageProvider("", "", "oss-cn-hangzhou", "test-bucket", "https://oss-cn-hangzhou.aliyuncs.com")
	if provider == nil {
		t.Error("Expected provider to be created with static credentials")
	}
}
