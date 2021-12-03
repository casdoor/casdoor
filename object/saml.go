// Copyright 2021 The casbin Authors. All Rights Reserved.
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
	"crypto/x509"
	"encoding/base64"
	"net/url"
	"regexp"
	"strings"

	"github.com/astaxie/beego"
	saml2 "github.com/russellhaering/gosaml2"
	dsig "github.com/russellhaering/goxmldsig"
)

func ParseSamlResponse(samlResponse string) string {
	certStore := dsig.MemoryX509CertificateStore{
		Roots: []*x509.Certificate{},
	}
	samlResponse, _ = url.QueryUnescape(samlResponse)
	de, err := base64.StdEncoding.DecodeString(samlResponse)
	if err != nil {
		panic(err)
	}
	deStr := strings.Replace(string(de), "\n", "", -1)
	res := regexp.MustCompile(`<ds:X509Certificate>(.*?)</ds:X509Certificate>`).FindAllStringSubmatch(deStr, -1)
	str := res[0][0]
	str = str[20 : len(str)-21]

	certData, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		panic(err)
	}
	idpCert, err := x509.ParseCertificate(certData)
	if err != nil {
		panic(err)
	}
	certStore.Roots = append(certStore.Roots, idpCert)

	samlOrigin := beego.AppConfig.String("samlOrigin")
	sp := &saml2.SAMLServiceProvider{
		ServiceProviderIssuer:       samlOrigin + "/api/acs",
		AssertionConsumerServiceURL: samlOrigin + "/api/acs",
		IDPCertificateStore:         &certStore,
	}
	assertionInfo, err := sp.RetrieveAssertionInfo(samlResponse)
	if err != nil {
		panic(err)
	}
	return assertionInfo.NameID
}

func GenerateSamlLoginUrl(id string) string {
	certStore := dsig.MemoryX509CertificateStore{
		Roots: []*x509.Certificate{},
	}
	provider := GetProvider(id)
	certData, err := base64.StdEncoding.DecodeString(provider.IdP)
	if err != nil {
		panic(err)
	}
	idpCert, err := x509.ParseCertificate(certData)
	if err != nil {
		panic(err)
	}
	certStore.Roots = append(certStore.Roots, idpCert)
	randomKeyStore := dsig.RandomKeyStoreForTest()
	samlOrigin := beego.AppConfig.String("samlOrigin")
	sp := &saml2.SAMLServiceProvider{
		IdentityProviderSSOURL:      provider.Endpoint,
		IdentityProviderIssuer:      provider.IssuerUrl,
		ServiceProviderIssuer:       samlOrigin + "/api/acs",
		AssertionConsumerServiceURL: samlOrigin + "/api/acs",
		SignAuthnRequests:           false,
		IDPCertificateStore:         &certStore,
		SPKeyStore:                  randomKeyStore,
	}
	authURL, err := sp.BuildAuthURL("")
	if err != nil {
		panic(err)
	}
	return authURL
}
