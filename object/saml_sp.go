// Copyright 2021 The Casdoor Authors. All Rights Reserved.
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
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/casdoor/casdoor/idp"
	"github.com/mitchellh/mapstructure"

	"github.com/casdoor/casdoor/i18n"
	saml2 "github.com/russellhaering/gosaml2"
	dsig "github.com/russellhaering/goxmldsig"
)

func ParseSamlResponse(samlResponse string, provider *Provider, host string) (*idp.UserInfo, error) {
	samlResponse, _ = url.QueryUnescape(samlResponse)
	sp, err := buildSp(provider, samlResponse, host)
	if err != nil {
		return nil, err
	}

	assertionInfo, err := sp.RetrieveAssertionInfo(samlResponse)
	if err != nil {
		return nil, err
	}

	userInfoMap := make(map[string]string)
	for spAttr, idpAttr := range provider.UserMapping {
		for _, attr := range assertionInfo.Values {
			if attr.Name == idpAttr {
				userInfoMap[spAttr] = attr.Values[0].Value
			}
		}
	}
	userInfoMap["id"] = assertionInfo.NameID

	customUserInfo := &idp.CustomUserInfo{}
	err = mapstructure.Decode(userInfoMap, customUserInfo)
	if err != nil {
		return nil, err
	}
	userInfo := &idp.UserInfo{
		Id:          customUserInfo.Id,
		Username:    customUserInfo.Username,
		DisplayName: customUserInfo.DisplayName,
		Email:       customUserInfo.Email,
		AvatarUrl:   customUserInfo.AvatarUrl,
	}
	return userInfo, err
}

func GenerateSamlRequest(id, relayState, host, lang string) (auth string, method string, err error) {
	provider, err := GetProvider(id)
	if err != nil {
		return "", "", err
	}
	if provider.Category != "SAML" {
		return "", "", fmt.Errorf(i18n.Translate(lang, "saml_sp:provider %s's category is not SAML"), provider.Name)
	}

	sp, err := buildSp(provider, "", host)
	if err != nil {
		return "", "", err
	}

	if provider.EnableSignAuthnRequest {
		post, err := sp.BuildAuthBodyPost(relayState)
		if err != nil {
			return "", "", err
		}
		auth = string(post[:])
		method = "POST"
	} else {
		auth, err = sp.BuildAuthURL(relayState)
		if err != nil {
			return "", "", err
		}
		method = "GET"
	}
	return auth, method, nil
}

func buildSp(provider *Provider, samlResponse string, host string) (*saml2.SAMLServiceProvider, error) {
	_, origin := getOriginFromHost(host)

	certStore, err := buildSpCertificateStore(provider, samlResponse)
	if err != nil {
		return nil, err
	}

	sp := &saml2.SAMLServiceProvider{
		ServiceProviderIssuer:       fmt.Sprintf("%s/api/acs", origin),
		AssertionConsumerServiceURL: fmt.Sprintf("%s/api/acs", origin),
		SignAuthnRequests:           false,
		IDPCertificateStore:         &certStore,
		SPKeyStore:                  dsig.RandomKeyStoreForTest(),
	}

	if provider.Endpoint != "" {
		sp.IdentityProviderSSOURL = provider.Endpoint
		sp.IdentityProviderIssuer = provider.IssuerUrl
	}
	if provider.EnableSignAuthnRequest {
		sp.SignAuthnRequests = true
		sp.SPKeyStore, err = buildSpKeyStore()
		if err != nil {
			return nil, err
		}
	}

	return sp, nil
}

func buildSpKeyStore() (dsig.X509KeyStore, error) {
	keyPair, err := tls.LoadX509KeyPair("object/token_jwt_key.pem", "object/token_jwt_key.key")
	if err != nil {
		return nil, err
	}
	return &dsig.TLSCertKeyStore{
		PrivateKey:  keyPair.PrivateKey,
		Certificate: keyPair.Certificate,
	}, nil
}

func buildSpCertificateStore(provider *Provider, samlResponse string) (certStore dsig.MemoryX509CertificateStore, err error) {
	certEncodedData := ""
	if samlResponse != "" {
		certEncodedData, err = getCertificateFromSamlResponse(samlResponse, provider.Type)
		if err != nil {
			return
		}
	} else if provider.IdP != "" {
		certEncodedData = provider.IdP
	}

	var certData []byte
	block, _ := pem.Decode([]byte(certEncodedData))
	if block != nil {
		// this was a PEM file
		// block.Bytes are DER encoded so the following code block should happily accept it
		certData = block.Bytes
	} else {
		certData, err = base64.StdEncoding.DecodeString(certEncodedData)
		if err != nil {
			return dsig.MemoryX509CertificateStore{}, err
		}
	}

	idpCert, err := x509.ParseCertificate(certData)
	if err != nil {
		return dsig.MemoryX509CertificateStore{}, err
	}

	certStore = dsig.MemoryX509CertificateStore{
		Roots: []*x509.Certificate{idpCert},
	}
	return certStore, nil
}

func getCertificateFromSamlResponse(samlResponse string, providerType string) (string, error) {
	de, err := base64.StdEncoding.DecodeString(samlResponse)
	if err != nil {
		return "", err
	}
	var (
		expression string
		deStr      = strings.Replace(string(de), "\n", "", -1)
		tagMap     = map[string]string{
			"Aliyun IDaaS": "ds",
			"Keycloak":     "dsig",
		}
	)
	tag := tagMap[providerType]
	if tag == "" {
		// <ds:X509Certificate>...</ds:X509Certificate>
		// <dsig:X509Certificate>...</dsig:X509Certificate>
		// <X509Certificate>...</X509Certificate>
		// ...
		expression = "<[^>]*:?X509Certificate>([\\s\\S]*?)<[^>]*:?X509Certificate>"
	} else {
		expression = fmt.Sprintf("<%s:X509Certificate>([\\s\\S]*?)</%s:X509Certificate>", tag, tag)
	}
	res := regexp.MustCompile(expression).FindStringSubmatch(deStr)
	return res[1], nil
}
