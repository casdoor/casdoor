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
)

type Md5UserSaltCredManager struct{}

func getMd5(data []byte) []byte {
	hash := md5.Sum(data)
	return hash[:]
}

func getMd5HexDigest(s string) string {
	b := getMd5([]byte(s))
	res := hex.EncodeToString(b)
	return res
}

func NewMd5UserSaltCredManager() *Sha256SaltCredManager {
	cm := &Sha256SaltCredManager{}
	return cm
}

func (cm *Md5UserSaltCredManager) GetHashedPassword(password string, userSalt string, organizationSalt string) string {
	hash := getMd5HexDigest(password)
	res := getMd5HexDigest(hash + userSalt)
	return res
}

func (cm *Md5UserSaltCredManager) IsPasswordCorrect(plainPwd string, hashedPwd string, userSalt string, organizationSalt string) bool {
	return hashedPwd == cm.GetHashedPassword(plainPwd, userSalt, organizationSalt)
}
