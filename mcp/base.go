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
	"encoding/json"
	"fmt"

	"github.com/beego/beego/v2/server/web"
)

// MCP JSON-RPC 2.0 structures
type McpRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type McpResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *McpError   `json:"error,omitempty"`
}

type McpError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type McpInitializeParams struct {
	ProtocolVersion string                 `json:"protocolVersion"`
	Capabilities    map[string]interface{} `json:"capabilities"`
	ClientInfo      McpImplementation      `json:"clientInfo"`
}

type McpImplementation struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type McpInitializeResult struct {
	ProtocolVersion string                `json:"protocolVersion"`
	Capabilities    McpServerCapabilities `json:"capabilities"`
	ServerInfo      McpImplementation     `json:"serverInfo"`
}

type McpServerCapabilities struct {
	Tools map[string]interface{} `json:"tools,omitempty"`
}

type McpListToolsResult struct {
	Tools []McpTool `json:"tools"`
}

type McpTool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

type McpCallToolParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
}

type McpCallToolResult struct {
	Content []McpContent `json:"content"`
	IsError bool         `json:"isError,omitempty"`
}

type McpContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// MCPController handles MCP protocol requests
type McpController struct {
	web.Controller
}

// SendMcpResponse sends a successful MCP response
func (c *McpController) SendMcpResponse(id interface{}, result interface{}) {
	resp := McpResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
	c.Data["json"] = resp
	c.ServeJSON()
}

// SendMcpError sends an MCP error response
func (c *McpController) SendMcpError(id interface{}, code int, message string, data interface{}) {
	resp := McpResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &McpError{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}
	c.Data["json"] = resp
	c.ServeJSON()
}

// sendInvalidParamsError sends an invalid params error
func (c *McpController) sendInvalidParamsError(id interface{}, details string) {
	c.SendMcpError(id, -32602, "Invalid params", details)
}

// SendToolResult sends a successful tool execution result
func (c *McpController) SendToolResult(id interface{}, text string) {
	result := McpCallToolResult{
		Content: []McpContent{
			{
				Type: "text",
				Text: text,
			},
		},
		IsError: false,
	}
	c.SendMcpResponse(id, result)
}

// SendToolErrorResult sends a tool execution error result
func (c *McpController) SendToolErrorResult(id interface{}, errorMsg string) {
	result := McpCallToolResult{
		Content: []McpContent{
			{
				Type: "text",
				Text: errorMsg,
			},
		},
		IsError: true,
	}
	c.SendMcpResponse(id, result)
}

// FormatOperationResult formats the result of CRUD operations in a clear, descriptive way
func FormatOperationResult(operation, resourceType string, affected bool) string {
	if affected {
		// Map operation to past tense
		pastTense := operation + "d"
		if operation == "add" {
			pastTense = "added"
		} else if operation == "update" {
			pastTense = "updated"
		} else if operation == "delete" {
			pastTense = "deleted"
		}
		return fmt.Sprintf("Successfully %s %s", pastTense, resourceType)
	}
	return fmt.Sprintf("No changes were made to %s (may already exist or not found)", resourceType)
}

// HandleMcp handles MCP protocol requests
// @Title HandleMcp
// @Tag MCP API
// @Description handle MCP (Model Context Protocol) requests
// @Success 200 {object} McpResponse The Response object
// @router /mcp [post]
func (c *McpController) HandleMcp() {
	var req McpRequest
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &req)
	if err != nil {
		c.SendMcpError(nil, -32700, "Parse error", err.Error())
		return
	}

	// Handle different MCP methods
	switch req.Method {
	case "initialize":
		c.handleInitialize(req)
	case "tools/list":
		c.handleToolsList(req)
	case "tools/call":
		c.handleToolsCall(req)
	default:
		c.SendMcpError(req.ID, -32601, "Method not found", fmt.Sprintf("Method '%s' not found", req.Method))
	}
}

func (c *McpController) handleInitialize(req McpRequest) {
	var params McpInitializeParams
	if req.Params != nil {
		err := json.Unmarshal(req.Params, &params)
		if err != nil {
			c.sendInvalidParamsError(req.ID, err.Error())
			return
		}
	}

	result := McpInitializeResult{
		ProtocolVersion: "2024-11-05",
		Capabilities: McpServerCapabilities{
			Tools: map[string]interface{}{},
		},
		ServerInfo: McpImplementation{
			Name:    "Casdoor MCP Server",
			Version: "1.0.0",
		},
	}

	c.SendMcpResponse(req.ID, result)
}

func (c *McpController) handleToolsList(req McpRequest) {
	tools := []McpTool{
		{
			Name:        "get_applications",
			Description: "Get all applications for a specific owner",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"owner": map[string]interface{}{
						"type":        "string",
						"description": "The owner of applications",
					},
				},
				"required": []string{"owner"},
			},
		},
		{
			Name:        "get_application",
			Description: "Get the detail of a specific application",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"id": map[string]interface{}{
						"type":        "string",
						"description": "The id (owner/name) of the application",
					},
				},
				"required": []string{"id"},
			},
		},
		{
			Name:        "add_application",
			Description: "Add a new application",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"application": map[string]interface{}{
						"type":        "object",
						"description": "The application object to add",
					},
				},
				"required": []string{"application"},
			},
		},
		{
			Name:        "update_application",
			Description: "Update an existing application",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"id": map[string]interface{}{
						"type":        "string",
						"description": "The id (owner/name) of the application",
					},
					"application": map[string]interface{}{
						"type":        "object",
						"description": "The updated application object",
					},
				},
				"required": []string{"id", "application"},
			},
		},
		{
			Name:        "delete_application",
			Description: "Delete an application",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"application": map[string]interface{}{
						"type":        "object",
						"description": "The application object to delete",
					},
				},
				"required": []string{"application"},
			},
		},
	}

	result := McpListToolsResult{
		Tools: tools,
	}

	c.SendMcpResponse(req.ID, result)
}

func (c *McpController) handleToolsCall(req McpRequest) {
	var params McpCallToolParams
	err := json.Unmarshal(req.Params, &params)
	if err != nil {
		c.sendInvalidParamsError(req.ID, err.Error())
		return
	}

	// Convert ID to string for tool handlers
	idStr := fmt.Sprintf("%v", req.ID)

	// Route to the appropriate tool handler
	switch params.Name {
	case "get_applications":
		c.handleGetApplicationsTool(idStr, params.Arguments)
	case "get_application":
		c.handleGetApplicationTool(idStr, params.Arguments)
	case "add_application":
		c.handleAddApplicationTool(idStr, params.Arguments)
	case "update_application":
		c.handleUpdateApplicationTool(idStr, params.Arguments)
	case "delete_application":
		c.handleDeleteApplicationTool(idStr, params.Arguments)
	default:
		c.SendMcpError(req.ID, -32602, "Invalid tool name", fmt.Sprintf("Tool '%s' not found", params.Name))
	}
}
