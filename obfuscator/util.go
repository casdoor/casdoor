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

package obfuscator

import (
	"crypto/cipher"
	"encoding/hex"
	"fmt"
)

// PKCS7 反填充
func unPaddingPKCS7(s []byte) []byte {
	length := len(s)
	if length == 0 {
		return s
	}
	unPadding := int(s[length-1])
	return s[:(length - unPadding)]
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

	return string(unPaddingPKCS7(password)), nil
}
