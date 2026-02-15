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
	"fmt"
)

// OauthProtectedResourceMetadata represents RFC 9728 OAuth 2.0 Protected Resource Metadata
type OauthProtectedResourceMetadata struct {
	Resource               string   `json:"resource"`
	AuthorizationServers   []string `json:"authorization_servers"`
	BearerMethodsSupported []string `json:"bearer_methods_supported,omitempty"`
	ScopesSupported        []string `json:"scopes_supported,omitempty"`
	ResourceSigningAlg     []string `json:"resource_signing_alg_values_supported,omitempty"`
	ResourceDocumentation  string   `json:"resource_documentation,omitempty"`
}

// GetOauthProtectedResourceMetadata returns RFC 9728 Protected Resource Metadata for global discovery
func GetOauthProtectedResourceMetadata(host string) OauthProtectedResourceMetadata {
	_, originBackend := getOriginFromHost(host)

	return OauthProtectedResourceMetadata{
		Resource:               originBackend,
		AuthorizationServers:   []string{originBackend},
		BearerMethodsSupported: []string{"header"},
		ScopesSupported:        []string{"openid", "profile", "email", "read", "write"},
		ResourceSigningAlg:     []string{"RS256"},
	}
}

// GetOauthProtectedResourceMetadataByApplication returns RFC 9728 Protected Resource Metadata for application-specific discovery
func GetOauthProtectedResourceMetadataByApplication(host string, applicationName string) OauthProtectedResourceMetadata {
	_, originBackend := getOriginFromHost(host)

	// For application-specific discovery, the resource identifier includes the application name
	resourceIdentifier := fmt.Sprintf("%s/.well-known/%s", originBackend, applicationName)
	authServer := fmt.Sprintf("%s/.well-known/%s", originBackend, applicationName)

	return OauthProtectedResourceMetadata{
		Resource:               resourceIdentifier,
		AuthorizationServers:   []string{authServer},
		BearerMethodsSupported: []string{"header"},
		ScopesSupported:        []string{"openid", "profile", "email", "read", "write"},
		ResourceSigningAlg:     []string{"RS256"},
	}
}
