// Copyright 2026 The Casdoor Authors. All Rights Reserved.
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
	"net/url"
	"strings"

	"github.com/casdoor/casdoor/util"
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

// McpScopesSupported are the scopes that Casdoor's MCP server advertises, "read" and "write" grant all its tools
var McpScopesSupported = []string{"openid", "profile", "email", "read", "write"}

// GetOauthProtectedResourceMetadata returns RFC 9728 Protected Resource Metadata for global discovery
func GetOauthProtectedResourceMetadata(host string) OauthProtectedResourceMetadata {
	_, originBackend := getOriginFromHost(host)

	return OauthProtectedResourceMetadata{
		Resource:               originBackend,
		AuthorizationServers:   []string{originBackend},
		BearerMethodsSupported: []string{"header"},
		ScopesSupported:        McpScopesSupported,
		ResourceSigningAlg:     []string{"RS256"},
	}
}

// GetOauthProtectedResourceMetadataByOrganization returns RFC 9728 Protected Resource Metadata for the built-in MCP server
// of an organization, "/api/mcp/orgs/:organization". Its authorization server is the issuer of the organization, so that
// MCP clients following the standard discovery register and sign in there instead of the default organization. It
// returns nil when the organization doesn't exist.
func GetOauthProtectedResourceMetadataByOrganization(host string, organizationName string) (*OauthProtectedResourceMetadata, error) {
	organization, err := getOrganization("admin", organizationName)
	if err != nil || organization == nil {
		return nil, err
	}

	return &OauthProtectedResourceMetadata{
		Resource:               GetBuiltinMcpUrl(host, organization.Name),
		AuthorizationServers:   []string{getOrganizationIssuerOrGlobal(host, organization)},
		BearerMethodsSupported: []string{"header"},
		ScopesSupported:        McpScopesSupported,
		ResourceSigningAlg:     []string{"RS256"},
	}, nil
}

// GetBuiltinMcpUrl returns the URL of the built-in MCP server, "/api/mcp/orgs/:organization" for an organization
func GetBuiltinMcpUrl(host string, organizationName string) string {
	_, originBackend := getOriginFromHost(host)
	if organizationName == "" {
		return fmt.Sprintf("%s/api/mcp", originBackend)
	}
	return fmt.Sprintf("%s/api/mcp/orgs/%s", originBackend, url.PathEscape(organizationName))
}

// IsBuiltinMcpPath reports whether a request path is the built-in MCP server: "/api/mcp" or "/api/mcp/orgs/:organization"
func IsBuiltinMcpPath(urlPath string) bool {
	if urlPath == "/api/mcp" {
		return true
	}

	organizationName, ok := strings.CutPrefix(urlPath, "/api/mcp/orgs/")
	return ok && organizationName != "" && !strings.Contains(organizationName, "/")
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
		ScopesSupported:        McpScopesSupported,
		ResourceSigningAlg:     []string{"RS256"},
	}
}

// GetOauthProtectedResourceMetadataByServer returns RFC 9728 Protected Resource Metadata for an MCP server behind the proxy
// of Casdoor. Its authorization server is the issuer of the server's organization, so that MCP clients register and sign
// in there. It returns nil when the server doesn't exist.
func GetOauthProtectedResourceMetadataByServer(host string, owner string, name string) (*OauthProtectedResourceMetadata, error) {
	server, err := GetServer(util.GetId(owner, name))
	if err != nil || server == nil {
		return nil, err
	}

	organization, err := getOrganization("admin", owner)
	if err != nil {
		return nil, err
	}

	authorizationServer := GetOidcIssuer(host)
	if organization != nil {
		authorizationServer = getOrganizationIssuerOrGlobal(host, organization)
	}

	return &OauthProtectedResourceMetadata{
		Resource:               GetServerProxyUrl(host, owner, name),
		AuthorizationServers:   []string{authorizationServer},
		BearerMethodsSupported: []string{"header"},
		ScopesSupported:        []string{"openid", "profile", "email", "offline_access"},
		ResourceSigningAlg:     []string{"RS256"},
	}, nil
}

// GetServerProxyUrl returns the URL MCP clients connect to for an MCP server behind the proxy of Casdoor
func GetServerProxyUrl(host string, owner string, name string) string {
	_, originBackend := getOriginFromHost(host)
	return fmt.Sprintf("%s/api/server/%s/%s", originBackend, url.PathEscape(owner), url.PathEscape(name))
}

// IsBuiltinMcpResource reports whether an RFC 8707 resource indicator names the built-in MCP server "/api/mcp": the
// resource of the global, an organization's or an application's protected resource metadata, or the URL of the MCP
// endpoint itself
func IsBuiltinMcpResource(host string, resource string) bool {
	_, originBackend := getOriginFromHost(host)
	originBackend = strings.TrimSuffix(originBackend, "/")
	resource = strings.TrimSuffix(resource, "/")

	if resource == originBackend || resource == originBackend+"/api/mcp" {
		return true
	}

	if organizationName, ok := strings.CutPrefix(resource, originBackend+"/api/mcp/orgs/"); ok {
		return organizationName != "" && !strings.Contains(organizationName, "/")
	}

	applicationName, ok := strings.CutPrefix(resource, originBackend+"/.well-known/")
	return ok && applicationName != "" && !strings.Contains(applicationName, "/")
}
