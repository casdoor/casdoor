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

	"github.com/beego/beego/v2/server/web/context"
	"github.com/casdoor/casdoor/conf"
	"github.com/casdoor/casdoor/i18n"
	"github.com/casdoor/casdoor/mcp"
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
	req := mcp.McpRequest{}
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

	resp := mcp.BuildMcpResponse(req.ID, nil, &mcp.McpError{
		Code:    -32001,
		Message: "Unauthorized",
		Data:    T(ctx, "auth:Unauthorized operation"),
	})

	ctx.Output.SetStatus(http.StatusUnauthorized)
	_ = ctx.Output.JSON(resp, true, false)
}

// getUsernameByClientCert authenticates a client using mTLS (RFC 8705)
// It extracts the client certificate from the TLS connection and validates it
// against the certificate stored in the application configuration
func getUsernameByClientCert(ctx *context.Context, clientId string) (string, error) {
	// Check if TLS is being used and if peer certificates are available
	if ctx.Request.TLS == nil || len(ctx.Request.TLS.PeerCertificates) == 0 {
		return "", nil
	}

	// Get the client certificate (first in the chain)
	clientCert := ctx.Request.TLS.PeerCertificates[0]

	// Get the application by clientId
	application, err := object.GetApplicationByClientId(clientId)
	if err != nil {
		return "", err
	}
	if application == nil {
		return "", nil
	}

	// Check if mTLS is enabled for this application
	if !application.EnableClientCert {
		return "", nil
	}

	// Check if a client certificate is configured
	if application.ClientCert == "" {
		return "", fmt.Errorf("mTLS is enabled but no client certificate is configured for application: %s", application.Name)
	}

	// Get the stored certificate configuration
	certConfig, err := object.GetCert(util.GetId(application.Owner, application.ClientCert))
	if err != nil {
		return "", fmt.Errorf("failed to get client certificate for application %s: %w", application.Name, err)
	}
	if certConfig == nil {
		return "", fmt.Errorf("client certificate not found: %s", application.ClientCert)
	}

	// Parse the stored certificate
	storedCert, err := util.ParseCertificate(certConfig.Certificate)
	if err != nil {
		return "", fmt.Errorf("failed to parse stored certificate for application %s: %w", application.Name, err)
	}

	// Validate the client certificate
	err = util.ValidateCertificate(clientCert, storedCert)
	if err != nil {
		return "", fmt.Errorf("certificate validation failed for application %s: %w", application.Name, err)
	}

	// Authentication successful
	return fmt.Sprintf("app/%s", application.Name), nil
}

func getUsernameByClientIdSecret(ctx *context.Context) (string, error) {
	clientId, clientSecret, ok := ctx.Request.BasicAuth()
	if !ok {
		clientId = ctx.Input.Query("clientId")
		clientSecret = ctx.Input.Query("clientSecret")
	}

	if clientId == "" {
		return "", nil
	}

	// Try certificate-based authentication first (RFC 8705 - tls_client_auth)
	if ctx.Request.TLS != nil && len(ctx.Request.TLS.PeerCertificates) > 0 {
		username, err := getUsernameByClientCert(ctx, clientId)
		if err != nil {
			return "", err
		}
		if username != "" {
			return username, nil
		}
	}

	// Fall back to client secret authentication
	if clientSecret == "" {
		return "", nil
	}

	application, err := object.GetApplicationByClientId(clientId)
	if err != nil {
		return "", err
	}
	if application == nil {
		return "", fmt.Errorf("Application not found for client ID: %s", clientId)
	}

	if application.ClientSecret != clientSecret {
		return "", fmt.Errorf("Incorrect client secret for application: %s", application.Name)
	}

	return fmt.Sprintf("app/%s", application.Name), nil
}

func getUsernameByKeys(ctx *context.Context) (string, error) {
	accessKey, accessSecret := getKeys(ctx)
	user, err := object.GetUserByAccessKey(accessKey)
	if err != nil {
		return "", err
	}

	if user == nil {
		return "", fmt.Errorf("user not found for access key: %s", accessKey)
	}

	if accessSecret != user.AccessSecret {
		return "", fmt.Errorf("incorrect access secret for user: %s", user.Name)
	}

	return user.GetId(), nil
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

	prefix := tokens[0]
	if prefix != "Bearer" {
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
