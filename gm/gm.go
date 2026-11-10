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

// Package gm has the Chinese national cryptographic algorithms: SM2 signatures for tokens and certificates, SM3
// hashes and SM4 encryption
package gm

import (
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/emmansun/gmsm/sm2"
	"github.com/emmansun/gmsm/sm3"
	"github.com/emmansun/gmsm/sm4"
	"github.com/emmansun/gmsm/smx509"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/cryptobyte"
	"golang.org/x/crypto/cryptobyte/asn1"
)

// Alg is the JWS algorithm of SM2 signatures: SM3 with SM2 and the default user ID 1234567812345678 (GB/T 32918),
// the signature being r and s of 32 bytes each like ES256
const Alg = "SM2"

const sm2KeySize = 32

type signingMethodSm2 struct{}

var SigningMethodSm2 jwt.SigningMethod = &signingMethodSm2{}

func init() {
	jwt.RegisterSigningMethod(Alg, func() jwt.SigningMethod {
		return SigningMethodSm2
	})
}

func (m *signingMethodSm2) Alg() string {
	return Alg
}

func (m *signingMethodSm2) Sign(signingString string, key interface{}) ([]byte, error) {
	privateKey, ok := key.(*sm2.PrivateKey)
	if !ok {
		return nil, jwt.ErrInvalidKeyType
	}

	asn1Signature, err := privateKey.Sign(rand.Reader, []byte(signingString), sm2.DefaultSM2SignerOpts)
	if err != nil {
		return nil, err
	}

	r, s, err := parseAsn1Signature(asn1Signature)
	if err != nil {
		return nil, err
	}

	res := make([]byte, 2*sm2KeySize)
	r.FillBytes(res[:sm2KeySize])
	s.FillBytes(res[sm2KeySize:])
	return res, nil
}

func (m *signingMethodSm2) Verify(signingString string, signature []byte, key interface{}) error {
	publicKey, ok := key.(*ecdsa.PublicKey)
	if !ok || !sm2.IsSM2PublicKey(publicKey) {
		return jwt.ErrInvalidKeyType
	}
	if len(signature) != 2*sm2KeySize {
		return jwt.ErrSignatureInvalid
	}

	r := new(big.Int).SetBytes(signature[:sm2KeySize])
	s := new(big.Int).SetBytes(signature[sm2KeySize:])
	asn1Signature, err := marshalAsn1Signature(r, s)
	if err != nil {
		return err
	}

	if !sm2.VerifyASN1WithSM2(publicKey, nil, []byte(signingString), asn1Signature) {
		return jwt.ErrSignatureInvalid
	}
	return nil
}

func parseAsn1Signature(signature []byte) (*big.Int, *big.Int, error) {
	r, s := new(big.Int), new(big.Int)
	var inner cryptobyte.String
	input := cryptobyte.String(signature)
	if !input.ReadASN1(&inner, asn1.SEQUENCE) || !input.Empty() || !inner.ReadASN1Integer(r) || !inner.ReadASN1Integer(s) || !inner.Empty() {
		return nil, nil, errors.New("invalid SM2 signature")
	}
	return r, s, nil
}

func marshalAsn1Signature(r *big.Int, s *big.Int) ([]byte, error) {
	var b cryptobyte.Builder
	b.AddASN1(asn1.SEQUENCE, func(b *cryptobyte.Builder) {
		b.AddASN1BigInt(r)
		b.AddASN1BigInt(s)
	})
	return b.Bytes()
}

// GenerateCertificate creates an SM2 key pair and a self-signed certificate signed with SM3 with SM2, both in PEM
func GenerateCertificate(expireInYears int, commonName string, organization string) (string, string, error) {
	privateKey, err := sm2.GenerateKey(rand.Reader)
	if err != nil {
		return "", "", err
	}

	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return "", "", err
	}

	template := &smx509.Certificate{
		SerialNumber:          serialNumber,
		Subject:               pkix.Name{CommonName: commonName, Organization: []string{organization}},
		NotBefore:             time.Now().Add(-time.Minute),
		NotAfter:              time.Now().AddDate(expireInYears, 0, 0),
		KeyUsage:              smx509.KeyUsageDigitalSignature | smx509.KeyUsageCertSign,
		IsCA:                  true,
		BasicConstraintsValid: true,
	}

	certDer, err := smx509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return "", "", err
	}

	keyDer, err := smx509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return "", "", err
	}

	certPem := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDer})
	keyPem := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyDer})
	return string(certPem), string(keyPem), nil
}

// ParsePrivateKey reads an SM2 private key in PEM, PKCS #8 or SEC 1
func ParsePrivateKey(privateKeyPem string) (*sm2.PrivateKey, error) {
	block, _ := pem.Decode([]byte(privateKeyPem))
	if block == nil {
		return nil, errors.New("the SM2 private key is not in PEM")
	}

	key, err := smx509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		key, err = smx509.ParseSM2PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
	}

	privateKey, ok := key.(*sm2.PrivateKey)
	if !ok {
		return nil, errors.New("the private key is not an SM2 key")
	}
	return privateKey, nil
}

// ParseCertificate reads an SM2 certificate in PEM
func ParseCertificate(certificatePem string) (*smx509.Certificate, error) {
	block, _ := pem.Decode([]byte(certificatePem))
	if block == nil {
		return nil, errors.New("the SM2 certificate is not in PEM")
	}
	return smx509.ParseCertificate(block.Bytes)
}

// ParsePublicKey reads the SM2 public key of a certificate or of a PUBLIC KEY block in PEM
func ParsePublicKey(certificatePem string) (*ecdsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(certificatePem))
	if block == nil {
		return nil, errors.New("the SM2 certificate is not in PEM")
	}

	var key any
	var err error
	if block.Type == "CERTIFICATE" {
		var certificate *smx509.Certificate
		certificate, err = smx509.ParseCertificate(block.Bytes)
		if err == nil {
			key = certificate.PublicKey
		}
	} else {
		key, err = smx509.ParsePKIXPublicKey(block.Bytes)
	}
	if err != nil {
		return nil, err
	}

	publicKey, ok := key.(*ecdsa.PublicKey)
	if !ok || !sm2.IsSM2PublicKey(publicKey) {
		return nil, errors.New("the public key is not an SM2 key")
	}
	return publicKey, nil
}

// IsSm2Certificate tells whether the PEM certificate has an SM2 public key
func IsSm2Certificate(certificatePem string) bool {
	_, err := ParsePublicKey(certificatePem)
	return err == nil
}

// GetJsonWebKey returns the JWK of an SM2 certificate, kty "EC" with crv "SM2"
func GetJsonWebKey(certificatePem string, keyId string) (map[string]interface{}, error) {
	certificate, err := ParseCertificate(certificatePem)
	if err != nil {
		return nil, err
	}
	publicKey, ok := certificate.PublicKey.(*ecdsa.PublicKey)
	if !ok || !sm2.IsSM2PublicKey(publicKey) {
		return nil, errors.New("the certificate is not an SM2 certificate")
	}

	x := make([]byte, sm2KeySize)
	y := make([]byte, sm2KeySize)
	publicKey.X.FillBytes(x)
	publicKey.Y.FillBytes(y)
	return map[string]interface{}{
		"kty": "EC",
		"crv": "SM2",
		"x":   base64.RawURLEncoding.EncodeToString(x),
		"y":   base64.RawURLEncoding.EncodeToString(y),
		"kid": keyId,
		"alg": Alg,
		"use": "sig",
		"x5c": []string{base64.StdEncoding.EncodeToString(certificate.Raw)},
	}, nil
}

// GetCertExpireTime returns the NotAfter of an SM2 certificate in RFC 3339
func GetCertExpireTime(certificatePem string) (string, error) {
	certificate, err := ParseCertificate(certificatePem)
	if err != nil {
		return "", err
	}
	return certificate.NotAfter.Local().Format(time.RFC3339), nil
}

// Sm3Hex returns the SM3 hash of data in hex
func Sm3Hex(data []byte) string {
	sum := sm3.Sum(data)
	return hex.EncodeToString(sum[:])
}

// NewSm4Gcm returns SM4 in GCM mode for a 16-byte key
func NewSm4Gcm(key []byte) (cipher.AEAD, error) {
	if len(key) != sm4.BlockSize {
		return nil, fmt.Errorf("an SM4 key is %d bytes, not %d", sm4.BlockSize, len(key))
	}
	block, err := sm4.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}
