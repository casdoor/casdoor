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

package mcpself

import (
	"encoding/json"

	"github.com/beego/beego/v2/server/web/context"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

// ExtraTool is a tool that an extension adds to the MCP server, with the scope that grants it
type ExtraTool struct {
	Tool  McpTool
	Scope string
}

// Extension adds tools to the MCP server and wraps its tool calls
type Extension interface {
	// GetTools returns every tool that the extension adds, whether it is available or not
	GetTools() []*ExtraTool
	// ListTools returns the tools that tools/list shows, given the server's own tools
	ListTools(tools []McpTool) []McpTool
	// CallTool handles a tool call, next runs a tool of the server itself
	CallTool(c *McpController, id interface{}, name string, arguments json.RawMessage, next func())
}

// McpExtension is set by the editions that add tools to the MCP server
var McpExtension Extension

// AuthorizeRequestHook authorizes a request that a tool makes to the REST API inside the server, the same way as the
// request would be authorized from outside
var AuthorizeRequestHook func(ctx *context.Context) error

func getExtraTools() []*ExtraTool {
	if McpExtension == nil {
		return nil
	}
	return McpExtension.GetTools()
}

func isBuiltinTool(name string) bool {
	for _, tool := range (&McpController{}).getAllTools() {
		if tool.Name == name {
			return true
		}
	}
	return false
}

// IsExtraTool tells whether a tool is added by an extension: its calls are authorized request by request. The
// server's own tools are authorized by their arguments, an extension can't take their names.
func IsExtraTool(name string) bool {
	if isBuiltinTool(name) {
		return false
	}
	for _, tool := range getExtraTools() {
		if tool.Tool.Name == name {
			return true
		}
	}
	return false
}

// getScopeRegistry is BuiltinScopes with the tools of the extension added to the scopes that grant them
func getScopeRegistry() []*object.ScopeItem {
	extraTools := getExtraTools()
	if len(extraTools) == 0 {
		return BuiltinScopes
	}

	registry := make([]*object.ScopeItem, 0, len(BuiltinScopes))
	for _, item := range BuiltinScopes {
		scopeItem := *item
		scopeItem.Tools = append([]string{}, item.Tools...)
		for _, tool := range extraTools {
			if tool.Scope == item.Name && !util.InSlice(scopeItem.Tools, tool.Tool.Name) {
				scopeItem.Tools = append(scopeItem.Tools, tool.Tool.Name)
			}
		}
		registry = append(registry, &scopeItem)
	}
	return registry
}
