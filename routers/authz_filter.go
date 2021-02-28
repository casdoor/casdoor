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

	"github.com/astaxie/beego/context"
	"github.com/casdoor/casdoor/authz"
)

type Object struct {
	Owner string `json:"owner"`
	Name  string `json:"name"`
}

func getUserId(ctx *context.Context) string {
	username := ctx.Input.Session("username")
	userId := strings.TrimLeft(username.(string), "/")
	if userId == "" {
		userId = "anonymous"
	}
	return userId
}

func getObject(ctx *context.Context) (string, string) {
	method := ctx.Request.Method
	if method == http.MethodGet {
		// query = "id=built-in/admin"
		query := ctx.Request.URL.RawQuery
		if query == "" {
			return "", ""
		}

		query = strings.TrimLeft(query, "id=")
		tokens := strings.Split(query, "/")
		owner := tokens[0]
		name := tokens[1]
		return owner, name
	} else {
		body := ctx.Input.RequestBody

		if len(body) == 0 {
			return "", ""
		}

		var obj Object
		err := json.Unmarshal(body, &obj)
		if err != nil {
			panic(err)
		}
		return obj.Owner, obj.Name
	}
}

func denyRequest(ctx *context.Context) {
	w := ctx.ResponseWriter
	w.WriteHeader(403)
	_, err := w.Write([]byte("403 Forbidden\n"))
	if err != nil {
		panic(err)
	}
}

func AuthzFilter(ctx *context.Context) {
	userId := getUserId(ctx)
	method := ctx.Request.Method
	urlPath := ctx.Request.URL.Path
	objOwner, objName := getObject(ctx)

	isAllowed := authz.IsAllowed(userId, method, urlPath, objOwner, objName)

	fmt.Printf("userId = %s, method = %s, urlPath = %s, obj.Owner = %s, obj.Name = %s, isAllowed = %v\n",
		userId, method, urlPath, objOwner, objName, isAllowed)
	//if !isAllowed {
	//	denyRequest(ctx)
	//}
}
