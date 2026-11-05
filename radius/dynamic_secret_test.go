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

package radius

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/casdoor/casdoor/object"
	layehr "layeh.com/radius"
	"layeh.com/radius/rfc2865"
)

// startTestServer runs a real layeh RADIUS server backed by
// dynamicSecretSource and returns its local UDP address.
func startTestServer(t *testing.T, src *dynamicSecretSource) string {
	t.Helper()
	pc, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen packet: %v", err)
	}
	server := &layehr.PacketServer{
		Handler: layehr.HandlerFunc(func(w layehr.ResponseWriter, r *layehr.Request) {
			_ = w.Write(r.Packet.Response(layehr.CodeAccessAccept))
		}),
		SecretSource: src,
	}
	go func() {
		_ = server.Serve(pc)
	}()
	t.Cleanup(func() {
		_ = server.Shutdown(context.Background())
	})
	return pc.LocalAddr().String()
}

func exchange(t *testing.T, addr string, secret string) (*layehr.Packet, error) {
	t.Helper()
	packet := layehr.New(layehr.CodeAccessRequest, []byte(secret))
	if err := rfc2865.UserName_SetString(packet, "alice"); err != nil {
		t.Fatalf("set username: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	client := &layehr.Client{Retry: 200 * time.Millisecond}
	return client.Exchange(ctx, packet, addr)
}

// TestDynamicSecretSourceEndToEnd spins up a real RADIUS server with
// dynamicSecretSource and validates the three secret-resolution paths:
//  1. known client source address -> per-client secret verifies
//  2. wrong secret -> server silently drops, client times out
//  3. unknown source address -> global secret fallback verifies
func TestDynamicSecretSourceEndToEnd(t *testing.T) {
	object.SetRadiusClientCacheForTest([]*object.RadiusClient{
		{Owner: "admin", Name: "vpn-gateway", Secret: "client-secret-1", IpCidr: "127.0.0.1", State: "Active"},
	})
	src := &dynamicSecretSource{globalSecret: []byte("global-secret")}
	addr := startTestServer(t, src)

	t.Run("per client secret", func(t *testing.T) {
		resp, err := exchange(t, addr, "client-secret-1")
		if err != nil {
			t.Fatalf("exchange with per-client secret: %v", err)
		}
		if resp.Code != layehr.CodeAccessAccept {
			t.Fatalf("got %v, want Access-Accept", resp.Code)
		}
	})

	t.Run("wrong secret is dropped", func(t *testing.T) {
		if _, err := exchange(t, addr, "totally-wrong"); err == nil {
			t.Fatalf("exchange with wrong secret should fail (server must drop it), but got a reply")
		}
	})

	t.Run("unknown client falls back to global secret", func(t *testing.T) {
		object.SetRadiusClientCacheForTest([]*object.RadiusClient{
			{Owner: "admin", Name: "elsewhere", Secret: "other-secret", IpCidr: "10.9.9.9", State: "Active"},
		})
		resp, err := exchange(t, addr, "global-secret")
		if err != nil {
			t.Fatalf("exchange with global secret: %v", err)
		}
		if resp.Code != layehr.CodeAccessAccept {
			t.Fatalf("got %v, want Access-Accept", resp.Code)
		}
	})
}

func TestIpFromAddr(t *testing.T) {
	cases := []struct {
		addr string
		want string
	}{
		{"127.0.0.1:38912", "127.0.0.1"},
		{"[::1]:38912", "::1"},
	}
	for _, c := range cases {
		ua, err := net.ResolveUDPAddr("udp", c.addr)
		if err != nil {
			t.Fatalf("resolve %s: %v", c.addr, err)
		}
		got := ipFromAddr(ua)
		if got == nil || got.String() != c.want {
			t.Errorf("ipFromAddr(%s) = %v, want %s", c.addr, got, c.want)
		}
	}
}
