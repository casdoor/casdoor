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
	"crypto/subtle"
	"encoding/base64"

	"github.com/emmansun/gmsm/sm3"
	"golang.org/x/crypto/pbkdf2"
)

// Pbkdf2Sm3CredManager hashes passwords with PBKDF2 and HMAC-SM3 (GB/T 32905), for the Chinese cryptography
// evaluation (密评) of the enterprise edition
type Pbkdf2Sm3CredManager struct{}

const pbkdf2Sm3Iterations = 27500

func NewPbkdf2Sm3CredManager() *Pbkdf2Sm3CredManager {
	return &Pbkdf2Sm3CredManager{}
}

func (cm *Pbkdf2Sm3CredManager) GetHashedPassword(password string, salt string) string {
	res := pbkdf2.Key([]byte(password), []byte(salt), pbkdf2Sm3Iterations, 32, sm3.New)
	return base64.StdEncoding.EncodeToString(res)
}

func (cm *Pbkdf2Sm3CredManager) IsPasswordCorrect(plainPwd string, hashedPwd string, salt string) bool {
	return subtle.ConstantTimeCompare([]byte(hashedPwd), []byte(cm.GetHashedPassword(plainPwd, salt))) == 1
}
