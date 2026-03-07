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

package object

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func generateTestCACert() (*x509.Certificate, []byte, interface{}, error) {
	caKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, nil, err
	}

	caTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName:   "Test CA",
			Organization: []string{"Casdoor Test"},
		},
		NotBefore:             time.Now().Add(-1 * time.Hour),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	caCertDER, err := x509.CreateCertificate(rand.Reader, caTemplate, caTemplate, &caKey.PublicKey, caKey)
	if err != nil {
		return nil, nil, nil, err
	}

	caCert, err := x509.ParseCertificate(caCertDER)
	if err != nil {
		return nil, nil, nil, err
	}

	caCertPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: caCertDER})

	return caCert, caCertPEM, caKey, nil
}

func generateTestClientCert(caCert *x509.Certificate, caKey interface{}) (*x509.Certificate, error) {
	clientKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}

	clientTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject: pkix.Name{
			CommonName:   "Test Client",
			Organization: []string{"Casdoor Test Client"},
		},
		NotBefore: time.Now().Add(-1 * time.Hour),
		NotAfter:  time.Now().Add(24 * time.Hour),
		KeyUsage:  x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageClientAuth,
		},
	}

	clientCertDER, err := x509.CreateCertificate(rand.Reader, clientTemplate, caCert, &clientKey.PublicKey, caKey)
	if err != nil {
		return nil, err
	}

	clientCert, err := x509.ParseCertificate(clientCertDER)
	if err != nil {
		return nil, err
	}

	return clientCert, nil
}

func TestVerifyClientCertAgainstCACert(t *testing.T) {
	caCert, caCertPEM, caKey, err := generateTestCACert()
	assert.Nil(t, err)

	t.Run("valid client cert signed by CA", func(t *testing.T) {
		clientCert, err := generateTestClientCert(caCert, caKey)
		assert.Nil(t, err)

		err = verifyClientCertAgainstCACert(clientCert, string(caCertPEM))
		assert.Nil(t, err)
	})

	t.Run("self-signed cert validates against itself", func(t *testing.T) {
		err = verifyClientCertAgainstCACert(caCert, string(caCertPEM))
		assert.Nil(t, err)
	})

	t.Run("client cert from different CA fails", func(t *testing.T) {
		otherCACert, _, otherCAKey, err := generateTestCACert()
		assert.Nil(t, err)

		clientCert, err := generateTestClientCert(otherCACert, otherCAKey)
		assert.Nil(t, err)

		err = verifyClientCertAgainstCACert(clientCert, string(caCertPEM))
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "client certificate verification failed")
	})

	t.Run("invalid PEM data fails", func(t *testing.T) {
		clientCert, err := generateTestClientCert(caCert, caKey)
		assert.Nil(t, err)

		err = verifyClientCertAgainstCACert(clientCert, "not-a-valid-pem")
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "failed to decode CA certificate PEM")
	})

	t.Run("expired client cert fails", func(t *testing.T) {
		clientKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		assert.Nil(t, err)

		expiredTemplate := &x509.Certificate{
			SerialNumber: big.NewInt(3),
			Subject: pkix.Name{
				CommonName: "Expired Client",
			},
			NotBefore: time.Now().Add(-48 * time.Hour),
			NotAfter:  time.Now().Add(-24 * time.Hour),
			KeyUsage:  x509.KeyUsageDigitalSignature,
			ExtKeyUsage: []x509.ExtKeyUsage{
				x509.ExtKeyUsageClientAuth,
			},
		}

		expiredCertDER, err := x509.CreateCertificate(rand.Reader, expiredTemplate, caCert, &clientKey.PublicKey, caKey)
		assert.Nil(t, err)

		expiredCert, err := x509.ParseCertificate(expiredCertDER)
		assert.Nil(t, err)

		err = verifyClientCertAgainstCACert(expiredCert, string(caCertPEM))
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "client certificate verification failed")
	})
}
