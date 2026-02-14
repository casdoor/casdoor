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
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"time"
)

// ParseCertificate parses a PEM-encoded certificate string into an x509.Certificate
func ParseCertificate(certPEM string) (*x509.Certificate, error) {
	if certPEM == "" {
		return nil, fmt.Errorf("certificate is empty")
	}

	block, _ := pem.Decode([]byte(certPEM))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM certificate")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	return cert, nil
}

// ValidateCertificate validates a client certificate against a stored certificate
// It checks:
// - Certificate expiration
// - Certificate matches (by comparing public keys or full certificate)
func ValidateCertificate(clientCert, storedCert *x509.Certificate) error {
	if clientCert == nil {
		return fmt.Errorf("client certificate is nil")
	}

	if storedCert == nil {
		return fmt.Errorf("stored certificate is nil")
	}

	// Check if certificate is expired
	now := time.Now()
	if now.Before(clientCert.NotBefore) {
		return fmt.Errorf("client certificate is not yet valid (NotBefore: %v)", clientCert.NotBefore)
	}
	if now.After(clientCert.NotAfter) {
		return fmt.Errorf("client certificate has expired (NotAfter: %v)", clientCert.NotAfter)
	}

	// Compare certificates by their fingerprint (more reliable than just public key)
	if !clientCert.Equal(storedCert) {
		return fmt.Errorf("client certificate does not match stored certificate")
	}

	return nil
}

// GetCertificateFingerprint returns a string representation of the certificate for logging
func GetCertificateFingerprint(cert *x509.Certificate) string {
	if cert == nil {
		return ""
	}
	return fmt.Sprintf("Subject=%s, SerialNumber=%s, NotAfter=%v",
		cert.Subject.String(), cert.SerialNumber.String(), cert.NotAfter)
}
