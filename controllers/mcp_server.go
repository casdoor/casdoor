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
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/beego/beego/v2/server/web/pagination"
	"github.com/casdoor/casdoor/conf"
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
	if err = edition.Require(edition.Commercial, edition.FeatureMcpGateway); err != nil {
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

	caller, err := c.getMcpCaller()
	if err != nil {
		c.McpResponseError(mcpReq.ID, -32603, "failed to identify the caller", err.Error())
		return
	}

	// one trace ID ties the agent, the gateway, the MCP server and the business systems behind it together
	traceId, upstreamTraceparent := object.GetMcpTraceContext(c.Ctx.Request.Header.Get("traceparent"), c.Ctx.Request.Header.Get(mcpTraceIdHeader))
	c.Ctx.Output.Header(mcpTraceIdHeader, traceId)
	call := &mcpCall{server: server, caller: caller, traceId: traceId}

	if mcpReq.Method == "tools/call" {
		var params mcpself.McpCallToolParams
		err = json.Unmarshal(mcpReq.Params, &params)
		if err != nil {
			c.McpResponseError(mcpReq.ID, -32600, "Invalid request", err.Error())
			return
		}
		call.params = params

		if len(server.Tools) > 0 {
			tool := server.GetTool(params.Name)
			if tool == nil {
				c.blockMcpToolCall(mcpReq, call, "tool is not in the allowlist")
				return
			}
			if !tool.IsAllowed {
				c.blockMcpToolCall(mcpReq, call, "tool is forbidden")
				return
			}
		}

		arguments := map[string]interface{}{}
		if len(params.Arguments) > 0 {
			_ = json.Unmarshal(params.Arguments, &arguments)
		}
		reason, err := object.CheckMcpToolPermission(server, caller, params.Name, arguments)
		if err != nil {
			c.McpResponseError(mcpReq.ID, -32603, "failed to check the tool permission", err.Error())
			return
		}
		if reason != "" {
			c.blockMcpToolCall(mcpReq, call, reason)
			return
		}

		if tool := server.GetTool(params.Name); tool != nil && tool.RequireApproval {
			approval, isApproved, err := object.CheckMcpApproval(server, caller, params.Name, string(params.Arguments), traceId, util.GetClientIpFromRequest(c.Ctx.Request))
			if err != nil {
				c.blockMcpToolCall(mcpReq, call, err.Error())
				return
			}
			if !isApproved {
				c.holdMcpToolCall(mcpReq, call, approval)
				return
			}
			call.approval = approval.GetId()
		}
	}

	identityToken := ""
	if server.IdentityHeader != "" {
		identityToken, call.identityTokenId, err = object.GetMcpIdentityToken(server, caller, traceId, call.params.Name, c.Ctx.Request.Host)
		if err != nil {
			c.McpResponseError(mcpReq.ID, -32603, "failed to sign the identity token", err.Error())
			return
		}
	}

	proxy := httputil.NewSingleHostReverseProxy(targetUrl)
	// the MCP servers of the organizations are on the Internet, unless a private deployment lets them be on its intranet
	if server.Owner != "built-in" && !conf.GetConfigBool("mcpGatewayAllowIntranet") {
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

		request.Header.Set("traceparent", upstreamTraceparent)
		request.Header.Set(mcpTraceIdHeader, traceId)
		if identityToken != "" {
			if strings.EqualFold(server.IdentityHeader, "Authorization") {
				request.Header.Set("Authorization", "Bearer "+identityToken)
			} else {
				request.Header.Set(server.IdentityHeader, identityToken)
			}
		}
	}

	// the proxy writes the response itself, Beego would otherwise look for a template afterwards and crash before the
	// audit record of the request is written
	c.EnableRender = false

	if mcpReq.Method != "tools/call" {
		c.setMcpAuditRecord(call, "", "")
		proxy.ServeHTTP(c.Ctx.ResponseWriter, c.Ctx.Request)
		return
	}

	recorder := &mcpResponseRecorder{ResponseWriter: c.Ctx.ResponseWriter.ResponseWriter, statusCode: http.StatusOK, limit: object.GetMcpRecordBodyLimit()}
	startTime := time.Now()
	proxy.ServeHTTP(recorder, c.Ctx.Request)
	c.recordMcpToolCall(call, recorder, time.Since(startTime))
}

const mcpTraceIdHeader = "X-Casdoor-Trace-Id"

// mcpCall is what the gateway knows about a request: the server, who calls and the trace, for the records
type mcpCall struct {
	server          *object.Server
	caller          *object.McpCaller
	params          mcpself.McpCallToolParams
	traceId         string
	identityTokenId string
	approval        string
}

// getMcpCaller tells who calls the gateway from the credential that the AutoSigninFilter has already verified
func (c *ApiController) getMcpCaller() (*object.McpCaller, error) {
	userId := c.GetSessionUsername()

	accessKey := c.Ctx.Request.Header.Get("X-Casdoor-Access-Key")
	accessToken := c.Ctx.Input.Query("accessToken")
	if accessToken == "" {
		accessToken = c.Ctx.Input.Query("access_token")
	}
	if accessToken == "" {
		if scheme, value, ok := strings.Cut(c.Ctx.Request.Header.Get("Authorization"), " "); ok && strings.EqualFold(scheme, "Bearer") {
			accessToken = strings.TrimSpace(value)
		}
	}
	if key, _, ok := strings.Cut(accessToken, ":"); ok {
		accessKey, accessToken = key, ""
	}

	if accessToken != "" {
		token, err := object.GetTokenByAccessToken(accessToken)
		if err != nil {
			return nil, err
		}
		if token != nil {
			return object.NewMcpCallerByToken(userId, token)
		}
	}
	if accessKey != "" {
		key, err := object.GetKeyByAccessKey(accessKey)
		if err != nil {
			return nil, err
		}
		if key != nil {
			return object.NewMcpCallerByKey(userId, key)
		}
	}
	return &object.McpCaller{User: userId, AuthMethod: "Session"}, nil
}

// mcpResponseRecorder keeps the status and the beginning of the upstream response for the tool call record
type mcpResponseRecorder struct {
	http.ResponseWriter
	statusCode int
	body       bytes.Buffer
	limit      int
}

func (r *mcpResponseRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *mcpResponseRecorder) Write(data []byte) (int, error) {
	if r.body.Len() < r.limit {
		r.body.Write(data[:min(len(data), r.limit-r.body.Len())])
	}
	return r.ResponseWriter.Write(data)
}

func (r *mcpResponseRecorder) Flush() {
	if flusher, ok := r.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (c *ApiController) newMcpRecord(call *mcpCall) *object.McpRecord {
	return &object.McpRecord{
		Owner:           call.server.Owner,
		Server:          call.server.Name,
		UserName:        call.caller.User,
		ClientIp:        util.GetClientIpFromRequest(c.Ctx.Request),
		UserAgent:       c.Ctx.Request.UserAgent(),
		Tool:            call.params.Name,
		Arguments:       string(call.params.Arguments),
		Application:     call.caller.Application,
		Agent:           call.caller.GetAgentId(),
		Actor:           call.caller.GetActorJson(),
		AuthMethod:      call.caller.AuthMethod,
		TokenId:         call.caller.TokenId,
		TraceId:         call.traceId,
		IdentityTokenId: call.identityTokenId,
		Approval:        call.approval,
	}
}

// setMcpAuditRecord makes the request's audit record (see routers/record.go) a "mcp-tool-call" that tells who called
// which tool in which trace, so that it is in the tamper-proof chain and sent to the SIEM like the other records
func (c *ApiController) setMcpAuditRecord(call *mcpCall, result string, detail string) {
	c.Ctx.Input.SetParam("recordMcpTraceId", call.traceId)
	if call.params.Name == "" {
		return
	}

	summary := map[string]interface{}{
		"server":      call.server.GetId(),
		"tool":        call.params.Name,
		"arguments":   json.RawMessage(nonEmptyJson(call.params.Arguments)),
		"user":        call.caller.User,
		"application": call.caller.Application,
		"agent":       call.caller.GetAgentId(),
		"actor":       call.caller.Actor,
		"authMethod":  call.caller.AuthMethod,
		"traceId":     call.traceId,
		"result":      result,
		"detail":      detail,
		"approval":    call.approval,
	}
	data, err := json.Marshal(summary)
	if err != nil {
		return
	}

	status := "ok"
	if result != "Success" {
		status = "error"
	}
	c.Ctx.Input.SetParam("recordMcpObject", string(data))
	c.Ctx.Input.SetParam("recordMcpResponse", fmt.Sprintf("{status:\"%s\", msg:\"%s\"}", status, strings.ReplaceAll(result+" "+detail, "\"", "'")))
}

func nonEmptyJson(data json.RawMessage) []byte {
	if len(data) == 0 || !json.Valid(data) {
		return []byte("null")
	}
	return data
}

func (c *ApiController) recordMcpToolCall(call *mcpCall, recorder *mcpResponseRecorder, duration time.Duration) {
	response := recorder.body.String()
	result := "Success"
	if recorder.statusCode >= 400 {
		result = "Failed"
	} else if strings.Contains(response, `"error"`) || strings.Contains(response, `"isError":true`) {
		result = "Error"
	}

	c.setMcpAuditRecord(call, result, "")
	if !object.IsMcpToolCallRecorded(call.server.Owner) {
		return
	}

	record := c.newMcpRecord(call)
	record.Result = result
	record.StatusCode = recorder.statusCode
	record.Response = response
	record.DurationMs = duration.Milliseconds()
	object.AddMcpRecord(record)
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

func (c *ApiController) blockMcpToolCall(mcpReq *mcpself.McpRequest, call *mcpCall, reason string) {
	c.setMcpAuditRecord(call, "Blocked", reason)
	if object.IsMcpToolCallRecorded(call.server.Owner) {
		record := c.newMcpRecord(call)
		record.Result = "Blocked"
		record.Response = reason
		object.AddMcpRecord(record)
	}
	c.McpResponseError(mcpReq.ID, -32600, reason, map[string]string{"traceId": call.traceId})
}

// holdMcpToolCall answers a call that waits for approval with a tool result, so that the agent tells the user to get
// it approved and to call again
func (c *ApiController) holdMcpToolCall(mcpReq *mcpself.McpRequest, call *mcpCall, approval *object.McpApproval) {
	call.approval = approval.GetId()
	approvalUrl := object.GetMcpApprovalUrl(c.Ctx.Request.Host, approval)
	text := fmt.Sprintf("The call of the tool \"%s\" needs approval before it runs. An administrator can approve it at %s, then call the tool again with the same arguments within an hour of the approval. Approval: %s, trace ID: %s",
		call.params.Name, approvalUrl, approval.GetId(), call.traceId)

	c.setMcpAuditRecord(call, "Pending approval", approval.GetId())
	if object.IsMcpToolCallRecorded(call.server.Owner) {
		record := c.newMcpRecord(call)
		record.Result = "Pending approval"
		record.Response = approvalUrl
		object.AddMcpRecord(record)
	}

	result := mcpself.McpCallToolResult{
		Content: []mcpself.TextContent{{Type: "text", Text: text}},
		IsError: true,
	}
	c.Ctx.Output.Header("Content-Type", "application/json")
	c.Data["json"] = mcpself.BuildMcpResponse(mcpReq.ID, result, nil)
	c.ServeJSON()
}

// GetMcpRecords
// @Title GetMcpRecords
// @Tag Server API
// @Description get the tool calls through the MCP gateway of an organization
// @Param   owner    query string true  "The organization"
// @Param   server   query string false "The server name"
// @Param   tool     query string false "The tool name"
// @Param   agent    query string false "The agent ID"
// @Param   user     query string false "The user ID"
// @Param   traceId  query string false "The trace ID"
// @Param   result   query string false "The result"
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

	filter := &object.McpRecordFilter{
		Server:  c.Ctx.Input.Query("server"),
		Tool:    c.Ctx.Input.Query("tool"),
		Agent:   c.Ctx.Input.Query("agent"),
		User:    c.Ctx.Input.Query("user"),
		TraceId: strings.TrimSpace(c.Ctx.Input.Query("traceId")),
		Result:  c.Ctx.Input.Query("result"),
	}
	limit := util.ParseInt(c.Ctx.Input.Query("pageSize"))
	if limit <= 0 || limit > 500 {
		limit = 100
	}

	count, err := object.GetMcpRecordCount(owner, filter)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	paginator := pagination.SetPaginator(c.Ctx, limit, count)
	records, err := object.GetPaginationMcpRecords(owner, filter, paginator.Offset(), limit)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	c.ResponseOk(records, paginator.Nums())
}

func (c *ApiController) requireMcpApprovalAdmin(owner string) (*object.User, bool) {
	user := c.getCurrentUser()
	if user == nil || !(c.IsGlobalAdmin() || (user.IsAdmin && user.Owner == owner)) {
		c.ResponseError(c.T("auth:Unauthorized operation"))
		return nil, false
	}
	if err := edition.Require(edition.Enterprise, edition.FeatureMcpApproval); err != nil {
		c.ResponseError(err.Error())
		return nil, false
	}
	return user, true
}

// GetMcpApprovals
// @Title GetMcpApprovals
// @Tag Server API
// @Description get the tool calls through the MCP gateway that wait for or got an approval
// @Param   owner query string true  "The organization"
// @Param   state query string false "Pending, Approved, Rejected or Used"
// @Success 200 {array} object.McpApproval The Response object
// @router /get-mcp-approvals [get]
func (c *ApiController) GetMcpApprovals() {
	owner := c.Ctx.Input.Query("owner")
	if _, ok := c.requireMcpApprovalAdmin(owner); !ok {
		return
	}

	approvals, err := object.GetMcpApprovals(owner, c.Ctx.Input.Query("state"))
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	c.ResponseOk(approvals)
}

// GetMcpApproval
// @Title GetMcpApproval
// @Tag Server API
// @Description get a tool call that waits for or got an approval
// @Param   id query string true "The ID of the approval"
// @Success 200 {object} object.McpApproval The Response object
// @router /get-mcp-approval [get]
func (c *ApiController) GetMcpApproval() {
	id := c.Ctx.Input.Query("id")
	owner, _ := util.GetOwnerAndNameFromIdNoCheck(id)
	if _, ok := c.requireMcpApprovalAdmin(owner); !ok {
		return
	}

	approval, err := object.GetMcpApproval(id)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	c.ResponseOk(approval)
}

// DecideMcpApproval
// @Title DecideMcpApproval
// @Tag Server API
// @Description approve or reject a tool call that waits for approval
// @Param   id       query string true  "The ID of the approval"
// @Param   approved query string true  "true to approve, false to reject"
// @Param   comment  query string false "The comment"
// @Success 200 {object} controllers.Response The Response object
// @router /decide-mcp-approval [post]
func (c *ApiController) DecideMcpApproval() {
	id := c.Ctx.Input.Query("id")
	owner, _ := util.GetOwnerAndNameFromIdNoCheck(id)
	user, ok := c.requireMcpApprovalAdmin(owner)
	if !ok {
		return
	}

	affected, err := object.DecideMcpApproval(id, user, c.Ctx.Input.Query("approved") == "true", c.Ctx.Input.Query("comment"))
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	c.Data["json"] = wrapActionResponse(affected)
	c.ServeJSON()
}

// GetMyMcpServers
// @Title GetMyMcpServers
// @Tag Server API
// @Description get the MCP servers of the signed-in user's organization with their gateway addresses and allowed tools, for the agents that act for the user
// @Success 200 {array} object.UserMcpServer The Response object
// @router /get-my-mcp-servers [get]
func (c *ApiController) GetMyMcpServers() {
	user := c.getCurrentUser()
	if user == nil || object.IsAppUser(user.GetId()) {
		c.ResponseError(c.T("general:Please login first"))
		return
	}
	if err := edition.Require(edition.Commercial, edition.FeatureMcpGateway); err != nil {
		c.ResponseError(err.Error())
		return
	}

	servers, err := object.GetUserMcpServers(user, c.Ctx.Request.Host)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	c.ResponseOk(servers)
}
