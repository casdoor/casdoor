// Copyright 2026 Casdoor Inc. All Rights Reserved.
//
// This file is part of the commercial and enterprise editions of Casdoor and is licensed under the Casdoor Enterprise
// License, not the Apache License 2.0. You may not use this file except in compliance with that license, see
// ee/LICENSE.

package mcptools

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"

	"github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/context"
	"github.com/casdoor/casdoor/controllers"
	"github.com/casdoor/casdoor/mcpself"
)

// apiResponse is the response of a REST API of Casdoor
type apiResponse struct {
	Status string          `json:"status"`
	Msg    string          `json:"msg"`
	Code   string          `json:"code"`
	Data   json.RawMessage `json:"data"`
	Data2  json.RawMessage `json:"data2"`
}

func (resp *apiResponse) err() error {
	if resp.Status == "ok" {
		return nil
	}
	if resp.Msg == "" {
		return fmt.Errorf("the request failed")
	}
	return fmt.Errorf("%s", resp.Msg)
}

// isAffected tells whether an add, update or delete API changed anything, it answers "Affected" or "Unaffected"
func (resp *apiResponse) isAffected() bool {
	var data string
	return json.Unmarshal(resp.Data, &data) == nil && data == "Affected"
}

// apiAction is a handler of the REST API, e.g. (*controllers.ApiController).GetProvider
type apiAction func(c *controllers.ApiController)

var forwardedHeaders = []string{"Accept-Language", "User-Agent", "X-Forwarded-For", "X-Forwarded-Proto", "X-Forwarded-Host", "X-Real-Ip", "Origin", "Referer"}

// callApi makes a request to the REST API inside the server for the user, after authorizing it as the same request
// from outside would be, so that a tool can do only what the user may do in the console and the API's own checks,
// masking and edition limits apply to it
func callApi(parent *context.Context, userId string, method string, path string, query url.Values, body interface{}, action apiAction) (*apiResponse, error) {
	var bodyBytes []byte
	if body != nil {
		var err error
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return nil, err
		}
	}

	target := path
	if len(query) > 0 {
		target += "?" + query.Encode()
	}
	request, err := http.NewRequestWithContext(parent.Request.Context(), method, target, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	request.Host = parent.Request.Host
	request.RemoteAddr = parent.Request.RemoteAddr
	request.TLS = parent.Request.TLS
	for _, header := range forwardedHeaders {
		if value := parent.Request.Header.Get(header); value != "" {
			request.Header.Set(header, value)
		}
	}
	if bodyBytes != nil {
		request.Header.Set("Content-Type", "application/json")
	}

	recorder := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(recorder, request)
	ctx.Input.RequestBody = bodyBytes
	ctx.Input.CruSession = parent.Input.CruSession
	ctx.Input.SetData("currentUserId", userId)
	// the request carries its own credential, the session's pending password change or MFA doesn't hold it back
	ctx.Input.SetData("requestCredentialUser", userId)
	if impersonating := parent.Input.GetData("impersonating"); impersonating != nil {
		ctx.Input.SetData("impersonating", impersonating)
	}

	if mcpself.AuthorizeRequestHook == nil {
		return nil, fmt.Errorf("the requests of the MCP tools can't be authorized")
	}
	if err = mcpself.AuthorizeRequestHook(ctx); err != nil {
		return &apiResponse{Status: "error", Msg: err.Error()}, nil
	}

	c := &controllers.ApiController{}
	c.Init(ctx, "ApiController", "", c)
	err = runAction(c, action)
	if err != nil {
		return nil, err
	}

	resp := &apiResponse{}
	if err = json.Unmarshal(recorder.Body.Bytes(), resp); err != nil {
		return nil, fmt.Errorf("unexpected response of %s: %s", path, truncate(recorder.Body.String(), 200))
	}
	return resp, nil
}

func runAction(c *controllers.ApiController, action apiAction) (err error) {
	defer func() {
		if r := recover(); r != nil && r != web.ErrAbort {
			err = fmt.Errorf("%v", r)
		}
	}()

	c.Prepare()
	action(c)
	return nil
}

func truncate(text string, limit int) string {
	runes := []rune(text)
	if len(runes) <= limit {
		return text
	}
	return string(runes[:limit]) + "..."
}
