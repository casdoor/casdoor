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
	"crypto/tls"
	"crypto/x509"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
)

func (syncer *Syncer) getPostgresTLSConfig(dataSourceName string) (*pgx.ConnConfig, error) {
	if syncer.SslMode == "" || syncer.SslMode == "disable" {
		return nil, errors.New("SSL certificates require an enabled SSL mode")
	}
	if (syncer.SslCert == "") != (syncer.SslKey == "") {
		return nil, errors.New("both SSL client certificate and private key are required")
	}

	// Inline credentials must not inherit certificate file paths from the process.
	dataSourceName += " sslrootcert='' sslcert='' sslkey=''"
	if syncer.SslMode == "require" && syncer.SslRootCert != "" {
		dataSourceName += " sslmode=verify-ca"
	}
	if !isCloudIntranet {
		dataSourceName = strings.ReplaceAll(dataSourceName, "dbi.", "db.")
	}
	config, err := pgx.ParseConfig(dataSourceName)
	if err != nil {
		return nil, errors.New("invalid PostgreSQL connection configuration")
	}
	if config.TLSConfig == nil {
		return nil, errors.New("SSL certificates require a TCP connection with TLS")
	}
	config.Fallbacks = nil
	if syncer.SslRootCert != "" {
		roots := x509.NewCertPool()
		if !roots.AppendCertsFromPEM([]byte(syncer.SslRootCert)) {
			return nil, errors.New("invalid SSL root certificate PEM")
		}
		config.TLSConfig.RootCAs = roots
	}
	if syncer.SslCert != "" {
		cert, err := tls.X509KeyPair([]byte(syncer.SslCert), []byte(syncer.SslKey))
		if err != nil {
			return nil, errors.New("invalid SSL client certificate or private key PEM")
		}
		config.TLSConfig.Certificates = []tls.Certificate{cert}
	}
	return config, nil
}
