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
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/beevik/etree"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	saml "github.com/russellhaering/gosaml2"
	dsig "github.com/russellhaering/goxmldsig"
)

// SAMLTimeFormat is the time format for SAML assertions, compliant with xs:dateTime
// Format: YYYY-MM-DDTHH:MM:SSZ (ISO 8601 / RFC 3339 compatible)
const SAMLTimeFormat = "2006-01-02T15:04:05Z"

// NewSamlResponse
// returns a saml2 response
func NewSamlResponse(application *Application, user *User, host string, certificate string, destination string, iss string, requestId string, redirectUri []string) (*etree.Element, error) {
	samlResponse := &etree.Element{
		Space: "samlp",
		Tag:   "Response",
	}
	now := time.Now().UTC().Format(SAMLTimeFormat)
	expireTime := time.Now().UTC().Add(time.Hour * 24).Format(SAMLTimeFormat)
	samlResponse.CreateAttr("xmlns:samlp", "urn:oasis:names:tc:SAML:2.0:protocol")
	samlResponse.CreateAttr("xmlns:saml", "urn:oasis:names:tc:SAML:2.0:assertion")
	samlResponse.CreateAttr("xmlns:xsi", "http://www.w3.org/2001/XMLSchema-instance")
	samlResponse.CreateAttr("xmlns:xs", "http://www.w3.org/2001/XMLSchema")
	arId := uuid.New()

	samlResponse.CreateAttr("ID", fmt.Sprintf("_%s", arId))
	samlResponse.CreateAttr("Version", "2.0")
	samlResponse.CreateAttr("IssueInstant", now)
	samlResponse.CreateAttr("Destination", destination)
	samlResponse.CreateAttr("InResponseTo", requestId)
	samlResponse.CreateElement("saml:Issuer").SetText(host)

	samlResponse.CreateElement("samlp:Status").CreateElement("samlp:StatusCode").CreateAttr("Value", "urn:oasis:names:tc:SAML:2.0:status:Success")

	assertion := samlResponse.CreateElement("saml:Assertion")
	assertion.CreateAttr("xmlns:saml", "urn:oasis:names:tc:SAML:2.0:assertion")
	assertion.CreateAttr("xmlns:xsi", "http://www.w3.org/2001/XMLSchema-instance")
	assertion.CreateAttr("xmlns:xs", "http://www.w3.org/2001/XMLSchema")
	assertion.CreateAttr("ID", fmt.Sprintf("_%s", uuid.New()))
	assertion.CreateAttr("Version", "2.0")
	assertion.CreateAttr("IssueInstant", now)
	assertion.CreateElement("saml:Issuer").SetText(host)
	subject := assertion.CreateElement("saml:Subject")
	nameIDValue := user.Name
	if application.UseEmailAsSamlNameId {
		nameIDValue = user.Email
	}
	nameId := subject.CreateElement("saml:NameID")
	if application.UseEmailAsSamlNameId {
		nameId.CreateAttr("Format", "urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress")
	} else {
		nameId.CreateAttr("Format", "urn:oasis:names:tc:SAML:1.1:nameid-format:unspecified")
	}
	nameId.SetText(nameIDValue)
	subjectConfirmation := subject.CreateElement("saml:SubjectConfirmation")
	subjectConfirmation.CreateAttr("Method", "urn:oasis:names:tc:SAML:2.0:cm:bearer")
	subjectConfirmationData := subjectConfirmation.CreateElement("saml:SubjectConfirmationData")
	subjectConfirmationData.CreateAttr("InResponseTo", requestId)
	subjectConfirmationData.CreateAttr("Recipient", destination)
	subjectConfirmationData.CreateAttr("NotOnOrAfter", expireTime)
	condition := assertion.CreateElement("saml:Conditions")
	condition.CreateAttr("NotBefore", now)
	condition.CreateAttr("NotOnOrAfter", expireTime)
	audience := condition.CreateElement("saml:AudienceRestriction")
	audience.CreateElement("saml:Audience").SetText(iss)
	// Add redirect URIs as audiences, but skip duplicates and empty values
	for _, value := range redirectUri {
		if value != "" && value != iss {
			audience.CreateElement("saml:Audience").SetText(value)
		}
	}
	authnStatement := assertion.CreateElement("saml:AuthnStatement")
	authnStatement.CreateAttr("AuthnInstant", now)
	authnStatement.CreateAttr("SessionIndex", fmt.Sprintf("_%s", uuid.New()))
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

	err := ExtendUserWithRolesAndPermissions(user)
	if err != nil {
		return nil, err
	}

	for _, item := range application.SamlAttributes {
		role := attributes.CreateElement("saml:Attribute")
		role.CreateAttr("Name", item.Name)
		role.CreateAttr("NameFormat", item.NameFormat)

		valueList := replaceAttributeValue(user, item.Value)
		for _, value := range valueList {
			av := role.CreateElement("saml:AttributeValue")
			av.CreateAttr("xsi:type", "xs:string").Element().SetText(value)
		}
	}

	roles := attributes.CreateElement("saml:Attribute")
	roles.CreateAttr("Name", "Roles")
	roles.CreateAttr("NameFormat", "urn:oasis:names:tc:SAML:2.0:attrname-format:basic")

	for _, role := range user.Roles {
		roles.CreateElement("saml:AttributeValue").CreateAttr("xsi:type", "xs:string").Element().SetText(role.Name)
	}

	return samlResponse, nil
}

// ensureNamespaces ensures that xsi and xs namespaces are present on Response and Assertion elements
// This is needed because C14N10 Exclusive Canonicalization may remove namespace declarations
// during the canonicalization process, even if they are used in attributes like xsi:type="xs:string"
func ensureNamespaces(samlResponse *etree.Element) {
	xsiNS := "http://www.w3.org/2001/XMLSchema-instance"
	xsNS := "http://www.w3.org/2001/XMLSchema"

	// Ensure namespaces on Response element
	// Check if namespaces exist and update/add them
	setNamespaceAttr(samlResponse, "xmlns:xsi", xsiNS)
	setNamespaceAttr(samlResponse, "xmlns:xs", xsNS)

	// Find and ensure namespaces on Assertion element
	assertion := samlResponse.FindElement("./Assertion")
	if assertion != nil {
		setNamespaceAttr(assertion, "xmlns:xsi", xsiNS)
		setNamespaceAttr(assertion, "xmlns:xs", xsNS)
	}
}

// setNamespaceAttr sets a namespace attribute on an element, removing any existing one first
func setNamespaceAttr(elem *etree.Element, key, value string) {
	// Remove existing attribute if present by filtering the Attr slice
	newAttrs := []etree.Attr{}
	for _, attr := range elem.Attr {
		if attr.Key != key {
			newAttrs = append(newAttrs, attr)
		}
	}
	elem.Attr = newAttrs
	// Add the new attribute
	elem.CreateAttr(key, value)
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

// IdpEntityDescriptor
// SAML METADATA
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
	// XMLName xml.Name
	Value string `xml:",innerxml"`
}

type SingleSignOnService struct {
	// XMLName  xml.Name
	Binding  string `xml:"Binding,attr"`
	Location string `xml:"Location,attr"`
}

type Attribute struct {
	// XMLName      xml.Name
	Xmlns        string   `xml:"xmlns,attr"`
	Name         string   `xml:"Name,attr"`
	NameFormat   string   `xml:"NameFormat,attr"`
	FriendlyName string   `xml:"FriendlyName,attr"`
	Values       []string `xml:"AttributeValue"`
}

func GetSamlMeta(application *Application, host string, enablePostBinding bool) (*IdpEntityDescriptor, error) {
	cert, err := getCertByApplication(application)
	if err != nil {
		return nil, err
	}

	if cert == nil {
		return nil, errors.New("please set a cert for the application first")
	}

	if cert.Certificate == "" {
		return nil, fmt.Errorf("the certificate field should not be empty for the cert: %v", cert)
	}

	block, _ := pem.Decode([]byte(cert.Certificate))
	certificate := base64.StdEncoding.EncodeToString(block.Bytes)

	originFrontend, originBackend := getOriginFromHost(host)

	idpLocation := ""
	idpBinding := ""
	if enablePostBinding {
		idpLocation = fmt.Sprintf("%s/api/saml/redirect/%s/%s", originBackend, application.Owner, application.Name)
		idpBinding = "urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST"
	} else {
		idpLocation = fmt.Sprintf("%s/login/saml/authorize/%s/%s", originFrontend, application.Owner, application.Name)
		idpBinding = "urn:oasis:names:tc:SAML:2.0:bindings:HTTP-Redirect"
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
							Cert: certificate,
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
				Binding:  idpBinding,
				Location: idpLocation,
			},
			ProtocolSupportEnumeration: "urn:oasis:names:tc:SAML:2.0:protocol",
		},
	}

	return &d, nil
}

// GetSamlResponse generates a SAML2.0 response
// parameter samlRequest is saml request in base64 format
func GetSamlResponse(application *Application, user *User, samlRequest string, host string) (string, string, string, error) {
	// request type
	method := "GET"
	samlRequest = strings.ReplaceAll(samlRequest, " ", "+")
	// base64 decode
	defated, err := base64.StdEncoding.DecodeString(samlRequest)
	if err != nil {
		return "", "", "", fmt.Errorf("err: Failed to decode SAML request, %s", err.Error())
	}

	var requestByte []byte

	if strings.Contains(string(defated), "xmlns:") {
		requestByte = defated
	} else {
		// decompress
		var buffer bytes.Buffer
		rdr := flate.NewReader(bytes.NewReader(defated))

		for {

			_, err = io.CopyN(&buffer, rdr, 1024)
			if err != nil {
				if err == io.EOF {
					break
				}
				return "", "", "", err
			}
		}

		requestByte = buffer.Bytes()
	}

	var authnRequest saml.AuthNRequest
	err = xml.Unmarshal(requestByte, &authnRequest)
	if err != nil {
		return "", "", "", fmt.Errorf("err: Failed to unmarshal AuthnRequest, please check the SAML request, %s", err.Error())
	}

	// verify samlRequest
	if isValid := application.IsRedirectUriValid(authnRequest.Issuer); !isValid {
		return "", "", "", fmt.Errorf("err: Issuer URI: %s doesn't exist in the allowed Redirect URI list", authnRequest.Issuer)
	}

	// get certificate string
	cert, err := getCertByApplication(application)
	if err != nil {
		return "", "", "", err
	}

	if cert.Certificate == "" {
		return "", "", "", fmt.Errorf("the certificate field should not be empty for the cert: %v", cert)
	}

	block, _ := pem.Decode([]byte(cert.Certificate))
	certificate := base64.StdEncoding.EncodeToString(block.Bytes)

	// redirect Url (Assertion Consumer Url)
	if application.SamlReplyUrl != "" {
		method = "POST"
		authnRequest.AssertionConsumerServiceURL = application.SamlReplyUrl
	} else if authnRequest.AssertionConsumerServiceURL == "" {
		return "", "", "", fmt.Errorf("err: SAML request don't has attribute 'AssertionConsumerServiceURL' in <samlp:AuthnRequest>")
	}
	if authnRequest.ProtocolBinding == "urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST" {
		method = "POST"
	}

	_, originBackend := getOriginFromHost(host)

	// build signedResponse
	samlResponse, err := NewSamlResponse(application, user, originBackend, certificate, authnRequest.AssertionConsumerServiceURL, authnRequest.Issuer, authnRequest.ID, application.RedirectUris)
	if err != nil {
		return "", "", "", fmt.Errorf("err: NewSamlResponse() error, %s", err.Error())
	}

	randomKeyStore := &X509Key{
		PrivateKey:      cert.PrivateKey,
		X509Certificate: certificate,
	}
	ctx := dsig.NewDefaultSigningContext(randomKeyStore)
	if application.SamlHashAlgorithm == "" || application.SamlHashAlgorithm == "SHA1" {
		ctx.Hash = crypto.SHA1
	} else if application.SamlHashAlgorithm == "SHA256" {
		ctx.Hash = crypto.SHA256
	} else if application.SamlHashAlgorithm == "SHA512" {
		ctx.Hash = crypto.SHA512
	}

	if application.EnableSamlC14n10 {
		ctx.Canonicalizer = dsig.MakeC14N10ExclusiveCanonicalizerWithPrefixList("")
		// Ensure xsi and xs namespaces are present on Response and Assertion elements BEFORE signing
		// This is critical for C14N10 which may remove namespace declarations during canonicalization
		// If we add namespaces after signing, the XML won't match the signature
		ensureNamespaces(samlResponse)
	}

	// signedXML, err := ctx.SignEnvelopedLimix(samlResponse)
	// if err != nil {
	//	return "", "", fmt.Errorf("err: %s", err.Error())
	// }

	// Sign the assertion (SAML 2.0 best practice)
	assertion := samlResponse.FindElement("./Assertion")
	if assertion != nil {
		assertionSig, err := ctx.ConstructSignature(assertion, true)
		if err != nil {
			return "", "", "", fmt.Errorf("err: Failed to sign SAML assertion, %s", err.Error())
		}
		// Insert signature as the second child of assertion (after Issuer)
		assertion.InsertChildAt(1, assertionSig)
	}

	// Sign the response
	sig, err := ctx.ConstructSignature(samlResponse, true)
	if err != nil {
		return "", "", "", fmt.Errorf("err: Failed to serializes the SAML request into bytes, %s", err.Error())
	}

	samlResponse.InsertChildAt(1, sig)

	doc := etree.NewDocument()
	doc.SetRoot(samlResponse)

	// Write to bytes and ensure namespaces are preserved in the final XML
	xmlBytes, err := doc.WriteToBytes()
	if err != nil {
		return "", "", "", fmt.Errorf("err: Failed to serializes the SAML request into bytes, %s", err.Error())
	}

	// compress
	if application.EnableSamlCompress {
		flated := bytes.NewBuffer(nil)
		writer, err := flate.NewWriter(flated, flate.DefaultCompression)
		if err != nil {
			return "", "", "", err
		}

		_, err = writer.Write(xmlBytes)
		if err != nil {
			return "", "", "", err
		}

		err = writer.Close()
		if err != nil {
			return "", "", "", err
		}

		xmlBytes = flated.Bytes()
	}
	// base64 encode
	res := base64.StdEncoding.EncodeToString(xmlBytes)
	return res, authnRequest.AssertionConsumerServiceURL, method, err
}

// NewSamlResponse11 return a saml1.1 response(not 2.0)
func NewSamlResponse11(application *Application, user *User, requestID string, host string) (*etree.Element, error) {
	samlResponse := &etree.Element{
		Space: "samlp",
		Tag:   "Response",
	}

	samlResponse.CreateAttr("xmlns:samlp", "urn:oasis:names:tc:SAML:1.0:protocol")
	samlResponse.CreateAttr("xmlns:xsi", "http://www.w3.org/2001/XMLSchema-instance")
	samlResponse.CreateAttr("xmlns:xs", "http://www.w3.org/2001/XMLSchema")
	samlResponse.CreateAttr("MajorVersion", "1")
	samlResponse.CreateAttr("MinorVersion", "1")

	responseID := uuid.New()
	samlResponse.CreateAttr("ResponseID", fmt.Sprintf("_%s", responseID))
	samlResponse.CreateAttr("InResponseTo", requestID)

	now := time.Now().UTC().Format(SAMLTimeFormat)
	expireTime := time.Now().UTC().Add(time.Hour * 24).Format(SAMLTimeFormat)

	samlResponse.CreateAttr("IssueInstant", now)

	samlResponse.CreateElement("samlp:Status").CreateElement("samlp:StatusCode").CreateAttr("Value", "samlp:Success")

	// create assertion which is inside the response
	assertion := samlResponse.CreateElement("saml:Assertion")
	assertion.CreateAttr("xmlns:saml", "urn:oasis:names:tc:SAML:1.0:assertion")
	assertion.CreateAttr("MajorVersion", "1")
	assertion.CreateAttr("MinorVersion", "1")
	assertion.CreateAttr("AssertionID", uuid.New().String())
	assertion.CreateAttr("Issuer", host)
	assertion.CreateAttr("IssueInstant", now)

	condition := assertion.CreateElement("saml:Conditions")
	condition.CreateAttr("NotBefore", now)
	condition.CreateAttr("NotOnOrAfter", expireTime)

	// AuthenticationStatement inside assertion
	authenticationStatement := assertion.CreateElement("saml:AuthenticationStatement")
	authenticationStatement.CreateAttr("AuthenticationMethod", "urn:oasis:names:tc:SAML:1.0:am:password")
	authenticationStatement.CreateAttr("AuthenticationInstant", now)

	// subject inside AuthenticationStatement
	subject := assertion.CreateElement("saml:Subject")
	// nameIdentifier inside subject
	nameIdentifier := subject.CreateElement("saml:NameIdentifier")
	// nameIdentifier.CreateAttr("Format", "urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress")
	if application.UseEmailAsSamlNameId {
		nameIdentifier.SetText(user.Email)
	} else {
		nameIdentifier.SetText(user.Name)
	}

	// subjectConfirmation inside subject
	subjectConfirmation := subject.CreateElement("saml:SubjectConfirmation")
	subjectConfirmation.CreateElement("saml:ConfirmationMethod").SetText("urn:oasis:names:tc:SAML:1.0:cm:artifact")

	attributeStatement := assertion.CreateElement("saml:AttributeStatement")
	subjectInAttribute := attributeStatement.CreateElement("saml:Subject")
	nameIdentifierInAttribute := subjectInAttribute.CreateElement("saml:NameIdentifier")
	if application.UseEmailAsSamlNameId {
		nameIdentifierInAttribute.SetText(user.Email)
	} else {
		nameIdentifierInAttribute.SetText(user.Name)
	}

	subjectConfirmationInAttribute := subjectInAttribute.CreateElement("saml:SubjectConfirmation")
	subjectConfirmationInAttribute.CreateElement("saml:ConfirmationMethod").SetText("urn:oasis:names:tc:SAML:1.0:cm:artifact")

	data, err := json.Marshal(user)
	if err != nil {
		return nil, err
	}

	tmp := map[string]string{}
	err = json.Unmarshal(data, &tmp)
	if err != nil {
		return nil, err
	}

	for k, v := range tmp {
		if v != "" {
			attr := attributeStatement.CreateElement("saml:Attribute")
			attr.CreateAttr("saml:AttributeName", k)
			attr.CreateAttr("saml:AttributeNamespace", "http://www.ja-sig.org/products/cas/")
			attr.CreateElement("saml:AttributeValue").SetText(v)
		}
	}

	return samlResponse, nil
}

func GetSamlRedirectAddress(owner string, application string, relayState string, samlRequest string, host string, username string, loginHint string) string {
	originF, _ := getOriginFromHost(host)
	baseURL := fmt.Sprintf("%s/login/saml/authorize/%s/%s?relayState=%s&samlRequest=%s", originF, owner, application, relayState, samlRequest)
	if username != "" {
		baseURL += fmt.Sprintf("&username=%s", url.QueryEscape(username))
	}
	if loginHint != "" {
		baseURL += fmt.Sprintf("&login_hint=%s", url.QueryEscape(loginHint))
	}
	return baseURL
}
