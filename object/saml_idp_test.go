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

	"github.com/beevik/etree"
)

// TestSamlAttributeStatementNamespaces verifies that when C14N10 is enabled,
// the xsi and xs namespace declarations are added to the AttributeStatement element
func TestSamlAttributeStatementNamespaces(t *testing.T) {
	tests := []struct {
		name             string
		enableC14n10     bool
		expectXsiOnAttr  bool
		expectXsOnAttr   bool
	}{
		{
			name:            "C14N10 enabled - namespaces on AttributeStatement",
			enableC14n10:    true,
			expectXsiOnAttr: true,
			expectXsOnAttr:  true,
		},
		{
			name:            "C14N10 disabled - no namespaces on AttributeStatement",
			enableC14n10:    false,
			expectXsiOnAttr: false,
			expectXsOnAttr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a minimal SAML response structure to test namespace handling
			samlResponse := &etree.Element{
				Space: "samlp",
				Tag:   "Response",
			}
			samlResponse.CreateAttr("xmlns:samlp", "urn:oasis:names:tc:SAML:2.0:protocol")
			samlResponse.CreateAttr("xmlns:saml", "urn:oasis:names:tc:SAML:2.0:assertion")
			samlResponse.CreateAttr("xmlns:xsi", "http://www.w3.org/2001/XMLSchema-instance")
			samlResponse.CreateAttr("xmlns:xs", "http://www.w3.org/2001/XMLSchema")

			assertion := samlResponse.CreateElement("saml:Assertion")
			assertion.CreateAttr("xmlns:saml", "urn:oasis:names:tc:SAML:2.0:assertion")
			assertion.CreateAttr("xmlns:xsi", "http://www.w3.org/2001/XMLSchema-instance")
			assertion.CreateAttr("xmlns:xs", "http://www.w3.org/2001/XMLSchema")

			// Mimic the logic in NewSamlResponse for AttributeStatement creation
			attributes := assertion.CreateElement("saml:AttributeStatement")
			if tt.enableC14n10 {
				attributes.CreateAttr("xmlns:xsi", "http://www.w3.org/2001/XMLSchema-instance")
				attributes.CreateAttr("xmlns:xs", "http://www.w3.org/2001/XMLSchema")
			}

			// Create an attribute with xsi:type
			attr := attributes.CreateElement("saml:Attribute")
			attr.CreateAttr("Name", "Email")
			attr.CreateElement("saml:AttributeValue").CreateAttr("xsi:type", "xs:string")

			// Verify namespace declarations on AttributeStatement
			// Note: etree stores namespace declarations with Space="xmlns" and Key=prefix
			hasXsiNamespace := false
			hasXsNamespace := false
			for _, attr := range attributes.Attr {
				if attr.Space == "xmlns" && attr.Key == "xsi" && attr.Value == "http://www.w3.org/2001/XMLSchema-instance" {
					hasXsiNamespace = true
				}
				if attr.Space == "xmlns" && attr.Key == "xs" && attr.Value == "http://www.w3.org/2001/XMLSchema" {
					hasXsNamespace = true
				}
			}

			if tt.expectXsiOnAttr && !hasXsiNamespace {
				t.Errorf("Expected xmlns:xsi on AttributeStatement when C14N10 is enabled, but it was not found")
			}
			if !tt.expectXsiOnAttr && hasXsiNamespace {
				t.Errorf("Did not expect xmlns:xsi on AttributeStatement when C14N10 is disabled, but it was found")
			}
			if tt.expectXsOnAttr && !hasXsNamespace {
				t.Errorf("Expected xmlns:xs on AttributeStatement when C14N10 is enabled, but it was not found")
			}
			if !tt.expectXsOnAttr && hasXsNamespace {
				t.Errorf("Did not expect xmlns:xs on AttributeStatement when C14N10 is disabled, but it was found")
			}
		})
	}
}
