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

var orgOwnerObject = []string{
	"-organization",
	"-syncer",
	"-webhook",
	"-application",
	"-token",
}

type Object struct {
	Owner string `json:"owner"`
	Name  string `json:"name"`
}

type ObjectWithOrg struct {
	Object
	Organization string `json:"organization"`
}

// ownerNameFromForm parses form or multipart body for authorization checks when the
// request is not JSON (e.g. MFA APIs use FormData). RequestBodyFilter caches the raw
// body but leaves Request.Body restorable for ParseForm/ParseMultipartForm.
func ownerNameFromForm(ctx *context.Context) (string, string) {
	ct := ctx.Request.Header.Get("Content-Type")
	if strings.Contains(ct, "multipart/form-data") {
		_ = ctx.Request.ParseMultipartForm(32 << 20)
	} else {
		_ = ctx.Request.ParseForm()
	}
	return ctx.Request.Form.Get("owner"), ctx.Request.Form.Get("name")
}

func checkIsOrgOwnerObject(urlPath string) bool {
	for _, suffix := range orgOwnerObject {
		if strings.HasSuffix(urlPath, suffix) || strings.Contains(urlPath, suffix+"s") {
			return true
		}
	}
	return false
}

func getUsername(ctx *context.Context) (username string) {
	username, ok := ctx.Input.Session("username").(string)
	if !ok || username == "" {
		username, _ = getUsernameByClientIdSecret(ctx)
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
		return getMcpObject(ctx)
	}

	if strings.HasPrefix(path, "/api/server/") {
		return ctx.Input.Param(":owner"), ctx.Input.Param(":name"), nil
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

		organization := ctx.Input.Query("organization")

		if !(strings.HasPrefix(ctx.Request.URL.Path, "/api/get-") && strings.HasSuffix(ctx.Request.URL.Path, "s")) {
			// query == "?id=built-in/admin"
			id := ctx.Input.Query("id")
			if id != "" {
				owner, name, err := util.GetOwnerAndNameFromIdWithError(id)
				if err != nil {
					return owner, name, err
				}
				if organization != "" {
					return organization, name, nil
				}

				if strings.HasSuffix(ctx.Request.URL.Path, "organization") {
					return name, name, nil
				}
				return owner, name, nil
			}
		}

		owner := ctx.Input.Query("owner")
		if organization != "" {
			return organization, "", nil
		}
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

		isOwnerObjPath := checkIsOrgOwnerObject(path)

		// For non-GET requests, if the `id` query param is present it is the
		// authoritative identifier of the object being operated on.  Use it
		// instead of the request body so that an attacker cannot spoof the
		// object owner by injecting "owner":"admin" (or any other value) into
		// the request body while pointing the URL at a different organization's
		// resource.
		if id := ctx.Input.Query("id"); id != "" && (!isOwnerObjPath || strings.HasSuffix(path, "update-organization")) {
			owner, name, err := util.GetOwnerAndNameFromIdWithError(id)
			if err == nil {
				return owner, name, nil
			}
		}

		body := ctx.Input.RequestBody
		if len(body) == 0 {
			return ctx.Request.Form.Get("owner"), ctx.Request.Form.Get("name"), nil
		}

		var obj Object

		if isOwnerObjPath && !strings.HasSuffix(path, "-organization") {
			var objWithOrg ObjectWithOrg
			err := json.Unmarshal(body, &objWithOrg)
			if err != nil {
				o, n := ownerNameFromForm(ctx)
				return o, n, nil
			}
			return objWithOrg.Organization, objWithOrg.Name, nil
		}

		err := json.Unmarshal(body, &obj)
		if err != nil {
			// Form-urlencoded, multipart, or other non-JSON body (common for web FormData).
			o, n := ownerNameFromForm(ctx)
			return o, n, nil
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

func willLog(subOwner string, subName string, method string, urlPath string, objOwner string, objName string) bool {
	if subOwner == "anonymous" && subName == "anonymous" && method == "GET" && (urlPath == "/api/get-account" || urlPath == "/api/get-app-login") && objOwner == "" && objName == "" {
		return false
	}
	return true
}

func getUrlPath(ctx *context.Context) string {
	urlPath := ctx.Request.URL.Path

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

func getExtraInfo(ctx *context.Context, urlPath string) map[string]interface{} {
	var extra map[string]interface{}
	if urlPath == "/api/mcp" {
		var m map[string]interface{}
		if err := json.Unmarshal(ctx.Input.RequestBody, &m); err != nil {
			return nil
		}

		method, ok := m["method"].(string)
		if !ok {
			return nil
		}

		return map[string]interface{}{
			"detailPathUrl": method,
		}
	}
	return extra
}

func getImpersonateUser(ctx *context.Context, subOwner, subName, username string) (string, string, string) {
	impersonateUser, ok := ctx.Input.Session("impersonateUser").(string)
	impersonateUserCookie := ctx.GetCookie("impersonateUser")
	if ok && impersonateUser != "" && impersonateUserCookie != "" {
		user, err := object.GetUser(util.GetId(subOwner, subName))
		if err != nil {
			panic(err)
		}

		if user != nil {
			impUserOwner, impUserName, err := util.GetOwnerAndNameFromIdWithError(impersonateUser)
			if err != nil {
				panic(err)
			}

			if user.IsAdmin && impUserOwner == user.Owner {
				ctx.Input.SetData("impersonating", true)
				return impUserOwner, impUserName, impersonateUser
			}
		}
	}

	return subOwner, subName, username
}

func ApiFilter(ctx *context.Context) {
	subOwner, subName := getSubject(ctx)
	// stash current user info into request context for controllers
	username := ""
	if !(subOwner == "anonymous" && subName == "anonymous") {
		username = fmt.Sprintf("%s/%s", subOwner, subName)
		subOwner, subName, username = getImpersonateUser(ctx, subOwner, subName, username)
	}
	ctx.Input.SetData("currentUserId", username)

	method := ctx.Request.Method
	urlPath := getUrlPath(ctx)
	extraInfo := getExtraInfo(ctx, urlPath)

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

	isAllowed, err := authz.IsAllowed(subOwner, subName, method, urlPath, objOwner, objName, extraInfo)
	if err != nil {
		responseError(ctx, err.Error())
		return
	}

	if method != "GET" && !strings.HasSuffix(urlPath, "-entry") {
		util.SafeGoroutine(func() {
			writePermissionLog(objOwner, subOwner, subName, method, urlPath, isAllowed)
		})
	}

	result := "deny"
	if isAllowed {
		result = "allow"
	}

	if willLog(subOwner, subName, method, urlPath, objOwner, objName) {
		logLine := fmt.Sprintf("subOwner = %s, subName = %s, method = %s, urlPath = %s, obj.Owner = %s, obj.Name = %s, result = %s",
			subOwner, subName, method, urlPath, objOwner, objName, result)
		extra := formatExtraInfo(extraInfo)
		if extra != "" {
			logLine += fmt.Sprintf(", extraInfo = %s", extra)
		}
		fmt.Println(logLine)
		util.LogInfo(ctx, logLine)
	}

	if !isAllowed {
		if urlPath == "/api/mcp" || strings.HasPrefix(urlPath, "/api/server/") {
			denyMcpRequest(ctx)
		} else {
			denyRequest(ctx)
		}
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

func writePermissionLog(objOwner, subOwner, subName, method, urlPath string, allowed bool) {
	providers, err := object.GetProvidersByCategory(objOwner, "Log")
	if err != nil {
		return
	}

	severity := "info"
	if !allowed {
		severity = "warning"
	}
	message := fmt.Sprintf("sub=%s/%s method=%s url=%s objOwner=%s allowed=%v", subOwner, subName, method, urlPath, objOwner, allowed)

	for _, provider := range providers {
		// System Log is a pull-based collector; it does not accept Write calls.
		if provider.Type == "System Log" {
			continue
		}
		if provider.State == "Disabled" {
			continue
		}
		logProvider, err := object.GetLogProviderFromProvider(provider)
		if err != nil {
			continue
		}
		_ = logProvider.Write(severity, message)
	}
}

func formatExtraInfo(extra map[string]interface{}) string {
	if extra == nil {
		return ""
	}
	b, err := json.Marshal(extra)
	if err != nil {
		return ""
	}
	return string(b)
}
