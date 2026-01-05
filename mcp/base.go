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
	"github.com/casdoor/casdoor/object"
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
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments,omitempty"`
}

// Tool-specific argument structs
type GetApplicationsArgs struct {
	Owner string `json:"owner"`
}

type GetApplicationArgs struct {
	Id string `json:"id"`
}

type AddApplicationArgs struct {
	Application object.Application `json:"application"`
}

type UpdateApplicationArgs struct {
	Id          string             `json:"id"`
	Application object.Application `json:"application"`
}

type DeleteApplicationArgs struct {
	Application object.Application `json:"application"`
}

type McpCallToolResult struct {
	Content []TextContent `json:"content"`
	IsError bool          `json:"isError,omitempty"`
}

type TextContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// MCPController handles MCP protocol requests
type McpController struct {
	web.Controller
}

func (c *McpController) Prepare() {
	c.EnableRender = false
}

// SendMcpResponse sends a successful MCP response
func (c *McpController) SendMcpResponse(id interface{}, result interface{}) {
	resp := McpResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}

	// Set proper HTTP headers for MCP responses
	c.Ctx.Output.Header("Content-Type", "application/json")
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

	// Set proper HTTP headers for MCP responses
	c.Ctx.Output.Header("Content-Type", "application/json")
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
		Content: []TextContent{
			{
				Type: "text",
				Text: text,
			},
		},
	}
	c.SendMcpResponse(id, result)
}

// SendToolErrorResult sends a tool execution error result
func (c *McpController) SendToolErrorResult(id interface{}, errorMsg string) {
	result := McpCallToolResult{
		Content: []TextContent{
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
	case "notifications/initialized":
		c.handleNotificationsInitialized(req)
	case "ping":
		c.handlePing(req)
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
			Tools: map[string]interface{}{
				"listChanged": true,
			},
		},
		ServerInfo: McpImplementation{
			Name:    "Casdoor MCP Server",
			Version: "1.0.0",
		},
	}

	c.SendMcpResponse(req.ID, result)
}

func (c *McpController) handleNotificationsInitialized(req McpRequest) {
	// notifications/initialized is a notification from client indicating
	// that the initialization process is complete and the client is ready
	// to start using the server. No response is expected for notifications.
	// We can log this event or perform any post-initialization setup here.
}

func (c *McpController) handlePing(req McpRequest) {
	// ping method is used to check if the server is alive and responsive
	// Return an empty object as result to indicate server is active
	c.SendMcpResponse(req.ID, map[string]interface{}{})
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

	// Route to the appropriate tool handler
	switch params.Name {
	case "get_applications":
		var args GetApplicationsArgs
		if err := json.Unmarshal(params.Arguments, &args); err != nil {
			c.sendInvalidParamsError(req.ID, err.Error())
			return
		}
		c.handleGetApplicationsTool(req.ID, args)
	case "get_application":
		var args GetApplicationArgs
		if err := json.Unmarshal(params.Arguments, &args); err != nil {
			c.sendInvalidParamsError(req.ID, err.Error())
			return
		}
		c.handleGetApplicationTool(req.ID, args)
	case "add_application":
		var args AddApplicationArgs
		if err := json.Unmarshal(params.Arguments, &args); err != nil {
			c.sendInvalidParamsError(req.ID, err.Error())
			return
		}
		c.handleAddApplicationTool(req.ID, args)
	case "update_application":
		var args UpdateApplicationArgs
		if err := json.Unmarshal(params.Arguments, &args); err != nil {
			c.sendInvalidParamsError(req.ID, err.Error())
			return
		}
		c.handleUpdateApplicationTool(req.ID, args)
	case "delete_application":
		var args DeleteApplicationArgs
		if err := json.Unmarshal(params.Arguments, &args); err != nil {
			c.sendInvalidParamsError(req.ID, err.Error())
			return
		}
		c.handleDeleteApplicationTool(req.ID, args)
	default:
		c.SendMcpError(req.ID, -32602, "Invalid tool name", fmt.Sprintf("Tool '%s' not found", params.Name))
	}
}
