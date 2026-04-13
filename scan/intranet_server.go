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

package scan

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/casdoor/casdoor/mcp"
)

const (
	defaultIntranetSyncTimeoutMs      = 1200
	defaultIntranetSyncMaxConcurrency = 32
	maxIntranetSyncHosts              = 1024

	ScanProviderCategory        = "Scan"
	McpScanProviderType         = "MCP Scan"
	IntranetScanProviderSubType = "Intranet Scan"
)

var (
	defaultIntranetSyncPorts = []int{3000, 8080, 80}
	defaultIntranetSyncPaths = []string{"/", "/mcp", "/sse", "/mcp/sse"}
)

type SyncInnerServersRequest struct {
	CIDR  []string `json:"scopes"`
	Ports []string `json:"content"`
	Paths []string `json:"endpoint"`
}

type SyncIntranetServersRequest struct {
	Provider string `json:"provider"`
}

type SyncInnerServersResult struct {
	CIDR         []string              `json:"cidr"`
	ScannedHosts int                   `json:"scannedHosts"`
	OnlineHosts  []string              `json:"onlineHosts"`
	Servers      []*mcp.InnerMcpServer `json:"servers"`
}

type IntranetServerProvider struct{}

func NewIntranetServerProvider() *IntranetServerProvider {
	return &IntranetServerProvider{}
}

func (p *IntranetServerProvider) Scan(target string, command string) (string, error) {
	req, err := parseIntranetScanRequest(target, command)
	if err != nil {
		return "", err
	}

	hosts, err := mcp.ParseScanTargets(req.CIDR, maxIntranetSyncHosts)
	if err != nil {
		return "", err
	}

	timeout := mcp.SanitizeTimeout(0, defaultIntranetSyncTimeoutMs, 10000)
	concurrency := mcp.SanitizeConcurrency(0, defaultIntranetSyncMaxConcurrency, 256)
	ports := mcp.SanitizePorts(req.Ports, defaultIntranetSyncPorts)
	paths := mcp.SanitizePaths(req.Paths, defaultIntranetSyncPaths)
	scheme := mcp.SanitizeScheme("")

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

	result := SyncInnerServersResult{
		CIDR:         req.CIDR,
		ScannedHosts: len(hosts),
		OnlineHosts:  onlineHosts,
		Servers:      servers,
	}

	resultBytes, err := json.Marshal(result)
	if err != nil {
		return "", err
	}

	return string(resultBytes), nil
}

func (p *IntranetServerProvider) ParseResult(rawResult string) (string, error) {
	var result SyncInnerServersResult
	if err := json.Unmarshal([]byte(rawResult), &result); err != nil {
		return "", err
	}

	resultBytes, err := json.Marshal(result)
	if err != nil {
		return "", err
	}

	return string(resultBytes), nil
}

func (p *IntranetServerProvider) GetResultSummary(result string) string {
	var parsed SyncInnerServersResult
	if err := json.Unmarshal([]byte(result), &parsed); err != nil {
		return fmt.Sprintf("invalid result: %v", err)
	}

	return fmt.Sprintf("scannedHosts=%d onlineHosts=%d servers=%d", parsed.ScannedHosts, len(parsed.OnlineHosts), len(parsed.Servers))
}

func parseIntranetScanRequest(target string, command string) (*SyncInnerServersRequest, error) {
	req := &SyncInnerServersRequest{}
	if strings.TrimSpace(command) != "" {
		var payload map[string]json.RawMessage
		if err := json.Unmarshal([]byte(command), &payload); err != nil {
			return nil, err
		}

		req.CIDR = append(req.CIDR, parseListField(payload, "scopes", "cidr")...)
		req.Ports = append(req.Ports, parseListField(payload, "content", "ports")...)
		req.Paths = append(req.Paths, parseListField(payload, "endpoint", "paths")...)
	}

	if strings.TrimSpace(target) != "" {
		req.CIDR = append(req.CIDR, target)
	}

	for i := range req.CIDR {
		req.CIDR[i] = strings.TrimSpace(req.CIDR[i])
	}

	trimmedCIDR := make([]string, 0, len(req.CIDR))
	for _, item := range req.CIDR {
		if item == "" {
			continue
		}
		trimmedCIDR = append(trimmedCIDR, item)
	}
	req.CIDR = trimmedCIDR

	if len(req.CIDR) == 0 {
		return nil, fmt.Errorf("scan target (CIDR/IP) is required")
	}

	return req, nil
}

func splitAndTrim(raw string) []string {
	parts := strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == '\n' || r == '\r' || r == '\t'
	})

	res := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		res = append(res, trimmed)
	}

	return res
}

func parseListField(payload map[string]json.RawMessage, keys ...string) []string {
	for _, key := range keys {
		raw, ok := payload[key]
		if !ok || len(raw) == 0 {
			continue
		}

		var list []string
		if err := json.Unmarshal(raw, &list); err == nil {
			for i := range list {
				list[i] = strings.TrimSpace(list[i])
			}
			return list
		}

		var single string
		if err := json.Unmarshal(raw, &single); err == nil {
			return splitAndTrim(single)
		}
	}

	return nil
}
