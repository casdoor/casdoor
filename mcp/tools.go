// Copyright 2024 The Casdoor Authors. All Rights Reserved.
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
	"encoding/json"
	"fmt"

	"github.com/casdoor/casdoor/object"
)

// ToolHandler handles MCP tool invocations
type ToolHandler struct{}

// CallTool executes a tool and returns the result
func (h *ToolHandler) CallTool(toolName string, params map[string]interface{}) (*CallToolResult, error) {
	switch toolName {
	case "get_user":
		return h.getUser(params)
	case "list_users":
		return h.listUsers(params)
	case "get_organization":
		return h.getOrganization(params)
	case "list_organizations":
		return h.listOrganizations(params)
	case "get_application":
		return h.getApplication(params)
	case "list_applications":
		return h.listApplications(params)
	case "get_role":
		return h.getRole(params)
	case "list_roles":
		return h.listRoles(params)
	default:
		return nil, fmt.Errorf("unknown tool: %s", toolName)
	}
}

// getUser retrieves a user by ID
func (h *ToolHandler) getUser(params map[string]interface{}) (*CallToolResult, error) {
	userId, ok := params["userId"].(string)
	if !ok {
		return &CallToolResult{
			Content: []ContentItem{{Type: "text", Text: "Missing or invalid userId parameter"}},
			IsError: true,
		}, nil
	}

	user, err := object.GetUser(userId)
	if err != nil {
		return &CallToolResult{
			Content: []ContentItem{{Type: "text", Text: fmt.Sprintf("Error retrieving user: %v", err)}},
			IsError: true,
		}, nil
	}

	if user == nil {
		return &CallToolResult{
			Content: []ContentItem{{Type: "text", Text: fmt.Sprintf("User not found: %s", userId)}},
			IsError: true,
		}, nil
	}

	jsonData, err := json.MarshalIndent(user, "", "  ")
	if err != nil {
		return &CallToolResult{
			Content: []ContentItem{{Type: "text", Text: fmt.Sprintf("Error marshaling user data: %v", err)}},
			IsError: true,
		}, nil
	}

	return &CallToolResult{
		Content: []ContentItem{{Type: "text", Text: string(jsonData)}},
	}, nil
}

// listUsers lists all users
func (h *ToolHandler) listUsers(params map[string]interface{}) (*CallToolResult, error) {
	owner := ""
	if ownerParam, ok := params["owner"].(string); ok {
		owner = ownerParam
	}

	users, err := object.GetUsers(owner)
	if err != nil {
		return &CallToolResult{
			Content: []ContentItem{{Type: "text", Text: fmt.Sprintf("Error retrieving users: %v", err)}},
			IsError: true,
		}, nil
	}

	jsonData, err := json.MarshalIndent(users, "", "  ")
	if err != nil {
		return &CallToolResult{
			Content: []ContentItem{{Type: "text", Text: fmt.Sprintf("Error marshaling users data: %v", err)}},
			IsError: true,
		}, nil
	}

	return &CallToolResult{
		Content: []ContentItem{{Type: "text", Text: string(jsonData)}},
	}, nil
}

// getOrganization retrieves an organization by ID
func (h *ToolHandler) getOrganization(params map[string]interface{}) (*CallToolResult, error) {
	orgId, ok := params["organizationId"].(string)
	if !ok {
		return &CallToolResult{
			Content: []ContentItem{{Type: "text", Text: "Missing or invalid organizationId parameter"}},
			IsError: true,
		}, nil
	}

	org, err := object.GetOrganization(orgId)
	if err != nil {
		return &CallToolResult{
			Content: []ContentItem{{Type: "text", Text: fmt.Sprintf("Error retrieving organization: %v", err)}},
			IsError: true,
		}, nil
	}

	if org == nil {
		return &CallToolResult{
			Content: []ContentItem{{Type: "text", Text: fmt.Sprintf("Organization not found: %s", orgId)}},
			IsError: true,
		}, nil
	}

	jsonData, err := json.MarshalIndent(org, "", "  ")
	if err != nil {
		return &CallToolResult{
			Content: []ContentItem{{Type: "text", Text: fmt.Sprintf("Error marshaling organization data: %v", err)}},
			IsError: true,
		}, nil
	}

	return &CallToolResult{
		Content: []ContentItem{{Type: "text", Text: string(jsonData)}},
	}, nil
}

// listOrganizations lists all organizations
func (h *ToolHandler) listOrganizations(params map[string]interface{}) (*CallToolResult, error) {
	owner := ""
	if ownerParam, ok := params["owner"].(string); ok {
		owner = ownerParam
	}

	orgs, err := object.GetOrganizations(owner)
	if err != nil {
		return &CallToolResult{
			Content: []ContentItem{{Type: "text", Text: fmt.Sprintf("Error retrieving organizations: %v", err)}},
			IsError: true,
		}, nil
	}

	jsonData, err := json.MarshalIndent(orgs, "", "  ")
	if err != nil {
		return &CallToolResult{
			Content: []ContentItem{{Type: "text", Text: fmt.Sprintf("Error marshaling organizations data: %v", err)}},
			IsError: true,
		}, nil
	}

	return &CallToolResult{
		Content: []ContentItem{{Type: "text", Text: string(jsonData)}},
	}, nil
}

// getApplication retrieves an application by ID
func (h *ToolHandler) getApplication(params map[string]interface{}) (*CallToolResult, error) {
	appId, ok := params["applicationId"].(string)
	if !ok {
		return &CallToolResult{
			Content: []ContentItem{{Type: "text", Text: "Missing or invalid applicationId parameter"}},
			IsError: true,
		}, nil
	}

	app, err := object.GetApplication(appId)
	if err != nil {
		return &CallToolResult{
			Content: []ContentItem{{Type: "text", Text: fmt.Sprintf("Error retrieving application: %v", err)}},
			IsError: true,
		}, nil
	}

	if app == nil {
		return &CallToolResult{
			Content: []ContentItem{{Type: "text", Text: fmt.Sprintf("Application not found: %s", appId)}},
			IsError: true,
		}, nil
	}

	jsonData, err := json.MarshalIndent(app, "", "  ")
	if err != nil {
		return &CallToolResult{
			Content: []ContentItem{{Type: "text", Text: fmt.Sprintf("Error marshaling application data: %v", err)}},
			IsError: true,
		}, nil
	}

	return &CallToolResult{
		Content: []ContentItem{{Type: "text", Text: string(jsonData)}},
	}, nil
}

// listApplications lists all applications
func (h *ToolHandler) listApplications(params map[string]interface{}) (*CallToolResult, error) {
	owner := ""
	if ownerParam, ok := params["owner"].(string); ok {
		owner = ownerParam
	}

	apps, err := object.GetApplications(owner)
	if err != nil {
		return &CallToolResult{
			Content: []ContentItem{{Type: "text", Text: fmt.Sprintf("Error retrieving applications: %v", err)}},
			IsError: true,
		}, nil
	}

	jsonData, err := json.MarshalIndent(apps, "", "  ")
	if err != nil {
		return &CallToolResult{
			Content: []ContentItem{{Type: "text", Text: fmt.Sprintf("Error marshaling applications data: %v", err)}},
			IsError: true,
		}, nil
	}

	return &CallToolResult{
		Content: []ContentItem{{Type: "text", Text: string(jsonData)}},
	}, nil
}

// getRole retrieves a role by ID
func (h *ToolHandler) getRole(params map[string]interface{}) (*CallToolResult, error) {
	roleId, ok := params["roleId"].(string)
	if !ok {
		return &CallToolResult{
			Content: []ContentItem{{Type: "text", Text: "Missing or invalid roleId parameter"}},
			IsError: true,
		}, nil
	}

	role, err := object.GetRole(roleId)
	if err != nil {
		return &CallToolResult{
			Content: []ContentItem{{Type: "text", Text: fmt.Sprintf("Error retrieving role: %v", err)}},
			IsError: true,
		}, nil
	}

	if role == nil {
		return &CallToolResult{
			Content: []ContentItem{{Type: "text", Text: fmt.Sprintf("Role not found: %s", roleId)}},
			IsError: true,
		}, nil
	}

	jsonData, err := json.MarshalIndent(role, "", "  ")
	if err != nil {
		return &CallToolResult{
			Content: []ContentItem{{Type: "text", Text: fmt.Sprintf("Error marshaling role data: %v", err)}},
			IsError: true,
		}, nil
	}

	return &CallToolResult{
		Content: []ContentItem{{Type: "text", Text: string(jsonData)}},
	}, nil
}

// listRoles lists all roles
func (h *ToolHandler) listRoles(params map[string]interface{}) (*CallToolResult, error) {
	owner := ""
	if ownerParam, ok := params["owner"].(string); ok {
		owner = ownerParam
	}

	roles, err := object.GetRoles(owner)
	if err != nil {
		return &CallToolResult{
			Content: []ContentItem{{Type: "text", Text: fmt.Sprintf("Error retrieving roles: %v", err)}},
			IsError: true,
		}, nil
	}

	jsonData, err := json.MarshalIndent(roles, "", "  ")
	if err != nil {
		return &CallToolResult{
			Content: []ContentItem{{Type: "text", Text: fmt.Sprintf("Error marshaling roles data: %v", err)}},
			IsError: true,
		}, nil
	}

	return &CallToolResult{
		Content: []ContentItem{{Type: "text", Text: string(jsonData)}},
	}, nil
}
