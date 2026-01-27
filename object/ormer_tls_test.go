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
	"os"
	"testing"

	"github.com/beego/beego/v2/server/web"
	"github.com/stretchr/testify/assert"
)

func TestAppendMySQLTLSParam(t *testing.T) {
	// Load config
	err := web.LoadAppConfig("ini", "../conf/app.conf")
	assert.Nil(t, err)

	scenarios := []struct {
		description    string
		dsn            string
		caCert         string
		clientCert     string
		clientKey      string
		expectedSuffix string
	}{
		{
			description:    "No TLS certificates configured",
			dsn:            "root:password@tcp(localhost:3306)/",
			caCert:         "",
			clientCert:     "",
			clientKey:      "",
			expectedSuffix: "",
		},
		{
			description:    "TLS certificates configured - DSN without params",
			dsn:            "root:password@tcp(localhost:3306)/",
			caCert:         "/path/to/ca.pem",
			clientCert:     "/path/to/client.pem",
			clientKey:      "/path/to/client-key.pem",
			expectedSuffix: "?tls=custom-mtls",
		},
		{
			description:    "TLS certificates configured - DSN with existing params",
			dsn:            "root:password@tcp(localhost:3306)/?charset=utf8mb4",
			caCert:         "/path/to/ca.pem",
			clientCert:     "/path/to/client.pem",
			clientKey:      "/path/to/client-key.pem",
			expectedSuffix: "&tls=custom-mtls",
		},
		{
			description:    "Only CA cert configured",
			dsn:            "root:password@tcp(localhost:3306)/",
			caCert:         "/path/to/ca.pem",
			clientCert:     "",
			clientKey:      "",
			expectedSuffix: "?tls=custom-mtls",
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.description, func(t *testing.T) {
			// Set environment variables for test
			os.Setenv("dbCaCert", scenario.caCert)
			os.Setenv("dbClientCert", scenario.clientCert)
			os.Setenv("dbClientKey", scenario.clientKey)

			result := appendMySQLTLSParam(scenario.dsn)

			if scenario.expectedSuffix == "" {
				assert.Equal(t, scenario.dsn, result)
			} else {
				assert.Contains(t, result, scenario.expectedSuffix)
			}

			// Clean up environment variables
			os.Unsetenv("dbCaCert")
			os.Unsetenv("dbClientCert")
			os.Unsetenv("dbClientKey")
		})
	}
}

func TestSetupMySQLTLS_NoCertificates(t *testing.T) {
	// Load config
	err := web.LoadAppConfig("ini", "../conf/app.conf")
	assert.Nil(t, err)

	// Ensure no certificates are configured
	os.Unsetenv("dbCaCert")
	os.Unsetenv("dbClientCert")
	os.Unsetenv("dbClientKey")

	err = setupMySQLTLS()
	assert.Nil(t, err)
}

func TestSetupMySQLTLS_InvalidCertPath(t *testing.T) {
	// Load config
	err := web.LoadAppConfig("ini", "../conf/app.conf")
	assert.Nil(t, err)

	// Set an invalid certificate path
	os.Setenv("dbCaCert", "/invalid/path/ca.pem")
	defer os.Unsetenv("dbCaCert")

	err = setupMySQLTLS()
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "failed to read CA certificate")
}

func TestSetupMySQLTLS_MismatchedClientCerts(t *testing.T) {
	// Load config
	err := web.LoadAppConfig("ini", "../conf/app.conf")
	assert.Nil(t, err)

	// Set only client cert without key
	os.Setenv("dbClientCert", "/path/to/client.pem")
	defer os.Unsetenv("dbClientCert")

	err = setupMySQLTLS()
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "both dbClientCert and dbClientKey must be provided together")
}
