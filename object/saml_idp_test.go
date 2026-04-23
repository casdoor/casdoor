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
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"testing"

	"github.com/beevik/etree"
	"github.com/stretchr/testify/require"
)

func testSamlSigningKeyStore(t *testing.T) *X509Key {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 1024)
	require.NoError(t, err)

	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	return &X509Key{
		PrivateKey:      string(privateKeyPEM),
		X509Certificate: base64.StdEncoding.EncodeToString([]byte("test-certificate")),
	}
}

func TestNewSamlSigningContextUsesExclusiveCanonicalization(t *testing.T) {
	keyStore := testSamlSigningKeyStore(t)

	tests := []struct {
		name              string
		enableSamlC14n10  bool
		expectedAlgorithm string
		unexpected        []string
	}{
		{
			name:              "disabled uses c14n11",
			enableSamlC14n10:  false,
			expectedAlgorithm: "http://www.w3.org/2006/12/xml-c14n11",
			unexpected: []string{
				"http://www.w3.org/2001/10/xml-exc-c14n#",
				"InclusiveNamespaces",
				"PrefixList=\"xs\"",
			},
		},
		{
			name:              "enabled uses exclusive c14n10",
			enableSamlC14n10:  true,
			expectedAlgorithm: "http://www.w3.org/2001/10/xml-exc-c14n#",
			unexpected: []string{
				"http://www.w3.org/2006/12/xml-c14n11",
				"InclusiveNamespaces",
				"PrefixList=\"xs\"",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewSamlSigningContext(&Application{EnableSamlC14n10: tt.enableSamlC14n10}, keyStore)

			signedElement := etree.NewElement("Assertion")
			signedElement.CreateAttr("ID", "_test")
			signedElement.CreateElement("Issuer").SetText("https://example.com")

			signature, err := ctx.ConstructSignature(signedElement, true)
			require.NoError(t, err)

			doc := etree.NewDocument()
			doc.SetRoot(signature)

			signatureXML, err := doc.WriteToBytes()
			require.NoError(t, err)

			xml := string(signatureXML)
			require.Contains(t, xml, tt.expectedAlgorithm)
			for _, unexpected := range tt.unexpected {
				require.NotContains(t, xml, unexpected)
			}
		})
	}
}
