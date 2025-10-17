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

//go:build !skipCi
// +build !skipCi

package deployment

import (
	"testing"

	"github.com/casdoor/casdoor/v2/object"
	"github.com/casdoor/casdoor/v2/util"
)

func TestDeployStaticFiles(t *testing.T) {
	object.InitConfig()

	provider, err := object.GetProvider(util.GetId("admin", "provider_storage_aliyun_oss"))
	if err != nil {
		panic(err)
	}

	deployStaticFiles(provider)
}
