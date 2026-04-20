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
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/beego/beego/v2/server/web/context"
	"github.com/casdoor/casdoor/conf"
	"github.com/casdoor/casdoor/i18n"
	"github.com/casdoor/casdoor/mcpself"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

type Response struct {
	Status string      `json:"status"`
	Msg    string      `json:"msg"`
	Data   interface{} `json:"data"`
	Data2  interface{} `json:"data2"`
}

func responseError(ctx *context.Context, error string, data ...interface{}) {
	// ctx.ResponseWriter.WriteHeader(http.StatusForbidden)
	urlPath := ctx.Request.URL.Path
	if urlPath == "/api/mcp" {
		denyMcpRequest(ctx)
		return
	}

	resp := Response{Status: "error", Msg: error}
	switch len(data) {
	case 2:
		resp.Data2 = data[1]
		fallthrough
	case 1:
		resp.Data = data[0]
	}

	err := ctx.Output.JSON(resp, true, false)
	if err != nil {
		panic(err)
	}
}

func getAcceptLanguage(ctx *context.Context) string {
	language := ctx.Request.Header.Get("Accept-Language")
	return conf.GetLanguage(language)
}

func T(ctx *context.Context, error string) string {
	return i18n.Translate(getAcceptLanguage(ctx), error)
}

func denyRequest(ctx *context.Context) {
	responseError(ctx, T(ctx, "auth:Unauthorized operation"))
}

func denyMcpRequest(ctx *context.Context) {
	req := mcpself.McpRequest{}
	err := json.Unmarshal(ctx.Input.RequestBody, &req)
	if err != nil {
		ctx.Output.SetStatus(http.StatusBadRequest)
		return
	}

	if req.ID == nil {
		ctx.Output.SetStatus(http.StatusAccepted)
		ctx.Output.Body([]byte{})
		return
	}

	resp := mcpself.BuildMcpResponse(req.ID, nil, &mcpself.McpError{
		Code:    -32001,
		Message: "Unauthorized",
		Data:    T(ctx, "auth:Unauthorized operation"),
	})

	// Add WWW-Authenticate header per MCP Authorization spec (RFC 9728)
	// Use the same logic as getOriginFromHost to determine the scheme
	host := ctx.Request.Host
	scheme := "https"
	if !strings.Contains(host, ".") {
		// localhost:8000 or computer-name:80
		scheme = "http"
	}
	resourceMetadataUrl := fmt.Sprintf("%s://%s/.well-known/oauth-protected-resource", scheme, host)
	ctx.Output.Header("WWW-Authenticate", fmt.Sprintf("Bearer realm=\"casdoor\", resource_metadata=\"%s\"", resourceMetadataUrl))

	ctx.Output.SetStatus(http.StatusUnauthorized)
	_ = ctx.Output.JSON(resp, true, false)
}

func getUsernameByClientIdSecret(ctx *context.Context) (string, error) {
	clientId, clientSecret, fromBasicAuth := ctx.Request.BasicAuth()
	if !fromBasicAuth {
		clientId = ctx.Input.Query("clientId")
		clientSecret = ctx.Input.Query("clientSecret")
	}

	if clientId == "" || clientSecret == "" {
		return "", nil
	}

	application, err := object.GetApplicationByClientId(clientId)
	if err != nil {
		return "", err
	}
	if application == nil {
		if fromBasicAuth {
			// The Basic Auth credentials may come from a reverse proxy protecting Casdoor with
			// HTTP Basic Auth. In that case, the username is not an OAuth client ID, so we
			// silently ignore it instead of returning an error that would break the whole system.
			return "", nil
		}
		return "", fmt.Errorf("Application not found for client ID: %s", clientId)
	}

	if application.ClientSecret != clientSecret {
		if fromBasicAuth {
			// Same as above: the secret mismatch may be due to proxy-level Basic Auth credentials.
			return "", nil
		}
		return "", fmt.Errorf("Incorrect client secret for application: %s", application.Name)
	}

	for _, tag := range application.Tags {
		if tag == "dcr" {
			return fmt.Sprintf("app-dcr/%s", application.Name), nil
		}
	}
	return fmt.Sprintf("app/%s", application.Name), nil
}

func getUsernameByAccessKey(ctx *context.Context) (string, error) {
	accessKey := ctx.Input.Query("accessKey")
	accessSecret := ctx.Input.Query("accessSecret")

	if accessKey == "" || accessSecret == "" {
		return "", nil
	}

	key, err := object.GetKeyByAccessKey(accessKey)
	if err != nil {
		return "", err
	}
	if key == nil {
		return "", fmt.Errorf("Access key not found: %s", accessKey)
	}

	if key.AccessSecret != accessSecret {
		return "", fmt.Errorf("Incorrect access secret for key: %s", key.Name)
	}

	if key.State != "Active" {
		return "", fmt.Errorf("Access key is not active: %s", key.Name)
	}

	if key.ExpireTime != "" {
		expireTime, err := time.Parse(time.RFC3339, key.ExpireTime)
		if err != nil {
			return "", fmt.Errorf("Invalid expire time format for key: %s", key.Name)
		}
		if time.Now().After(expireTime) {
			return "", fmt.Errorf("Access key has expired, expireTime = %s", key.ExpireTime)
		}
	}

	if key.User != "" {
		return util.GetId(key.Organization, key.User), nil
	}

	if key.Application != "" {
		return fmt.Sprintf("app/%s", key.Application), nil
	}

	return "", nil
}

func getSessionUser(ctx *context.Context) string {
	user := ctx.Input.CruSession.Get(stdcontext.Background(), "username")
	if user == nil {
		return ""
	}

	return user.(string)
}

func setSessionUser(ctx *context.Context, user string) {
	err := ctx.Input.CruSession.Set(stdcontext.Background(), "username", user)
	if err != nil {
		panic(err)
	}

	// https://github.com/beego/beego/issues/3445#issuecomment-455411915
	ctx.Input.CruSession.SessionRelease(stdcontext.Background(), ctx.ResponseWriter)
}

func setSessionExpire(ctx *context.Context, ExpireTime int64) {
	SessionData := struct{ ExpireTime int64 }{ExpireTime: ExpireTime}
	err := ctx.Input.CruSession.Set(stdcontext.Background(), "SessionData", util.StructToJson(SessionData))
	if err != nil {
		panic(err)
	}
	ctx.Input.CruSession.SessionRelease(stdcontext.Background(), ctx.ResponseWriter)
}

func setSessionOidc(ctx *context.Context, scope string, aud string) {
	err := ctx.Input.CruSession.Set(stdcontext.Background(), "scope", scope)
	if err != nil {
		panic(err)
	}
	err = ctx.Input.CruSession.Set(stdcontext.Background(), "aud", aud)
	if err != nil {
		panic(err)
	}
	ctx.Input.CruSession.SessionRelease(stdcontext.Background(), ctx.ResponseWriter)
}

func parseBearerToken(ctx *context.Context) string {
	header := ctx.Request.Header.Get("Authorization")
	tokens := strings.Split(header, " ")
	if len(tokens) != 2 {
		return ""
	}

	// Accept both "Bearer" (RFC 6750) and "DPoP" (RFC 9449) authorization schemes.
	prefix := tokens[0]
	if prefix != "Bearer" && prefix != "DPoP" {
		return ""
	}

	return tokens[1]
}

func getHostname(s string) string {
	if s == "" {
		return ""
	}

	l, err := url.Parse(s)
	if err != nil {
		panic(err)
	}

	res := l.Hostname()
	return res
}

func removePort(s string) string {
	ipStr, _, err := net.SplitHostPort(s)
	if err != nil {
		ipStr = s
	}
	return ipStr
}
