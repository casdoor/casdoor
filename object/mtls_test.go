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
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"net/http"
	"testing"
	"time"
)

// Helper function to generate a test certificate
func generateTestCertificate() (*x509.Certificate, *rsa.PrivateKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}

	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, nil, err
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Test Organization"},
			CommonName:   "Test Client",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, nil, err
	}

	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		return nil, nil, err
	}

	return cert, privateKey, nil
}

func TestGetCertificateFingerprint(t *testing.T) {
	cert, _, err := generateTestCertificate()
	if err != nil {
		t.Fatalf("Failed to generate test certificate: %v", err)
	}

	fingerprint := GetCertificateFingerprint(cert)
	if fingerprint == "" {
		t.Error("Expected non-empty fingerprint")
	}

	// Test with nil certificate
	nilFingerprint := GetCertificateFingerprint(nil)
	if nilFingerprint != "" {
		t.Error("Expected empty fingerprint for nil certificate")
	}

	// Fingerprint should be consistent
	fingerprint2 := GetCertificateFingerprint(cert)
	if fingerprint != fingerprint2 {
		t.Error("Fingerprint should be consistent for same certificate")
	}
}

func TestValidateClientCertificate(t *testing.T) {
	cert, _, err := generateTestCertificate()
	if err != nil {
		t.Fatalf("Failed to generate test certificate: %v", err)
	}

	tests := []struct {
		name        string
		cert        *x509.Certificate
		app         *Application
		shouldError bool
	}{
		{
			name:        "nil certificate",
			cert:        nil,
			app:         &Application{EnableMtls: true, MtlsAuthMethod: "tls_client_auth"},
			shouldError: true,
		},
		{
			name: "valid certificate - self_signed_tls_client_auth",
			cert: cert,
			app:  &Application{EnableMtls: true, MtlsAuthMethod: "self_signed_tls_client_auth"},
			shouldError: false,
		},
		{
			name: "valid certificate - tls_client_auth without issuer check",
			cert: cert,
			app:  &Application{EnableMtls: true, MtlsAuthMethod: "tls_client_auth", AllowedClientCertIssuers: []string{}},
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateClientCertificate(tt.cert, tt.app)
			if tt.shouldError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.shouldError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestIsMtlsEnabled(t *testing.T) {
	tests := []struct {
		name string
		app  *Application
		want bool
	}{
		{
			name: "mTLS enabled",
			app:  &Application{EnableMtls: true},
			want: true,
		},
		{
			name: "mTLS disabled",
			app:  &Application{EnableMtls: false},
			want: false,
		},
		{
			name: "nil application",
			app:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsMtlsEnabled(tt.app); got != tt.want {
				t.Errorf("IsMtlsEnabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSupportsMtlsAuthMethod(t *testing.T) {
	tests := []struct {
		name   string
		app    *Application
		method string
		want   bool
	}{
		{
			name:   "supports tls_client_auth",
			app:    &Application{EnableMtls: true, MtlsAuthMethod: "tls_client_auth"},
			method: "tls_client_auth",
			want:   true,
		},
		{
			name:   "supports self_signed_tls_client_auth",
			app:    &Application{EnableMtls: true, MtlsAuthMethod: "self_signed_tls_client_auth"},
			method: "self_signed_tls_client_auth",
			want:   true,
		},
		{
			name:   "method mismatch",
			app:    &Application{EnableMtls: true, MtlsAuthMethod: "tls_client_auth"},
			method: "self_signed_tls_client_auth",
			want:   false,
		},
		{
			name:   "mTLS disabled",
			app:    &Application{EnableMtls: false, MtlsAuthMethod: "tls_client_auth"},
			method: "tls_client_auth",
			want:   false,
		},
		{
			name:   "no method configured",
			app:    &Application{EnableMtls: true, MtlsAuthMethod: ""},
			method: "tls_client_auth",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SupportsMtlsAuthMethod(tt.app, tt.method); got != tt.want {
				t.Errorf("SupportsMtlsAuthMethod() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetClientCertificate(t *testing.T) {
	cert, key, err := generateTestCertificate()
	if err != nil {
		t.Fatalf("Failed to generate test certificate: %v", err)
	}

	// Test with TLS request containing certificate
	req := &http.Request{
		TLS: &tls.ConnectionState{
			PeerCertificates: []*x509.Certificate{cert},
		},
	}

	retrievedCert, err := GetClientCertificate(req)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if retrievedCert == nil {
		t.Error("Expected certificate but got nil")
	}
	if retrievedCert != cert {
		t.Error("Retrieved certificate doesn't match original")
	}

	// Test with request without TLS
	req2 := &http.Request{}
	retrievedCert2, err := GetClientCertificate(req2)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if retrievedCert2 != nil {
		t.Error("Expected nil certificate for non-TLS request")
	}

	// Suppress unused variable warning
	_ = key
}

func TestGetCertificateSubject(t *testing.T) {
	cert, _, err := generateTestCertificate()
	if err != nil {
		t.Fatalf("Failed to generate test certificate: %v", err)
	}

	subject := GetCertificateSubject(cert)
	if subject == "" {
		t.Error("Expected non-empty subject")
	}

	if subject != cert.Subject.String() {
		t.Errorf("Subject mismatch: got %s, want %s", subject, cert.Subject.String())
	}

	// Test with nil certificate
	nilSubject := GetCertificateSubject(nil)
	if nilSubject != "" {
		t.Error("Expected empty subject for nil certificate")
	}
}

func TestGetClientCredentialsTokenWithCert(t *testing.T) {
	// This test requires database initialization, so we skip it in unit tests
	// It should be tested in integration tests
	t.Skip("Skipping database-dependent test")

	app := &Application{
		Owner:          "test-org",
		Name:           "test-app",
		EnableMtls:     true,
		MtlsAuthMethod: "tls_client_auth",
		ExpireInHours:  1,
	}

	cert, _, err := generateTestCertificate()
	if err != nil {
		t.Fatalf("Failed to generate test certificate: %v", err)
	}

	fingerprint := GetCertificateFingerprint(cert)

	token, tokenError, err := GetClientCredentialsTokenWithCert(app, "", "read write", "localhost", fingerprint)
	
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if tokenError != nil {
		t.Errorf("Unexpected token error: %v", tokenError)
	}
	if token != nil && token.CertFingerprint != fingerprint {
		t.Errorf("Token fingerprint mismatch: got %s, want %s", token.CertFingerprint, fingerprint)
	}
}

func TestGetTLSConfig(t *testing.T) {
	config := GetTLSConfig()
	if config == nil {
		t.Error("Expected non-nil TLS config")
	}
	if config.ClientAuth != tls.RequestClientCert {
		t.Errorf("Expected ClientAuth to be RequestClientCert, got %v", config.ClientAuth)
	}
	if config.MinVersion != tls.VersionTLS12 {
		t.Errorf("Expected MinVersion to be TLS 1.2, got %v", config.MinVersion)
	}
}
