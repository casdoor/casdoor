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
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/beego/beego/v2/server/web/pagination"
	"github.com/casdoor/casdoor/edition"
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
// @router /server/:owner/:name [get,post]
func (c *ApiController) ProxyServer() {
	owner := c.Ctx.Input.Param(":owner")
	name := c.Ctx.Input.Param(":name")

	var mcpReq *mcpself.McpRequest
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &mcpReq)
	if err != nil {
		c.McpResponseError(1, -32700, "Parse error", err.Error())
		return
	}
	if mcpReq == nil || isAmbiguousMcpRequest(c.Ctx.Input.RequestBody, mcpReq) {
		c.McpResponseError(1, -32600, "Invalid request", "duplicate or case-variant keys are not allowed")
		return
	}
	if util.IsStringsEmpty(owner, name) {
		c.McpResponseError(1, -32600, "invalid server identifier", nil)
		return
	}
	if err = edition.Require(edition.Commercial, "MCP gateway"); err != nil {
		c.McpResponseError(mcpReq.ID, -32600, err.Error(), nil)
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

		if len(server.Tools) > 0 {
			toolAllowed := false
			for _, tool := range server.Tools {
				if tool.Name == params.Name {
					if !tool.IsAllowed {
						c.recordBlockedMcpToolCall(server, params, "tool is forbidden")
						c.McpResponseError(mcpReq.ID, -32600, "tool is forbidden", nil)
						return
					}
					toolAllowed = true
					break
				}
			}
			if !toolAllowed {
				c.recordBlockedMcpToolCall(server, params, "tool is not in the allowlist")
				c.McpResponseError(mcpReq.ID, -32600, "tool is not in the allowlist", nil)
				return
			}
		}
	}

	proxy := httputil.NewSingleHostReverseProxy(targetUrl)
	if server.Owner != "built-in" {
		proxy.Transport = util.NewInternetOnlyHttpClient(30 * time.Second).Transport
	}
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

		// don't forward the caller's own credentials to the upstream MCP server
		request.Header.Del("Cookie")
		request.Header.Del("X-Casdoor-Access-Key")
		request.Header.Del("X-Casdoor-Access-Secret")
		request.Header.Del("Authorization")
		if server.Token != "" {
			// some platforms take the key in their own header, e.g. "X-Api-Key"
			if server.TokenHeader != "" && !strings.EqualFold(server.TokenHeader, "Authorization") {
				request.Header.Set(server.TokenHeader, server.Token)
			} else {
				request.Header.Set("Authorization", "Bearer "+server.Token)
			}
		}
	}

	if mcpReq.Method != "tools/call" || !object.IsMcpToolCallRecorded(server.Owner) {
		proxy.ServeHTTP(c.Ctx.ResponseWriter, c.Ctx.Request)
		return
	}

	recorder := &mcpResponseRecorder{ResponseWriter: c.Ctx.ResponseWriter.ResponseWriter, statusCode: http.StatusOK}
	startTime := time.Now()
	proxy.ServeHTTP(recorder, c.Ctx.Request)
	c.recordMcpToolCall(server, mcpReq, recorder, time.Since(startTime))
}

// mcpResponseRecorder keeps the status and the beginning of the upstream response for the tool call record
type mcpResponseRecorder struct {
	http.ResponseWriter
	statusCode int
	body       bytes.Buffer
}

func (r *mcpResponseRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *mcpResponseRecorder) Write(data []byte) (int, error) {
	if r.body.Len() < 8192 {
		r.body.Write(data[:min(len(data), 8192-r.body.Len())])
	}
	return r.ResponseWriter.Write(data)
}

func (r *mcpResponseRecorder) Flush() {
	if flusher, ok := r.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (c *ApiController) recordMcpToolCall(server *object.Server, mcpReq *mcpself.McpRequest, recorder *mcpResponseRecorder, duration time.Duration) {
	var params mcpself.McpCallToolParams
	_ = json.Unmarshal(mcpReq.Params, &params)
	arguments, _ := json.Marshal(params.Arguments)

	response := recorder.body.String()
	result := "Success"
	if recorder.statusCode >= 400 {
		result = "Failed"
	} else if strings.Contains(response, `"error"`) || strings.Contains(response, `"isError":true`) {
		result = "Error"
	}

	object.AddMcpRecord(&object.McpRecord{
		Owner:      server.Owner,
		Server:     server.Name,
		UserName:   c.GetSessionUsername(),
		ClientIp:   util.GetClientIpFromRequest(c.Ctx.Request),
		UserAgent:  c.Ctx.Request.UserAgent(),
		Tool:       params.Name,
		Arguments:  string(arguments),
		Result:     result,
		StatusCode: recorder.statusCode,
		Response:   response,
		DurationMs: duration.Milliseconds(),
	})
}

func isAmbiguousMcpRequest(body []byte, mcpReq *mcpself.McpRequest) bool {
	if hasAmbiguousJsonKeys(body, "method", "params") {
		return true
	}
	return len(mcpReq.Params) > 0 && hasAmbiguousJsonKeys(mcpReq.Params, "name")
}

func hasAmbiguousJsonKeys(data []byte, keys ...string) bool {
	decoder := json.NewDecoder(bytes.NewReader(data))
	token, err := decoder.Token()
	if err != nil || token != json.Delim('{') {
		return false
	}

	seen := map[string]bool{}
	for decoder.More() {
		token, err = decoder.Token()
		if err != nil {
			return true
		}
		key, ok := token.(string)
		if !ok {
			return true
		}

		for _, k := range keys {
			if strings.EqualFold(key, k) {
				if key != k || seen[k] {
					return true
				}
				seen[k] = true
			}
		}

		var value json.RawMessage
		if err = decoder.Decode(&value); err != nil {
			return true
		}
	}
	return false
}

func (c *ApiController) recordBlockedMcpToolCall(server *object.Server, params mcpself.McpCallToolParams, reason string) {
	if !object.IsMcpToolCallRecorded(server.Owner) {
		return
	}
	object.AddMcpRecord(&object.McpRecord{
		Owner:     server.Owner,
		Server:    server.Name,
		UserName:  c.GetSessionUsername(),
		ClientIp:  util.GetClientIpFromRequest(c.Ctx.Request),
		UserAgent: c.Ctx.Request.UserAgent(),
		Tool:      params.Name,
		Arguments: string(params.Arguments),
		Result:    "Blocked",
		Response:  reason,
	})
}

// GetMcpRecords
// @Title GetMcpRecords
// @Tag Server API
// @Description get the tool calls through the MCP gateway of an organization
// @Param   owner    query string true  "The organization"
// @Param   server   query string false "The server name"
// @Param   tool     query string false "The tool name"
// @Param   p        query string false "The page"
// @Param   pageSize query string false "The page size"
// @Success 200 {array} object.McpRecord The Response object
// @router /get-mcp-records [get]
func (c *ApiController) GetMcpRecords() {
	owner := c.Ctx.Input.Query("owner")
	user := c.getCurrentUser()
	if user == nil || !(c.IsGlobalAdmin() || (user.IsAdmin && user.Owner == owner)) {
		c.ResponseError(c.T("auth:Unauthorized operation"))
		return
	}

	server := c.Ctx.Input.Query("server")
	tool := c.Ctx.Input.Query("tool")
	limit := util.ParseInt(c.Ctx.Input.Query("pageSize"))
	if limit <= 0 || limit > 500 {
		limit = 100
	}

	count, err := object.GetMcpRecordCount(owner, server, tool)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	paginator := pagination.SetPaginator(c.Ctx, limit, count)
	records, err := object.GetPaginationMcpRecords(owner, server, tool, paginator.Offset(), limit)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	c.ResponseOk(records, paginator.Nums())
}
