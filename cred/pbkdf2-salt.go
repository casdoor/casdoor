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

package cred

import (
	"crypto/sha256"
	"encoding/base64"

	"golang.org/x/crypto/pbkdf2"
)

type Pbkdf2SaltCredManager struct{}

func NewPbkdf2SaltCredManager() *Pbkdf2SaltCredManager {
	cm := &Pbkdf2SaltCredManager{}
	return cm
}

func (cm *Pbkdf2SaltCredManager) GetHashedPassword(password string, passwordSalt string) string {
	// https://www.keycloak.org/docs/latest/server_admin/index.html#password-database-compromised
	decodedSalt, _ := base64.StdEncoding.DecodeString(passwordSalt)
	res := pbkdf2.Key([]byte(password), decodedSalt, 27500, 64, sha256.New)
	return base64.StdEncoding.EncodeToString(res)
}

func (cm *Pbkdf2SaltCredManager) IsPasswordCorrect(plainPwd string, hashedPwd string, passwordSalt string) bool {
	return hashedPwd == cm.GetHashedPassword(plainPwd, passwordSalt)
}
