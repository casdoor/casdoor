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

//go:build !skipCi

package object

import (
	"testing"

	"github.com/beego/beego/v2/server/web"
	"github.com/casdoor/casdoor/certificate"
	"github.com/casdoor/casdoor/proxy"
	"github.com/casdoor/casdoor/util"
	"github.com/stretchr/testify/assert"
)

func TestGenerateEccKey(t *testing.T) {
	eccKeyStr := encodeEccKey(generateEccKey())
	println(eccKeyStr)
	util.WriteStringToPath(eccKeyStr, "acme_account.key")
}

func TestGetAcmeClient(t *testing.T) {
	err := web.LoadAppConfig("ini", "../conf/app.conf")
	if err != nil {
		panic(err)
	}

	proxy.InitHttpClient()

	eccKey := util.ReadStringFromPath("acme_account.key")
	println(eccKey)

	client, err := getAcmeClient("acme2@casbin.org", eccKey, false, true)
	assert.Nil(t, err)
	pem, key, err := certificate.ObtainCertificateAli(client, "casbin.com", "", "")
	assert.Nil(t, err)
	println(pem)
	println()
	println(key)
}
