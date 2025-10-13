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
	"strconv"
	"strings"

	"github.com/alexedwards/argon2id"
)

type Argon2idCredManager struct{}

func NewArgon2idCredManager() *Argon2idCredManager {
	cm := &Argon2idCredManager{}
	return cm
}

// parseArgon2idSalt parses the salt field to extract pepper and parameters
// Format: "pepper|m=65536|t=1|p=2" or just "pepper"
// Returns: pepper, params (or nil for defaults)
func parseArgon2idSalt(salt string) (string, *argon2id.Params) {
	if salt == "" {
		return "", nil
	}

	parts := strings.Split(salt, "|")
	pepper := parts[0]

	// If no parameters specified, use defaults
	if len(parts) == 1 {
		return pepper, nil
	}

	// Parse parameters
	params := &argon2id.Params{
		Memory:      64 * 1024, // default
		Iterations:  1,         // default
		Parallelism: 2,         // default
		SaltLength:  16,
		KeyLength:   32,
	}

	for i := 1; i < len(parts); i++ {
		param := parts[i]
		if strings.HasPrefix(param, "m=") {
			if val, err := strconv.Atoi(strings.TrimPrefix(param, "m=")); err == nil {
				params.Memory = uint32(val)
			}
		} else if strings.HasPrefix(param, "t=") {
			if val, err := strconv.Atoi(strings.TrimPrefix(param, "t=")); err == nil {
				params.Iterations = uint32(val)
			}
		} else if strings.HasPrefix(param, "p=") {
			if val, err := strconv.Atoi(strings.TrimPrefix(param, "p=")); err == nil {
				params.Parallelism = uint8(val)
			}
		}
	}

	return pepper, params
}

func (cm *Argon2idCredManager) GetHashedPassword(password string, salt string) string {
	// Parse salt to extract pepper and optional parameters
	// Format: "pepper|m=65536|t=1|p=2" or just "pepper"
	pepper, params := parseArgon2idSalt(salt)

	// Use pepper: prepend it to the password before hashing
	// This allows migration of users from systems that used a pepper
	passwordWithPepper := password
	if pepper != "" {
		passwordWithPepper = pepper + password
	}

	// Use custom parameters if provided, otherwise use defaults
	if params == nil {
		params = argon2id.DefaultParams
	}

	hash, err := argon2id.CreateHash(passwordWithPepper, params)
	if err != nil {
		return ""
	}
	return hash
}

func (cm *Argon2idCredManager) IsPasswordCorrect(plainPwd string, hashedPwd string, salt string) bool {
	// Parse salt to extract pepper and optional parameters
	// Format: "pepper|m=65536|t=1|p=2" or just "pepper"
	pepper, _ := parseArgon2idSalt(salt)

	// Use pepper: prepend it to the password before verification
	// This allows migration of users from systems that used a pepper
	passwordWithPepper := plainPwd
	if pepper != "" {
		passwordWithPepper = pepper + plainPwd
	}

	// The argon2id library automatically uses the parameters embedded in the hash
	match, _ := argon2id.ComparePasswordAndHash(passwordWithPepper, hashedPwd)
	return match
}
