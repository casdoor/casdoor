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

package util

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

const obfuscatorTestKey = "00112233445566778899aabbccddeeff"

var obfuscatorTestIv = []byte("0123456789abcdef")

// encryptPkcs7 encrypts a plaintext that already carries its PKCS#7 padding, and
// prepends the random IV, which is the format of the password ciphertext.
func encryptPkcs7(t *testing.T, plainPassword []byte, padding []byte) string {
	t.Helper()

	key, err := hex.DecodeString(obfuscatorTestKey)
	assert.Nil(t, err)

	block, err := aes.NewCipher(key)
	assert.Nil(t, err)

	password := append(append([]byte{}, plainPassword...), padding...)
	ciphertext := make([]byte, len(obfuscatorTestIv)+len(password))
	copy(ciphertext, obfuscatorTestIv)
	mode := cipher.NewCBCEncrypter(block, obfuscatorTestIv)
	mode.CryptBlocks(ciphertext[len(obfuscatorTestIv):], password)

	return hex.EncodeToString(ciphertext)
}

func TestGetUnobfuscatedPassword(t *testing.T) {
	scenarios := []struct {
		description  string
		ciphertext   string
		expected     interface{}
		expectedPass string
	}{
		{
			description:  "Should be return the password of a valid ciphertext",
			ciphertext:   encryptPkcs7(t, []byte("casdoor"), bytes.Repeat([]byte{0x09}, 9)),
			expected:     nil,
			expectedPass: "casdoor",
		},
		{
			description:  "Should be return an error for a ciphertext that is not a multiple of the block size",
			ciphertext:   "00112233445566778899aabbccddeeff00112233445566",
			expected:     "error",
			expectedPass: "",
		},
		{
			description:  "Should be return an error for a ciphertext with a PKCS#7 padding longer than the password",
			ciphertext:   encryptPkcs7(t, []byte("0123456789abcde"), []byte{0xff}),
			expected:     "error",
			expectedPass: "",
		},
	}
	for _, scenery := range scenarios {
		t.Run(scenery.description, func(t *testing.T) {
			var password string
			var err error
			assert.NotPanics(t, func() {
				password, err = GetUnobfuscatedPassword("AES", obfuscatorTestKey, scenery.ciphertext)
			}, "GetUnobfuscatedPassword() should not panic")

			if scenery.expected == "error" {
				assert.Error(t, err, "The malformed ciphertext should return an error")
				assert.Equal(t, scenery.expectedPass, password, "The password should be empty")
			} else {
				assert.Equal(t, scenery.expected, err, "The returned error not is expected")
				assert.Equal(t, scenery.expectedPass, password, "The returned password not is expected")
			}
		})
	}
}
