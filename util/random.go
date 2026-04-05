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

package util

import (
	"crypto/rand"
	"math/big"

	"github.com/google/uuid"
	"github.com/thanhpk/randstr"
)

func GenerateClientId() string {
	return randstr.Hex(10)
}

func GenerateClientSecret() string {
	return randstr.Hex(20)
}

func GeneratePasswordSalt() string {
	return randstr.Hex(10)
}

// RandomIntn returns a cryptographically secure random int in [0, n).
func RandomIntn(n int) int {
	val, err := rand.Int(rand.Reader, big.NewInt(int64(n)))
	if err != nil {
		panic(err)
	}
	return int(val.Int64())
}

// GenerateUUID returns a random UUID v4 string.
func GenerateUUID() string {
	return uuid.NewString()
}

// RandomStringFromCharset returns a cryptographically secure random string
// of the given length drawn from charset.
func RandomStringFromCharset(charset string, length int) string {
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[RandomIntn(len(charset))]
	}
	return string(result)
}

func GetRandomName() string {
	return RandomStringFromCharset("0123456789abcdefghijklmnopqrstuvwxyz", 6)
}

func generateRandomString(length int) (string, error) {
	return RandomStringFromCharset("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789", length), nil
}

func GenerateTwoUniqueRandomStrings() (string, string, error) {
	len1 := 16 + int(big.NewInt(17).Int64())
	len2 := 16 + int(big.NewInt(17).Int64())

	str1, err := generateRandomString(len1)
	if err != nil {
		return "", "", err
	}
	str2, err := generateRandomString(len2)
	if err != nil {
		return "", "", err
	}

	for str1 == str2 {
		str2, err = generateRandomString(len2)
		if err != nil {
			return "", "", err
		}
	}
	return str1, str2, nil
}
