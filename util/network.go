// Copyright 2025 The Casdoor Authors. All Rights Reserved.
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
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"syscall"
	"time"
)

func GetHostname() string {
	name, err := os.Hostname()
	if err != nil {
		panic(err)
	}

	return name
}

func IsInternetIp(ip string) bool {
	ipStr, _, err := net.SplitHostPort(ip)
	if err != nil {
		ipStr = ip
	}

	parsedIP := net.ParseIP(ipStr)
	if parsedIP == nil {
		return false
	}

	return !parsedIP.IsPrivate() && !parsedIP.IsLoopback() && !parsedIP.IsMulticast() && !parsedIP.IsUnspecified()
}

func IsHostIntranet(ip string) bool {
	ipStr, _, err := net.SplitHostPort(ip)
	if err != nil {
		ipStr = ip
	}

	parsedIP := net.ParseIP(ipStr)
	if parsedIP == nil {
		return false
	}

	return parsedIP.IsPrivate() || parsedIP.IsLoopback() || parsedIP.IsLinkLocalUnicast() || parsedIP.IsLinkLocalMulticast()
}

// IsCredentialedOrigin reports whether the browser origin may act with the Casdoor session cookie,
// the same origins CorsFilter answers with "Access-Control-Allow-Credentials"
func IsCredentialedOrigin(origin string, originConf string, host string) bool {
	if origin == originConf {
		return true
	}

	originUrl, err := url.Parse(origin)
	if err != nil || originUrl.Host == "" {
		return false
	}

	hostname, _, err := net.SplitHostPort(host)
	if err != nil {
		hostname = host
	}

	originHostname := originUrl.Hostname()
	return originHostname == hostname || IsHostIntranet(hostname) && IsHostIntranet(originHostname)
}

// NewInternetOnlyHttpClient returns a client that refuses to connect to intranet, loopback,
// link-local (e.g. cloud metadata) and unspecified addresses. The check runs on the resolved
// IP of every connection, so DNS rebinding and redirects to such addresses are refused too.
func NewInternetOnlyHttpClient(timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout:   timeout,
		Transport: newRestrictedHttpTransport(timeout, isNonInternetAddress, "only Internet addresses can be requested"),
	}
}

func isNonInternetAddress(address string) bool {
	return IsHostIntranet(address) || isUnspecifiedIp(address) || isSharedAddressSpaceIp(address)
}

// CheckInternetHost refuses a host (a hostname, "host:port" or URL) that resolves to an address
// NewInternetOnlyHttpClient would refuse, for the clients that cannot dial through it, e.g. SMTP.
// A bare path such as "/v3/mail/send" names no host and passes.
func CheckInternetHost(host string) error {
	if strings.Contains(host, "://") {
		urlObj, err := url.Parse(host)
		if err != nil {
			return err
		}
		host = urlObj.Hostname()
	} else if hostname, _, err := net.SplitHostPort(host); err == nil {
		host = hostname
	}

	host = strings.Trim(host, "[]")
	if host == "" || strings.HasPrefix(host, "/") {
		return nil
	}

	ips, err := net.LookupIP(host)
	if err != nil {
		return err
	}
	for _, ip := range ips {
		if isNonInternetAddress(ip.String()) {
			return fmt.Errorf("the host: %s is not allowed, only Internet addresses can be requested", host)
		}
	}
	return nil
}

func NewNonLocalHttpTransport(timeout time.Duration) *http.Transport {
	isRefused := func(address string) bool {
		ip := parseDialIp(address)
		return ip != nil && (ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified() || isSharedAddressSpaceIp(address))
	}

	return newRestrictedHttpTransport(timeout, isRefused, "loopback, link-local and metadata addresses cannot be requested")
}

func newRestrictedHttpTransport(timeout time.Duration, isRefused func(address string) bool, reason string) *http.Transport {
	dialer := &net.Dialer{
		Timeout: timeout,
		Control: func(network, address string, _ syscall.RawConn) error {
			if isRefused(address) {
				return fmt.Errorf("the address: %s is not allowed, %s", address, reason)
			}
			return nil
		},
	}

	return &http.Transport{
		// a proxy would be dialed instead of the target, bypassing the check
		Proxy:       nil,
		DialContext: dialer.DialContext,
	}
}

func parseDialIp(address string) net.IP {
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		host = address
	}
	return net.ParseIP(host)
}

var sharedAddressSpace = &net.IPNet{IP: net.IPv4(100, 64, 0, 0), Mask: net.CIDRMask(10, 32)}

func isSharedAddressSpaceIp(address string) bool {
	ip := parseDialIp(address)
	return ip != nil && sharedAddressSpace.Contains(ip)
}

func isUnspecifiedIp(address string) bool {
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		host = address
	}

	parsedIP := net.ParseIP(host)
	return parsedIP != nil && parsedIP.IsUnspecified()
}

func ResolveDomainToIp(domain string) string {
	ips, err := net.LookupIP(domain)
	if err != nil {
		if strings.Contains(err.Error(), "no such host") {
			return "(empty)"
		}

		fmt.Printf("resolveDomainToIp() error: %s\n", err.Error())
		return err.Error()
	}

	for _, ip := range ips {
		if ipv4 := ip.To4(); ipv4 != nil {
			return ipv4.String()
		}
	}
	return "(empty)"
}

func PingUrl(url string) (bool, string) {
	client := http.Client{
		Timeout: 5 * time.Second,
	}
	resp, err := client.Get(url)
	if err != nil {
		return false, err.Error()
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		return true, ""
	}
	return false, fmt.Sprintf("Status: %s", resp.Status)
}

func IsIntranetIp(ip string) bool {
	ipStr, _, err := net.SplitHostPort(ip)
	if err != nil {
		ipStr = ip
	}

	parsedIP := net.ParseIP(ipStr)
	if parsedIP == nil {
		return false
	}

	return parsedIP.IsPrivate() ||
		parsedIP.IsLoopback() ||
		parsedIP.IsLinkLocalUnicast() ||
		parsedIP.IsLinkLocalMulticast()
}
