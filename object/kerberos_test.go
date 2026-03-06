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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateKerberosToken_IncompleteConfig(t *testing.T) {
	tests := []struct {
		name string
		org  *Organization
	}{
		{
			name: "missing realm",
			org: &Organization{
				Name:            "test-org",
				KerberosKdcHost: "kdc.example.com",
				KerberosKeytab:  "dGVzdA==",
			},
		},
		{
			name: "missing kdc host",
			org: &Organization{
				Name:           "test-org",
				KerberosRealm:  "EXAMPLE.COM",
				KerberosKeytab: "dGVzdA==",
			},
		},
		{
			name: "missing keytab",
			org: &Organization{
				Name:            "test-org",
				KerberosRealm:   "EXAMPLE.COM",
				KerberosKdcHost: "kdc.example.com",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ValidateKerberosToken(tt.org, "dGVzdA==")
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "kerberos configuration is incomplete")
		})
	}
}

func TestValidateKerberosToken_InvalidKeytab(t *testing.T) {
	org := &Organization{
		Name:            "test-org",
		KerberosRealm:   "EXAMPLE.COM",
		KerberosKdcHost: "kdc.example.com",
		KerberosKeytab:  "!!!invalid-base64!!!",
	}

	_, err := ValidateKerberosToken(org, "dGVzdA==")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode keytab")
}

func TestValidateKerberosToken_InvalidSPNEGOToken(t *testing.T) {
	org := &Organization{
		Name:                "test-org",
		KerberosRealm:       "EXAMPLE.COM",
		KerberosKdcHost:     "kdc.example.com",
		KerberosKeytab:      "BQIAAAA=", // minimal valid base64 but invalid keytab
		KerberosServiceName: "HTTP",
	}

	_, err := ValidateKerberosToken(org, "dGVzdA==")
	assert.Error(t, err)
}
