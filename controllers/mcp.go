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

package controllers

import (
	"encoding/json"
	"fmt"

	"github.com/casdoor/casdoor/object"
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
	ProtocolVersion string                 `json:"protocolVersion"`
	Capabilities    MCPServerCapabilities  `json:"capabilities"`
	ServerInfo      MCPImplementation      `json:"serverInfo"`
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

// HandleMCP handles MCP protocol requests
// @Title HandleMCP
// @Tag MCP API
// @Description handle MCP (Model Context Protocol) requests
// @Success 200 {object} MCPResponse The Response object
// @router /mcp [post]
func (c *ApiController) HandleMCP() {
	var req MCPRequest
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &req)
	if err != nil {
		c.sendMCPError(nil, -32700, "Parse error", err.Error())
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
		c.sendMCPError(req.ID, -32601, "Method not found", fmt.Sprintf("Method '%s' not found", req.Method))
	}
}

func (c *ApiController) handleInitialize(req MCPRequest) {
	var params MCPInitializeParams
	if req.Params != nil {
		err := json.Unmarshal(req.Params, &params)
		if err != nil {
			c.sendMCPError(req.ID, -32602, "Invalid params", err.Error())
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

	c.sendMCPResponse(req.ID, result)
}

func (c *ApiController) handleToolsList(req MCPRequest) {
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

	c.sendMCPResponse(req.ID, result)
}

func (c *ApiController) handleToolsCall(req MCPRequest) {
	var params MCPCallToolParams
	err := json.Unmarshal(req.Params, &params)
	if err != nil {
		c.sendMCPError(req.ID, -32602, "Invalid params", err.Error())
		return
	}

	// Route to the appropriate tool handler
	switch params.Name {
	case "get_applications":
		c.handleGetApplicationsTool(req.ID, params.Arguments)
	case "get_application":
		c.handleGetApplicationTool(req.ID, params.Arguments)
	case "add_application":
		c.handleAddApplicationTool(req.ID, params.Arguments)
	case "update_application":
		c.handleUpdateApplicationTool(req.ID, params.Arguments)
	case "delete_application":
		c.handleDeleteApplicationTool(req.ID, params.Arguments)
	default:
		c.sendMCPError(req.ID, -32602, "Invalid tool name", fmt.Sprintf("Tool '%s' not found", params.Name))
	}
}

func (c *ApiController) handleGetApplicationsTool(id interface{}, args map[string]interface{}) {
	userId := c.GetSessionUsername()
	owner, ok := args["owner"].(string)
	if !ok {
		c.sendMCPError(id, -32602, "Invalid params", "Missing or invalid 'owner' parameter")
		return
	}

	applications, err := object.GetApplications(owner)
	if err != nil {
		c.sendToolErrorResult(id, err.Error())
		return
	}

	maskedApps := object.GetMaskedApplications(applications, userId)
	jsonData, err := json.MarshalIndent(maskedApps, "", "  ")
	if err != nil {
		c.sendToolErrorResult(id, err.Error())
		return
	}

	c.sendToolResult(id, string(jsonData))
}

func (c *ApiController) handleGetApplicationTool(id interface{}, args map[string]interface{}) {
	userId := c.GetSessionUsername()
	appId, ok := args["id"].(string)
	if !ok {
		c.sendMCPError(id, -32602, "Invalid params", "Missing or invalid 'id' parameter")
		return
	}

	application, err := object.GetApplication(appId)
	if err != nil {
		c.sendToolErrorResult(id, err.Error())
		return
	}

	maskedApp := object.GetMaskedApplication(application, userId)
	jsonData, err := json.MarshalIndent(maskedApp, "", "  ")
	if err != nil {
		c.sendToolErrorResult(id, err.Error())
		return
	}

	c.sendToolResult(id, string(jsonData))
}

func (c *ApiController) handleAddApplicationTool(id interface{}, args map[string]interface{}) {
	appData, ok := args["application"].(map[string]interface{})
	if !ok {
		c.sendMCPError(id, -32602, "Invalid params", "Missing or invalid 'application' parameter")
		return
	}

	jsonBytes, err := json.Marshal(appData)
	if err != nil {
		c.sendToolErrorResult(id, err.Error())
		return
	}

	var application object.Application
	err = json.Unmarshal(jsonBytes, &application)
	if err != nil {
		c.sendToolErrorResult(id, err.Error())
		return
	}

	count, err := object.GetApplicationCount("", "", "")
	if err != nil {
		c.sendToolErrorResult(id, err.Error())
		return
	}

	if err := checkQuotaForApplication(int(count)); err != nil {
		c.sendToolErrorResult(id, err.Error())
		return
	}

	if err = object.CheckIpWhitelist(application.IpWhitelist, c.GetAcceptLanguage()); err != nil {
		c.sendToolErrorResult(id, err.Error())
		return
	}

	affected, err := object.AddApplication(&application)
	if err != nil {
		c.sendToolErrorResult(id, err.Error())
		return
	}

	result := "Unaffected"
	if affected {
		result = "Affected"
	}
	c.sendToolResult(id, fmt.Sprintf("Application added successfully: %s", result))
}

func (c *ApiController) handleUpdateApplicationTool(id interface{}, args map[string]interface{}) {
	appId, ok := args["id"].(string)
	if !ok {
		c.sendMCPError(id, -32602, "Invalid params", "Missing or invalid 'id' parameter")
		return
	}

	appData, ok := args["application"].(map[string]interface{})
	if !ok {
		c.sendMCPError(id, -32602, "Invalid params", "Missing or invalid 'application' parameter")
		return
	}

	jsonBytes, err := json.Marshal(appData)
	if err != nil {
		c.sendToolErrorResult(id, err.Error())
		return
	}

	var application object.Application
	err = json.Unmarshal(jsonBytes, &application)
	if err != nil {
		c.sendToolErrorResult(id, err.Error())
		return
	}

	if err = object.CheckIpWhitelist(application.IpWhitelist, c.GetAcceptLanguage()); err != nil {
		c.sendToolErrorResult(id, err.Error())
		return
	}

	affected, err := object.UpdateApplication(appId, &application, c.IsGlobalAdmin(), c.GetAcceptLanguage())
	if err != nil {
		c.sendToolErrorResult(id, err.Error())
		return
	}

	result := "Unaffected"
	if affected {
		result = "Affected"
	}
	c.sendToolResult(id, fmt.Sprintf("Application updated successfully: %s", result))
}

func (c *ApiController) handleDeleteApplicationTool(id interface{}, args map[string]interface{}) {
	appData, ok := args["application"].(map[string]interface{})
	if !ok {
		c.sendMCPError(id, -32602, "Invalid params", "Missing or invalid 'application' parameter")
		return
	}

	jsonBytes, err := json.Marshal(appData)
	if err != nil {
		c.sendToolErrorResult(id, err.Error())
		return
	}

	var application object.Application
	err = json.Unmarshal(jsonBytes, &application)
	if err != nil {
		c.sendToolErrorResult(id, err.Error())
		return
	}

	affected, err := object.DeleteApplication(&application)
	if err != nil {
		c.sendToolErrorResult(id, err.Error())
		return
	}

	result := "Unaffected"
	if affected {
		result = "Affected"
	}
	c.sendToolResult(id, fmt.Sprintf("Application deleted successfully: %s", result))
}

func (c *ApiController) sendMCPResponse(id interface{}, result interface{}) {
	resp := MCPResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
	c.Data["json"] = resp
	c.ServeJSON()
}

func (c *ApiController) sendMCPError(id interface{}, code int, message string, data interface{}) {
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

func (c *ApiController) sendToolResult(id interface{}, text string) {
	result := MCPCallToolResult{
		Content: []MCPContent{
			{
				Type: "text",
				Text: text,
			},
		},
		IsError: false,
	}
	c.sendMCPResponse(id, result)
}

func (c *ApiController) sendToolErrorResult(id interface{}, errorMsg string) {
	result := MCPCallToolResult{
		Content: []MCPContent{
			{
				Type: "text",
				Text: errorMsg,
			},
		},
		IsError: true,
	}
	c.sendMCPResponse(id, result)
}
