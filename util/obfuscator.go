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

package util

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/des"
	"encoding/hex"
	"fmt"
)

func GetPlainPassword(passwordObfuscatorType string, passwordObfuscatorKey string, passwordCipher string) (string, error) {
	if passwordObfuscatorType == "DES" || passwordObfuscatorType == "AES" {
		key, err := hex.DecodeString(passwordObfuscatorKey)
		if err != nil {
			return "", err
		}
		if passwordObfuscatorType == "DES" {
			block, err := des.NewCipher(key)
			if err != nil {
				return "", err
			}
			return Decrypt(passwordCipher, block)
		} else {
			block, err := aes.NewCipher(key)
			if err != nil {
				return "", err
			}
			return Decrypt(passwordCipher, block)
		}
	}
	return passwordCipher, nil
}

func Decrypt(passwordCipherStr string, block cipher.Block) (string, error) {
	passwordCipherBytes, err := hex.DecodeString(passwordCipherStr)
	if err != nil {
		return "", err
	}

	if len(passwordCipherBytes) < block.BlockSize() {
		return "", fmt.Errorf("ciphertext too short")
	}

	iv := passwordCipherBytes[:block.BlockSize()]
	password := make([]byte, len(passwordCipherBytes)-block.BlockSize())

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(password, passwordCipherBytes[block.BlockSize():])

	return string(unPaddingPkcs7(password)), nil
}

func unPaddingPkcs7(s []byte) []byte {
	length := len(s)
	if length == 0 {
		return s
	}
	unPadding := int(s[length-1])
	return s[:(length - unPadding)]
}
