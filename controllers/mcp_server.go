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

package controllers

import (
	"encoding/json"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/casdoor/casdoor/mcpself"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

// ProxyServer
// @Title ProxyServer
// @Tag Server API
// @Description proxy request to the upstream MCP server by Server URL
// @Param   owner    path    string  true        "The owner name of the server"
// @Param   name     path    string  true        "The name of the server"
// @Success 200 {object} mcp.McpResponse The Response object
// @router /server/:owner/:name [post]
func (c *ApiController) ProxyServer() {
	owner := c.Ctx.Input.Param(":owner")
	name := c.Ctx.Input.Param(":name")

	var mcpReq *mcpself.McpRequest
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &mcpReq)
	if err != nil {
		c.McpResponseError(1, -32700, "Parse error", err.Error())
		return
	}
	if util.IsStringsEmpty(owner, name) {
		c.McpResponseError(1, -32600, "invalid server identifier", nil)
		return
	}

	server, err := object.GetServer(util.GetId(owner, name))
	if err != nil {
		c.McpResponseError(mcpReq.ID, -32600, "server not found", err.Error())
		return
	}
	if server == nil {
		c.McpResponseError(mcpReq.ID, -32600, "server not found", nil)
		return
	}
	if server.Url == "" {
		c.McpResponseError(mcpReq.ID, -32600, "server URL is empty", nil)
		return
	}

	targetUrl, err := url.Parse(server.Url)
	if err != nil || !targetUrl.IsAbs() || targetUrl.Host == "" {
		c.McpResponseError(mcpReq.ID, -32600, "server URL is invalid", nil)
		return
	}
	if targetUrl.Scheme != "http" && targetUrl.Scheme != "https" {
		c.McpResponseError(mcpReq.ID, -32600, "server URL scheme is invalid", nil)
		return
	}

	if mcpReq.Method == "tools/call" {
		var params mcpself.McpCallToolParams
		err = json.Unmarshal(mcpReq.Params, &params)
		if err != nil {
			c.McpResponseError(mcpReq.ID, -32600, "Invalid request", err.Error())
			return
		}

		for _, tool := range server.Tools {
			if tool.Name == params.Name && !tool.IsAllowed {
				c.McpResponseError(mcpReq.ID, -32600, "tool is forbidden", nil)
				return
			} else if tool.Name == params.Name {
				break
			}
		}
	}

	proxy := httputil.NewSingleHostReverseProxy(targetUrl)
	proxy.ErrorHandler = func(writer http.ResponseWriter, request *http.Request, proxyErr error) {
		c.Ctx.Output.SetStatus(http.StatusBadGateway)
		c.McpResponseError(mcpReq.ID, -32603, "failed to proxy server request: %s", proxyErr.Error())
	}
	proxy.Director = func(request *http.Request) {
		request.URL.Scheme = targetUrl.Scheme
		request.URL.Host = targetUrl.Host
		request.Host = targetUrl.Host
		request.URL.Path = targetUrl.Path
		request.URL.RawPath = ""
		request.URL.RawQuery = targetUrl.RawQuery

		if server.Token != "" {
			request.Header.Set("Authorization", "Bearer "+server.Token)
		}
	}

	proxy.ServeHTTP(c.Ctx.ResponseWriter, c.Ctx.Request)
}
