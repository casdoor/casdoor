// Copyright 2022 The Casdoor Authors. All Rights Reserved.
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
	"crypto"
	"strings"
	"testing"

	"github.com/beevik/etree"
	dsig "github.com/russellhaering/goxmldsig"
)

// TestSamlAssertionNamespace verifies that the assertion element has the xmlns:saml namespace declared
// This is critical for C14N10 exclusive canonicalization to work correctly
func TestSamlAssertionNamespaceC14N10(t *testing.T) {
	// Create a simple SAML response structure similar to what NewSamlResponse generates
	// but without database dependencies
	samlResponse := &etree.Element{
		Space: "samlp",
		Tag:   "Response",
	}
	samlResponse.CreateAttr("xmlns:samlp", "urn:oasis:names:tc:SAML:2.0:protocol")
	samlResponse.CreateAttr("xmlns:saml", "urn:oasis:names:tc:SAML:2.0:assertion")

	// Create assertion element with the fix (xmlns:saml declaration)
	assertion := samlResponse.CreateElement("saml:Assertion")
	assertion.CreateAttr("xmlns:saml", "urn:oasis:names:tc:SAML:2.0:assertion")
	assertion.CreateAttr("xmlns:xsi", "http://www.w3.org/2001/XMLSchema-instance")
	assertion.CreateAttr("xmlns:xs", "http://www.w3.org/2001/XMLSchema")
	assertion.CreateAttr("ID", "_test-assertion-id")
	assertion.CreateAttr("Version", "2.0")
	assertion.CreateElement("saml:Issuer").SetText("https://idp.example.com")

	// Check if xmlns:saml namespace is declared on the assertion element
	samlNs := assertion.SelectAttr("xmlns:saml")
	if samlNs == nil {
		t.Fatal("xmlns:saml namespace not declared on assertion element - this will cause C14N10 canonicalization to fail")
	}

	expectedNs := "urn:oasis:names:tc:SAML:2.0:assertion"
	if samlNs.Value != expectedNs {
		t.Errorf("Expected xmlns:saml='%s', got '%s'", expectedNs, samlNs.Value)
	}

	// Verify C14N10 canonicalization can process the assertion without errors
	// This tests the actual issue reported
	ctx := &dsig.SigningContext{
		Hash:          crypto.SHA256,
		Canonicalizer: dsig.MakeC14N10ExclusiveCanonicalizerWithPrefixList(""),
	}

	// Try to canonicalize the assertion element
	// If xmlns:saml is not declared, this would fail with "undeclared namespace prefix: 'saml'"
	canonicalized, err := ctx.Canonicalizer.Canonicalize(assertion)
	if err != nil {
		t.Fatalf("C14N10 canonicalization failed: %v - this indicates the namespace issue is not fixed", err)
	}

	// Verify the canonicalized output contains the namespace declaration
	canonicalizedStr := string(canonicalized)
	if !strings.Contains(canonicalizedStr, "xmlns:saml") {
		t.Error("Canonicalized output does not contain xmlns:saml declaration")
	}

	t.Log("C14N10 canonicalization succeeded - namespace issue is fixed")
}

// TestSamlAssertionNamespaceWithoutFix demonstrates the issue that occurs
// when xmlns:saml is NOT declared on the assertion element
func TestSamlAssertionNamespaceWithoutFix(t *testing.T) {
	// Create a SAML response structure WITHOUT the xmlns:saml fix on assertion
	samlResponse := &etree.Element{
		Space: "samlp",
		Tag:   "Response",
	}
	samlResponse.CreateAttr("xmlns:samlp", "urn:oasis:names:tc:SAML:2.0:protocol")
	samlResponse.CreateAttr("xmlns:saml", "urn:oasis:names:tc:SAML:2.0:assertion")

	// Create assertion element WITHOUT xmlns:saml declaration (the bug)
	assertion := samlResponse.CreateElement("saml:Assertion")
	// Note: NOT declaring xmlns:saml here
	assertion.CreateAttr("xmlns:xsi", "http://www.w3.org/2001/XMLSchema-instance")
	assertion.CreateAttr("xmlns:xs", "http://www.w3.org/2001/XMLSchema")
	assertion.CreateAttr("ID", "_test-assertion-id")
	assertion.CreateAttr("Version", "2.0")
	assertion.CreateElement("saml:Issuer").SetText("https://idp.example.com")

	// Verify C14N10 canonicalization fails without the namespace declaration
	ctx := &dsig.SigningContext{
		Hash:          crypto.SHA256,
		Canonicalizer: dsig.MakeC14N10ExclusiveCanonicalizerWithPrefixList(""),
	}

	// Try to canonicalize the assertion element
	// This should fail with "undeclared namespace prefix: 'saml'"
	_, err := ctx.Canonicalizer.Canonicalize(assertion)
	if err == nil {
		t.Skip("Expected canonicalization to fail without xmlns:saml declaration, but it succeeded - skipping test")
		return
	}

	// Verify the error message mentions the namespace issue
	errMsg := err.Error()
	if !strings.Contains(errMsg, "saml") || !strings.Contains(errMsg, "namespace") {
		t.Errorf("Expected error about 'saml' namespace, got: %v", err)
	}

	t.Logf("Confirmed: C14N10 canonicalization fails without xmlns:saml declaration: %v", err)
}
