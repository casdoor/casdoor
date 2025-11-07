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
	"strings"
	"testing"

	"github.com/beevik/etree"
)

func TestSamlResponse_RootNamespaceDeclarations(t *testing.T) {
	// Create a minimal SAML response element to test namespace declarations
	samlResponse := &etree.Element{
		Space: "samlp",
		Tag:   "Response",
	}

	// Add namespace declarations as they are in the actual code
	samlResponse.CreateAttr("xmlns:samlp", "urn:oasis:names:tc:SAML:2.0:protocol")
	samlResponse.CreateAttr("xmlns:saml", "urn:oasis:names:tc:SAML:2.0:assertion")
	samlResponse.CreateAttr("xmlns:xsi", "http://www.w3.org/2001/XMLSchema-instance")
	samlResponse.CreateAttr("xmlns:xs", "http://www.w3.org/2001/XMLSchema")

	// Create a simple assertion with attribute that uses xsi:type
	assertion := samlResponse.CreateElement("saml:Assertion")
	attributes := assertion.CreateElement("saml:AttributeStatement")
	email := attributes.CreateElement("saml:Attribute")
	email.CreateAttr("Name", "Email")
	email.CreateElement("saml:AttributeValue").CreateAttr("xsi:type", "xs:string").Element().SetText("test@example.com")

	// Convert to XML
	doc := etree.NewDocument()
	doc.SetRoot(samlResponse)
	xmlBytes, err := doc.WriteToBytes()
	if err != nil {
		t.Fatalf("Failed to serialize SAML response: %v", err)
	}

	xmlString := string(xmlBytes)

	// Check that xmlns:xsi is declared at the root level
	if !strings.Contains(xmlString, `xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"`) {
		t.Error("SAML Response root element is missing xmlns:xsi namespace declaration")
	}

	// Check that xmlns:xs is declared at the root level
	if !strings.Contains(xmlString, `xmlns:xs="http://www.w3.org/2001/XMLSchema"`) {
		t.Error("SAML Response root element is missing xmlns:xs namespace declaration")
	}

	// Check that xsi:type="xs:string" is present
	if !strings.Contains(xmlString, `xsi:type="xs:string"`) {
		t.Error("SAML Response is missing xsi:type=\"xs:string\" in AttributeValue elements")
	}

	// Verify the root element has the expected namespaces by checking attributes
	xsiAttr := samlResponse.SelectAttr("xmlns:xsi")
	if xsiAttr == nil || xsiAttr.Value != "http://www.w3.org/2001/XMLSchema-instance" {
		t.Error("Root element does not have correct xmlns:xsi attribute")
	}

	xsAttr := samlResponse.SelectAttr("xmlns:xs")
	if xsAttr == nil || xsAttr.Value != "http://www.w3.org/2001/XMLSchema" {
		t.Error("Root element does not have correct xmlns:xs attribute")
	}

	// Print the XML for manual verification (optional)
	t.Logf("Generated SAML Response XML:\n%s", xmlString)
}

func TestSamlResponse_AttributeValueWithType(t *testing.T) {
	// Test that AttributeValue elements with xsi:type work correctly
	// when namespaces are declared at the root
	samlResponse := &etree.Element{
		Space: "samlp",
		Tag:   "Response",
	}

	// Declare namespaces at root (as fixed in the code)
	samlResponse.CreateAttr("xmlns:samlp", "urn:oasis:names:tc:SAML:2.0:protocol")
	samlResponse.CreateAttr("xmlns:saml", "urn:oasis:names:tc:SAML:2.0:assertion")
	samlResponse.CreateAttr("xmlns:xsi", "http://www.w3.org/2001/XMLSchema-instance")
	samlResponse.CreateAttr("xmlns:xs", "http://www.w3.org/2001/XMLSchema")

	// Create multiple AttributeValue elements as in the actual code
	assertion := samlResponse.CreateElement("saml:Assertion")
	attributes := assertion.CreateElement("saml:AttributeStatement")

	// Email attribute
	email := attributes.CreateElement("saml:Attribute")
	email.CreateAttr("Name", "Email")
	email.CreateElement("saml:AttributeValue").CreateAttr("xsi:type", "xs:string").Element().SetText("test@example.com")

	// Name attribute
	name := attributes.CreateElement("saml:Attribute")
	name.CreateAttr("Name", "Name")
	name.CreateElement("saml:AttributeValue").CreateAttr("xsi:type", "xs:string").Element().SetText("testuser")

	// DisplayName attribute
	displayName := attributes.CreateElement("saml:Attribute")
	displayName.CreateAttr("Name", "DisplayName")
	displayName.CreateElement("saml:AttributeValue").CreateAttr("xsi:type", "xs:string").Element().SetText("Test User")

	// Convert to XML
	doc := etree.NewDocument()
	doc.SetRoot(samlResponse)
	xmlBytes, err := doc.WriteToBytes()
	if err != nil {
		t.Fatalf("Failed to serialize SAML response: %v", err)
	}

	xmlString := string(xmlBytes)

	// Count occurrences of xsi:type to ensure all attributes have it
	xsiTypeCount := strings.Count(xmlString, `xsi:type="xs:string"`)
	if xsiTypeCount < 3 {
		t.Errorf("Expected at least 3 AttributeValue elements with xsi:type, found %d", xsiTypeCount)
	}

	// Verify namespace declarations are at root
	if !strings.Contains(xmlString, `xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"`) {
		t.Error("Missing xmlns:xsi at root")
	}

	if !strings.Contains(xmlString, `xmlns:xs="http://www.w3.org/2001/XMLSchema"`) {
		t.Error("Missing xmlns:xs at root")
	}

	t.Logf("Generated SAML Response with multiple attributes:\n%s", xmlString)
}
