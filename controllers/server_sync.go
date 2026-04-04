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
	"context"
	"encoding/json"
	"net/http"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/casdoor/casdoor/mcp"
)

const (
	defaultSyncTimeoutMs      = 1200
	defaultSyncMaxConcurrency = 32
	maxSyncHosts              = 1024
)

var (
	defaultSyncPorts = []int{3000, 8080, 80}
	defaultSyncPaths = []string{"/", "/mcp", "/sse", "/mcp/sse"}
)

type SyncInnerServersRequest struct {
	CIDR           []string `json:"cidr"`
	Scheme         string   `json:"scheme"`
	Ports          []string `json:"ports"`
	Paths          []string `json:"paths"`
	TimeoutMs      int      `json:"timeoutMs"`
	MaxConcurrency int      `json:"maxConcurrency"`
}

type SyncInnerServersResult struct {
	CIDR         []string              `json:"cidr"`
	ScannedHosts int                   `json:"scannedHosts"`
	OnlineHosts  []string              `json:"onlineHosts"`
	Servers      []*mcp.InnerMcpServer `json:"servers"`
}

// SyncIntranetServers
// @Title SyncIntranetServers
// @Tag Server API
// @Description scan intranet IP/CIDR targets and detect MCP servers by probing common ports and paths
// @Param   body    body   controllers.SyncInnerServersRequest  true  "Intranet MCP server scan request"
// @Success 200 {object} controllers.Response The Response object
// @router /sync-intranet-servers [post]
func (c *ApiController) SyncIntranetServers() {
	_, ok := c.RequireAdmin()
	if !ok {
		return
	}

	var req SyncInnerServersRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		c.ResponseError(err.Error())
		return
	}

	for i := range req.CIDR {
		req.CIDR[i] = strings.TrimSpace(req.CIDR[i])
	}
	if len(req.CIDR) == 0 {
		c.ResponseError("scan target (CIDR/IP) is required")
		return
	}

	hosts, err := mcp.ParseScanTargets(req.CIDR, maxSyncHosts)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	timeout := mcp.SanitizeTimeout(req.TimeoutMs, defaultSyncTimeoutMs, 10000)
	concurrency := mcp.SanitizeConcurrency(req.MaxConcurrency, defaultSyncMaxConcurrency, 256)
	ports := mcp.SanitizePorts(req.Ports, defaultSyncPorts)
	paths := mcp.SanitizePaths(req.Paths, defaultSyncPaths)
	scheme := mcp.SanitizeScheme(req.Scheme)

	client := &http.Client{
		Timeout: timeout,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	onlineHostSet := map[string]struct{}{}
	serverMap := map[string]*mcp.InnerMcpServer{}
	mutex := sync.Mutex{}
	waitGroup := sync.WaitGroup{}
	sem := make(chan struct{}, concurrency)

	for _, host := range hosts {
		host := host.String()
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()

			select {
			case sem <- struct{}{}:
			case <-ctx.Done():
				return
			}
			defer func() { <-sem }()

			isOnline, servers := mcp.ProbeHost(ctx, client, scheme, host, ports, paths, timeout)
			if !isOnline {
				return
			}

			mutex.Lock()
			onlineHostSet[host] = struct{}{}
			for _, server := range servers {
				serverMap[server.Url] = server
			}
			mutex.Unlock()
		}()
	}

	waitGroup.Wait()

	onlineHosts := make([]string, 0, len(onlineHostSet))
	for host := range onlineHostSet {
		onlineHosts = append(onlineHosts, host)
	}
	slices.Sort(onlineHosts)

	servers := make([]*mcp.InnerMcpServer, 0, len(serverMap))
	for _, server := range serverMap {
		servers = append(servers, server)
	}
	slices.SortFunc(servers, func(a, b *mcp.InnerMcpServer) int {
		if a.Url < b.Url {
			return -1
		}
		if a.Url > b.Url {
			return 1
		}
		return 0
	})

	c.ResponseOk(&SyncInnerServersResult{
		CIDR:         req.CIDR,
		ScannedHosts: len(hosts),
		OnlineHosts:  onlineHosts,
		Servers:      servers,
	})
}

func (c *ApiController) SyncInnerServers() {
	c.SyncIntranetServers()
}
