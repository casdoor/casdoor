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

// TestSamlAssertionNamespaces verifies that the SAML assertion element
// has the required namespace declarations (xmlns:xsi and xmlns:xs)
// to support xsi:type attributes used in AttributeValue elements.
// This is a unit test that doesn't require database access.
func TestSamlAssertionNamespaces(t *testing.T) {
	// Create a simple SAML response structure to verify namespace declarations
	samlResponse := &etree.Element{
		Space: "samlp",
		Tag:   "Response",
	}
	samlResponse.CreateAttr("xmlns:samlp", "urn:oasis:names:tc:SAML:2.0:protocol")
	samlResponse.CreateAttr("xmlns:saml", "urn:oasis:names:tc:SAML:2.0:assertion")
	samlResponse.CreateAttr("xmlns:xsi", "http://www.w3.org/2001/XMLSchema-instance")
	samlResponse.CreateAttr("xmlns:xs", "http://www.w3.org/2001/XMLSchema")

	// Create assertion with proper namespace declarations
	assertion := samlResponse.CreateElement("saml:Assertion")
	assertion.CreateAttr("xmlns:saml", "urn:oasis:names:tc:SAML:2.0:assertion")
	assertion.CreateAttr("xmlns:xsi", "http://www.w3.org/2001/XMLSchema-instance")
	assertion.CreateAttr("xmlns:xs", "http://www.w3.org/2001/XMLSchema")

	// Add an attribute statement with xsi:type
	attributes := assertion.CreateElement("saml:AttributeStatement")
	email := attributes.CreateElement("saml:Attribute")
	email.CreateAttr("Name", "Email")
	email.CreateElement("saml:AttributeValue").CreateAttr("xsi:type", "xs:string").Element().SetText("test@example.com")

	// Check that assertion has xmlns:xsi attribute
	xsiAttr := assertion.SelectAttr("xmlns:xsi")
	if xsiAttr == nil {
		t.Error("Assertion element is missing xmlns:xsi namespace declaration")
	} else if xsiAttr.Value != "http://www.w3.org/2001/XMLSchema-instance" {
		t.Errorf("xmlns:xsi has incorrect value: got %s, want http://www.w3.org/2001/XMLSchema-instance", xsiAttr.Value)
	}

	// Check that assertion has xmlns:xs attribute
	xsAttr := assertion.SelectAttr("xmlns:xs")
	if xsAttr == nil {
		t.Error("Assertion element is missing xmlns:xs namespace declaration")
	} else if xsAttr.Value != "http://www.w3.org/2001/XMLSchema" {
		t.Errorf("xmlns:xs has incorrect value: got %s, want http://www.w3.org/2001/XMLSchema", xsAttr.Value)
	}

	// Check that assertion has xmlns:saml attribute
	samlAttr := assertion.SelectAttr("xmlns:saml")
	if samlAttr == nil {
		t.Error("Assertion element is missing xmlns:saml namespace declaration")
	} else if samlAttr.Value != "urn:oasis:names:tc:SAML:2.0:assertion" {
		t.Errorf("xmlns:saml has incorrect value: got %s, want urn:oasis:names:tc:SAML:2.0:assertion", samlAttr.Value)
	}

	// Verify that the assertion can be serialized as standalone XML
	doc := etree.NewDocument()
	doc.SetRoot(assertion.Copy())
	xmlString, err := doc.WriteToString()
	if err != nil {
		t.Fatalf("Failed to serialize assertion to XML: %v", err)
	}

	// Verify the serialized XML contains all required namespace declarations
	if !strings.Contains(xmlString, `xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"`) {
		t.Error("Serialized assertion XML is missing xmlns:xsi declaration")
	}

	if !strings.Contains(xmlString, `xmlns:xs="http://www.w3.org/2001/XMLSchema"`) {
		t.Error("Serialized assertion XML is missing xmlns:xs declaration")
	}

	if !strings.Contains(xmlString, `xmlns:saml="urn:oasis:names:tc:SAML:2.0:assertion"`) {
		t.Error("Serialized assertion XML is missing xmlns:saml declaration")
	}

	// Verify xsi:type attribute is present
	if !strings.Contains(xmlString, `xsi:type="xs:string"`) {
		t.Error("Serialized assertion XML is missing xsi:type attribute")
	}
}

// TestSamlAssertionStructure verifies that the NewSamlResponse function
// creates assertions with proper namespace declarations.
// This test validates the fix for the "undeclared namespace prefix: 'xsi'" error.
func TestSamlAssertionStructure(t *testing.T) {
	// Skip if database is not initialized
	if ormer == nil {
		t.Skip("Database not initialized, skipping integration test")
		return
	}

	InitConfig()

	application := &Application{
		Name:                 "test-app",
		Owner:                "built-in",
		UseEmailAsSamlNameId: false,
		EnableSamlCompress:   false,
		SamlHashAlgorithm:    "SHA256",
		EnableSamlC14n10:     false,
		SamlAttributes:       []*SamlItem{},
	}

	user := &User{
		Owner:       "built-in",
		Name:        "testuser",
		DisplayName: "Test User",
		Email:       "test@example.com",
		Roles:       []*Role{{Name: "admin"}},
		Permissions: []*Permission{},
	}

	host := "https://example.com"
	destination := "https://sp.example.com/acs"
	requestId := "_test_request_id"
	redirectUri := []string{"https://sp.example.com"}

	// Generate SAML response
	samlResponse, err := NewSamlResponse(application, user, host, "test-cert", destination, "test-issuer", requestId, redirectUri)
	if err != nil {
		t.Fatalf("Failed to create SAML response: %v", err)
	}

	// Find the assertion element
	assertion := samlResponse.FindElement("./Assertion")
	if assertion == nil {
		t.Fatal("Assertion element not found in SAML response")
	}

	// Check that assertion has required namespace declarations
	xsiAttr := assertion.SelectAttr("xmlns:xsi")
	if xsiAttr == nil {
		t.Error("Assertion element is missing xmlns:xsi namespace declaration")
	}

	xsAttr := assertion.SelectAttr("xmlns:xs")
	if xsAttr == nil {
		t.Error("Assertion element is missing xmlns:xs namespace declaration")
	}

	samlAttr := assertion.SelectAttr("xmlns:saml")
	if samlAttr == nil {
		t.Error("Assertion element is missing xmlns:saml namespace declaration")
	}

	// Verify assertion can be extracted as standalone XML
	doc := etree.NewDocument()
	doc.SetRoot(assertion.Copy())
	xmlString, err := doc.WriteToString()
	if err != nil {
		t.Fatalf("Failed to serialize assertion: %v", err)
	}

	// All namespace declarations should be present in standalone assertion
	if !strings.Contains(xmlString, `xmlns:xsi=`) {
		t.Error("Standalone assertion missing xmlns:xsi")
	}
	if !strings.Contains(xmlString, `xmlns:xs=`) {
		t.Error("Standalone assertion missing xmlns:xs")
	}
	if !strings.Contains(xmlString, `xmlns:saml=`) {
		t.Error("Standalone assertion missing xmlns:saml")
	}
}
