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

package routers

import (
	stdcontext "context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/beego/beego/v2/core/logs"
	"github.com/casdoor/casdoor/controllers"
	"github.com/casdoor/casdoor/object"

	"github.com/beego/beego/v2/server/web/context"
	"github.com/casdoor/casdoor/authz"
	"github.com/casdoor/casdoor/util"
)

type Object struct {
	Owner        string `json:"owner"`
	Name         string `json:"name"`
	AccessKey    string `json:"accessKey"`
	AccessSecret string `json:"accessSecret"`
}

type ObjectWithOrg struct {
	Object
	Organization string `json:"organization"`
}

func getUsername(ctx *context.Context) (username string) {
	username, ok := ctx.Input.Session("username").(string)
	if !ok || username == "" {
		username, _ = getUsernameByClientIdSecret(ctx)
	}

	if username == "" {
		username, _ = getUsernameByKeys(ctx)
	}

	session := ctx.Input.Session("SessionData")
	if session == nil {
		return
	}

	sessionData := &controllers.SessionData{}
	err := util.JsonToStruct(session.(string), sessionData)
	if err != nil {
		logs.Error("GetSessionData failed, error: %s", err)
		return ""
	}

	if sessionData.ExpireTime != 0 &&
		sessionData.ExpireTime < time.Now().Unix() {
		err = ctx.Input.CruSession.Set(stdcontext.Background(), "username", "")
		if err != nil {
			logs.Error("Failed to clear expired session, error: %s", err)
			return ""
		}
		err = ctx.Input.CruSession.Delete(stdcontext.Background(), "SessionData")
		if err != nil {
			logs.Error("Failed to clear expired session, error: %s", err)
		}
		return ""
	}

	return
}

func getSubject(ctx *context.Context) (string, string) {
	username := getUsername(ctx)
	if username == "" {
		return "anonymous", "anonymous"
	}

	// username == "built-in/admin"
	owner, name, err := util.GetOwnerAndNameFromIdWithError(username)
	if err != nil {
		panic(err)
	}
	return owner, name
}

func getObject(ctx *context.Context) (string, string, error) {
	method := ctx.Request.Method
	path := ctx.Request.URL.Path

	// Special handling for MCP requests
	if path == "/api/mcp" && method == http.MethodPost {
		return getMCPObject(ctx)
	}

	if method == http.MethodGet {
		if ctx.Request.URL.Path == "/api/get-policies" {
			if ctx.Input.Query("id") == "/" {
				adapterId := ctx.Input.Query("adapterId")
				if adapterId != "" {
					return util.GetOwnerAndNameFromIdWithError(adapterId)
				}
			} else {
				// query == "?id=built-in/admin"
				id := ctx.Input.Query("id")
				if id != "" {
					return util.GetOwnerAndNameFromIdWithError(id)
				}
			}
		}

		if !(strings.HasPrefix(ctx.Request.URL.Path, "/api/get-") && strings.HasSuffix(ctx.Request.URL.Path, "s")) {
			// query == "?id=built-in/admin"
			id := ctx.Input.Query("id")
			if id != "" {
				return util.GetOwnerAndNameFromIdWithError(id)
			}
		}

		owner := ctx.Input.Query("owner")
		if owner != "" {
			return owner, "", nil
		}

		return "", "", nil
	} else {
		if path == "/api/add-policy" || path == "/api/remove-policy" || path == "/api/update-policy" || path == "/api/send-invitation" {
			id := ctx.Input.Query("id")
			if id != "" {
				return util.GetOwnerAndNameFromIdWithError(id)
			}
		}

		body := ctx.Input.RequestBody
		if len(body) == 0 {
			return ctx.Request.Form.Get("owner"), ctx.Request.Form.Get("name"), nil
		}

		var obj Object

		if strings.HasSuffix(path, "-application") || strings.HasSuffix(path, "-token") ||
			strings.HasSuffix(path, "-syncer") || strings.HasSuffix(path, "-webhook") {
			var objWithOrg ObjectWithOrg
			err := json.Unmarshal(body, &objWithOrg)
			if err != nil {
				return "", "", nil
			}
			return objWithOrg.Organization, objWithOrg.Name, nil
		}

		err := json.Unmarshal(body, &obj)
		if err != nil {
			// this is not error
			return "", "", nil
		}

		if strings.HasSuffix(path, "-organization") {
			return obj.Name, obj.Name, nil
		}

		if path == "/api/delete-resource" {
			tokens := strings.Split(obj.Name, "/")
			if len(tokens) >= 5 {
				obj.Name = tokens[4]
			}
		}

		return obj.Owner, obj.Name, nil
	}
}

func getMCPObject(ctx *context.Context) (string, string, error) {
	body := ctx.Input.RequestBody
	if len(body) == 0 {
		return "", "", nil
	}

	// Parse MCP request to determine tool name
	type MCPRequest struct {
		Method string          `json:"method"`
		Params json.RawMessage `json:"params,omitempty"`
	}

	type MCPCallToolParams struct {
		Name      string                 `json:"name"`
		Arguments map[string]interface{} `json:"arguments,omitempty"`
	}

	var mcpReq MCPRequest
	err := json.Unmarshal(body, &mcpReq)
	if err != nil {
		return "", "", nil
	}

	// Only extract object for tool calls
	if mcpReq.Method != "tools/call" {
		return "", "", nil
	}

	var params MCPCallToolParams
	err = json.Unmarshal(mcpReq.Params, &params)
	if err != nil {
		return "", "", nil
	}

	// Extract owner/id from arguments based on tool
	switch params.Name {
	case "get_applications":
		if owner, ok := params.Arguments["owner"].(string); ok {
			return owner, "", nil
		}
	case "get_application", "update_application":
		if id, ok := params.Arguments["id"].(string); ok {
			return util.GetOwnerAndNameFromIdWithError(id)
		}
	case "add_application", "delete_application":
		if appData, ok := params.Arguments["application"].(map[string]interface{}); ok {
			return extractOwnerNameFromAppData(appData)
		}
	}

	return "", "", nil
}

// extractOwnerNameFromAppData extracts owner and name from application data
// Prioritizes organization field over owner field for consistency
func extractOwnerNameFromAppData(appData map[string]interface{}) (string, string, error) {
	// Try organization field first (used in application APIs)
	if org, ok := appData["organization"].(string); ok {
		if name, ok := appData["name"].(string); ok {
			return org, name, nil
		}
		return org, "", nil
	}
	// Fall back to owner field
	if owner, ok := appData["owner"].(string); ok {
		if name, ok := appData["name"].(string); ok {
			return owner, name, nil
		}
		return owner, "", nil
	}
	return "", "", nil
}

func getMCPUrlPath(ctx *context.Context) string {
	body := ctx.Input.RequestBody
	if len(body) == 0 {
		return "/api/mcp"
	}

	type MCPRequest struct {
		Method string          `json:"method"`
		Params json.RawMessage `json:"params,omitempty"`
	}

	type MCPCallToolParams struct {
		Name string `json:"name"`
	}

	var mcpReq MCPRequest
	err := json.Unmarshal(body, &mcpReq)
	if err != nil {
		return "/api/mcp"
	}

	// Map initialize and tools/list to public endpoints
	// These operations don't require special permissions beyond authentication
	// We use /api/get-application as it's a read-only operation that authenticated users can access
	if mcpReq.Method == "initialize" || mcpReq.Method == "tools/list" {
		return "/api/get-application"
	}

	if mcpReq.Method != "tools/call" {
		return "/api/mcp"
	}

	var params MCPCallToolParams
	err = json.Unmarshal(mcpReq.Params, &params)
	if err != nil {
		return "/api/mcp"
	}

	// Map MCP tool names to corresponding API paths
	switch params.Name {
	case "get_applications":
		return "/api/get-applications"
	case "get_application":
		return "/api/get-application"
	case "add_application":
		return "/api/add-application"
	case "update_application":
		return "/api/update-application"
	case "delete_application":
		return "/api/delete-application"
	default:
		return "/api/mcp"
	}
}

func getKeys(ctx *context.Context) (string, string) {
	method := ctx.Request.Method

	if method == http.MethodGet {
		accessKey := ctx.Input.Query("accessKey")
		accessSecret := ctx.Input.Query("accessSecret")
		return accessKey, accessSecret
	} else {
		body := ctx.Input.RequestBody

		if len(body) == 0 {
			return ctx.Request.Form.Get("accessKey"), ctx.Request.Form.Get("accessSecret")
		}

		var obj Object
		err := json.Unmarshal(body, &obj)
		if err != nil {
			return "", ""
		}

		return obj.AccessKey, obj.AccessSecret
	}
}

func willLog(subOwner string, subName string, method string, urlPath string, objOwner string, objName string) bool {
	if subOwner == "anonymous" && subName == "anonymous" && method == "GET" && (urlPath == "/api/get-account" || urlPath == "/api/get-app-login") && objOwner == "" && objName == "" {
		return false
	}
	return true
}

func getUrlPath(urlPath string, ctx *context.Context) string {
	// Special handling for MCP requests
	if urlPath == "/api/mcp" {
		return getMCPUrlPath(ctx)
	}

	if strings.HasPrefix(urlPath, "/cas") && (strings.HasSuffix(urlPath, "/serviceValidate") || strings.HasSuffix(urlPath, "/proxy") || strings.HasSuffix(urlPath, "/proxyValidate") || strings.HasSuffix(urlPath, "/validate") || strings.HasSuffix(urlPath, "/p3/serviceValidate") || strings.HasSuffix(urlPath, "/p3/proxyValidate") || strings.HasSuffix(urlPath, "/samlValidate")) {
		return "/cas"
	}

	if strings.HasPrefix(urlPath, "/scim") {
		return "/scim"
	}

	if strings.HasPrefix(urlPath, "/api/login/oauth") {
		return "/api/login/oauth"
	}

	if strings.HasPrefix(urlPath, "/api/webauthn") {
		return "/api/webauthn"
	}

	if strings.HasPrefix(urlPath, "/api/saml/redirect") {
		return "/api/saml/redirect"
	}

	return urlPath
}

func ApiFilter(ctx *context.Context) {
	subOwner, subName := getSubject(ctx)
	method := ctx.Request.Method
	urlPath := getUrlPath(ctx.Request.URL.Path, ctx)

	objOwner, objName := "", ""
	if urlPath != "/api/get-app-login" && urlPath != "/api/get-resource" {
		var err error
		objOwner, objName, err = getObject(ctx)
		if err != nil {
			responseError(ctx, err.Error())
			return
		}
	}

	if strings.HasPrefix(urlPath, "/api/notify-payment") {
		urlPath = "/api/notify-payment"
	}

	isAllowed := authz.IsAllowed(subOwner, subName, method, urlPath, objOwner, objName)

	result := "deny"
	if isAllowed {
		result = "allow"
	}

	if willLog(subOwner, subName, method, urlPath, objOwner, objName) {
		logLine := fmt.Sprintf("subOwner = %s, subName = %s, method = %s, urlPath = %s, obj.Owner = %s, obj.Name = %s, result = %s",
			subOwner, subName, method, urlPath, objOwner, objName, result)
		fmt.Println(logLine)
		util.LogInfo(ctx, logLine)
	}

	if !isAllowed {
		denyRequest(ctx)
		record, err := object.NewRecord(ctx)
		if err != nil {
			return
		}

		record.Organization = subOwner
		record.User = subName // auth:Unauthorized operation
		record.Response = fmt.Sprintf("{status:\"error\", msg:\"%s\"}", T(ctx, "auth:Unauthorized operation"))

		util.SafeGoroutine(func() {
			object.AddRecord(record)
		})
	}
}
