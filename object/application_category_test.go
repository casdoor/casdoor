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
	"testing"
)

func TestApplicationCategoryAndType(t *testing.T) {
	// Test default application
	defaultApp := &Application{
		Owner:        "admin",
		Name:         "test-default-app",
		Organization: "test-org",
		Category:     "Default",
		Type:         "",
	}

	if defaultApp.Category != "Default" {
		t.Errorf("Expected Category to be 'Default', got '%s'", defaultApp.Category)
	}

	// Test agent application
	agentApp := &Application{
		Owner:        "admin",
		Name:         "test-agent-app",
		Organization: "test-org",
		Category:     "Agent",
		Type:         "MCP",
		Scopes: []*ScopeItem{
			{Name: "files:read", DisplayName: "Read Files", Description: "Allow reading files"},
			{Name: "calendar:manage", DisplayName: "Manage Calendar", Description: "Allow managing calendar events"},
		},
	}

	if agentApp.Category != "Agent" {
		t.Errorf("Expected Category to be 'Agent', got '%s'", agentApp.Category)
	}

	if agentApp.Type != "MCP" {
		t.Errorf("Expected Type to be 'MCP', got '%s'", agentApp.Type)
	}

	if len(agentApp.Scopes) != 2 {
		t.Errorf("Expected 2 scopes, got %d", len(agentApp.Scopes))
	}

	if agentApp.Scopes[0].Name != "files:read" {
		t.Errorf("Expected first scope name to be 'files:read', got '%s'", agentApp.Scopes[0].Name)
	}
}

func TestScopeItem(t *testing.T) {
	scope := &ScopeItem{
		Name:        "database:query",
		DisplayName: "Query Database",
		Description: "Allow executing read-only database queries",
	}

	if scope.Name != "database:query" {
		t.Errorf("Expected scope name to be 'database:query', got '%s'", scope.Name)
	}

	if scope.DisplayName != "Query Database" {
		t.Errorf("Expected scope display name to be 'Query Database', got '%s'", scope.DisplayName)
	}

	if scope.Description != "Allow executing read-only database queries" {
		t.Errorf("Expected scope description, got '%s'", scope.Description)
	}
}
