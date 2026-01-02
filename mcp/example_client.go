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

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

// JSONRPCRequest represents a JSON-RPC 2.0 request
type JSONRPCRequest struct {
	JSONRPC string                 `json:"jsonrpc"`
	ID      interface{}            `json:"id,omitempty"`
	Method  string                 `json:"method"`
	Params  map[string]interface{} `json:"params,omitempty"`
}

// JSONRPCResponse represents a JSON-RPC 2.0 response
type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Error   *RPCError   `json:"error,omitempty"`
}

// RPCError represents a JSON-RPC error
type RPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run example_client.go <server_url> [auth_token]")
		fmt.Println("Example: go run example_client.go http://localhost:8000/mcp")
		os.Exit(1)
	}

	serverURL := os.Args[1]
	authToken := ""
	if len(os.Args) >= 3 {
		authToken = os.Args[2]
	}

	fmt.Println("=== Casdoor MCP Client Example ===\n")

	// Example 1: Initialize
	fmt.Println("1. Initializing connection...")
	initRequest := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params: map[string]interface{}{
			"protocolVersion": "2024-11-05",
		},
	}
	response, err := sendRequest(serverURL, authToken, initRequest)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		printResponse(response)
	}

	// Example 2: List Tools
	fmt.Println("\n2. Listing available tools...")
	listToolsRequest := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      2,
		Method:  "tools/list",
		Params:  map[string]interface{}{},
	}
	response, err = sendRequest(serverURL, authToken, listToolsRequest)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		printResponse(response)
	}

	// Example 3: Call a tool (list organizations)
	fmt.Println("\n3. Calling tool: list_organizations...")
	callToolRequest := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      3,
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name": "list_organizations",
			"arguments": map[string]interface{}{
				"owner": "",
			},
		},
	}
	response, err = sendRequest(serverURL, authToken, callToolRequest)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		printResponse(response)
	}

	fmt.Println("\n=== Example completed ===")
}

func sendRequest(serverURL, authToken string, request JSONRPCRequest) (*JSONRPCResponse, error) {
	body, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	req, err := http.NewRequest("POST", serverURL, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if authToken != "" {
		req.Header.Set("Authorization", "Bearer "+authToken)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	var jsonRPCResponse JSONRPCResponse
	if err := json.Unmarshal(responseBody, &jsonRPCResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}

	return &jsonRPCResponse, nil
}

func printResponse(response *JSONRPCResponse) {
	if response.Error != nil {
		fmt.Printf("Error: Code=%d, Message=%s\n", response.Error.Code, response.Error.Message)
		return
	}

	resultJSON, err := json.MarshalIndent(response.Result, "", "  ")
	if err != nil {
		fmt.Printf("Failed to format result: %v\n", err)
		return
	}

	fmt.Printf("Result:\n%s\n", string(resultJSON))
}
