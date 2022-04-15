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
	"bytes"
	"compress/flate"
	"crypto"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"encoding/xml"
	"fmt"
	"io"
	"time"

	"github.com/RobotsAndPencils/go-saml"
	"github.com/astaxie/beego"
	"github.com/beevik/etree"
	"github.com/golang-jwt/jwt/v4"
	dsig "github.com/russellhaering/goxmldsig"
	uuid "github.com/satori/go.uuid"
)

//returns a saml2 response
func NewSamlResponse(user *User, host string, publicKey string, destination string, iss string, redirectUri []string) (*etree.Element, error) {
	samlResponse := &etree.Element{
		Space: "samlp",
		Tag:   "Response",
	}
	now := time.Now().UTC().Format(time.RFC3339)
	expireTime := time.Now().UTC().Add(time.Hour * 24).Format(time.RFC3339)
	samlResponse.CreateAttr("xmlns:samlp", "urn:oasis:names:tc:SAML:2.0:protocol")
	samlResponse.CreateAttr("xmlns:saml", "urn:oasis:names:tc:SAML:2.0:assertion")
	arId := uuid.NewV4()

	samlResponse.CreateAttr("ID", fmt.Sprintf("_%s", arId))
	samlResponse.CreateAttr("Version", "2.0")
	samlResponse.CreateAttr("IssueInstant", now)
	samlResponse.CreateAttr("Destination", destination)
	samlResponse.CreateAttr("InResponseTo", fmt.Sprintf("Casdoor_%s", arId))
	samlResponse.CreateElement("saml:Issuer").SetText(host)

	samlResponse.CreateElement("samlp:Status").CreateElement("samlp:StatusCode").CreateAttr("Value", "urn:oasis:names:tc:SAML:2.0:status:Success")

	assertion := samlResponse.CreateElement("saml:Assertion")
	assertion.CreateAttr("xmlns:xsi", "http://www.w3.org/2001/XMLSchema-instance")
	assertion.CreateAttr("xmlns:xs", "http://www.w3.org/2001/XMLSchema")
	assertion.CreateAttr("ID", fmt.Sprintf("_%s", uuid.NewV4()))
	assertion.CreateAttr("Version", "2.0")
	assertion.CreateAttr("IssueInstant", now)
	assertion.CreateElement("saml:Issuer").SetText(host)
	subject := assertion.CreateElement("saml:Subject")
	subject.CreateElement("saml:NameID").SetText(user.Email)
	subjectConfirmation := subject.CreateElement("saml:SubjectConfirmation")
	subjectConfirmation.CreateAttr("Method", "urn:oasis:names:tc:SAML:2.0:cm:bearer")
	subjectConfirmationData := subjectConfirmation.CreateElement("saml:SubjectConfirmationData")
	subjectConfirmationData.CreateAttr("InResponseTo", fmt.Sprintf("_%s", arId))
	subjectConfirmationData.CreateAttr("Recipient", destination)
	subjectConfirmationData.CreateAttr("NotOnOrAfter", expireTime)
	condition := assertion.CreateElement("saml:Conditions")
	condition.CreateAttr("NotBefore", now)
	condition.CreateAttr("NotOnOrAfter", expireTime)
	audience := condition.CreateElement("saml:AudienceRestriction")
	audience.CreateElement("saml:Audience").SetText(iss)
	for _, value := range redirectUri {
		audience.CreateElement("saml:Audience").SetText(value)
	}
	authnStatement := assertion.CreateElement("saml:AuthnStatement")
	authnStatement.CreateAttr("AuthnInstant", now)
	authnStatement.CreateAttr("SessionIndex", fmt.Sprintf("_%s", uuid.NewV4()))
	authnStatement.CreateAttr("SessionNotOnOrAfter", expireTime)
	authnStatement.CreateElement("saml:AuthnContext").CreateElement("saml:AuthnContextClassRef").SetText("urn:oasis:names:tc:SAML:2.0:ac:classes:PasswordProtectedTransport")

	attributes := assertion.CreateElement("saml:AttributeStatement")
	email := attributes.CreateElement("saml:Attribute")
	email.CreateAttr("Name", "Email")
	email.CreateAttr("NameFormat", "urn:oasis:names:tc:SAML:2.0:attrname-format:basic")
	email.CreateElement("saml:AttributeValue").CreateAttr("xsi:type", "xs:string").Element().SetText(user.Email)
	name := attributes.CreateElement("saml:Attribute")
	name.CreateAttr("Name", "Name")
	name.CreateAttr("NameFormat", "urn:oasis:names:tc:SAML:2.0:attrname-format:basic")
	name.CreateElement("saml:AttributeValue").CreateAttr("xsi:type", "xs:string").Element().SetText(user.Name)
	displayName := attributes.CreateElement("saml:Attribute")
	displayName.CreateAttr("Name", "DisplayName")
	displayName.CreateAttr("NameFormat", "urn:oasis:names:tc:SAML:2.0:attrname-format:basic")
	displayName.CreateElement("saml:AttributeValue").CreateAttr("xsi:type", "xs:string").Element().SetText(user.DisplayName)

	return samlResponse, nil

}

type X509Key struct {
	X509Certificate string
	PrivateKey      string
}

func (x X509Key) GetKeyPair() (privateKey *rsa.PrivateKey, cert []byte, err error) {
	cert, _ = base64.StdEncoding.DecodeString(x.X509Certificate)
	privateKey, err = jwt.ParseRSAPrivateKeyFromPEM([]byte(x.PrivateKey))
	return privateKey, cert, err
}

//SAML METADATA
type IdpEntityDescriptor struct {
	XMLName  xml.Name `xml:"EntityDescriptor"`
	DS       string   `xml:"xmlns:ds,attr"`
	XMLNS    string   `xml:"xmlns,attr"`
	MD       string   `xml:"xmlns:md,attr"`
	EntityId string   `xml:"entityID,attr"`

	IdpSSODescriptor IdpSSODescriptor `xml:"IDPSSODescriptor"`
}

type KeyInfo struct {
	XMLName  xml.Name `xml:"http://www.w3.org/2000/09/xmldsig# KeyInfo"`
	X509Data X509Data `xml:",innerxml"`
}

type X509Data struct {
	XMLName         xml.Name        `xml:"http://www.w3.org/2000/09/xmldsig# X509Data"`
	X509Certificate X509Certificate `xml:",innerxml"`
}

type X509Certificate struct {
	XMLName xml.Name `xml:"http://www.w3.org/2000/09/xmldsig# X509Certificate"`
	Cert    string   `xml:",innerxml"`
}

type KeyDescriptor struct {
	XMLName xml.Name `xml:"KeyDescriptor"`
	Use     string   `xml:"use,attr"`
	KeyInfo KeyInfo  `xml:"KeyInfo"`
}

type IdpSSODescriptor struct {
	XMLName                    xml.Name `xml:"urn:oasis:names:tc:SAML:2.0:metadata IDPSSODescriptor"`
	ProtocolSupportEnumeration string   `xml:"protocolSupportEnumeration,attr"`
	SigningKeyDescriptor       KeyDescriptor
	NameIDFormats              []NameIDFormat      `xml:"NameIDFormat"`
	SingleSignOnService        SingleSignOnService `xml:"SingleSignOnService"`
	Attribute                  []Attribute         `xml:"Attribute"`
}

type NameIDFormat struct {
	XMLName xml.Name
	Value   string `xml:",innerxml"`
}

type SingleSignOnService struct {
	XMLName  xml.Name
	Binding  string `xml:"Binding,attr"`
	Location string `xml:"Location,attr"`
}

type Attribute struct {
	XMLName      xml.Name
	Name         string `xml:"Name,attr"`
	NameFormat   string `xml:"NameFormat,attr"`
	FriendlyName string `xml:"FriendlyName,attr"`
	Xmlns        string `xml:"xmlns,attr"`
}

func GetSamlMeta(application *Application, host string) (*IdpEntityDescriptor, error) {
	//_, originBackend := getOriginFromHost(host)
	cert := getCertByApplication(application)
	block, _ := pem.Decode([]byte(cert.PublicKey))
	publicKey := base64.StdEncoding.EncodeToString(block.Bytes)

	origin := beego.AppConfig.String("origin")
	originFrontend, originBackend := getOriginFromHost(host)
	if origin != "" {
		originBackend = origin
	}
	d := IdpEntityDescriptor{
		XMLName: xml.Name{
			Local: "md:EntityDescriptor",
		},
		DS:       "http://www.w3.org/2000/09/xmldsig#",
		XMLNS:    "urn:oasis:names:tc:SAML:2.0:metadata",
		MD:       "urn:oasis:names:tc:SAML:2.0:metadata",
		EntityId: originBackend,
		IdpSSODescriptor: IdpSSODescriptor{
			SigningKeyDescriptor: KeyDescriptor{
				Use: "signing",
				KeyInfo: KeyInfo{
					X509Data: X509Data{
						X509Certificate: X509Certificate{
							Cert: publicKey,
						},
					},
				},
			},
			NameIDFormats: []NameIDFormat{
				{Value: "urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress"},
				{Value: "urn:oasis:names:tc:SAML:2.0:nameid-format:persistent"},
				{Value: "urn:oasis:names:tc:SAML:2.0:nameid-format:transient"},
			},
			Attribute: []Attribute{
				{Xmlns: "urn:oasis:names:tc:SAML:2.0:assertion", Name: "Email", NameFormat: "urn:oasis:names:tc:SAML:2.0:attrname-format:basic", FriendlyName: "E-Mail"},
				{Xmlns: "urn:oasis:names:tc:SAML:2.0:assertion", Name: "DisplayName", NameFormat: "urn:oasis:names:tc:SAML:2.0:attrname-format:basic", FriendlyName: "displayName"},
				{Xmlns: "urn:oasis:names:tc:SAML:2.0:assertion", Name: "Name", NameFormat: "urn:oasis:names:tc:SAML:2.0:attrname-format:basic", FriendlyName: "Name"},
			},
			SingleSignOnService: SingleSignOnService{
				Binding:  "urn:oasis:names:tc:SAML:2.0:bindings:HTTP-Redirect",
				Location: fmt.Sprintf("%s/login/saml/authorize/%s/%s", originFrontend, application.Owner, application.Name),
			},
			ProtocolSupportEnumeration: "urn:oasis:names:tc:SAML:2.0:protocol",
		},
	}

	return &d, nil
}

//GenerateSamlResponse generates a SAML2.0 response
//parameter samlRequest is saml request in base64 format
func GetSamlResponse(application *Application, user *User, samlRequest string, host string) (string, string, error) {
	//decode samlRequest
	defated, err := base64.StdEncoding.DecodeString(samlRequest)
	if err != nil {
		return "", "", fmt.Errorf("err: %s", err.Error())
	}
	var buffer bytes.Buffer
	rdr := flate.NewReader(bytes.NewReader(defated))
	io.Copy(&buffer, rdr)
	var authnRequest saml.AuthnRequest
	err = xml.Unmarshal(buffer.Bytes(), &authnRequest)
	if err != nil {
		return "", "", fmt.Errorf("err: %s", err.Error())
	}
	//verify samlRequest
	if valid := CheckRedirectUriValid(application, authnRequest.Issuer.Url); !valid {
		return "", "", fmt.Errorf("err: invalid issuer url")
	}

	//get publickey string
	cert := getCertByApplication(application)
	block, _ := pem.Decode([]byte(cert.PublicKey))
	publicKey := base64.StdEncoding.EncodeToString(block.Bytes)

	_, originBackend := getOriginFromHost(host)

	//build signedResponse
	samlResponse, _ := NewSamlResponse(user, originBackend, publicKey, authnRequest.AssertionConsumerServiceURL, authnRequest.Issuer.Url, application.RedirectUris)
	randomKeyStore := &X509Key{
		PrivateKey:      cert.PrivateKey,
		X509Certificate: publicKey,
	}
	ctx := dsig.NewDefaultSigningContext(randomKeyStore)
	ctx.Hash = crypto.SHA1
	signedXML, err := ctx.SignEnveloped(samlResponse)
	if err != nil {
		return "", "", fmt.Errorf("err: %s", err.Error())
	}

	doc := etree.NewDocument()
	doc.SetRoot(signedXML)
	xmlStr, err := doc.WriteToString()
	if err != nil {
		return "", "", fmt.Errorf("err: %s", err.Error())
	}
	res := base64.StdEncoding.EncodeToString([]byte(xmlStr))
	return res, authnRequest.AssertionConsumerServiceURL, nil
}

//return a saml1.1 response(not 2.0)
func NewSamlResponse11(user *User, requestID string, host string) *etree.Element {
	samlResponse := &etree.Element{
		Space: "samlp",
		Tag:   "Response",
	}
	//create samlresponse
	samlResponse.CreateAttr("xmlns:samlp", "urn:oasis:names:tc:SAML:1.0:protocol")
	samlResponse.CreateAttr("MajorVersion", "1")
	samlResponse.CreateAttr("MinorVersion", "1")

	responseID := uuid.NewV4()
	samlResponse.CreateAttr("ResponseID", fmt.Sprintf("_%s", responseID))
	samlResponse.CreateAttr("InResponseTo", requestID)

	now := time.Now().UTC().Format(time.RFC3339)
	expireTime := time.Now().UTC().Add(time.Hour * 24).Format(time.RFC3339)

	samlResponse.CreateAttr("IssueInstant", now)

	samlResponse.CreateElement("samlp:Status").CreateElement("samlp:StatusCode").CreateAttr("Value", "samlp:Success")

	//create assertion which is inside the response
	assertion := samlResponse.CreateElement("saml:Assertion")
	assertion.CreateAttr("xmlns:saml", "urn:oasis:names:tc:SAML:1.0:assertion")
	assertion.CreateAttr("MajorVersion", "1")
	assertion.CreateAttr("MinorVersion", "1")
	assertion.CreateAttr("AssertionID", uuid.NewV4().String())
	assertion.CreateAttr("Issuer", host)
	assertion.CreateAttr("IssueInstant", now)

	condition := assertion.CreateElement("saml:Conditions")
	condition.CreateAttr("NotBefore", now)
	condition.CreateAttr("NotOnOrAfter", expireTime)

	//AuthenticationStatement inside assertion
	authenticationStatement := assertion.CreateElement("saml:AuthenticationStatement")
	authenticationStatement.CreateAttr("AuthenticationMethod", "urn:oasis:names:tc:SAML:1.0:am:password")
	authenticationStatement.CreateAttr("AuthenticationInstant", now)

	//subject inside AuthenticationStatement
	subject := assertion.CreateElement("saml:Subject")
	//nameIdentifier inside subject
	nameIdentifier := subject.CreateElement("saml:NameIdentifier")
	//nameIdentifier.CreateAttr("Format", "urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress")
	nameIdentifier.SetText(user.Name)

	//subjectConfirmation inside subject
	subjectConfirmation := subject.CreateElement("saml:SubjectConfirmation")
	subjectConfirmation.CreateElement("saml:ConfirmationMethod").SetText("urn:oasis:names:tc:SAML:1.0:cm:artifact")

	attributeStatement := assertion.CreateElement("saml:AttributeStatement")
	subjectInAttribute := attributeStatement.CreateElement("saml:Subject")
	nameIdentifierInAttribute := subjectInAttribute.CreateElement("saml:NameIdentifier")
	nameIdentifierInAttribute.SetText(user.Name)

	subjectConfirmationInAttribute := subjectInAttribute.CreateElement("saml:SubjectConfirmation")
	subjectConfirmationInAttribute.CreateElement("saml:ConfirmationMethod").SetText("urn:oasis:names:tc:SAML:1.0:cm:artifact")

	data, _ := json.Marshal(user)
	tmp := map[string]string{}
	json.Unmarshal(data, &tmp)

	for k, v := range tmp {
		if v != "" {
			attr := attributeStatement.CreateElement("saml:Attribute")
			attr.CreateAttr("saml:AttributeName", k)
			attr.CreateAttr("saml:AttributeNamespace", "http://www.ja-sig.org/products/cas/")
			attr.CreateElement("saml:AttributeValue").SetText(v)
		}
	}

	return samlResponse
}
