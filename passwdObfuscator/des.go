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

package passwdObfuscator

import (
	"crypto/cipher"
	"crypto/des"
	"encoding/hex"
	"fmt"
)

type DESObfuscator struct {
	key string
}

func NewDESObfuscator(key string) *DESObfuscator {
	obfuscator := &DESObfuscator{key: key}
	return obfuscator
}

func (obfuscator *DESObfuscator) Decrypte(passwdCipherStr string) (string, error) {
	key, err := hex.DecodeString(obfuscator.key)
	if err != nil {
		return "", err
	}

	passwdCipherBytes, err := hex.DecodeString(passwdCipherStr)
	if err != nil {
		return "", err
	}

	block, err := des.NewCipher(key)
	if err != nil {
		return "", err
	}

	if len(passwdCipherBytes) < block.BlockSize() {
		return "", fmt.Errorf("ciphertext too short")
	}

	iv := passwdCipherBytes[:block.BlockSize()]
	password := make([]byte, len(passwdCipherBytes)-block.BlockSize())

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(password, passwdCipherBytes[block.BlockSize():])

	return string(unPaddingPKCS7(password)), nil
}
