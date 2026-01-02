// Copyright 2021 The Casdoor Authors. All Rights Reserved.
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
type MCPRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type MCPResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *MCPError   `json:"error,omitempty"`
}

type MCPError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type MCPInitializeParams struct {
	ProtocolVersion string                 `json:"protocolVersion"`
	Capabilities    map[string]interface{} `json:"capabilities"`
	ClientInfo      MCPImplementation      `json:"clientInfo"`
}

type MCPImplementation struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type MCPInitializeResult struct {
	ProtocolVersion string                `json:"protocolVersion"`
	Capabilities    MCPServerCapabilities `json:"capabilities"`
	ServerInfo      MCPImplementation     `json:"serverInfo"`
}

type MCPServerCapabilities struct {
	Tools map[string]interface{} `json:"tools,omitempty"`
}

type MCPListToolsResult struct {
	Tools []MCPTool `json:"tools"`
}

type MCPTool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

type MCPCallToolParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
}

type MCPCallToolResult struct {
	Content []MCPContent `json:"content"`
	IsError bool         `json:"isError,omitempty"`
}

type MCPContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// MCPController handles MCP protocol requests
type MCPController struct {
	web.Controller
}

// SendMCPResponse sends a successful MCP response
func (c *MCPController) SendMCPResponse(id interface{}, result interface{}) {
	resp := MCPResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
	c.Data["json"] = resp
	c.ServeJSON()
}

// SendMCPError sends an MCP error response
func (c *MCPController) SendMCPError(id interface{}, code int, message string, data interface{}) {
	resp := MCPResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &MCPError{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}
	c.Data["json"] = resp
	c.ServeJSON()
}

// SendToolResult sends a successful tool execution result
func (c *MCPController) SendToolResult(id interface{}, text string) {
	result := MCPCallToolResult{
		Content: []MCPContent{
			{
				Type: "text",
				Text: text,
			},
		},
		IsError: false,
	}
	c.SendMCPResponse(id, result)
}

// SendToolErrorResult sends a tool execution error result
func (c *MCPController) SendToolErrorResult(id interface{}, errorMsg string) {
	result := MCPCallToolResult{
		Content: []MCPContent{
			{
				Type: "text",
				Text: errorMsg,
			},
		},
		IsError: true,
	}
	c.SendMCPResponse(id, result)
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

// HandleMCP handles MCP protocol requests
// @Title HandleMCP
// @Tag MCP API
// @Description handle MCP (Model Context Protocol) requests
// @Success 200 {object} MCPResponse The Response object
// @router /mcp [post]
func (c *MCPController) HandleMCP() {
	var req MCPRequest
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &req)
	if err != nil {
		c.SendMCPError(nil, -32700, "Parse error", err.Error())
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
		c.SendMCPError(req.ID, -32601, "Method not found", fmt.Sprintf("Method '%s' not found", req.Method))
	}
}

func (c *MCPController) handleInitialize(req MCPRequest) {
	var params MCPInitializeParams
	if req.Params != nil {
		err := json.Unmarshal(req.Params, &params)
		if err != nil {
			c.SendMCPError(req.ID, -32602, "Invalid params", err.Error())
			return
		}
	}

	result := MCPInitializeResult{
		ProtocolVersion: "2024-11-05",
		Capabilities: MCPServerCapabilities{
			Tools: map[string]interface{}{},
		},
		ServerInfo: MCPImplementation{
			Name:    "Casdoor MCP Server",
			Version: "1.0.0",
		},
	}

	c.SendMCPResponse(req.ID, result)
}

func (c *MCPController) handleToolsList(req MCPRequest) {
	tools := []MCPTool{
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

	result := MCPListToolsResult{
		Tools: tools,
	}

	c.SendMCPResponse(req.ID, result)
}

func (c *MCPController) handleToolsCall(req MCPRequest) {
	var params MCPCallToolParams
	err := json.Unmarshal(req.Params, &params)
	if err != nil {
		c.SendMCPError(req.ID, -32602, "Invalid params", err.Error())
		return
	}

	// Route to the appropriate tool handler
	switch params.Name {
	case "get_applications":
		c.HandleGetApplicationsTool(req.ID, params.Arguments)
	case "get_application":
		c.HandleGetApplicationTool(req.ID, params.Arguments)
	case "add_application":
		c.HandleAddApplicationTool(req.ID, params.Arguments)
	case "update_application":
		c.HandleUpdateApplicationTool(req.ID, params.Arguments)
	case "delete_application":
		c.HandleDeleteApplicationTool(req.ID, params.Arguments)
	default:
		c.SendMCPError(req.ID, -32602, "Invalid tool name", fmt.Sprintf("Tool '%s' not found", params.Name))
	}
}
