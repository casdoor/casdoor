// Copyright 2026 The Casdoor Authors. All Rights Reserved.
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

package hsm

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"

	"github.com/golang-jwt/jwt/v5"
)

// jwtSigningMethod signs RS*, PS* and ES* tokens with a crypto.Signer, which golang-jwt's own methods don't take;
// it has the same alg and verifies like them
type jwtSigningMethod struct {
	jwt.SigningMethod
}

// JwtSigningMethod wraps a golang-jwt RSA, RSA-PSS or ECDSA signing method to sign with an HSM signer, the other
// methods are returned as they are
func JwtSigningMethod(method jwt.SigningMethod) jwt.SigningMethod {
	switch method.(type) {
	case *jwt.SigningMethodRSA, *jwt.SigningMethodRSAPSS, *jwt.SigningMethodECDSA:
		return &jwtSigningMethod{SigningMethod: method}
	}
	return method
}

func (m *jwtSigningMethod) Sign(signingString string, key interface{}) ([]byte, error) {
	signer, ok := key.(*Signer)
	if !ok {
		return m.SigningMethod.Sign(signingString, key)
	}

	switch method := m.SigningMethod.(type) {
	case *jwt.SigningMethodRSA:
		if _, ok := signer.Public().(*rsa.PublicKey); !ok {
			return nil, jwt.ErrInvalidKeyType
		}
		return signer.Sign(rand.Reader, hashOf(method.Hash, signingString), method.Hash)
	case *jwt.SigningMethodRSAPSS:
		if _, ok := signer.Public().(*rsa.PublicKey); !ok {
			return nil, jwt.ErrInvalidKeyType
		}
		opts := &rsa.PSSOptions{SaltLength: rsa.PSSSaltLengthEqualsHash, Hash: method.Hash}
		if method.Options != nil {
			opts.SaltLength = method.Options.SaltLength
		}
		return signer.Sign(rand.Reader, hashOf(method.Hash, signingString), opts)
	case *jwt.SigningMethodECDSA:
		public, ok := signer.Public().(*ecdsa.PublicKey)
		if !ok || public.Curve.Params().BitSize != method.CurveBits {
			return nil, jwt.ErrInvalidKeyType
		}
		der, err := signer.Sign(rand.Reader, hashOf(method.Hash, signingString), method.Hash)
		if err != nil {
			return nil, err
		}
		r, s, err := parseEcSignature(der)
		if err != nil {
			return nil, err
		}
		keyBytes := (method.CurveBits + 7) / 8
		res := make([]byte, 2*keyBytes)
		r.FillBytes(res[:keyBytes])
		s.FillBytes(res[keyBytes:])
		return res, nil
	}
	return nil, jwt.ErrInvalidKeyType
}

func hashOf(hash crypto.Hash, data string) []byte {
	hasher := hash.New()
	hasher.Write([]byte(data))
	return hasher.Sum(nil)
}
