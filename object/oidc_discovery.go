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
	"encoding/pem"
	"fmt"

	"github.com/astaxie/beego"
	"gopkg.in/square/go-jose.v2"
)

type OidcDiscovery struct {
	Issuer                                 string   `json:"issuer"`
	AuthorizationEndpoint                  string   `json:"authorization_endpoint"`
	TokenEndpoint                          string   `json:"token_endpoint"`
	UserinfoEndpoint                       string   `json:"userinfo_endpoint"`
	JwksUri                                string   `json:"jwks_uri"`
	ResponseTypesSupported                 []string `json:"response_types_supported"`
	ResponseModesSupported                 []string `json:"response_modes_supported"`
	GrantTypesSupported                    []string `json:"grant_types_supported"`
	SubjectTypesSupported                  []string `json:"subject_types_supported"`
	IdTokenSigningAlgValuesSupported       []string `json:"id_token_signing_alg_values_supported"`
	ScopesSupported                        []string `json:"scopes_supported"`
	ClaimsSupported                        []string `json:"claims_supported"`
	RequestParameterSupported              bool     `json:"request_parameter_supported"`
	RequestObjectSigningAlgValuesSupported []string `json:"request_object_signing_alg_values_supported"`
}

var oidcDiscovery OidcDiscovery

func init() {
	origin := beego.AppConfig.String("origin")

	// Examples:
	// https://login.okta.com/.well-known/openid-configuration
	// https://auth0.auth0.com/.well-known/openid-configuration
	// https://accounts.google.com/.well-known/openid-configuration
	// https://access.line.me/.well-known/openid-configuration
	oidcDiscovery = OidcDiscovery{
		Issuer:                                 origin,
		AuthorizationEndpoint:                  fmt.Sprintf("%s/login/oauth/authorize", origin),
		TokenEndpoint:                          fmt.Sprintf("%s/api/login/oauth/access_token", origin),
		UserinfoEndpoint:                       fmt.Sprintf("%s/api/get-account", origin),
		JwksUri:                                fmt.Sprintf("%s/api/certs", origin),
		ResponseTypesSupported:                 []string{"id_token"},
		ResponseModesSupported:                 []string{"login", "code", "link"},
		GrantTypesSupported:                    []string{"password", "authorization_code"},
		SubjectTypesSupported:                  []string{"public"},
		IdTokenSigningAlgValuesSupported:       []string{"RS256"},
		ScopesSupported:                        []string{"openid", "email", "profile", "address", "phone", "offline_access"},
		ClaimsSupported:                        []string{"iss", "ver", "sub", "aud", "iat", "exp", "id", "type", "displayName", "avatar", "permanentAvatar", "email", "phone", "location", "affiliation", "title", "homepage", "bio", "tag", "region", "language", "score", "ranking", "isOnline", "isAdmin", "isGlobalAdmin", "isForbidden", "signupApplication", "ldap"},
		RequestParameterSupported:              true,
		RequestObjectSigningAlgValuesSupported: []string{"HS256", "HS384", "HS512"},
	}
}

func GetOidcDiscovery() OidcDiscovery {
	return oidcDiscovery
}

func GetJsonWebKeySet() (jose.JSONWebKeySet, error) {
	cert := GetDefaultCert()

	//follows the protocol rfc 7517(draft)
	//link here: https://self-issued.info/docs/draft-ietf-jose-json-web-key.html
	//or https://datatracker.ietf.org/doc/html/draft-ietf-jose-json-web-key
	certPemBlock := []byte(cert.PublicKey)
	certDerBlock, _ := pem.Decode(certPemBlock)
	x509Cert, _ := x509.ParseCertificate(certDerBlock.Bytes)

	var jwk jose.JSONWebKey
	jwk.Key = x509Cert.PublicKey
	jwk.Certificates = []*x509.Certificate{x509Cert}

	var jwks jose.JSONWebKeySet
	jwks.Keys = []jose.JSONWebKey{jwk}
	return jwks, nil
}
