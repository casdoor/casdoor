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
	"encoding/xml"
	"fmt"
	"time"

	uuid "github.com/satori/go.uuid"
)

/*
SAMLReponse struct
<?xml version="1.0"?>
<samlp:Response xmlns:samlp="urn:oasis:names:tc:SAML:2.0:protocol" xmlns:saml="urn:oasis:names:tc:SAML:2.0:assertion" ID="pfxf0c4d762-2e55-7305-7632-bf64d5cbb640" Version="2.0" IssueInstant="2014-07-17T01:01:48Z" Destination="http://sp.example.com/demo1/index.php?acs" InResponseTo="ONELOGIN_4fee3b046395c4e751011e97f8900b5273d56685">
  <saml:Issuer>http://idp.example.com/metadata.php</saml:Issuer><ds:Signature xmlns:ds="http://www.w3.org/2000/09/xmldsig#">
  <ds:SignedInfo><ds:CanonicalizationMethod Algorithm="http://www.w3.org/2001/10/xml-exc-c14n#"/>
    <ds:SignatureMethod Algorithm="http://www.w3.org/2000/09/xmldsig#rsa-sha1"/>
  <ds:Reference URI="#pfxf0c4d762-2e55-7305-7632-bf64d5cbb640"><ds:Transforms><ds:Transform Algorithm="http://www.w3.org/2000/09/xmldsig#enveloped-signature"/><ds:Transform Algorithm="http://www.w3.org/2001/10/xml-exc-c14n#"/></ds:Transforms><ds:DigestMethod Algorithm="http://www.w3.org/2000/09/xmldsig#sha1"/><ds:DigestValue>AfGB5paadu7xpFpCBmHWP+TmGmc=</ds:DigestValue></ds:Reference></ds:SignedInfo><ds:SignatureValue>B0SXxAzfb9y0qIu92rd652wrj1h/Mj83Au0Sx3kFarJiPcLmZlFWvVOwohrTqEq6iwhR0ZWo3YgzO50wP9ebFVbLOwW/Wv8m6v57/b3WzUXsBhRcDb+ZCM5aomX/GWLCa4viBYQpL/V+MZYs8dEjN/kVdxBTLqlWML/DYEs3xIw=</ds:SignatureValue>
<ds:KeyInfo><ds:X509Data><ds:X509Certificate>MIICajCCAdOgAwIBAgIBADANBgkqhkiG9w0BAQ0FADBSMQswCQYDVQQGEwJ1czETMBEGA1UECAwKQ2FsaWZvcm5pYTEVMBMGA1UECgwMT25lbG9naW4gSW5jMRcwFQYDVQQDDA5zcC5leGFtcGxlLmNvbTAeFw0xNDA3MTcxNDEyNTZaFw0xNTA3MTcxNDEyNTZaMFIxCzAJBgNVBAYTAnVzMRMwEQYDVQQIDApDYWxpZm9ybmlhMRUwEwYDVQQKDAxPbmVsb2dpbiBJbmMxFzAVBgNVBAMMDnNwLmV4YW1wbGUuY29tMIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDZx+ON4IUoIWxgukTb1tOiX3bMYzYQiwWPUNMp+Fq82xoNogso2bykZG0yiJm5o8zv/sd6pGouayMgkx/2FSOdc36T0jGbCHuRSbtia0PEzNIRtmViMrt3AeoWBidRXmZsxCNLwgIV6dn2WpuE5Az0bHgpZnQxTKFek0BMKU/d8wIDAQABo1AwTjAdBgNVHQ4EFgQUGHxYqZYyX7cTxKVODVgZwSTdCnwwHwYDVR0jBBgwFoAUGHxYqZYyX7cTxKVODVgZwSTdCnwwDAYDVR0TBAUwAwEB/zANBgkqhkiG9w0BAQ0FAAOBgQByFOl+hMFICbd3DJfnp2Rgd/dqttsZG/tyhILWvErbio/DEe98mXpowhTkC04ENprOyXi7ZbUqiicF89uAGyt1oqgTUCD1VsLahqIcmrzgumNyTwLGWo17WDAa1/usDhetWAMhgzF/Cnf5ek0nK00m0YZGyc4LzgD0CROMASTWNg==</ds:X509Certificate></ds:X509Data></ds:KeyInfo></ds:Signature>
  <samlp:Status>
    <samlp:StatusCode Value="urn:oasis:names:tc:SAML:2.0:status:Success"/>
  </samlp:Status>
  <saml:Assertion xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:xs="http://www.w3.org/2001/XMLSchema" ID="_d71a3a8e9fcc45c9e9d248ef7049393fc8f04e5f75" Version="2.0" IssueInstant="2014-07-17T01:01:48Z">
    <saml:Issuer>http://idp.example.com/metadata.php</saml:Issuer>
    <saml:Subject>
      <saml:NameID SPNameQualifier="http://sp.example.com/demo1/metadata.php" Format="urn:oasis:names:tc:SAML:2.0:nameid-format:transient">_ce3d2948b4cf20146dee0a0b3dd6f69b6cf86f62d7</saml:NameID>
      <saml:SubjectConfirmation Method="urn:oasis:names:tc:SAML:2.0:cm:bearer">
        <saml:SubjectConfirmationData NotOnOrAfter="2024-01-18T06:21:48Z" Recipient="http://sp.example.com/demo1/index.php?acs" InResponseTo="ONELOGIN_4fee3b046395c4e751011e97f8900b5273d56685"/>
      </saml:SubjectConfirmation>
    </saml:Subject>
    <saml:Conditions NotBefore="2014-07-17T01:01:18Z" NotOnOrAfter="2024-01-18T06:21:48Z">
      <saml:AudienceRestriction>
        <saml:Audience>http://sp.example.com/demo1/metadata.php</saml:Audience>
      </saml:AudienceRestriction>
    </saml:Conditions>
    <saml:AuthnStatement AuthnInstant="2014-07-17T01:01:48Z" SessionNotOnOrAfter="2024-07-17T09:01:48Z" SessionIndex="_be9967abd904ddcae3c0eb4189adbe3f71e327cf93">
      <saml:AuthnContext>
        <saml:AuthnContextClassRef>urn:oasis:names:tc:SAML:2.0:ac:classes:Password</saml:AuthnContextClassRef>
      </saml:AuthnContext>
    </saml:AuthnStatement>
    <saml:AttributeStatement>
      <saml:Attribute Name="uid" NameFormat="urn:oasis:names:tc:SAML:2.0:attrname-format:basic">
        <saml:AttributeValue xsi:type="xs:string">test</saml:AttributeValue>
      </saml:Attribute>
      <saml:Attribute Name="mail" NameFormat="urn:oasis:names:tc:SAML:2.0:attrname-format:basic">
        <saml:AttributeValue xsi:type="xs:string">test@example.com</saml:AttributeValue>
      </saml:Attribute>
      <saml:Attribute Name="eduPersonAffiliation" NameFormat="urn:oasis:names:tc:SAML:2.0:attrname-format:basic">
        <saml:AttributeValue xsi:type="xs:string">users</saml:AttributeValue>
        <saml:AttributeValue xsi:type="xs:string">examplerole1</saml:AttributeValue>
      </saml:Attribute>
    </saml:AttributeStatement>
  </saml:Assertion>
</samlp:Response>
*/

type SamlResponse struct {
	XMLName      xml.Name  `xml:"samlp:Response"`
	XmlnsSamlp   string    `xml:"xmlns:samlp,attr"`
	XmlnsSaml    string    `xml:"xmlns:saml,attr"`
	ID           string    `xml:"ID,attr"`
	Version      string    `xml:"Version,attr"`
	IssueInstant string    `xml:"IssueInstant,attr"`
	Destination  string    `xml:"Destination,attr"`
	InResponseTo string    `xml:"InResponseTo,attr"`
	Issuer       string    `xml:"saml:Issuer"`
	Signature    Signature `xml:"ds:Signature"`
	Status       Status    `xml:"samlp:Status"`
	Assertion    Assertion `xml:"saml:Assertion"`
}

type Status struct {
	StatusCode StatusCode `xml:"samlp:StatusCode"`
}

type StatusCode struct {
	Value string `xml:"Value,attr"`
}

type Assertion struct {
	XMLName xml.Name `xml:"saml:Assertion"`
	//Xmlns              string             `xml:"xmlns,attr"`
	XmlnsXsi           string             `xml:"xmlns:xsi,attr"`
	XmlnsXs            string             `xml:"xmlns:xs,attr"`
	ID                 string             `xml:"ID,attr"`
	Version            string             `xml:"Version,attr"`
	IssueInstant       string             `xml:"IssueInstant,attr"`
	Issuer             string             `xml:"saml:Issuer"`
	Subject            Subject            `xml:"saml:Subject"`
	Conditions         Conditions         `xml:"saml:Conditions"`
	AuthnStatement     AuthnStatement     `xml:"saml:AuthnStatement"`
	AttributeStatement AttributeStatement `xml:"saml:AttributeStatement"`
}

type Signature struct {
	XMLName        xml.Name        `xml:"ds:Signature"`
	Xmlns          string          `xml:"xmlns:ds,attr"`
	SignedInfo     SignedInfo      `xml:"ds:SignedInfo"`
	SignatureValue string          `xml:"ds:SignatureValue"`
	KeyInfo        ResponseKeyInfo `xml:"ds:KeyInfo"`
}

type SignedInfo struct {
	XMLName                xml.Name               `xml:"ds:SignedInfo"`
	CanonicalizationMethod CanonicalizationMethod `xml:"ds:CanonicalizationMethod"`
	SignatureMethod        SignatureMethod        `xml:"ds:SignatureMethod"`
	SamlsigReference       SamlsigReference       `xml:"ds:Reference"`
}

type CanonicalizationMethod struct {
	XMLName   xml.Name `xml:"ds:CanonicalizationMethod"`
	Algorithm string   `xml:"Algorithm,attr"`
}

type SignatureMethod struct {
	XMLName   xml.Name `xml:"ds:SignatureMethod"`
	Algorithm string   `xml:"Algorithm,attr"`
}

type SamlsigReference struct {
	XMLName      xml.Name     `xml:"ds:Reference"`
	URI          string       `xml:"URI,attr"`
	Transforms   Transforms   `xml:"ds:Transforms"`
	DigestMethod DigestMethod `xml:"ds:DigestMethod"`
	DigestValue  string       `xml:"ds:DigestValue"`
}

type Transforms struct {
	XMLName   xml.Name    `xml:"ds:Transforms"`
	Transform []Transform `xml:"ds:Transform"`
}

type Transform struct {
	XMLName   xml.Name `xml:"ds:Transform"`
	Algorithm string   `xml:"Algorithm,attr"`
}

type DigestMethod struct {
	XMLName   xml.Name `xml:"ds:DigestMethod"`
	Algorithm string   `xml:"Algorithm,attr"`
}

type ResponseKeyInfo struct {
	XMLName          xml.Name         `xml:"ds:KeyInfo"`
	ResponseX509Data ResponseX509Data `xml:"ds:X509Data"`
}

type ResponseX509Data struct {
	XMLName         xml.Name `xml:"ds:X509Data"`
	X509Certificate string   `xml:"ds:X509Certificate"`
}

type Subject struct {
	NameID              NameID              `xml:"saml:NameID"`
	SubjectConfirmation SubjectConfirmation `xml:"saml:SubjectConfirmation"`
}

type NameID struct {
	SPNameQualifier string `xml:"SPNameQualifier,attr"`
	Format          string `xml:"Format,attr"`
	Value           string `xml:",innerxml"`
}

type SubjectConfirmation struct {
	Method                  string                  `xml:"Method,attr"`
	SubjectConfirmationData SubjectConfirmationData `xml:"saml:SubjectConfirmationData"`
}

type SubjectConfirmationData struct {
	NotOnOrAfter string `xml:"NotOnOrAfter,attr"`
	Recipient    string `xml:"Recipient,attr"`
	InResponseTo string `xml:"InResponseTo,attr"`
}

type Conditions struct {
	NotBefore           string              `xml:"NotBefore,attr"`
	NotOnOrAfter        string              `xml:"NotOnOrAfter,attr"`
	AudienceRestriction AudienceRestriction `xml:"saml:AudienceRestriction"`
}

type AudienceRestriction struct {
	Audience string `xml:"saml:Audience"`
}

type AuthnStatement struct {
	AuthnInstant        string       `xml:"AuthnInstant,attr"`
	SessionNotOnOrAfter string       `xml:"SessionNotOnOrAfter,attr"`
	SessionIndex        string       `xml:"SessionIndex,attr"`
	AuthnContext        AuthnContext `xml:"saml:AuthnContext"`
}

type AuthnContext struct {
	AuthnContextClassRef string `xml:"saml:AuthnContextClassRef"`
}

type AttributeStatement struct {
	Attribute []ResponseAttribute `xml:"saml:Attribute"`
}

type ResponseAttribute struct {
	Name           string   `xml:"Name,attr"`
	NameFormat     string   `xml:"NameFormat,attr"`
	AttributeValue []string `xml:"saml:AttributeValue"`
}

func NewSamlResponse(user *User, host string, publicKey string) *SamlResponse {
	now := time.Now().UTC()
	return &SamlResponse{
		XMLName: xml.Name{
			Local: "samlp:Response",
		},
		XmlnsSamlp:   "urn:oasis:names:tc:SAML:2.0:protocol",
		XmlnsSaml:    "urn:oasis:names:tc:SAML:2.0:assertion",
		ID:           SamlID(),
		Version:      "2.0",
		IssueInstant: now.Format(time.RFC3339),
		Issuer:       host,
		Signature: Signature{
			Xmlns: "http://www.w3.org/2000/09/xmldsig#",
			SignedInfo: SignedInfo{
				CanonicalizationMethod: CanonicalizationMethod{
					Algorithm: "http://www.w3.org/2001/10/xml-exc-c14n#",
				},
				SignatureMethod: SignatureMethod{
					Algorithm: "http://www.w3.org/2000/09/xmldsig#rsa-sha1",
				},
				SamlsigReference: SamlsigReference{
					URI: "", // caller must populate "#" + ar.Id,
					Transforms: Transforms{
						Transform: []Transform{{
							Algorithm: "http://www.w3.org/2000/09/xmldsig#enveloped-signature",
						}},
					},
					DigestMethod: DigestMethod{
						Algorithm: "http://www.w3.org/2000/09/xmldsig#sha1",
					},
				},
			},
			KeyInfo: ResponseKeyInfo{
				ResponseX509Data: ResponseX509Data{
					X509Certificate: publicKey,
				},
			},
		},
		Status: Status{
			StatusCode: StatusCode{
				Value: "urn:oasis:names:tc:SAML:2.0:status:Success",
			},
		},
		Assertion: Assertion{
			XMLName: xml.Name{
				Local: "saml:Assertion",
			},
			//Xmlns:        "urn:oasis:names:tc:SAML:2.0:assertion",
			XmlnsXsi:     "http://www.w3.org/2001/XMLSchema-instance",
			XmlnsXs:      "http://www.w3.org/2001/XMLSchema",
			ID:           SamlID(),
			Version:      "2.0",
			IssueInstant: now.Format(time.RFC3339),
			Issuer:       host,
			Subject: Subject{
				NameID: NameID{
					SPNameQualifier: host,
					Format:          "urn:oasis:names:tc:SAML:2.0:nameid-format:transient",
					Value:           user.Id,
				},
				SubjectConfirmation: SubjectConfirmation{
					Method: "urn:oasis:names:tc:SAML:2.0:cm:bearer",
					SubjectConfirmationData: SubjectConfirmationData{
						NotOnOrAfter: now.Add(time.Hour * 24).Format(time.RFC3339),
						Recipient:    host,
						InResponseTo: fmt.Sprintf("Casdoor_%s", user.Id),
					},
				},
			},
			Conditions: Conditions{
				NotBefore:    now.Format(time.RFC3339),
				NotOnOrAfter: now.Add(time.Hour * 24).Format(time.RFC3339),
				AudienceRestriction: AudienceRestriction{
					Audience: host,
				},
			},
			AuthnStatement: AuthnStatement{
				AuthnInstant:        now.Format(time.RFC3339),
				SessionNotOnOrAfter: now.Add(time.Hour * 24).Format(time.RFC3339),
				SessionIndex:        SamlID(),
				AuthnContext: AuthnContext{
					AuthnContextClassRef: "urn:oasis:names:tc:SAML:2.0:ac:classes:Password",
				},
			},
			AttributeStatement: AttributeStatement{
				Attribute: []ResponseAttribute{
					{
						Name:           "uid",
						NameFormat:     "urn:oasis:names:tc:SAML:2.0:attrname-format:basic",
						AttributeValue: []string{user.Id},
					},
					{
						Name:           "mail",
						NameFormat:     "urn:oasis:names:tc:SAML:2.0:attrname-format:basic",
						AttributeValue: []string{user.Email},
					},
				},
			},
		},
		Destination:  host,
		InResponseTo: fmt.Sprintf("Casdoor_%s", user.Id),
	}
}

// UUID generate a new V4 UUID
func SamlID() string {
	u := uuid.NewV4()
	return "_" + u.String()
}
