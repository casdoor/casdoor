// Copyright 2023 The casbin Authors. All Rights Reserved.
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
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"strings"
)

// generateEccKey generates a public and private key pair.(NIST P-256)
func generateEccKey() *ecdsa.PrivateKey {
	privateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	return privateKey
}

// encodeEccKey Return the input private key object as string type private key
func encodeEccKey(privateKey *ecdsa.PrivateKey) string {
	x509Encoded, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		panic(err)
	}

	pemEncoded := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: x509Encoded})
	return string(pemEncoded)
}

// decodeEccKey Return the entered private key string as a private key object that can be used
func decodeEccKey(pemEncoded string) (*ecdsa.PrivateKey, error) {
	pemEncoded = strings.ReplaceAll(pemEncoded, "\\n", "\n")
	block, _ := pem.Decode([]byte(pemEncoded))
	if block == nil {
		return nil, fmt.Errorf("decodeEccKey() error, block should not be nil")
	}

	x509Encoded := block.Bytes
	privateKey, err := x509.ParseECPrivateKey(x509Encoded)
	if err != nil {
		return nil, err
	}

	return privateKey, nil
}
