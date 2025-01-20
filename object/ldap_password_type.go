// Copyright 2025 The Casdoor Authors. All Rights Reserved.
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

package object

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
)

func generateSSHA(password string) (string, error) {
	salt := make([]byte, 4)
	_, err := rand.Read(salt)
	if err != nil {
		return "", err
	}

	combined := append([]byte(password), salt...)
	hash := sha1.Sum(combined)
	hashWithSalt := append(hash[:], salt...)
	encoded := base64.StdEncoding.EncodeToString(hashWithSalt)

	return "{SSHA}" + encoded, nil
}
