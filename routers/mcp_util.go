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

package routers

import (
	"encoding/json"

	"github.com/beego/beego/v2/server/web/context"
	"github.com/casdoor/casdoor/util"
)

func getMcpObject(ctx *context.Context) (string, string, error) {
	body := ctx.Input.RequestBody
	if len(body) == 0 {
		return "", "", nil
	}

	// Parse MCP request to determine tool name
	type MCPRequest struct {
		Method string          `json:"method"`
		Params json.RawMessage `json:"params,omitempty"`
	}

	type MCPCallToolParams struct {
		Name      string                 `json:"name"`
		Arguments map[string]interface{} `json:"arguments,omitempty"`
	}

	var mcpReq MCPRequest
	err := json.Unmarshal(body, &mcpReq)
	if err != nil {
		return "", "", nil
	}

	// Only extract object for tool calls
	if mcpReq.Method != "tools/call" {
		return "", "", nil
	}

	var params MCPCallToolParams
	err = json.Unmarshal(mcpReq.Params, &params)
	if err != nil {
		return "", "", nil
	}

	// Extract owner/id from arguments based on tool
	switch params.Name {
	case "get_applications":
		if owner, ok := params.Arguments["owner"].(string); ok {
			return owner, "", nil
		}
	case "get_application", "update_application":
		if id, ok := params.Arguments["id"].(string); ok {
			return util.GetOwnerAndNameFromIdWithError(id)
		}
	case "add_application", "delete_application":
		if appData, ok := params.Arguments["application"].(map[string]interface{}); ok {
			return extractOwnerNameFromAppData(appData)
		}
	}

	return "", "", nil
}

// extractOwnerNameFromAppData extracts owner and name from application data
// Prioritizes organization field over owner field for consistency
func extractOwnerNameFromAppData(appData map[string]interface{}) (string, string, error) {
	// Try organization field first (used in application APIs)
	if org, ok := appData["organization"].(string); ok {
		if name, ok := appData["name"].(string); ok {
			return org, name, nil
		}
		return org, "", nil
	}
	// Fall back to owner field
	if owner, ok := appData["owner"].(string); ok {
		if name, ok := appData["name"].(string); ok {
			return owner, name, nil
		}
		return owner, "", nil
	}
	return "", "", nil
}

func getMcpUrlPath(ctx *context.Context) string {
	body := ctx.Input.RequestBody
	if len(body) == 0 {
		return "/api/mcp"
	}

	type MCPRequest struct {
		Method string          `json:"method"`
		Params json.RawMessage `json:"params,omitempty"`
	}

	type MCPCallToolParams struct {
		Name string `json:"name"`
	}

	var mcpReq MCPRequest
	err := json.Unmarshal(body, &mcpReq)
	if err != nil {
		return "/api/mcp"
	}

	// Map initialize and tools/list to public endpoints
	// These operations don't require special permissions beyond authentication
	// We use /api/get-application as it's a read-only operation that authenticated users can access
	if mcpReq.Method == "initialize" || mcpReq.Method == "tools/list" {
		return "/api/get-application"
	}

	if mcpReq.Method != "tools/call" {
		return "/api/mcp"
	}

	var params MCPCallToolParams
	err = json.Unmarshal(mcpReq.Params, &params)
	if err != nil {
		return "/api/mcp"
	}

	// Map MCP tool names to corresponding API paths
	switch params.Name {
	case "get_applications":
		return "/api/get-applications"
	case "get_application":
		return "/api/get-application"
	case "add_application":
		return "/api/add-application"
	case "update_application":
		return "/api/update-application"
	case "delete_application":
		return "/api/delete-application"
	default:
		return "/api/mcp"
	}
}
