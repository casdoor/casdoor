// Copyright 2021 The casbin Authors. All Rights Reserved.
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

package cred

import (
	"crypto/md5"
	"encoding/hex"
	"github.com/thanhpk/randstr"
)

type Md5SaltCredManager struct{}

func generateMd5Salt() string {
	return randstr.Hex(8)
}

func getMd5(data []byte) []byte {
	hash := md5.Sum(data)
	return hash[:]
}

func getMd5HexDigest(s string) string {
	b := getMd5([]byte(s))
	res := hex.EncodeToString(b)
	return res
}

func NewMd5SaltCredManager() *Md5SaltCredManager {
	cm := &Md5SaltCredManager{}
	return cm
}

func (cm *Md5SaltCredManager) GetSealedPassword(password string, organizationSalt string) string {
	res := new(StandardPassword)
	res.Type = "md5-salt"
	res.OrganizationSalt = organizationSalt
	res.UserSalt = generateMd5Salt()

	hash := getMd5HexDigest(password)
	hash = getMd5HexDigest(hash + res.UserSalt)
	if res.OrganizationSalt != "" {
		hash = getMd5HexDigest(hash + res.OrganizationSalt)
	}
	res.PasswordHash = hash

	return res.String()
}

func (cm *Md5SaltCredManager) CheckSealedPassword(password string, sealedPassword string) bool {
	currentPassword, err := ParseStandardPassword(sealedPassword)
	if err != nil {
		panic(err)
	}

	hash := getMd5HexDigest(password)
	if currentPassword.UserSalt != "" {
		hash = getMd5HexDigest(hash + currentPassword.UserSalt)
	}
	if currentPassword.OrganizationSalt != "" {
		hash = getMd5HexDigest(hash + currentPassword.OrganizationSalt)
	}
	return hash == currentPassword.PasswordHash
}