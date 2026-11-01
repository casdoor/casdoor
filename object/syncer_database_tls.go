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

package object

import (
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/lib/pq"
)

func (syncer *Syncer) hasPostgresCerts() bool {
	return syncer.SslRootCert != "" || syncer.SslCert != "" || syncer.SslKey != ""
}

// registerPostgresTLSConfig registers the syncer's PEM certificates as a custom TLS config
// in lib/pq and returns the sslmode that selects it. The certificates are passed in memory,
// so no PGSSLROOTCERT or ~/.postgresql files are needed and each syncer can trust its own CA.
func (syncer *Syncer) registerPostgresTLSConfig(sslMode string, host string) (string, error) {
	if sslMode != "require" && sslMode != "verify-ca" && sslMode != "verify-full" {
		return "", fmt.Errorf("the SSL certificates of the syncer need the SSL mode: require, verify-ca or verify-full, got: %s", sslMode)
	}
	if (syncer.SslCert == "") != (syncer.SslKey == "") {
		return "", errors.New("the SSL client certificate and the SSL client private key of the syncer must be set together")
	}

	var roots *x509.CertPool
	if syncer.SslRootCert != "" {
		roots = x509.NewCertPool()
		if !roots.AppendCertsFromPEM([]byte(syncer.SslRootCert)) {
			return "", errors.New("the SSL root certificate of the syncer is not a valid PEM certificate")
		}
	}

	// lib/pq skips its own verification for a custom TLS config, and it may still overwrite RootCAs
	// with PGSSLROOTCERT or ~/.postgresql/root.crt, so the chain is verified here against our own pool.
	tlsConfig := &tls.Config{InsecureSkipVerify: true}
	if syncer.SslCert != "" {
		cert, err := tls.X509KeyPair([]byte(syncer.SslCert), []byte(syncer.SslKey))
		if err != nil {
			return "", fmt.Errorf("the SSL client certificate or private key of the syncer is invalid: %s", err.Error())
		}
		tlsConfig.GetClientCertificate = func(*tls.CertificateRequestInfo) (*tls.Certificate, error) {
			return &cert, nil
		}
	}

	// like libpq, sslmode=require with a root certificate verifies the CA
	if sslMode != "require" || roots != nil {
		dnsName := ""
		if sslMode == "verify-full" {
			dnsName = strings.Trim(host, "[]")
		}
		tlsConfig.VerifyConnection = func(state tls.ConnectionState) error {
			if len(state.PeerCertificates) == 0 {
				return errors.New("the PostgreSQL server did not send a certificate")
			}
			opts := x509.VerifyOptions{Roots: roots, DNSName: dnsName, Intermediates: x509.NewCertPool()}
			for _, cert := range state.PeerCertificates[1:] {
				opts.Intermediates.AddCert(cert)
			}
			_, err := state.PeerCertificates[0].Verify(opts)
			return err
		}
	}

	// the key changes with the settings, so a connection test never replaces the config of a running syncer
	hash := sha256.Sum256([]byte(strings.Join([]string{syncer.GetId(), sslMode, host, syncer.SslRootCert, syncer.SslCert, syncer.SslKey}, "\x00")))
	key := "syncer-" + hex.EncodeToString(hash[:16])
	err := pq.RegisterTLSConfig(key, tlsConfig)
	if err != nil {
		return "", err
	}

	return "pqgo-" + key, nil
}
