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

package cred

import (
	"crypto/sha256"
	"encoding/hex"
)

type Sha256SaltCredManager struct{}

func getSha256(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}

func getSha256HexDigest(s string) string {
	b := getSha256([]byte(s))
	res := hex.EncodeToString(b)
	return res
}

func NewSha256SaltCredManager() *Sha256SaltCredManager {
	cm := &Sha256SaltCredManager{}
	return cm
}

func (cm *Sha256SaltCredManager) GetHashedPassword(password string, salt string) string {
	if salt == "" {
		return getSha256HexDigest(password)
	}

	return getSha256HexDigest(getSha256HexDigest(password) + salt)
}

func (cm *Sha256SaltCredManager) IsPasswordCorrect(plainPwd string, hashedPwd string, salt string) bool {
	// For backward-compatibility
	if salt == "" {
		if hashedPwd == cm.GetHashedPassword(getSha256HexDigest(plainPwd), salt) {
			return true
		}
	}

	return hashedPwd == cm.GetHashedPassword(plainPwd, salt)
}
