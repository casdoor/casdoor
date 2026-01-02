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

// getOptionalStringParam extracts an optional string parameter from params
func getOptionalStringParam(params map[string]interface{}, key string) string {
	if value, ok := params[key].(string); ok {
		return value
	}
	return ""
}

// marshalToCallResult converts an interface to a CallToolResult with JSON marshaling
func marshalToCallResult(data interface{}) (*CallToolResult, error) {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return &CallToolResult{
			Content: []ContentItem{{Type: "text", Text: fmt.Sprintf("Error marshaling data: %v", err)}},
			IsError: true,
		}, nil
	}

	return &CallToolResult{
		Content: []ContentItem{{Type: "text", Text: string(jsonData)}},
	}, nil
}

// errorResult creates an error CallToolResult
func errorResult(message string) *CallToolResult {
	return &CallToolResult{
		Content: []ContentItem{{Type: "text", Text: message}},
		IsError: true,
	}
}

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
		return errorResult("Missing or invalid userId parameter"), nil
	}

	user, err := object.GetUser(userId)
	if err != nil {
		return errorResult(fmt.Sprintf("Error retrieving user: %v", err)), nil
	}

	if user == nil {
		return errorResult(fmt.Sprintf("User not found: %s", userId)), nil
	}

	return marshalToCallResult(user)
}

// listUsers lists all users
func (h *ToolHandler) listUsers(params map[string]interface{}) (*CallToolResult, error) {
	owner := getOptionalStringParam(params, "owner")

	users, err := object.GetUsers(owner)
	if err != nil {
		return errorResult(fmt.Sprintf("Error retrieving users: %v", err)), nil
	}

	return marshalToCallResult(users)
}

// getOrganization retrieves an organization by ID
func (h *ToolHandler) getOrganization(params map[string]interface{}) (*CallToolResult, error) {
	orgId, ok := params["organizationId"].(string)
	if !ok {
		return errorResult("Missing or invalid organizationId parameter"), nil
	}

	org, err := object.GetOrganization(orgId)
	if err != nil {
		return errorResult(fmt.Sprintf("Error retrieving organization: %v", err)), nil
	}

	if org == nil {
		return errorResult(fmt.Sprintf("Organization not found: %s", orgId)), nil
	}

	return marshalToCallResult(org)
}

// listOrganizations lists all organizations
func (h *ToolHandler) listOrganizations(params map[string]interface{}) (*CallToolResult, error) {
	owner := getOptionalStringParam(params, "owner")

	orgs, err := object.GetOrganizations(owner)
	if err != nil {
		return errorResult(fmt.Sprintf("Error retrieving organizations: %v", err)), nil
	}

	return marshalToCallResult(orgs)
}

// getApplication retrieves an application by ID
func (h *ToolHandler) getApplication(params map[string]interface{}) (*CallToolResult, error) {
	appId, ok := params["applicationId"].(string)
	if !ok {
		return errorResult("Missing or invalid applicationId parameter"), nil
	}

	app, err := object.GetApplication(appId)
	if err != nil {
		return errorResult(fmt.Sprintf("Error retrieving application: %v", err)), nil
	}

	if app == nil {
		return errorResult(fmt.Sprintf("Application not found: %s", appId)), nil
	}

	return marshalToCallResult(app)
}

// listApplications lists all applications
func (h *ToolHandler) listApplications(params map[string]interface{}) (*CallToolResult, error) {
	owner := getOptionalStringParam(params, "owner")

	apps, err := object.GetApplications(owner)
	if err != nil {
		return errorResult(fmt.Sprintf("Error retrieving applications: %v", err)), nil
	}

	return marshalToCallResult(apps)
}

// getRole retrieves a role by ID
func (h *ToolHandler) getRole(params map[string]interface{}) (*CallToolResult, error) {
	roleId, ok := params["roleId"].(string)
	if !ok {
		return errorResult("Missing or invalid roleId parameter"), nil
	}

	role, err := object.GetRole(roleId)
	if err != nil {
		return errorResult(fmt.Sprintf("Error retrieving role: %v", err)), nil
	}

	if role == nil {
		return errorResult(fmt.Sprintf("Role not found: %s", roleId)), nil
	}

	return marshalToCallResult(role)
}

// listRoles lists all roles
func (h *ToolHandler) listRoles(params map[string]interface{}) (*CallToolResult, error) {
	owner := getOptionalStringParam(params, "owner")

	roles, err := object.GetRoles(owner)
	if err != nil {
		return errorResult(fmt.Sprintf("Error retrieving roles: %v", err)), nil
	}

	return marshalToCallResult(roles)
}
