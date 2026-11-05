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
	"strings"

	"github.com/casdoor/casdoor/conf"
	"github.com/casdoor/casdoor/object"
)

// dynamicSecretSource resolves the RADIUS shared secret per request, based on
// the source address of the client (NAS). Known clients use their own secret
// (RadiusClient entity); unknown clients fall back to the global
// "radiusSecret" from app.conf, which keeps existing deployments unchanged.
//
// This implements layeh.com/radius' SecretSource interface:
//
//	type SecretSource interface {
//	    RADIUSSecret(ctx context.Context, remoteAddr net.Addr) ([]byte, error)
//	}
//
// layeh.com/radius resolves the secret through this interface before computing
// any HMAC (Request Authenticator, User-Password, Message-Authenticator), so
// per-client secrets work for every message type — including Access-Challenge
// (MFA) and Accounting-Request — without touching the packet handlers.
type dynamicSecretSource struct {
	globalSecret []byte
}

func newDynamicSecretSource() *dynamicSecretSource {
	return &dynamicSecretSource{
		globalSecret: []byte(conf.GetConfigString("radiusSecret")),
	}
}

func (s *dynamicSecretSource) RADIUSSecret(ctx context.Context, remoteAddr net.Addr) ([]byte, error) {
	if ip := ipFromAddr(remoteAddr); ip != nil {
		if secret := object.GetRadiusClientSecret(ip); secret != "" {
			return []byte(secret), nil
		}
	}
	// No per-client entry: keep the pre-existing behavior.
	return s.globalSecret, nil
}

func ipFromAddr(addr net.Addr) net.IP {
	host, _, err := net.SplitHostPort(addr.String())
	if err != nil {
		host = strings.Trim(addr.String(), "[]")
	}
	return net.ParseIP(host)
}
