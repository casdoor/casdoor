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
	"net/netip"
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
	addr, ok := parseAddr(ip)
	if !ok {
		return false
	}

	return !addr.IsPrivate() && !addr.IsLoopback() && !addr.IsMulticast() && !addr.IsUnspecified()
}

func IsHostIntranet(ip string) bool {
	addr, ok := parseAddr(ip)
	if !ok {
		return false
	}

	return addr.IsPrivate() || addr.IsLoopback() || addr.IsLinkLocalUnicast() || addr.IsLinkLocalMulticast()
}

func parseAddr(address string) (netip.Addr, bool) {
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		host = address
	}

	addr, err := netip.ParseAddr(strings.Trim(host, "[]"))
	if err != nil {
		return netip.Addr{}, false
	}

	addr = addr.WithZone("").Unmap()
	if nat64Prefix.Contains(addr) {
		b := addr.As16()
		addr = netip.AddrFrom4([4]byte{b[12], b[13], b[14], b[15]})
	}
	return addr, true
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
	return originHostname == hostname || isLoopbackHostname(hostname) && isLoopbackHostname(originHostname)
}

func isLoopbackHostname(hostname string) bool {
	hostname = strings.Trim(hostname, "[]")
	if strings.EqualFold(hostname, "localhost") {
		return true
	}

	parsedIP := net.ParseIP(hostname)
	return parsedIP != nil && parsedIP.IsLoopback()
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
	addr, ok := parseAddr(address)
	return !ok || isNonInternetAddr(addr)
}

var nat64Prefix = netip.MustParsePrefix("64:ff9b::/96")

var nonLocalBlockedPrefixes = []netip.Prefix{
	netip.MustParsePrefix("0.0.0.0/8"),
	netip.MustParsePrefix("100.64.0.0/10"),
}

var nonInternetPrefixes = append([]netip.Prefix{
	netip.MustParsePrefix("192.0.0.0/24"),
	netip.MustParsePrefix("198.18.0.0/15"),
	netip.MustParsePrefix("240.0.0.0/4"),
	netip.MustParsePrefix("fec0::/10"),
}, nonLocalBlockedPrefixes...)

func isInPrefixes(addr netip.Addr, prefixes []netip.Prefix) bool {
	for _, prefix := range prefixes {
		if prefix.Contains(addr) {
			return true
		}
	}
	return false
}

func isLocalAddr(addr netip.Addr) bool {
	return addr.IsLoopback() || addr.IsLinkLocalUnicast() || addr.IsLinkLocalMulticast() || addr.IsInterfaceLocalMulticast() || addr.IsUnspecified() || isInPrefixes(addr, nonLocalBlockedPrefixes)
}

func isNonInternetAddr(addr netip.Addr) bool {
	return isLocalAddr(addr) || addr.IsPrivate() || addr.IsMulticast() || isInPrefixes(addr, nonInternetPrefixes)
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
		addr, ok := parseAddr(address)
		return !ok || isLocalAddr(addr)
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
	addr, ok := parseAddr(ip)
	if !ok {
		return false
	}

	return addr.IsPrivate() ||
		addr.IsLoopback() ||
		addr.IsLinkLocalUnicast() ||
		addr.IsLinkLocalMulticast()
}
