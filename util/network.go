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
	"net"
)

// IsPrivateIP checks if the given IP is a private IP address (internal network IP)
func IsPrivateIP(ip string) bool {
	// Parse the IP address
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	// Check if it's an IPv4 address
	if parsedIP.To4() != nil {
		// IPv4 private address ranges:
		// 10.0.0.0/8     (10.0.0.0 - 10.255.255.255)
		// 172.16.0.0/12  (172.16.0.0 - 172.31.255.255)
		// 192.168.0.0/16 (192.168.0.0 - 192.168.255.255)

		// 10.0.0.0/8
		if parsedIP[12] == 10 {
			return true
		}

		// 172.16.0.0/12
		if parsedIP[12] == 172 && parsedIP[13] >= 16 && parsedIP[13] <= 31 {
			return true
		}

		// 192.168.0.0/16
		if parsedIP[12] == 192 && parsedIP[13] == 168 {
			return true
		}

		// 127.0.0.0/8 (loopback address)
		if parsedIP[12] == 127 {
			return true
		}

		// 169.254.0.0/16 (link-local address)
		if parsedIP[12] == 169 && parsedIP[13] == 254 {
			return true
		}
	} else {
		// IPv6 private addresses
		// fc00::/7 (unique local address)
		if parsedIP[0] >= 0xfc && parsedIP[0] <= 0xfd {
			return true
		}

		// fe80::/10 (link-local address)
		if parsedIP[0] == 0xfe && (parsedIP[1]&0xc0) == 0x80 {
			return true
		}

		// ::1/128 (loopback address)
		if parsedIP.IsLoopback() {
			return true
		}
	}

	return false
}

// IsInternetIp checks if the given IP is a public IP address (external network IP)
func IsInternetIp(ip string) bool {
	// Parse the IP address
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	if IsPrivateIP(ip) {
		return false
	}

	if parsedIP.IsMulticast() {
		return false
	}

	if parsedIP.IsUnspecified() {
		return false
	}

	// For IPv4, need to exclude some special address ranges
	if parsedIP.To4() != nil {
		// 0.0.0.0/8 (this network)
		if parsedIP[12] == 0 {
			return false
		}

		// 224.0.0.0/4 (multicast addresses, already checked by IsMulticast)
		// 240.0.0.0/4 (reserved addresses)
		if parsedIP[12] >= 240 {
			return false
		}
	}

	return true
}
