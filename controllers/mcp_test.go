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
	"testing"
)

func TestMCPInitialize(t *testing.T) {
	req := MCPRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params: json.RawMessage(`{
			"protocolVersion": "2024-11-05",
			"capabilities": {},
			"clientInfo": {
				"name": "test-client",
				"version": "1.0.0"
			}
		}`),
	}

	reqBytes, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	// Verify request structure
	var parsedReq MCPRequest
	err = json.Unmarshal(reqBytes, &parsedReq)
	if err != nil {
		t.Fatalf("Failed to unmarshal request: %v", err)
	}

	if parsedReq.Method != "initialize" {
		t.Errorf("Expected method 'initialize', got '%s'", parsedReq.Method)
	}
}

func TestMCPToolsList(t *testing.T) {
	req := MCPRequest{
		JSONRPC: "2.0",
		ID:      2,
		Method:  "tools/list",
	}

	reqBytes, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	// Verify request structure
	var parsedReq MCPRequest
	err = json.Unmarshal(reqBytes, &parsedReq)
	if err != nil {
		t.Fatalf("Failed to unmarshal request: %v", err)
	}

	if parsedReq.Method != "tools/list" {
		t.Errorf("Expected method 'tools/list', got '%s'", parsedReq.Method)
	}
}

func TestMCPToolsCall(t *testing.T) {
	params := MCPCallToolParams{
		Name: "get_applications",
		Arguments: map[string]interface{}{
			"owner": "built-in",
		},
	}

	paramsBytes, err := json.Marshal(params)
	if err != nil {
		t.Fatalf("Failed to marshal params: %v", err)
	}

	req := MCPRequest{
		JSONRPC: "2.0",
		ID:      3,
		Method:  "tools/call",
		Params:  json.RawMessage(paramsBytes),
	}

	reqBytes, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	// Verify request structure
	var parsedReq MCPRequest
	err = json.Unmarshal(reqBytes, &parsedReq)
	if err != nil {
		t.Fatalf("Failed to unmarshal request: %v", err)
	}

	if parsedReq.Method != "tools/call" {
		t.Errorf("Expected method 'tools/call', got '%s'", parsedReq.Method)
	}

	// Verify params
	var parsedParams MCPCallToolParams
	err = json.Unmarshal(parsedReq.Params, &parsedParams)
	if err != nil {
		t.Fatalf("Failed to unmarshal params: %v", err)
	}

	if parsedParams.Name != "get_applications" {
		t.Errorf("Expected tool name 'get_applications', got '%s'", parsedParams.Name)
	}

	if owner, ok := parsedParams.Arguments["owner"].(string); !ok || owner != "built-in" {
		t.Errorf("Expected owner 'built-in', got '%v'", parsedParams.Arguments["owner"])
	}
}

func TestMCPResponseStructure(t *testing.T) {
	// Test successful response
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

	resp := MCPResponse{
		JSONRPC: "2.0",
		ID:      1,
		Result:  result,
	}

	respBytes, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Failed to marshal response: %v", err)
	}

	// Verify response structure
	var parsedResp MCPResponse
	err = json.Unmarshal(respBytes, &parsedResp)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if parsedResp.JSONRPC != "2.0" {
		t.Errorf("Expected JSONRPC '2.0', got '%s'", parsedResp.JSONRPC)
	}

	// Test error response
	errResp := MCPResponse{
		JSONRPC: "2.0",
		ID:      2,
		Error: &MCPError{
			Code:    -32601,
			Message: "Method not found",
		},
	}

	errBytes, err := json.Marshal(errResp)
	if err != nil {
		t.Fatalf("Failed to marshal error response: %v", err)
	}

	// Verify error response structure
	var parsedErrResp MCPResponse
	err = json.Unmarshal(errBytes, &parsedErrResp)
	if err != nil {
		t.Fatalf("Failed to unmarshal error response: %v", err)
	}

	if parsedErrResp.Error == nil {
		t.Error("Expected error to be present")
	} else {
		if parsedErrResp.Error.Code != -32601 {
			t.Errorf("Expected error code -32601, got %d", parsedErrResp.Error.Code)
		}
		if parsedErrResp.Error.Message != "Method not found" {
			t.Errorf("Expected error message 'Method not found', got '%s'", parsedErrResp.Error.Message)
		}
	}
}
