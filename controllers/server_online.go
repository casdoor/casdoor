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
	"fmt"
	"net/http"
	"time"
)

const onlineServerListUrl = "https://mcp.casdoor.org/registry.json"

// GetOnlineServers
// @Title GetOnlineServers
// @Tag Server API
// @Description get online MCP server list
// @Success 200 {object} controllers.Response The Response object
// @router /get-online-servers [get]
func (c *ApiController) GetOnlineServers() {
	httpClient := &http.Client{Timeout: 10 * time.Second}
	resp, err := httpClient.Get(onlineServerListUrl)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		c.ResponseError(fmt.Sprintf("failed to get online server list, status code: %d", resp.StatusCode))
		return
	}

	var onlineServers interface{}
	err = json.NewDecoder(resp.Body).Decode(&onlineServers)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(onlineServers)
}
