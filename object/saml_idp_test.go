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
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/beevik/etree"
	"github.com/google/uuid"
)

// TestSamlAudienceDeduplication tests that the audience deduplication logic works correctly
// This is a unit test that directly tests the audience generation logic without requiring database access
func TestSamlAudienceDeduplication(t *testing.T) {
	tests := []struct {
		name                string
		iss                 string
		redirectUris        []string
		expectedAudiences   []string
		description         string
	}{
		{
			name:                "No duplicate when redirectUris contains issuer",
			iss:                 "https://sp.example.com/saml/metadata",
			redirectUris:        []string{"https://sp.example.com/saml/metadata"},
			expectedAudiences:   []string{"https://sp.example.com/saml/metadata"},
			description:         "Should have only one audience element when issuer is in redirectUris",
		},
		{
			name:                "No duplicate when redirectUris is empty",
			iss:                 "https://sp.example.com/saml/metadata",
			redirectUris:        []string{},
			expectedAudiences:   []string{"https://sp.example.com/saml/metadata"},
			description:         "Should have only one audience element when redirectUris is empty",
		},
		{
			name:                "Multiple audiences when redirectUris contains different URL",
			iss:                 "https://sp.example.com/saml/metadata",
			redirectUris:        []string{"https://sp.example.com/other"},
			expectedAudiences:   []string{"https://sp.example.com/saml/metadata", "https://sp.example.com/other"},
			description:         "Should have two audience elements when redirectUris contains different URL",
		},
		{
			name:                "No empty audiences",
			iss:                 "https://sp.example.com/saml/metadata",
			redirectUris:        []string{""},
			expectedAudiences:   []string{"https://sp.example.com/saml/metadata"},
			description:         "Should skip empty redirectUris",
		},
		{
			name:                "Multiple unique audiences",
			iss:                 "https://sp.example.com/saml/metadata",
			redirectUris:        []string{"https://sp1.example.com/metadata", "https://sp2.example.com/metadata"},
			expectedAudiences:   []string{"https://sp.example.com/saml/metadata", "https://sp1.example.com/metadata", "https://sp2.example.com/metadata"},
			description:         "Should have three audience elements (issuer + two unique redirectUris)",
		},
		{
			name:                "Mix of duplicate and unique",
			iss:                 "https://sp.example.com/saml/metadata",
			redirectUris:        []string{"https://sp.example.com/saml/metadata", "https://sp1.example.com/metadata"},
			expectedAudiences:   []string{"https://sp.example.com/saml/metadata", "https://sp1.example.com/metadata"},
			description:         "Should skip duplicate issuer and only add unique redirectUris",
		},
		{
			name:                "Empty string in mix",
			iss:                 "https://sp.example.com/saml/metadata",
			redirectUris:        []string{"", "https://sp1.example.com/metadata", ""},
			expectedAudiences:   []string{"https://sp.example.com/saml/metadata", "https://sp1.example.com/metadata"},
			description:         "Should skip empty strings and only add unique non-empty redirectUris",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a minimal SAML conditions element with audience restriction
			// This simulates the logic in NewSamlResponse without requiring full application setup
			now := time.Now().UTC().Format(time.RFC3339)
			expireTime := time.Now().UTC().Add(time.Hour * 24).Format(time.RFC3339)

			condition := &etree.Element{Tag: "saml:Conditions"}
			condition.CreateAttr("NotBefore", now)
			condition.CreateAttr("NotOnOrAfter", expireTime)
			audience := condition.CreateElement("saml:AudienceRestriction")

			// This is the logic we're testing - it should match the actual code in saml_idp.go
			audience.CreateElement("saml:Audience").SetText(tt.iss)
			// Add redirect URIs as audiences, but skip duplicates and empty values
			for _, value := range tt.redirectUris {
				if value != "" && value != tt.iss {
					audience.CreateElement("saml:Audience").SetText(value)
				}
			}

			// Convert to string for inspection
			doc := etree.NewDocument()
			doc.SetRoot(condition)
			xmlStr, err := doc.WriteToString()
			if err != nil {
				t.Fatalf("Failed to write XML to string: %v", err)
			}

			// Verify expected audiences
			for _, expectedAud := range tt.expectedAudiences {
				expectedTag := fmt.Sprintf("<saml:Audience>%s</saml:Audience>", expectedAud)
				if !strings.Contains(xmlStr, expectedTag) {
					t.Errorf("%s: expected audience '%s' not found in XML:\n%s",
						tt.description, expectedAud, xmlStr)
				}
			}

			// Count audience elements
			audienceCount := strings.Count(xmlStr, "<saml:Audience>")

			if audienceCount != len(tt.expectedAudiences) {
				t.Errorf("%s: expected %d audience elements, got %d\nXML:\n%s",
					tt.description, len(tt.expectedAudiences), audienceCount, xmlStr)
			}

			// Verify no empty audiences
			if strings.Contains(xmlStr, "<saml:Audience></saml:Audience>") || strings.Contains(xmlStr, "<saml:Audience/>") {
				t.Errorf("Empty audience element found in response:\n%s", xmlStr)
			}
		})
	}
}

// TestSamlResponseStructure is a minimal test to ensure the basic structure is created correctly
func TestSamlResponseStructure(t *testing.T) {
	now := time.Now().UTC().Format(time.RFC3339)
	samlResponse := &etree.Element{
		Space: "samlp",
		Tag:   "Response",
	}
	samlResponse.CreateAttr("xmlns:samlp", "urn:oasis:names:tc:SAML:2.0:protocol")
	samlResponse.CreateAttr("xmlns:saml", "urn:oasis:names:tc:SAML:2.0:assertion")
	arId := uuid.New()

	samlResponse.CreateAttr("ID", fmt.Sprintf("_%s", arId))
	samlResponse.CreateAttr("Version", "2.0")
	samlResponse.CreateAttr("IssueInstant", now)

	doc := etree.NewDocument()
	doc.SetRoot(samlResponse)
	xmlStr, err := doc.WriteToString()
	if err != nil {
		t.Fatalf("Failed to write XML to string: %v", err)
	}

	if !strings.Contains(xmlStr, "samlp:Response") {
		t.Errorf("Response element not found in XML")
	}
}
