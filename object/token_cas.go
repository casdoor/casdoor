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
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"encoding/xml"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/beevik/etree"
	"github.com/casdoor/casdoor/i18n"
	"github.com/casdoor/casdoor/util"
	dsig "github.com/russellhaering/goxmldsig"
)

type CasServiceResponse struct {
	XMLName      xml.Name `xml:"cas:serviceResponse" json:"-"`
	Xmlns        string   `xml:"xmlns:cas,attr"`
	Failure      *CasAuthenticationFailure
	Success      *CasAuthenticationSuccess
	ProxySuccess *CasProxySuccess
	ProxyFailure *CasProxyFailure
}

type CasAuthenticationFailure struct {
	XMLName xml.Name `xml:"cas:authenticationFailure" json:"-"`
	Code    string   `xml:"code,attr"`
	Message string   `xml:",innerxml"`
}

type CasAuthenticationSuccess struct {
	XMLName             xml.Name           `xml:"cas:authenticationSuccess" json:"-"`
	User                string             `xml:"cas:user"`
	ProxyGrantingTicket string             `xml:"cas:proxyGrantingTicket,omitempty"`
	Proxies             *CasProxies        `xml:"cas:proxies"`
	Attributes          *CasAttributes     `xml:"cas:attributes"`
	ExtraAttributes     []*CasAnyAttribute `xml:",any"`
}

type CasProxies struct {
	XMLName xml.Name `xml:"cas:proxies" json:"-"`
	Proxies []string `xml:"cas:proxy"`
}

type CasAttributes struct {
	XMLName                                xml.Name  `xml:"cas:attributes" json:"-"`
	AuthenticationDate                     time.Time `xml:"cas:authenticationDate"`
	LongTermAuthenticationRequestTokenUsed bool      `xml:"cas:longTermAuthenticationRequestTokenUsed"`
	IsFromNewLogin                         bool      `xml:"cas:isFromNewLogin"`
	MemberOf                               []string  `xml:"cas:memberOf"`
	FirstName                              string    `xml:"cas:firstName,omitempty"`
	LastName                               string    `xml:"cas:lastName,omitempty"`
	Title                                  string    `xml:"cas:title,omitempty"`
	Email                                  string    `xml:"cas:email,omitempty"`
	Affiliation                            string    `xml:"cas:affiliation,omitempty"`
	Avatar                                 string    `xml:"cas:avatar,omitempty"`
	Phone                                  string    `xml:"cas:phone,omitempty"`
	DisplayName                            string    `xml:"cas:displayName,omitempty"`
	UserAttributes                         *CasUserAttributes
	ExtraAttributes                        []*CasAnyAttribute `xml:",any"`
}

type CasUserAttributes struct {
	XMLName       xml.Name             `xml:"cas:userAttributes" json:"-"`
	Attributes    []*CasNamedAttribute `xml:"cas:attribute"`
	AnyAttributes []*CasAnyAttribute   `xml:",any"`
}

type CasNamedAttribute struct {
	XMLName xml.Name `xml:"cas:attribute" json:"-"`
	Name    string   `xml:"name,attr,omitempty"`
	Value   string   `xml:",innerxml"`
}

type CasAnyAttribute struct {
	XMLName xml.Name
	Value   string `xml:",chardata"`
}

type CasAuthenticationSuccessWrapper struct {
	AuthenticationSuccess *CasAuthenticationSuccess // the token we issued
	Service               string                    // to which service this token is issued
	UserId                string
}

type CasProxySuccess struct {
	XMLName     xml.Name `xml:"cas:proxySuccess" json:"-"`
	ProxyTicket string   `xml:"cas:proxyTicket"`
}
type CasProxyFailure struct {
	XMLName xml.Name `xml:"cas:proxyFailure" json:"-"`
	Code    string   `xml:"code,attr"`
	Message string   `xml:",innerxml"`
}

type Saml11Request struct {
	XMLName           xml.Name `xml:"Request"`
	SAMLP             string   `xml:"samlp,attr"`
	MajorVersion      string   `xml:"MajorVersion,attr"`
	MinorVersion      string   `xml:"MinorVersion,attr"`
	RequestID         string   `xml:"RequestID,attr"`
	IssueInstant      string   `xml:"IssueInstance,attr"`
	AssertionArtifact Saml11AssertionArtifact
}
type Saml11AssertionArtifact struct {
	XMLName  xml.Name `xml:"AssertionArtifact"`
	InnerXML string   `xml:",innerxml"`
}

// st is short for service ticket
var stToServiceResponse sync.Map

// pgt is short for proxy granting ticket
var pgtToServiceResponse sync.Map

func CheckCasLogin(application *Application, lang string, service string) error {
	if len(application.RedirectUris) > 0 && !application.IsRedirectUriValid(service) {
		return fmt.Errorf(i18n.Translate(lang, "token:Redirect URI: %s doesn't exist in the allowed Redirect URI list"), service)
	}
	return nil
}

func StoreCasTokenForPgt(token *CasAuthenticationSuccess, service, userId string) string {
	pgt := fmt.Sprintf("PGT-%s", util.GenerateId())
	pgtToServiceResponse.Store(pgt, &CasAuthenticationSuccessWrapper{
		AuthenticationSuccess: token,
		Service:               service,
		UserId:                userId,
	})
	return pgt
}

func GenerateId() {
	panic("unimplemented")
}

// GetCasTokenByPgt
/**
@ret1: whether a token is found
@ret2: token, nil if not found
@ret3: the service URL who requested to issue this token
@ret4: userIf of user who requested to issue this token
*/
func GetCasTokenByPgt(pgt string) (bool, *CasAuthenticationSuccess, string, string) {
	if responseWrapperType, ok := pgtToServiceResponse.LoadAndDelete(pgt); ok {
		responseWrapperTypeCast := responseWrapperType.(*CasAuthenticationSuccessWrapper)
		return true, responseWrapperTypeCast.AuthenticationSuccess, responseWrapperTypeCast.Service, responseWrapperTypeCast.UserId
	}
	return false, nil, "", ""
}

// GetCasTokenByTicket
/**
@ret1: whether a token is found
@ret2: token, nil if not found
@ret3: the service URL who requested to issue this token
@ret4: userIf of user who requested to issue this token
*/
func GetCasTokenByTicket(ticket string) (bool, *CasAuthenticationSuccess, string, string) {
	if responseWrapperType, ok := stToServiceResponse.LoadAndDelete(ticket); ok {
		responseWrapperTypeCast := responseWrapperType.(*CasAuthenticationSuccessWrapper)
		return true, responseWrapperTypeCast.AuthenticationSuccess, responseWrapperTypeCast.Service, responseWrapperTypeCast.UserId
	}
	return false, nil, "", ""
}

func StoreCasTokenForProxyTicket(token *CasAuthenticationSuccess, targetService, userId string) string {
	proxyTicket := fmt.Sprintf("PT-%s", util.GenerateId())
	stToServiceResponse.Store(proxyTicket, &CasAuthenticationSuccessWrapper{
		AuthenticationSuccess: token,
		Service:               targetService,
		UserId:                userId,
	})
	return proxyTicket
}

func escapeXMLText(input string) (string, error) {
	var sb strings.Builder
	err := xml.EscapeText(&sb, []byte(input))
	if err != nil {
		return "", err
	}
	return sb.String(), nil
}

func GenerateCasToken(userId string, service string) (string, error) {
	user, err := GetUser(userId)
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", fmt.Errorf("The user: %s doesn't exist", userId)
	}

	user, _ = GetMaskedUser(user, false)

	user.WebauthnCredentials = nil
	user.Properties = nil

	authenticationSuccess := CasAuthenticationSuccess{
		User: user.Name,
		Attributes: &CasAttributes{
			AuthenticationDate: time.Now(),
			UserAttributes:     &CasUserAttributes{},
		},
		ProxyGrantingTicket: fmt.Sprintf("PGTIOU-%s", util.GenerateId()),
	}

	data, err := json.Marshal(user)
	if err != nil {
		return "", err
	}

	tmp := map[string]interface{}{}
	err = json.Unmarshal(data, &tmp)
	if err != nil {
		return "", err
	}

	for k, v := range tmp {
		value := fmt.Sprintf("%v", v)
		if value == "<nil>" || value == "[]" || value == "map[]" {
			value = ""
		}

		if value != "" {
			if escapedValue, err := escapeXMLText(value); err != nil {
				return "", err
			} else {
				value = escapedValue
			}
			switch k {
			case "firstName":
				authenticationSuccess.Attributes.FirstName = value
			case "lastName":
				authenticationSuccess.Attributes.LastName = value
			case "title":
				authenticationSuccess.Attributes.Title = value
			case "email":
				authenticationSuccess.Attributes.Email = value
			case "affiliation":
				authenticationSuccess.Attributes.Affiliation = value
			case "avatar":
				authenticationSuccess.Attributes.Avatar = value
			case "phone":
				authenticationSuccess.Attributes.Phone = value
			case "displayName":
				authenticationSuccess.Attributes.DisplayName = value
			}
			authenticationSuccess.Attributes.UserAttributes.Attributes = append(authenticationSuccess.Attributes.UserAttributes.Attributes, &CasNamedAttribute{
				Name:  k,
				Value: value,
			})
		}
	}

	st := fmt.Sprintf("ST-%d", rand.Int())
	stToServiceResponse.Store(st, &CasAuthenticationSuccessWrapper{
		AuthenticationSuccess: &authenticationSuccess,
		Service:               service,
		UserId:                userId,
	})
	return st, nil
}

// GetValidationBySaml
/**
@ret1: saml response
@ret2: the service URL who requested to issue this token
@ret3: error
*/
func GetValidationBySaml(samlRequest string, host string) (string, string, error) {
	var request Saml11Request
	err := xml.Unmarshal([]byte(samlRequest), &request)
	if err != nil {
		return "", "", err
	}

	ticket := request.AssertionArtifact.InnerXML
	if ticket == "" {
		return "", "", fmt.Errorf("request.AssertionArtifact.InnerXML error, AssertionArtifact field not found")
	}

	ok, _, service, userId := GetCasTokenByTicket(ticket)
	if !ok {
		return "", "", fmt.Errorf("the CAS token for ticket %s is not found", ticket)
	}

	user, err := GetUser(userId)
	if err != nil {
		return "", "", err
	}

	if user == nil {
		return "", "", fmt.Errorf("the user %s is not found", userId)
	}

	application, err := GetApplicationByUser(user)
	if err != nil {
		return "", "", err
	}
	if application == nil {
		return "", "", fmt.Errorf("the application for user %s is not found", userId)
	}

	samlResponse, err := NewSamlResponse11(application, user, request.RequestID, host)
	if err != nil {
		return "", "", err
	}

	cert, err := getCertByApplication(application)
	if err != nil {
		return "", "", err
	}

	if cert.Certificate == "" {
		return "", "", fmt.Errorf("the certificate field should not be empty for the cert: %v", cert)
	}

	block, _ := pem.Decode([]byte(cert.Certificate))
	certificate := base64.StdEncoding.EncodeToString(block.Bytes)
	randomKeyStore := &X509Key{
		PrivateKey:      cert.PrivateKey,
		X509Certificate: certificate,
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
	return xmlStr, service, nil
}

func (c *CasAuthenticationSuccess) DeepCopy() CasAuthenticationSuccess {
	res := *c
	// copy proxy
	if c.Proxies != nil {
		tmp := c.Proxies.DeepCopy()
		res.Proxies = &tmp
	}
	if c.Attributes != nil {
		tmp := c.Attributes.DeepCopy()
		res.Attributes = &tmp
	}
	res.ExtraAttributes = make([]*CasAnyAttribute, len(c.ExtraAttributes))
	for i, e := range c.ExtraAttributes {
		tmp := *e
		res.ExtraAttributes[i] = &tmp
	}
	return res
}

func (c *CasProxies) DeepCopy() CasProxies {
	res := CasProxies{
		Proxies: make([]string, len(c.Proxies)),
	}
	copy(res.Proxies, c.Proxies)
	return res
}

func (c *CasAttributes) DeepCopy() CasAttributes {
	res := *c
	if c.MemberOf != nil {
		res.MemberOf = make([]string, len(c.MemberOf))
		copy(res.MemberOf, c.MemberOf)
	}
	tmp := c.UserAttributes.DeepCopy()
	res.UserAttributes = &tmp

	res.ExtraAttributes = make([]*CasAnyAttribute, len(c.ExtraAttributes))
	for i, e := range c.ExtraAttributes {
		tmp := *e
		res.ExtraAttributes[i] = &tmp
	}
	return res
}

func (c *CasUserAttributes) DeepCopy() CasUserAttributes {
	res := CasUserAttributes{
		AnyAttributes: make([]*CasAnyAttribute, len(c.AnyAttributes)),
		Attributes:    make([]*CasNamedAttribute, len(c.Attributes)),
	}
	for i, a := range c.AnyAttributes {
		tmp := *a
		res.AnyAttributes[i] = &tmp
	}
	for i, a := range c.Attributes {
		tmp := *a
		res.Attributes[i] = &tmp
	}
	return res
}
