// Copyright 2024 The Casdoor Authors. All Rights Reserved.
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
	"crypto/sha512"
	"encoding/hex"
)

type Sha512SaltCredManager struct{}

func getSha512(data []byte) []byte {
	hash := sha512.Sum512(data)
	return hash[:]
}

func getSha512HexDigest(s string) string {
	b := getSha512([]byte(s))
	res := hex.EncodeToString(b)
	return res
}

func NewSha512SaltCredManager() *Sha512SaltCredManager {
	cm := &Sha512SaltCredManager{}
	return cm
}

func (cm *Sha512SaltCredManager) GetHashedPassword(password string, salt string) string {
	if salt == "" {
		return getSha512HexDigest(password)
	}

	return getSha512HexDigest(getSha512HexDigest(password) + salt)
}

func (cm *Sha512SaltCredManager) IsPasswordCorrect(plainPwd string, hashedPwd string, salt string) bool {
	// For backward-compatibility
	if salt == "" {
		if hashedPwd == cm.GetHashedPassword(getSha512HexDigest(plainPwd), salt) {
			return true
		}
	}

	return hashedPwd == cm.GetHashedPassword(plainPwd, salt)
}
