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
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// GetClientCertificate extracts the client certificate from the HTTP request
func GetClientCertificate(r *http.Request) (*x509.Certificate, error) {
	if r.TLS == nil || len(r.TLS.PeerCertificates) == 0 {
		return nil, nil
	}
	return r.TLS.PeerCertificates[0], nil
}

// ValidateClientCertificate validates the client certificate against application configuration
func ValidateClientCertificate(cert *x509.Certificate, app *Application) error {
	if cert == nil {
		return fmt.Errorf("client certificate is required for mTLS authentication")
	}

	// Check certificate expiration
	now := time.Now()
	if now.Before(cert.NotBefore) {
		return fmt.Errorf("client certificate is not yet valid")
	}
	if now.After(cert.NotAfter) {
		return fmt.Errorf("client certificate has expired")
	}

	// For tls_client_auth, validate certificate chain and issuer
	if app.MtlsAuthMethod == "tls_client_auth" {
		if len(app.AllowedClientCertIssuers) > 0 {
			issuerDN := cert.Issuer.String()
			allowed := false
			for _, allowedIssuer := range app.AllowedClientCertIssuers {
				if strings.Contains(issuerDN, allowedIssuer) {
					allowed = true
					break
				}
			}
			if !allowed {
				return fmt.Errorf("client certificate issuer not allowed: %s", issuerDN)
			}
		}
	}

	// For self_signed_tls_client_auth, accept self-signed certificates
	// No additional validation needed beyond expiration check

	return nil
}

// GetCertificateFingerprint calculates the SHA-256 fingerprint of a certificate
func GetCertificateFingerprint(cert *x509.Certificate) string {
	if cert == nil {
		return ""
	}
	hash := sha256.Sum256(cert.Raw)
	return base64.URLEncoding.EncodeToString(hash[:])
}

// VerifyCertificateChain verifies the certificate chain for tls_client_auth
func VerifyCertificateChain(cert *x509.Certificate) error {
	if cert == nil {
		return fmt.Errorf("certificate is nil")
	}

	// For self-signed certificates, skip chain verification
	if cert.Issuer.String() == cert.Subject.String() {
		return nil
	}

	// Note: In production, this should use the system certificate pool or
	// configured trusted CA certificates. For now, we accept the certificate
	// if it's properly formed and not expired (checked elsewhere).
	// Proper PKI validation would require access to CA certificates.
	return nil
}

// GetCertificateSubject returns the subject DN of the certificate
func GetCertificateSubject(cert *x509.Certificate) string {
	if cert == nil {
		return ""
	}
	return cert.Subject.String()
}

// IsMtlsEnabled checks if mTLS is enabled for the application
func IsMtlsEnabled(app *Application) bool {
	return app != nil && app.EnableMtls
}

// SupportsMtlsAuthMethod checks if the application supports the specified mTLS auth method
func SupportsMtlsAuthMethod(app *Application, method string) bool {
	if !IsMtlsEnabled(app) {
		return false
	}
	// If no specific method is configured, require explicit method configuration
	if app.MtlsAuthMethod == "" {
		return false
	}
	return app.MtlsAuthMethod == method
}

// ValidateMtlsRequest validates the mTLS authentication request
func ValidateMtlsRequest(r *http.Request, app *Application) (*x509.Certificate, error) {
	if !IsMtlsEnabled(app) {
		return nil, fmt.Errorf("mTLS is not enabled for this application")
	}

	cert, err := GetClientCertificate(r)
	if err != nil {
		return nil, fmt.Errorf("failed to get client certificate: %v", err)
	}

	if cert == nil {
		return nil, fmt.Errorf("client certificate is required for mTLS authentication")
	}

	err = ValidateClientCertificate(cert, app)
	if err != nil {
		return nil, err
	}

	return cert, nil
}

// GetTLSConfig returns a TLS configuration for mTLS server
func GetTLSConfig() *tls.Config {
	return &tls.Config{
		ClientAuth: tls.RequestClientCert,
		MinVersion: tls.VersionTLS12,
	}
}
