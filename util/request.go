// Copyright 2023 The casbin Authors. All Rights Reserved.
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

package util

import (
	"net"
	"net/http"
	"strings"

	"github.com/casdoor/casdoor/conf"
)

func GetClientIp(r *http.Request) string {
	return GetClientIpFromRequest(r)
}

func getRemoteIp(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	return strings.Trim(host, "[]")
}

func isTrustedProxy(ip string) bool {
	parsedIp := net.ParseIP(ip)
	if parsedIp == nil {
		return false
	}

	trustedProxies := strings.TrimSpace(conf.GetConfigString("trustedProxies"))
	if trustedProxies == "" {
		return parsedIp.IsLoopback() || parsedIp.IsPrivate()
	}

	for _, proxy := range strings.Split(trustedProxies, ",") {
		proxy = strings.TrimSpace(proxy)
		if proxy == "*" {
			return true
		}
		if _, ipNet, err := net.ParseCIDR(proxy); err == nil {
			if ipNet.Contains(parsedIp) {
				return true
			}
		} else if proxyIp := net.ParseIP(proxy); proxyIp != nil && proxyIp.Equal(parsedIp) {
			return true
		}
	}
	return false
}

func parseForwardedIp(value string) string {
	value = strings.TrimSpace(value)
	if host, _, err := net.SplitHostPort(value); err == nil {
		value = host
	}
	return strings.Trim(value, "[]")
}

func getForwardedClientIp(r *http.Request) string {
	forwardedIps := []string{}
	for _, header := range r.Header.Values("X-Forwarded-For") {
		for _, value := range strings.Split(header, ",") {
			if ip := parseForwardedIp(value); ip != "" {
				forwardedIps = append(forwardedIps, ip)
			}
		}
	}

	for i := len(forwardedIps) - 1; i >= 0; i-- {
		if !isTrustedProxy(forwardedIps[i]) || i == 0 {
			return forwardedIps[i]
		}
	}

	return parseForwardedIp(r.Header.Get("X-Real-IP"))
}

func GetClientIpFromRequest(r *http.Request) string {
	remoteIp := getRemoteIp(r)
	if !isTrustedProxy(remoteIp) {
		return remoteIp
	}

	if clientIp := getForwardedClientIp(r); clientIp != "" {
		return clientIp
	}
	return remoteIp
}
