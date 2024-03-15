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

package object

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"time"
)

func generateRsaKeys(bitSize int, shaSize int, expireInYears int, commonName string, organization string) (string, string, error) {
	// https://stackoverflow.com/questions/64104586/use-golang-to-get-rsa-key-the-same-way-openssl-genrsa
	// https://stackoverflow.com/questions/43822945/golang-can-i-create-x509keypair-using-rsa-key

	// Generate RSA key.
	key, err := rsa.GenerateKey(rand.Reader, bitSize)
	if err != nil {
		return "", "", err
	}

	// Encode private key to PKCS#1 ASN.1 PEM.
	privateKeyPem := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(key),
		},
	)

	tml := x509.Certificate{
		// you can add any attr that you need
		NotBefore: time.Now(),
		NotAfter:  time.Now().AddDate(expireInYears, 0, 0),
		// you have to generate a different serial number each execution
		SerialNumber: big.NewInt(123456),
		Subject: pkix.Name{
			CommonName:   commonName,
			Organization: []string{organization},
		},
		BasicConstraintsValid: true,
	}

	switch shaSize {
	case 256:
		tml.SignatureAlgorithm = x509.SHA256WithRSA
	case 384:
		tml.SignatureAlgorithm = x509.SHA384WithRSA
	case 512:
		tml.SignatureAlgorithm = x509.SHA512WithRSA
	default:
		return "", "", fmt.Errorf("generateRsaKeys() error, unsupported SHA size: %d", shaSize)
	}

	cert, err := x509.CreateCertificate(rand.Reader, &tml, &tml, &key.PublicKey, key)
	if err != nil {
		return "", "", err
	}

	// Generate a pem block with the certificate
	certPem := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert,
	})

	return string(certPem), string(privateKeyPem), nil
}

func generateEsKeys(shaSize int, expireInYears int, commonName string, organization string) (string, string, error) {
	var curve elliptic.Curve
	switch shaSize {
	case 256:
		curve = elliptic.P256()
	case 384:
		curve = elliptic.P384()
	case 512:
		curve = elliptic.P521() // ES512(P521,SHA512)
	default:
		return "", "", fmt.Errorf("generateEsKeys() error, unsupported SHA size: %d", shaSize)
	}

	// Generate ECDSA key pair.
	privateKey, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		return "", "", err
	}

	// Encode private key to PEM format.
	privateKeyBytes, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		return "", "", err
	}
	privateKeyPem := pem.EncodeToMemory(&pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	// Generate certificate template.
	template := x509.Certificate{
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(expireInYears, 0, 0),
		SerialNumber: big.NewInt(time.Now().Unix()),
		Subject: pkix.Name{
			CommonName:   commonName,
			Organization: []string{organization},
		},
		BasicConstraintsValid: true,
	}

	// Generate certificate.
	certBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return "", "", err
	}

	// Encode certificate to PEM format.
	certPem := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})

	return string(certPem), string(privateKeyPem), nil
}

func generateRsaPssKeys(bitSize int, shaSize int, expireInYears int, commonName string, organization string) (string, string, error) {
	// Generate RSA key.
	key, err := rsa.GenerateKey(rand.Reader, bitSize)
	if err != nil {
		return "", "", err
	}

	// Encode private key to PKCS#8 ASN.1 PEM.
	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return "", "", err
	}

	privateKeyPem := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PSS PRIVATE KEY",
			Bytes: privateKeyBytes,
		},
	)

	tml := x509.Certificate{
		// you can add any attr that you need
		NotBefore: time.Now(),
		NotAfter:  time.Now().AddDate(expireInYears, 0, 0),
		// you have to generate a different serial number each execution
		SerialNumber: big.NewInt(123456),
		Subject: pkix.Name{
			CommonName:   commonName,
			Organization: []string{organization},
		},
		BasicConstraintsValid: true,
	}

	// Set the signature algorithm based on the hash function
	switch shaSize {
	case 256:
		tml.SignatureAlgorithm = x509.SHA256WithRSAPSS
	case 384:
		tml.SignatureAlgorithm = x509.SHA384WithRSAPSS
	case 512:
		tml.SignatureAlgorithm = x509.SHA512WithRSAPSS
	default:
		return "", "", fmt.Errorf("generateRsaPssKeys() error, unsupported SHA size: %d", shaSize)
	}

	cert, err := x509.CreateCertificate(rand.Reader, &tml, &tml, &key.PublicKey, key)
	if err != nil {
		return "", "", err
	}

	// Generate a pem block with the certificate
	certPem := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert,
	})

	return string(certPem), string(privateKeyPem), nil
}
