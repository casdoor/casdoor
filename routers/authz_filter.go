// Copyright 2021 The casbin Authors. All Rights Reserved.
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
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/beego/beego/v2/adapter/context"
	"github.com/casdoor/casdoor/authz"
	"github.com/casdoor/casdoor/controllers"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

type Object struct {
	Owner string `json:"owner"`
	Name  string `json:"name"`
}

func getUsernameByClientIdSecret(ctx *context.Context) string {
	requestUri := ctx.Request.RequestURI
	clientId := parseQuery(requestUri, "clientId")
	clientSecret := parseQuery(requestUri, "clientSecret")
	if len(clientId) == 0 || len(clientSecret) == 0 {
		return ""
	}

	app := object.GetApplicationByClientId(clientId)
	if app == nil || app.ClientSecret != clientSecret {
		return ""
	}
	return "built-in/service"
}

func getUsername(ctx *context.Context) (username string) {
	defer func() {
		if r := recover(); r != nil {
			username = getUsernameByClientIdSecret(ctx)
		}
	}()

	// bug in Beego: this call will panic when file session store is empty
	// so we catch the panic
	username = ctx.Input.Session("username").(string)

	if len(username) == 0 {
		username = getUsernameByClientIdSecret(ctx)
	}

	return
}

func getSubject(ctx *context.Context) (string, string) {
	username := getUsername(ctx)
	if username == "" {
		return "anonymous", "anonymous"
	}

	// username == "built-in/admin"
	tokens := strings.Split(username, "/")
	owner := tokens[0]
	name := tokens[1]
	return owner, name
}

func getObject(ctx *context.Context) (string, string) {
	method := ctx.Request.Method
	if method == http.MethodGet {
		query := ctx.Request.URL.RawQuery
		// query == "?id=built-in/admin"
		idParamValue := parseQuery(query, "id")
		if idParamValue == "" {
			return "", ""
		}
		return parseSlash(idParamValue)
	} else {
		body := ctx.Input.RequestBody

		if len(body) == 0 {
			return "", ""
		}

		var obj Object
		err := json.Unmarshal(body, &obj)
		if err != nil {
			//panic(err)
			return "", ""
		}
		return obj.Owner, obj.Name
	}
}

func denyRequest(ctx *context.Context) {
	w := ctx.ResponseWriter
	w.WriteHeader(403)
	resp := &controllers.Response{Status: "error", Msg: "Unauthorized operation"}
	_, err := w.Write([]byte(util.StructToJson(resp)))
	if err != nil {
		panic(err)
	}
}

func willLog(subOwner string, subName string, method string, urlPath string, objOwner string, objName string) bool {
	if subOwner == "anonymous" && subName == "anonymous" && method == "GET" && (urlPath == "/api/get-account" || urlPath == "/api/get-app-login") && objOwner == "" && objName == "" {
		return false
	}
	return true
}

func AuthzFilter(ctx *context.Context) {
	subOwner, subName := getSubject(ctx)
	method := ctx.Request.Method
	urlPath := ctx.Request.URL.Path
	objOwner, objName := getObject(ctx)

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
	}
}
