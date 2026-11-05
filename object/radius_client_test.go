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

package object

import (
	"net"
	"testing"
	"time"
)

func TestRadiusClientMatchesIp(t *testing.T) {
	cases := []struct {
		ipCidr string
		ip     string
		want   bool
	}{
		{"192.168.20.50", "192.168.20.50", true},
		{"192.168.20.50", "192.168.20.51", false},
		{"192.168.20.0/24", "192.168.20.77", true},
		{"192.168.20.0/24", "192.168.21.77", false},
		{"192.168.20.0/24, 10.1.1.5", "10.1.1.5", true},
		{"192.168.20.0/24,10.1.1.0/28", "10.1.1.9", true},
		{"", "192.168.20.50", false},
		{"not-an-ip", "192.168.20.50", false},
	}
	for _, c := range cases {
		radiusClient := &RadiusClient{IpCidr: c.ipCidr, State: "Active"}
		if got := radiusClientMatchesIp(radiusClient, net.ParseIP(c.ip)); got != c.want {
			t.Errorf("IpCidr=%q ip=%s: got %v, want %v", c.ipCidr, c.ip, got, c.want)
		}
	}
}

// TestGetRadiusClientSecretAndFallback verifies per-client lookup and the
// empty-string fallback (the RADIUS layer then uses the global secret).
// It injects the cache directly so no database is needed.
func TestGetRadiusClientSecretAndFallback(t *testing.T) {
	savedCache := radiusClientCache
	savedTime := radiusClientCacheTime
	t.Cleanup(func() {
		radiusClientCacheLock.Lock()
		defer radiusClientCacheLock.Unlock()
		radiusClientCache = savedCache
		radiusClientCacheTime = savedTime
	})

	radiusClientCacheLock.Lock()
	radiusClientCache = []*RadiusClient{
		{Owner: "admin", Name: "vpn-gateway", Secret: "secret-a", IpCidr: "10.0.0.1", State: "Active"},
		{Owner: "admin", Name: "wifi", Secret: "secret-b", IpCidr: "10.2.0.0/24", State: "Active"},
		{Owner: "admin", Name: "retired", Secret: "secret-c", IpCidr: "10.3.0.1", State: "Inactive"},
	}
	radiusClientCacheTime = time.Now()
	radiusClientCacheLock.Unlock()

	cases := []struct {
		ip   string
		want string
	}{
		{"10.0.0.1", "secret-a"},
		{"10.2.0.99", "secret-b"},
		{"10.2.1.99", ""},   // outside CIDR
		{"10.3.0.1", ""},    // client disabled
		{"192.168.1.1", ""}, // unknown -> fallback
	}
	for _, c := range cases {
		if got := GetRadiusClientSecret(net.ParseIP(c.ip)); got != c.want {
			t.Errorf("ip=%s: got %q, want %q", c.ip, got, c.want)
		}
	}
}
