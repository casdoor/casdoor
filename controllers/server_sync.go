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
	"strings"

	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/scan"
)

// SyncIntranetServers
// @Title SyncIntranetServers
// @Tag Server API
// @Description scan intranet IP/CIDR targets and detect MCP servers by probing common ports and paths
// @Param   owner   query  string  true  "The provider owner"
// @Param   name    query  string  true  "The provider name"
// @Success 200 {object} controllers.Response The Response object
// @router /sync-intranet-servers [post]
func (c *ApiController) SyncIntranetServers() {
	_, ok := c.RequireAdmin()
	if !ok {
		return
	}

	owner := strings.TrimSpace(c.GetString("owner"))
	name := strings.TrimSpace(c.GetString("name"))
	if owner == "" || name == "" {
		c.ResponseError("provider owner and name are required")
		return
	}

	providerId := fmt.Sprintf("%s/%s", owner, name)
	configuredProvider, err := object.GetProvider(providerId)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	if configuredProvider == nil {
		c.ResponseError("provider does not exist")
		return
	}

	provider, err := scan.GetScanProviderFromProvider(configuredProvider)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	commandBytes, err := json.Marshal(&scan.SyncInnerServersRequest{
		CIDR:  strings.Split(configuredProvider.Scopes, ","),
		Ports: strings.Split(configuredProvider.Content, ","),
		Paths: strings.Split(configuredProvider.Endpoint, ","),
	})
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	rawResult, err := provider.Scan("", string(commandBytes))
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	parsedResult, err := provider.ParseResult(rawResult)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	var result scan.SyncInnerServersResult
	if err := json.Unmarshal([]byte(parsedResult), &result); err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.ResponseOk(&scan.SyncInnerServersResult{
		CIDR:         result.CIDR,
		ScannedHosts: result.ScannedHosts,
		OnlineHosts:  result.OnlineHosts,
		Servers:      result.Servers,
	})
}
