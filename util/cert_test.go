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
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"testing"
	"time"
)

// generateTestCertificate generates a self-signed certificate for testing
func generateTestCertificate(notBefore, notAfter time.Time) (*x509.Certificate, string, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, "", err
	}

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Test Org"},
			CommonName:   "Test Client",
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, "", err
	}

	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		return nil, "", err
	}

	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	})

	return cert, string(certPEM), nil
}

func TestParseCertificate(t *testing.T) {
	now := time.Now()
	_, certPEM, err := generateTestCertificate(now, now.Add(365*24*time.Hour))
	if err != nil {
		t.Fatalf("Failed to generate test certificate: %v", err)
	}

	tests := []struct {
		name    string
		certPEM string
		wantErr bool
	}{
		{
			name:    "Valid certificate",
			certPEM: certPEM,
			wantErr: false,
		},
		{
			name:    "Empty certificate",
			certPEM: "",
			wantErr: true,
		},
		{
			name:    "Invalid PEM",
			certPEM: "not a valid PEM",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cert, err := ParseCertificate(tt.certPEM)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseCertificate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && cert == nil {
				t.Error("ParseCertificate() returned nil certificate")
			}
		})
	}
}

func TestValidateCertificate(t *testing.T) {
	now := time.Now()

	// Valid certificate
	validCert, _, err := generateTestCertificate(now.Add(-1*time.Hour), now.Add(365*24*time.Hour))
	if err != nil {
		t.Fatalf("Failed to generate valid certificate: %v", err)
	}

	// Expired certificate
	expiredCert, _, err := generateTestCertificate(now.Add(-365*24*time.Hour), now.Add(-1*time.Hour))
	if err != nil {
		t.Fatalf("Failed to generate expired certificate: %v", err)
	}

	// Not yet valid certificate
	futureNotBeforeCert, _, err := generateTestCertificate(now.Add(1*time.Hour), now.Add(365*24*time.Hour))
	if err != nil {
		t.Fatalf("Failed to generate future certificate: %v", err)
	}

	// Different valid certificate
	differentCert, _, err := generateTestCertificate(now.Add(-1*time.Hour), now.Add(365*24*time.Hour))
	if err != nil {
		t.Fatalf("Failed to generate different certificate: %v", err)
	}

	tests := []struct {
		name       string
		clientCert *x509.Certificate
		storedCert *x509.Certificate
		wantErr    bool
	}{
		{
			name:       "Valid matching certificates",
			clientCert: validCert,
			storedCert: validCert,
			wantErr:    false,
		},
		{
			name:       "Expired client certificate",
			clientCert: expiredCert,
			storedCert: expiredCert,
			wantErr:    true,
		},
		{
			name:       "Not yet valid client certificate",
			clientCert: futureNotBeforeCert,
			storedCert: futureNotBeforeCert,
			wantErr:    true,
		},
		{
			name:       "Certificates don't match",
			clientCert: validCert,
			storedCert: differentCert,
			wantErr:    true,
		},
		{
			name:       "Nil client certificate",
			clientCert: nil,
			storedCert: validCert,
			wantErr:    true,
		},
		{
			name:       "Nil stored certificate",
			clientCert: validCert,
			storedCert: nil,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCertificate(tt.clientCert, tt.storedCert)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCertificate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetCertificateFingerprint(t *testing.T) {
	now := time.Now()
	cert, _, err := generateTestCertificate(now, now.Add(365*24*time.Hour))
	if err != nil {
		t.Fatalf("Failed to generate test certificate: %v", err)
	}

	tests := []struct {
		name string
		cert *x509.Certificate
		want string
	}{
		{
			name: "Valid certificate",
			cert: cert,
			want: "", // We just check it's not empty
		},
		{
			name: "Nil certificate",
			cert: nil,
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetCertificateFingerprint(tt.cert)
			if tt.cert == nil && got != "" {
				t.Error("GetCertificateFingerprint() should return empty string for nil certificate")
			}
			if tt.cert != nil && got == "" {
				t.Error("GetCertificateFingerprint() should return non-empty string for valid certificate")
			}
		})
	}
}
