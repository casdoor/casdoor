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
	"io"
	"net/http"
)

const (
	// MCP Protocol Version
	ProtocolVersion = "2024-11-05"

	// Error codes
	ParseError     = -32700
	InvalidRequest = -32600
	MethodNotFound = -32601
	InvalidParams  = -32602
	InternalError  = -32603
)

// Server represents the MCP server
type Server struct {
	toolHandler *ToolHandler
}

// NewServer creates a new MCP server
func NewServer() *Server {
	return &Server{
		toolHandler: &ToolHandler{},
	}
}

// ServeHTTP handles HTTP requests for the MCP server
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		s.writeError(w, nil, MethodNotFound, "Only POST method is supported")
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		s.writeError(w, nil, ParseError, "Failed to read request body")
		return
	}
	defer r.Body.Close()

	var request JSONRPCRequest
	if err := json.Unmarshal(body, &request); err != nil {
		s.writeError(w, nil, ParseError, "Invalid JSON")
		return
	}

	if request.JSONRPC != "2.0" {
		s.writeError(w, request.ID, InvalidRequest, "Invalid JSON-RPC version")
		return
	}

	result, err := s.handleMethod(request.Method, request.Params)
	if err != nil {
		s.writeError(w, request.ID, InternalError, err.Error())
		return
	}

	response := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      request.ID,
		Result:  result,
	}

	responseData, err := json.Marshal(response)
	if err != nil {
		s.writeError(w, request.ID, InternalError, "Failed to marshal response")
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(responseData)
}

// handleMethod processes different MCP methods
func (s *Server) handleMethod(method string, params map[string]interface{}) (interface{}, error) {
	switch method {
	case "initialize":
		return s.handleInitialize(params)
	case "tools/list":
		return s.handleListTools()
	case "tools/call":
		return s.handleCallTool(params)
	case "resources/list":
		return s.handleListResources()
	case "prompts/list":
		return s.handleListPrompts()
	default:
		return nil, fmt.Errorf("method not found: %s", method)
	}
}

// handleInitialize handles the initialize method
func (s *Server) handleInitialize(params map[string]interface{}) (interface{}, error) {
	return InitializeResult{
		ProtocolVersion: ProtocolVersion,
		Capabilities: Capabilities{
			Tools: &ToolsCapability{
				ListChanged: false,
			},
			Resources: &ResourcesCapability{
				Subscribe:   false,
				ListChanged: false,
			},
			Prompts: &PromptsCapability{
				ListChanged: false,
			},
		},
		ServerInfo: ServerInfo{
			Name:    "casdoor-mcp-server",
			Version: "1.0.0",
		},
	}, nil
}

// handleListTools handles the tools/list method
func (s *Server) handleListTools() (interface{}, error) {
	tools := []ToolInfo{
		{
			Name:        "get_user",
			Description: "Retrieve a user by their user ID (format: owner/username)",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"userId": map[string]interface{}{
						"type":        "string",
						"description": "User ID in format owner/username",
					},
				},
				"required": []string{"userId"},
			},
		},
		{
			Name:        "list_users",
			Description: "List all users, optionally filtered by owner/organization",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"owner": map[string]interface{}{
						"type":        "string",
						"description": "Optional owner/organization to filter users",
					},
				},
			},
		},
		{
			Name:        "get_organization",
			Description: "Retrieve an organization by ID",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"organizationId": map[string]interface{}{
						"type":        "string",
						"description": "Organization ID",
					},
				},
				"required": []string{"organizationId"},
			},
		},
		{
			Name:        "list_organizations",
			Description: "List all organizations",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"owner": map[string]interface{}{
						"type":        "string",
						"description": "Optional owner to filter organizations",
					},
				},
			},
		},
		{
			Name:        "get_application",
			Description: "Retrieve an application by ID (format: owner/appname)",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"applicationId": map[string]interface{}{
						"type":        "string",
						"description": "Application ID in format owner/appname",
					},
				},
				"required": []string{"applicationId"},
			},
		},
		{
			Name:        "list_applications",
			Description: "List all applications, optionally filtered by owner",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"owner": map[string]interface{}{
						"type":        "string",
						"description": "Optional owner to filter applications",
					},
				},
			},
		},
		{
			Name:        "get_role",
			Description: "Retrieve a role by ID (format: owner/rolename)",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"roleId": map[string]interface{}{
						"type":        "string",
						"description": "Role ID in format owner/rolename",
					},
				},
				"required": []string{"roleId"},
			},
		},
		{
			Name:        "list_roles",
			Description: "List all roles, optionally filtered by owner",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"owner": map[string]interface{}{
						"type":        "string",
						"description": "Optional owner to filter roles",
					},
				},
			},
		},
	}

	return ListToolsResult{Tools: tools}, nil
}

// handleCallTool handles the tools/call method
func (s *Server) handleCallTool(params map[string]interface{}) (interface{}, error) {
	name, ok := params["name"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid tool name")
	}

	arguments, ok := params["arguments"].(map[string]interface{})
	if !ok {
		arguments = make(map[string]interface{})
	}

	return s.toolHandler.CallTool(name, arguments)
}

// handleListResources handles the resources/list method
func (s *Server) handleListResources() (interface{}, error) {
	// Resources represent read-only data sources
	// For now, we'll return an empty list, but this can be extended
	// to include things like system info, configuration, etc.
	return ListResourcesResult{Resources: []ResourceInfo{}}, nil
}

// handleListPrompts handles the prompts/list method
func (s *Server) handleListPrompts() (interface{}, error) {
	// Prompts represent template messages
	// For now, we'll return an empty list
	return ListPromptsResult{Prompts: []PromptInfo{}}, nil
}

// writeError writes an error response
func (s *Server) writeError(w http.ResponseWriter, id interface{}, code int, message string) {
	response := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &RPCError{
			Code:    code,
			Message: message,
		},
	}

	responseData, _ := json.Marshal(response)
	w.WriteHeader(http.StatusOK)
	w.Write(responseData)
}
