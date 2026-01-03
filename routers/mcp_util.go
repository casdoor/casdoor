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

// applicationStub is a lightweight struct for extracting owner/name from application data
type applicationStub struct {
	Owner        string `json:"owner"`
	Name         string `json:"name"`
	Organization string `json:"organization"`
}

func getMcpObject(ctx *context.Context) (string, string, error) {
	body := ctx.Input.RequestBody
	if len(body) == 0 {
		return "", "", nil
	}

	// Parse MCP request to determine tool name
	type mcpRequest struct {
		Method string          `json:"method"`
		Params json.RawMessage `json:"params,omitempty"`
	}

	type mcpCallToolParams struct {
		Name      string          `json:"name"`
		Arguments json.RawMessage `json:"arguments,omitempty"`
	}

	type getApplicationsArgs struct {
		Owner string `json:"owner"`
	}

	type getApplicationArgs struct {
		Id string `json:"id"`
	}

	type addApplicationArgs struct {
		Application applicationStub `json:"application"`
	}

	type updateApplicationArgs struct {
		Id string `json:"id"`
	}

	type deleteApplicationArgs struct {
		Application applicationStub `json:"application"`
	}

	var mcpReq mcpRequest
	err := json.Unmarshal(body, &mcpReq)
	if err != nil {
		return "", "", nil
	}

	// Only extract object for tool calls
	if mcpReq.Method != "tools/call" {
		return "", "", nil
	}

	var params mcpCallToolParams
	err = json.Unmarshal(mcpReq.Params, &params)
	if err != nil {
		return "", "", nil
	}

	// Extract owner/id from arguments based on tool
	switch params.Name {
	case "get_applications":
		var args getApplicationsArgs
		if err := json.Unmarshal(params.Arguments, &args); err == nil {
			return args.Owner, "", nil
		}
	case "get_application":
		var args getApplicationArgs
		if err := json.Unmarshal(params.Arguments, &args); err == nil {
			return util.GetOwnerAndNameFromIdWithError(args.Id)
		}
	case "update_application":
		var args updateApplicationArgs
		if err := json.Unmarshal(params.Arguments, &args); err == nil {
			return util.GetOwnerAndNameFromIdWithError(args.Id)
		}
	case "add_application":
		var args addApplicationArgs
		if err := json.Unmarshal(params.Arguments, &args); err == nil {
			return extractOwnerNameFromAppStub(args.Application)
		}
	case "delete_application":
		var args deleteApplicationArgs
		if err := json.Unmarshal(params.Arguments, &args); err == nil {
			return extractOwnerNameFromAppStub(args.Application)
		}
	}

	return "", "", nil
}

// extractOwnerNameFromAppStub extracts owner and name from application stub
// Prioritizes organization field over owner field for consistency
func extractOwnerNameFromAppStub(app applicationStub) (string, string, error) {
	// Try organization field first (used in application APIs)
	if app.Organization != "" {
		return app.Organization, app.Name, nil
	}
	// Fall back to owner field
	if app.Owner != "" {
		return app.Owner, app.Name, nil
	}
	return "", "", nil
}

func getMcpUrlPath(ctx *context.Context) string {
	body := ctx.Input.RequestBody
	if len(body) == 0 {
		return "/api/mcp"
	}

	type mcpRequest struct {
		Method string          `json:"method"`
		Params json.RawMessage `json:"params,omitempty"`
	}

	type mcpCallToolParams struct {
		Name string `json:"name"`
	}

	var mcpReq mcpRequest
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

	var params mcpCallToolParams
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
