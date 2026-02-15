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

package mcp

import (
	"github.com/casdoor/casdoor/object"
)

// BuiltinScopes defines the default scope-to-tool mappings for Casdoor's MCP server
var BuiltinScopes = []*object.ScopeItem{
	{
		Name:        "application:read",
		DisplayName: "Read Applications",
		Description: "View application list and details",
		Tools:       []string{"get_applications", "get_application"},
	},
	{
		Name:        "application:write",
		DisplayName: "Manage Applications",
		Description: "Create, update, and delete applications",
		Tools:       []string{"add_application", "update_application", "delete_application"},
	},
	{
		Name:        "user:read",
		DisplayName: "Read Users",
		Description: "View user list and details",
		Tools:       []string{"get_users", "get_user"},
	},
	{
		Name:        "user:write",
		DisplayName: "Manage Users",
		Description: "Create, update, and delete users",
		Tools:       []string{"add_user", "update_user", "delete_user"},
	},
	{
		Name:        "organization:read",
		DisplayName: "Read Organizations",
		Description: "View organization list and details",
		Tools:       []string{"get_organizations", "get_organization"},
	},
	{
		Name:        "organization:write",
		DisplayName: "Manage Organizations",
		Description: "Create, update, and delete organizations",
		Tools:       []string{"add_organization", "update_organization", "delete_organization"},
	},
	{
		Name:        "permission:read",
		DisplayName: "Read Permissions",
		Description: "View permission list and details",
		Tools:       []string{"get_permissions", "get_permission"},
	},
	{
		Name:        "permission:write",
		DisplayName: "Manage Permissions",
		Description: "Create, update, and delete permissions",
		Tools:       []string{"add_permission", "update_permission", "delete_permission"},
	},
	{
		Name:        "role:read",
		DisplayName: "Read Roles",
		Description: "View role list and details",
		Tools:       []string{"get_roles", "get_role"},
	},
	{
		Name:        "role:write",
		DisplayName: "Manage Roles",
		Description: "Create, update, and delete roles",
		Tools:       []string{"add_role", "update_role", "delete_role"},
	},
	{
		Name:        "provider:read",
		DisplayName: "Read Providers",
		Description: "View provider list and details",
		Tools:       []string{"get_providers", "get_provider"},
	},
	{
		Name:        "provider:write",
		DisplayName: "Manage Providers",
		Description: "Create, update, and delete providers",
		Tools:       []string{"add_provider", "update_provider", "delete_provider"},
	},
	{
		Name:        "token:read",
		DisplayName: "Read Tokens",
		Description: "View token list and details",
		Tools:       []string{"get_tokens", "get_token"},
	},
	{
		Name:        "token:write",
		DisplayName: "Manage Tokens",
		Description: "Delete tokens",
		Tools:       []string{"delete_token"},
	},
}

// ConvenienceScopes defines alias scopes that expand to multiple resource scopes
var ConvenienceScopes = map[string][]string{
	"read":  {"application:read", "user:read", "organization:read", "permission:read", "role:read", "provider:read", "token:read"},
	"write": {"application:write", "user:write", "organization:write", "permission:write", "role:write", "provider:write", "token:write"},
	"admin": {"application:read", "application:write", "user:read", "user:write", "organization:read", "organization:write", "permission:read", "permission:write", "role:read", "role:write", "provider:read", "provider:write", "token:read", "token:write"},
}

// GetToolsForScopes returns a map of tools allowed by the given scopes
// The grantedScopes are the scopes present in the token
// The registry contains the scope-to-tool mappings (either BuiltinScopes or Application.Scopes)
func GetToolsForScopes(grantedScopes []string, registry []*object.ScopeItem) map[string]bool {
	allowed := make(map[string]bool)

	// Expand convenience scopes first
	expandedScopes := make([]string, 0)
	for _, scopeName := range grantedScopes {
		if expansion, isConvenience := ConvenienceScopes[scopeName]; isConvenience {
			expandedScopes = append(expandedScopes, expansion...)
		} else {
			expandedScopes = append(expandedScopes, scopeName)
		}
	}

	// Map scopes to tools
	for _, scopeName := range expandedScopes {
		for _, item := range registry {
			if item.Name == scopeName {
				for _, tool := range item.Tools {
					allowed[tool] = true
				}
				break
			}
		}
	}

	return allowed
}

// GetRequiredScopeForTool returns the first scope that provides access to the given tool
// Returns an empty string if no scope is found for the tool
func GetRequiredScopeForTool(toolName string, registry []*object.ScopeItem) string {
	for _, scopeItem := range registry {
		for _, tool := range scopeItem.Tools {
			if tool == toolName {
				return scopeItem.Name
			}
		}
	}
	return ""
}
