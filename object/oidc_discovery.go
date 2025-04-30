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
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net"
	"strings"

	"github.com/casdoor/casdoor/conf"
	"gopkg.in/square/go-jose.v2"
)

type OidcDiscovery struct {
	Issuer                                 string   `json:"issuer"`
	AuthorizationEndpoint                  string   `json:"authorization_endpoint"`
	TokenEndpoint                          string   `json:"token_endpoint"`
	UserinfoEndpoint                       string   `json:"userinfo_endpoint"`
	DeviceAuthorizationEndpoint            string   `json:"device_authorization_endpoint"`
	JwksUri                                string   `json:"jwks_uri"`
	IntrospectionEndpoint                  string   `json:"introspection_endpoint"`
	ResponseTypesSupported                 []string `json:"response_types_supported"`
	ResponseModesSupported                 []string `json:"response_modes_supported"`
	GrantTypesSupported                    []string `json:"grant_types_supported"`
	SubjectTypesSupported                  []string `json:"subject_types_supported"`
	IdTokenSigningAlgValuesSupported       []string `json:"id_token_signing_alg_values_supported"`
	ScopesSupported                        []string `json:"scopes_supported"`
	ClaimsSupported                        []string `json:"claims_supported"`
	RequestParameterSupported              bool     `json:"request_parameter_supported"`
	RequestObjectSigningAlgValuesSupported []string `json:"request_object_signing_alg_values_supported"`
	EndSessionEndpoint                     string   `json:"end_session_endpoint"`
}

type WebFinger struct {
	Subject    string             `json:"subject"`
	Links      []WebFingerLink    `json:"links"`
	Aliases    *[]string          `json:"aliases,omitempty"`
	Properties *map[string]string `json:"properties,omitempty"`
}

type WebFingerLink struct {
	Rel  string `json:"rel"`
	Href string `json:"href"`
}

func isIpAddress(host string) bool {
	// Attempt to split the host and port, ignoring the error
	hostWithoutPort, _, err := net.SplitHostPort(host)
	if err != nil {
		// If an error occurs, it might be because there's no port
		// In that case, use the original host string
		hostWithoutPort = host
	}

	// Attempt to parse the host as an IP address (both IPv4 and IPv6)
	ip := net.ParseIP(hostWithoutPort)
	// if host is not nil is an IP address else is not an IP address
	return ip != nil
}

func getOriginFromHostInternal(host string) (string, string) {
	origin := conf.GetConfigString("origin")
	if origin != "" {
		return origin, origin
	}

	isDev := conf.GetConfigString("runmode") == "dev"
	// "door.casdoor.com"
	protocol := "https://"
	if !strings.Contains(host, ".") {
		// "localhost:8000" or "computer-name:80"
		protocol = "http://"
	} else if isIpAddress(host) {
		// "192.168.0.10"
		protocol = "http://"
	}

	if host == "localhost:8000" && isDev {
		return fmt.Sprintf("%s%s", protocol, "localhost:7001"), fmt.Sprintf("%s%s", protocol, "localhost:8000")
	} else {
		return fmt.Sprintf("%s%s", protocol, host), fmt.Sprintf("%s%s", protocol, host)
	}
}

func getOriginFromHost(host string) (string, string) {
	originF, originB := getOriginFromHostInternal(host)

	originFrontend := conf.GetConfigString("originFrontend")
	if originFrontend != "" {
		originF = originFrontend
	}

	return originF, originB
}

func GetOidcDiscovery(host string) OidcDiscovery {
	originFrontend, originBackend := getOriginFromHost(host)

	// Examples:
	// https://login.okta.com/.well-known/openid-configuration
	// https://auth0.auth0.com/.well-known/openid-configuration
	// https://accounts.google.com/.well-known/openid-configuration
	// https://access.line.me/.well-known/openid-configuration
	oidcDiscovery := OidcDiscovery{
		Issuer:                                 originBackend,
		AuthorizationEndpoint:                  fmt.Sprintf("%s/login/oauth/authorize", originFrontend),
		TokenEndpoint:                          fmt.Sprintf("%s/api/login/oauth/access_token", originBackend),
		UserinfoEndpoint:                       fmt.Sprintf("%s/api/userinfo", originBackend),
		DeviceAuthorizationEndpoint:            fmt.Sprintf("%s/api/device-auth", originBackend),
		JwksUri:                                fmt.Sprintf("%s/.well-known/jwks", originBackend),
		IntrospectionEndpoint:                  fmt.Sprintf("%s/api/login/oauth/introspect", originBackend),
		ResponseTypesSupported:                 []string{"code", "token", "id_token", "code token", "code id_token", "token id_token", "code token id_token", "none"},
		ResponseModesSupported:                 []string{"query", "fragment", "login", "code", "link"},
		GrantTypesSupported:                    []string{"password", "authorization_code"},
		SubjectTypesSupported:                  []string{"public"},
		IdTokenSigningAlgValuesSupported:       []string{"RS256", "RS512", "ES256", "ES384", "ES512"},
		ScopesSupported:                        []string{"openid", "email", "profile", "address", "phone", "offline_access"},
		ClaimsSupported:                        []string{"iss", "ver", "sub", "aud", "iat", "exp", "id", "type", "displayName", "avatar", "permanentAvatar", "email", "phone", "location", "affiliation", "title", "homepage", "bio", "tag", "region", "language", "score", "ranking", "isOnline", "isAdmin", "isForbidden", "signupApplication", "ldap"},
		RequestParameterSupported:              true,
		RequestObjectSigningAlgValuesSupported: []string{"HS256", "HS384", "HS512"},
		EndSessionEndpoint:                     fmt.Sprintf("%s/api/logout", originBackend),
	}

	return oidcDiscovery
}

func GetJsonWebKeySet() (jose.JSONWebKeySet, error) {
	jwks := jose.JSONWebKeySet{}
	certs, err := GetCerts("")
	if err != nil {
		return jwks, err
	}

	// follows the protocol rfc 7517(draft)
	// link here: https://self-issued.info/docs/draft-ietf-jose-json-web-key.html
	// or https://datatracker.ietf.org/doc/html/draft-ietf-jose-json-web-key
	for _, cert := range certs {
		if cert.Type != "x509" {
			continue
		}

		if cert.Certificate == "" {
			return jwks, fmt.Errorf("the certificate field should not be empty for the cert: %v", cert)
		}

		certPemBlock := []byte(cert.Certificate)
		certDerBlock, _ := pem.Decode(certPemBlock)
		x509Cert, err := x509.ParseCertificate(certDerBlock.Bytes)
		if err != nil {
			return jwks, err
		}

		var jwk jose.JSONWebKey
		jwk.Key = x509Cert.PublicKey
		jwk.Certificates = []*x509.Certificate{x509Cert}
		jwk.KeyID = cert.Name
		jwk.Algorithm = cert.CryptoAlgorithm
		jwk.Use = "sig"
		jwks.Keys = append(jwks.Keys, jwk)
	}

	return jwks, nil
}

func GetWebFinger(resource string, rels []string, host string) (WebFinger, error) {
	wf := WebFinger{}

	resourceSplit := strings.Split(resource, ":")

	if len(resourceSplit) != 2 {
		return wf, fmt.Errorf("invalid resource")
	}

	resourceType := resourceSplit[0]
	resourceValue := resourceSplit[1]

	oidcDiscovery := GetOidcDiscovery(host)

	switch resourceType {
	case "acct":
		user, err := GetUserByEmailOnly(resourceValue)
		if err != nil {
			return wf, err
		}

		if user == nil {
			return wf, fmt.Errorf("user not found")
		}

		wf.Subject = resource

		for _, rel := range rels {
			if rel == "http://openid.net/specs/connect/1.0/issuer" {
				wf.Links = append(wf.Links, WebFingerLink{
					Rel:  "http://openid.net/specs/connect/1.0/issuer",
					Href: oidcDiscovery.Issuer,
				})
			}
		}
	}

	return wf, nil
}

func GetDeviceAuthResponse(deviceCode string, userCode string, host string) DeviceAuthResponse {
	originFrontend, _ := getOriginFromHost(host)

	return DeviceAuthResponse{
		DeviceCode:      deviceCode,
		UserCode:        userCode,
		VerificationUri: fmt.Sprintf("%s/login/oauth/device/%s", originFrontend, userCode),
		ExpiresIn:       120,
	}
}
