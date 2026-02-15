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
	"net/http"

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

func (c *McpController) McpResponseOk(id interface{}, result interface{}) {
	resp := BuildMcpResponse(id, result, nil)
	c.Ctx.Output.Header("Content-Type", "application/json")
	c.Data["json"] = resp
	c.ServeJSON()
}

func (c *McpController) McpResponseError(id interface{}, code int, message string, data interface{}) {
	resp := BuildMcpResponse(id, nil, &McpError{
		Code:    code,
		Message: message,
		Data:    data,
	})
	c.Ctx.Output.Header("Content-Type", "application/json")
	c.Data["json"] = resp
	c.ServeJSON()
}

// GetMcpResponse returns a McpResponse object
func BuildMcpResponse(id interface{}, result interface{}, err *McpError) McpResponse {
	resp := McpResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
		Error:   err,
	}
	return resp
}

// sendInvalidParamsError sends an invalid params error
func (c *McpController) sendInvalidParamsError(id interface{}, details string) {
	c.McpResponseError(id, -32602, "Invalid params", details)
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
	c.McpResponseOk(id, result)
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
	c.McpResponseOk(id, result)
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
		c.McpResponseError(nil, -32700, "Parse error", err.Error())
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
		c.McpResponseError(req.ID, -32601, "Method not found", fmt.Sprintf("Method '%s' not found", req.Method))
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

	c.McpResponseOk(req.ID, result)
}

func (c *McpController) handleNotificationsInitialized(req McpRequest) {
	c.Ctx.Output.SetStatus(http.StatusAccepted)
	c.Ctx.Output.Body([]byte{})
}

func (c *McpController) handlePing(req McpRequest) {
	// ping method is used to check if the server is alive and responsive
	// Return an empty object as result to indicate server is active
	c.McpResponseOk(req.ID, map[string]interface{}{})
}

func (c *McpController) handleToolsList(req McpRequest) {
	allTools := c.getAllTools()

	// Get JWT claims from the request
	claims := c.GetClaimsFromToken()

	// If no token is present, check session authentication
	if claims == nil {
		username := c.GetSessionUsername()
		// If user is authenticated via session, return all tools (backward compatibility)
		if username != "" {
			result := McpListToolsResult{
				Tools: allTools,
			}
			c.McpResponseOk(req.ID, result)
			return
		}

		// Unauthenticated request - return all tools for discovery
		// This allows clients to see what tools are available before authenticating
		result := McpListToolsResult{
			Tools: allTools,
		}
		c.McpResponseOk(req.ID, result)
		return
	}

	// Token-based authentication - filter tools by scopes
	grantedScopes := GetScopesFromClaims(claims)
	allowedTools := GetToolsForScopes(grantedScopes, BuiltinScopes)

	// Filter tools based on allowed scopes
	var filteredTools []McpTool
	for _, tool := range allTools {
		if allowedTools[tool.Name] {
			filteredTools = append(filteredTools, tool)
		}
	}

	result := McpListToolsResult{
		Tools: filteredTools,
	}

	c.McpResponseOk(req.ID, result)
}

func (c *McpController) handleToolsCall(req McpRequest) {
	var params McpCallToolParams
	err := json.Unmarshal(req.Params, &params)
	if err != nil {
		c.sendInvalidParamsError(req.ID, err.Error())
		return
	}

	// Check scope-tool permission
	if !c.checkToolPermission(req.ID, params.Name) {
		return // Error already sent by checkToolPermission
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
		c.McpResponseError(req.ID, -32602, "Invalid tool name", fmt.Sprintf("Tool '%s' not found", params.Name))
	}
}

// checkToolPermission validates that the current token has the required scope for the tool
// Returns false and sends an error response if permission is denied
func (c *McpController) checkToolPermission(id interface{}, toolName string) bool {
	// Get JWT claims from the request
	claims := c.GetClaimsFromToken()

	// If no token is present, check if the user is authenticated via session
	if claims == nil {
		username := c.GetSessionUsername()
		// If user is authenticated via session (e.g., session cookie), allow access
		// This maintains backward compatibility with existing session-based auth
		if username != "" {
			return true
		}

		// No authentication present - deny access
		c.sendInsufficientScopeError(id, toolName, []string{})
		return false
	}

	// Extract scopes from claims
	grantedScopes := GetScopesFromClaims(claims)

	// Get allowed tools for the granted scopes
	allowedTools := GetToolsForScopes(grantedScopes, BuiltinScopes)

	// Check if the requested tool is allowed
	if !allowedTools[toolName] {
		c.sendInsufficientScopeError(id, toolName, grantedScopes)
		return false
	}

	return true
}

// sendInsufficientScopeError sends an error response for insufficient scope
func (c *McpController) sendInsufficientScopeError(id interface{}, toolName string, grantedScopes []string) {
	// Find required scope for this tool
	requiredScope := GetRequiredScopeForTool(toolName, BuiltinScopes)

	errorData := map[string]interface{}{
		"tool":           toolName,
		"granted_scopes": grantedScopes,
	}
	if requiredScope != "" {
		errorData["required_scope"] = requiredScope
	}

	c.McpResponseError(id, -32001, "insufficient_scope", errorData)
}

// getAllTools returns all available MCP tools
func (c *McpController) getAllTools() []McpTool {
	return []McpTool{
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
}
