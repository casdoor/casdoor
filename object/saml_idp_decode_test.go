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
	"bytes"
	"compress/flate"
	"encoding/base64"
	"encoding/xml"
	"testing"

	saml "github.com/russellhaering/gosaml2"
)

const (
	defaultNamespaceAuthnRequest  = `<AuthnRequest xmlns="urn:oasis:names:tc:SAML:2.0:protocol" ID="_default_ns" Version="2.0" AssertionConsumerServiceURL="https://sp.example.com/callback"><Issuer xmlns="urn:oasis:names:tc:SAML:2.0:assertion">https://sp.example.com</Issuer></AuthnRequest>`
	prefixedNamespaceAuthnRequest = `<samlp:AuthnRequest xmlns:samlp="urn:oasis:names:tc:SAML:2.0:protocol" xmlns:saml="urn:oasis:names:tc:SAML:2.0:assertion" ID="_prefixed_ns" Version="2.0" AssertionConsumerServiceURL="https://sp.example.com/callback"><saml:Issuer>https://sp.example.com</saml:Issuer></samlp:AuthnRequest>`
)

func deflateForTest(t *testing.T, data string, level int) []byte {
	t.Helper()
	var buffer bytes.Buffer
	writer, err := flate.NewWriter(&buffer, level)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = writer.Write([]byte(data)); err != nil {
		t.Fatal(err)
	}
	if err = writer.Close(); err != nil {
		t.Fatal(err)
	}
	return buffer.Bytes()
}

func TestDecodeSamlMessage(t *testing.T) {
	tests := []struct {
		name    string
		payload []byte
		wantID  string
	}{
		{"POST default namespace", []byte(defaultNamespaceAuthnRequest), "_default_ns"},
		{"POST prefixed namespace", []byte(prefixedNamespaceAuthnRequest), "_prefixed_ns"},
		{"POST leading whitespace", []byte("\n \t" + defaultNamespaceAuthnRequest), "_default_ns"},
		{"POST UTF-8 BOM", []byte("\xef\xbb\xbf" + defaultNamespaceAuthnRequest), "_default_ns"},
		{"POST XML declaration", []byte(`<?xml version="1.0" encoding="UTF-8"?>` + defaultNamespaceAuthnRequest), "_default_ns"},
	}
	for _, level := range []int{flate.HuffmanOnly, flate.BestSpeed, flate.DefaultCompression, flate.BestCompression} {
		tests = append(tests,
			struct {
				name    string
				payload []byte
				wantID  string
			}{"Redirect DEFLATE default namespace", deflateForTest(t, defaultNamespaceAuthnRequest, level), "_default_ns"},
			struct {
				name    string
				payload []byte
				wantID  string
			}{"Redirect DEFLATE prefixed namespace", deflateForTest(t, prefixedNamespaceAuthnRequest, level), "_prefixed_ns"},
		)
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			decoded, err := decodeSamlMessage(base64.StdEncoding.EncodeToString(test.payload))
			if err != nil {
				t.Fatalf("decodeSamlMessage() error = %v", err)
			}

			var authnRequest saml.AuthNRequest
			if err = xml.Unmarshal(decoded, &authnRequest); err != nil {
				t.Fatalf("xml.Unmarshal() error = %v", err)
			}
			if authnRequest.ID != test.wantID {
				t.Errorf("AuthnRequest.ID = %q, want %q", authnRequest.ID, test.wantID)
			}
		})
	}
}

func TestPlainSamlXmlRejectsBinaryStartingWithAngleBracket(t *testing.T) {
	// 0x3c is a valid first byte for a non-final dynamic Huffman DEFLATE block.
	if _, ok := plainSamlXml([]byte{0x3c, 0xff, 0x8d, 0x01}); ok {
		t.Error("plainSamlXml() accepted data that is not valid UTF-8")
	}
}
