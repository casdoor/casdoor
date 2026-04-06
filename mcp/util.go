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

package mcp

import (
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/casdoor/casdoor/util"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"golang.org/x/oauth2"
)

type InnerMcpServer struct {
	Host string `json:"host"`
	Port int    `json:"port"`
	Path string `json:"path"`
	Url  string `json:"url"`
}

func GetServerTools(owner, name, url, token string) ([]*mcpsdk.Tool, error) {
	var session *mcpsdk.ClientSession
	var err error

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*10)
	defer cancel()
	client := mcpsdk.NewClient(&mcpsdk.Implementation{Name: util.GetId(owner, name), Version: "1.0.0"}, nil)

	if strings.HasSuffix(url, "sse") {
		if token != "" {
			httpClient := oauth2.NewClient(ctx, oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token}))
			session, err = client.Connect(ctx, &mcpsdk.StreamableClientTransport{Endpoint: url, HTTPClient: httpClient}, nil)
		} else {
			session, err = client.Connect(ctx, &mcpsdk.StreamableClientTransport{Endpoint: url}, nil)
		}
	} else {
		if token != "" {
			httpClient := oauth2.NewClient(ctx, oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token}))
			session, err = client.Connect(ctx, &mcpsdk.StreamableClientTransport{Endpoint: url, HTTPClient: httpClient}, nil)
		} else {
			session, err = client.Connect(ctx, &mcpsdk.StreamableClientTransport{Endpoint: url}, nil)
		}
	}

	if err != nil {
		return nil, err
	}
	defer session.Close()

	toolResult, err := session.ListTools(ctx, nil)
	if err != nil {
		return nil, err
	}

	return toolResult.Tools, nil
}

func SanitizeScheme(scheme string) string {
	scheme = strings.ToLower(strings.TrimSpace(scheme))
	if scheme == "https" {
		return "https"
	}
	return "http"
}

func SanitizeTimeout(timeoutMs int, defaultTimeoutMs int, maxTimeoutMs int) time.Duration {
	if timeoutMs <= 0 {
		timeoutMs = defaultTimeoutMs
	}
	if timeoutMs > maxTimeoutMs {
		timeoutMs = maxTimeoutMs
	}
	return time.Duration(timeoutMs) * time.Millisecond
}

func SanitizeConcurrency(maxConcurrency int, defaultConcurrency int, maxAllowed int) int {
	if maxConcurrency <= 0 {
		maxConcurrency = defaultConcurrency
	}
	if maxConcurrency > maxAllowed {
		maxConcurrency = maxAllowed
	}
	return maxConcurrency
}

func SanitizePorts(portInputs []string, defaultPorts []int) []int {
	if len(portInputs) == 0 {
		return append([]int{}, defaultPorts...)
	}

	portSet := map[int]struct{}{}
	result := make([]int, 0, len(portInputs))
	for _, portInput := range portInputs {
		portInput = strings.TrimSpace(portInput)
		if portInput == "" {
			continue
		}

		if strings.Contains(portInput, "-") {
			parts := strings.SplitN(portInput, "-", 2)
			if len(parts) != 2 {
				continue
			}

			start, err := strconv.Atoi(strings.TrimSpace(parts[0]))
			if err != nil {
				continue
			}
			end, err := strconv.Atoi(strings.TrimSpace(parts[1]))
			if err != nil {
				continue
			}
			if start > end {
				continue
			}

			if start < 1 {
				start = 1
			}
			if end > 65535 {
				end = 65535
			}
			if start > end {
				continue
			}

			for port := start; port <= end; port++ {
				if _, ok := portSet[port]; ok {
					continue
				}
				portSet[port] = struct{}{}
				result = append(result, port)
			}
			continue
		}

		port, err := strconv.Atoi(portInput)
		if err != nil {
			continue
		}
		if port <= 0 || port > 65535 {
			continue
		}
		if _, ok := portSet[port]; ok {
			continue
		}
		portSet[port] = struct{}{}
		result = append(result, port)
	}
	if len(result) == 0 {
		return append([]int{}, defaultPorts...)
	}
	return result
}

func SanitizePaths(paths []string, defaultPaths []string) []string {
	if len(paths) == 0 {
		return append([]string{}, defaultPaths...)
	}

	pathSet := map[string]struct{}{}
	result := make([]string, 0, len(paths))
	for _, path := range paths {
		path = strings.TrimSpace(path)
		if path == "" {
			continue
		}
		if !strings.HasPrefix(path, "/") {
			path = "/" + path
		}
		if _, ok := pathSet[path]; ok {
			continue
		}
		pathSet[path] = struct{}{}
		result = append(result, path)
	}
	if len(result) == 0 {
		return append([]string{}, defaultPaths...)
	}
	return result
}

func ParseScanTargets(targets []string, maxHosts int) ([]net.IP, error) {
	hostSet := map[uint32]struct{}{}
	hosts := make([]net.IP, 0)

	addHost := func(ipv4 net.IP) error {
		value := binary.BigEndian.Uint32(ipv4)
		if _, ok := hostSet[value]; ok {
			return nil
		}
		if len(hosts) >= maxHosts {
			return fmt.Errorf("scan targets exceed max %d hosts", maxHosts)
		}
		hostSet[value] = struct{}{}
		host := make(net.IP, net.IPv4len)
		copy(host, ipv4)
		hosts = append(hosts, host)
		return nil
	}

	for _, target := range targets {
		target = strings.TrimSpace(target)
		if target == "" {
			continue
		}

		if ip := net.ParseIP(target); ip != nil {
			ipv4 := ip.To4()
			if ipv4 == nil {
				return nil, fmt.Errorf("only IPv4 is supported: %s", target)
			}
			if !util.IsIntranetIp(ipv4.String()) {
				return nil, fmt.Errorf("target must be intranet: %s", target)
			}
			if err := addHost(ipv4); err != nil {
				return nil, err
			}
			continue
		}

		cidrHosts, err := ParseCIDRHosts(target, maxHosts)
		if err != nil {
			return nil, err
		}
		for _, host := range cidrHosts {
			if !util.IsIntranetIp(host.String()) {
				return nil, fmt.Errorf("target must be intranet: %s", target)
			}
			if err = addHost(host.To4()); err != nil {
				return nil, err
			}
		}
	}

	if len(hosts) == 0 {
		return nil, fmt.Errorf("cidr is required")
	}

	return hosts, nil
}

func ParseCIDRHosts(cidr string, maxHosts int) ([]net.IP, error) {
	baseIp, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}

	ipv4 := baseIp.To4()
	if ipv4 == nil {
		return nil, fmt.Errorf("only IPv4 CIDR is supported")
	}
	if !util.IsIntranetIp(ipv4.String()) {
		return nil, fmt.Errorf("cidr must be intranet: %s", cidr)
	}

	ones, bits := ipNet.Mask.Size()
	hostBits := bits - ones
	if hostBits < 0 {
		return nil, fmt.Errorf("invalid cidr mask: %s", cidr)
	}

	if hostBits >= 63 {
		return nil, fmt.Errorf("cidr range is too large")
	}
	total := uint64(1) << hostBits
	if total > uint64(maxHosts)+2 {
		return nil, fmt.Errorf("cidr range is too large, max %d hosts", maxHosts)
	}

	totalInt := int(total)
	start := binary.BigEndian.Uint32(ipv4.Mask(ipNet.Mask))
	end := start + uint32(total) - 1
	hosts := make([]net.IP, 0, totalInt)
	for value := start; value <= end; value++ {
		if total > 2 && (value == start || value == end) {
			continue
		}

		candidate := make(net.IP, net.IPv4len)
		binary.BigEndian.PutUint32(candidate, value)
		if ipNet.Contains(candidate) {
			hosts = append(hosts, candidate)
		}
	}

	if len(hosts) == 0 {
		return nil, fmt.Errorf("cidr has no usable hosts: %s", cidr)
	}

	return hosts, nil
}

func ProbeHost(ctx context.Context, client *http.Client, scheme, host string, ports []int, paths []string, timeout time.Duration) (bool, []*InnerMcpServer) {
	if !util.IsIntranetIp(host) {
		return false, nil
	}

	dialer := &net.Dialer{Timeout: timeout}
	isOnline := false
	var servers []*InnerMcpServer

	for _, port := range ports {
		address := net.JoinHostPort(host, strconv.Itoa(port))
		conn, err := dialer.DialContext(ctx, "tcp", address)
		if err != nil {
			continue
		}
		_ = conn.Close()
		isOnline = true

		for _, path := range paths {
			server, ok := probeMcpInitialize(ctx, client, scheme, host, port, path)
			if ok {
				servers = append(servers, server)
			}
		}
	}

	return isOnline, servers
}

func probeMcpInitialize(ctx context.Context, client *http.Client, scheme, host string, port int, path string) (*InnerMcpServer, bool) {
	fullUrl := fmt.Sprintf("%s://%s%s", scheme, net.JoinHostPort(host, strconv.Itoa(port)), path)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, fullUrl, nil)
	if err != nil {
		return nil, false
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, false
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode == http.StatusNotFound {
		return nil, false
	}

	return &InnerMcpServer{
		Host: host,
		Port: port,
		Path: path,
		Url:  fullUrl,
	}, true
}
