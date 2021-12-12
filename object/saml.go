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
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/astaxie/beego"
	saml2 "github.com/russellhaering/gosaml2"
	dsig "github.com/russellhaering/goxmldsig"
)

func ParseSamlResponse(samlResponse string) (string, error) {
	samlResponse, _ = url.QueryUnescape(samlResponse)
	sp, err := buildSp(nil, samlResponse)
	if err != nil {
		return "", err
	}
	assertionInfo, err := sp.RetrieveAssertionInfo(samlResponse)
	if err != nil {
		panic(err)
	}
	return assertionInfo.NameID, nil
}

func GenerateSamlLoginUrl(id string) (string, error) {
	provider := GetProvider(id)
	if provider.Category != "SAML" {
		return "", fmt.Errorf("Provider %s's category is not SAML", provider.Name)
	}
	sp, err := buildSp(provider, "")
	if err != nil {
		return "", err
	}
	authURL, err := sp.BuildAuthURL("")
	if err != nil {
		return "", err
	}
	return authURL, nil
}

func buildSp(provider *Provider, samlResponse string) (*saml2.SAMLServiceProvider, error) {
	certStore := dsig.MemoryX509CertificateStore{
		Roots: []*x509.Certificate{},
	}
	origin := beego.AppConfig.String("origin")
	certEncodedData := ""
	if samlResponse != "" {
		de, err := base64.StdEncoding.DecodeString(samlResponse)
		if err != nil {
			panic(err)
		}
		deStr := strings.Replace(string(de), "\n", "", -1)
		res := regexp.MustCompile(`<ds:X509Certificate>(.*?)</ds:X509Certificate>`).FindAllStringSubmatch(deStr, -1)
		str := res[0][0]
		certEncodedData = str[20 : len(str)-21]
	} else if provider != nil {
		certEncodedData = provider.IdP
	}
	certData, err := base64.StdEncoding.DecodeString(certEncodedData)
	if err != nil {
		return nil, err
	}
	idpCert, err := x509.ParseCertificate(certData)
	if err != nil {
		return nil, err
	}
	certStore.Roots = append(certStore.Roots, idpCert)
	sp := &saml2.SAMLServiceProvider{
		ServiceProviderIssuer:       fmt.Sprintf("%s/api/acs", origin),
		AssertionConsumerServiceURL: fmt.Sprintf("%s/api/acs", origin),
		IDPCertificateStore:         &certStore,
	}
	if provider != nil {
		randomKeyStore := dsig.RandomKeyStoreForTest()
		sp.IdentityProviderSSOURL = provider.Endpoint
		sp.IdentityProviderIssuer = provider.IssuerUrl
		sp.SignAuthnRequests = false
		sp.SPKeyStore = randomKeyStore
	}
	return sp, nil
}
