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
	"testing"

	"github.com/casdoor/casdoor/object"
)

func TestGetToolsForScopes(t *testing.T) {
	tests := []struct {
		name            string
		grantedScopes   []string
		expectedTools   []string
		unexpectedTools []string
	}{
		{
			name:            "application:read scope",
			grantedScopes:   []string{"application:read"},
			expectedTools:   []string{"get_applications", "get_application"},
			unexpectedTools: []string{"add_application", "update_application", "delete_application"},
		},
		{
			name:            "application:write scope",
			grantedScopes:   []string{"application:write"},
			expectedTools:   []string{"add_application", "update_application", "delete_application"},
			unexpectedTools: []string{"get_applications", "get_application"},
		},
		{
			name:            "both application scopes",
			grantedScopes:   []string{"application:read", "application:write"},
			expectedTools:   []string{"get_applications", "get_application", "add_application", "update_application", "delete_application"},
			unexpectedTools: []string{},
		},
		{
			name:            "read convenience scope",
			grantedScopes:   []string{"read"},
			expectedTools:   []string{"get_applications", "get_application"},
			unexpectedTools: []string{"add_application", "update_application", "delete_application"},
		},
		{
			name:            "write convenience scope",
			grantedScopes:   []string{"write"},
			expectedTools:   []string{"add_application", "update_application", "delete_application"},
			unexpectedTools: []string{},
		},
		{
			name:            "admin convenience scope",
			grantedScopes:   []string{"admin"},
			expectedTools:   []string{"get_applications", "get_application", "add_application", "update_application", "delete_application"},
			unexpectedTools: []string{},
		},
		{
			name:            "no scopes",
			grantedScopes:   []string{},
			expectedTools:   []string{},
			unexpectedTools: []string{"get_applications", "add_application"},
		},
		{
			name:            "unknown scope",
			grantedScopes:   []string{"unknown:scope"},
			expectedTools:   []string{},
			unexpectedTools: []string{"get_applications", "add_application"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allowedTools := GetToolsForScopes(tt.grantedScopes, BuiltinScopes)

			// Check expected tools are present
			for _, tool := range tt.expectedTools {
				if !allowedTools[tool] {
					t.Errorf("Expected tool %s to be allowed, but it was not", tool)
				}
			}

			// Check unexpected tools are not present
			for _, tool := range tt.unexpectedTools {
				if allowedTools[tool] {
					t.Errorf("Expected tool %s to be disallowed, but it was allowed", tool)
				}
			}
		})
	}
}

func TestBuiltinScopesCompleteness(t *testing.T) {
	// Verify that all tools are covered by at least one scope
	allTools := map[string]bool{
		"get_applications":   false,
		"get_application":    false,
		"add_application":    false,
		"update_application": false,
		"delete_application": false,
	}

	for _, scopeItem := range BuiltinScopes {
		for _, tool := range scopeItem.Tools {
			if _, exists := allTools[tool]; exists {
				allTools[tool] = true
			}
		}
	}

	// Check if any tool is not covered
	for tool, covered := range allTools {
		if !covered {
			t.Errorf("Tool %s is not covered by any scope", tool)
		}
	}
}

func TestGetScopesFromClaims(t *testing.T) {
	tests := []struct {
		name          string
		claims        *object.Claims
		expectedCount int
		expectedScope []string
	}{
		{
			name: "single scope",
			claims: &object.Claims{
				Scope: "application:read",
			},
			expectedCount: 1,
			expectedScope: []string{"application:read"},
		},
		{
			name: "multiple scopes",
			claims: &object.Claims{
				Scope: "application:read application:write user:read",
			},
			expectedCount: 3,
			expectedScope: []string{"application:read", "application:write", "user:read"},
		},
		{
			name: "empty scope",
			claims: &object.Claims{
				Scope: "",
			},
			expectedCount: 0,
			expectedScope: []string{},
		},
		{
			name:          "nil claims",
			claims:        nil,
			expectedCount: 0,
			expectedScope: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scopes := GetScopesFromClaims(tt.claims)

			if len(scopes) != tt.expectedCount {
				t.Errorf("Expected %d scopes, got %d", tt.expectedCount, len(scopes))
			}

			for i, expectedScope := range tt.expectedScope {
				if i >= len(scopes) || scopes[i] != expectedScope {
					t.Errorf("Expected scope %s at index %d, got %s", expectedScope, i, scopes[i])
				}
			}
		})
	}
}

func TestGetRequiredScopeForTool(t *testing.T) {
	tests := []struct {
		name          string
		toolName      string
		expectedScope string
	}{
		{
			name:          "get_applications tool",
			toolName:      "get_applications",
			expectedScope: "application:read",
		},
		{
			name:          "add_application tool",
			toolName:      "add_application",
			expectedScope: "application:write",
		},
		{
			name:          "update_application tool",
			toolName:      "update_application",
			expectedScope: "application:write",
		},
		{
			name:          "unknown tool",
			toolName:      "unknown_tool",
			expectedScope: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scope := GetRequiredScopeForTool(tt.toolName, BuiltinScopes)

			if scope != tt.expectedScope {
				t.Errorf("Expected scope %s for tool %s, got %s", tt.expectedScope, tt.toolName, scope)
			}
		})
	}
}
