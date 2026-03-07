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
	"crypto/x509"
	"encoding/pem"
	"fmt"
)

// ValidateClientCertForApplication validates the given client certificate
// against the application's configured Cert. It checks whether the client
// certificate was signed by the CA certificate stored in the application's cert,
// or whether it matches the certificate directly (self-signed case).
func ValidateClientCertForApplication(clientCert *x509.Certificate, application *Application) error {
	if application.Cert == "" {
		return fmt.Errorf("application %s has no cert configured", application.Name)
	}

	cert, err := getCertByApplication(application)
	if err != nil {
		return fmt.Errorf("failed to get cert for application %s: %w", application.Name, err)
	}
	if cert == nil {
		return fmt.Errorf("cert not found for application %s", application.Name)
	}

	if cert.Certificate == "" {
		return fmt.Errorf("certificate field is empty for cert: %s", cert.Name)
	}

	return verifyClientCertAgainstCACert(clientCert, cert.Certificate)
}

// verifyClientCertAgainstCACert verifies the client certificate against the CA certificate PEM.
// It supports both CA-signed and self-signed certificate validation.
func verifyClientCertAgainstCACert(clientCert *x509.Certificate, caCertPem string) error {
	block, _ := pem.Decode([]byte(caCertPem))
	if block == nil {
		return fmt.Errorf("failed to decode CA certificate PEM")
	}

	caCert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse CA certificate: %w", err)
	}

	roots := x509.NewCertPool()
	roots.AddCert(caCert)

	_, err = clientCert.Verify(x509.VerifyOptions{
		Roots:     roots,
		KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageAny},
	})
	if err != nil {
		return fmt.Errorf("client certificate verification failed: %w", err)
	}

	return nil
}

// GetApplicationByClientCert validates a client certificate and returns the matching application.
// The clientId is used to look up the application, and the certificate is validated against
// the application's configured Cert.
func GetApplicationByClientCert(clientId string, clientCert *x509.Certificate) (*Application, error) {
	if clientId == "" {
		return nil, fmt.Errorf("clientId is required for mTLS authentication")
	}
	if clientCert == nil {
		return nil, fmt.Errorf("client certificate is required for mTLS authentication")
	}

	application, err := GetApplicationByClientId(clientId)
	if err != nil {
		return nil, err
	}
	if application == nil {
		return nil, fmt.Errorf("application not found for client ID: %s", clientId)
	}

	err = ValidateClientCertForApplication(clientCert, application)
	if err != nil {
		return nil, fmt.Errorf("mTLS authentication failed for application %s: %w", application.Name, err)
	}

	return application, nil
}
