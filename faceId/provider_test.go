// Copyright 2026 The Casdoor Authors. All Rights Reserved.
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

import "testing"

func TestGetFaceIdProviderLocalUniFace(t *testing.T) {
	provider := GetFaceIdProvider("Local UniFace", "", "secret", "http://127.0.0.1:8100")

	localProvider, ok := provider.(*LocalUniFaceProvider)
	if !ok {
		t.Fatalf("expected *LocalUniFaceProvider, got %T", provider)
	}
	if localProvider.ApiKey != "secret" {
		t.Fatalf("expected api key secret, got %s", localProvider.ApiKey)
	}
}

func TestGetFaceIdProviderAlibabaCloudFacebody(t *testing.T) {
	provider := GetFaceIdProvider("Alibaba Cloud Facebody", "accessKey", "accessSecret", "endpoint")

	if _, ok := provider.(*AliyunFaceIdProvider); !ok {
		t.Fatalf("expected *AliyunFaceIdProvider, got %T", provider)
	}
}
