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

package edition

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"
)

const upgradeOrderType = "upgrade"

// UpgradeOrder is the vendor asking this instance to move to a release, signed by the same key as the license; the
// checksums (by release package name) let it install a package from a mirror or one uploaded by hand
type UpgradeOrder struct {
	Type       string            `json:"type"`
	Id         string            `json:"id"`
	License    string            `json:"license"`
	InstanceId string            `json:"instanceId,omitempty"`
	Version    string            `json:"version"`
	Note       string            `json:"note,omitempty"`
	Checksums  map[string]string `json:"checksums,omitempty"`
	IssuedTime string            `json:"issuedTime"`
	ExpireTime string            `json:"expireTime,omitempty"`
}

var (
	orderMu      sync.RWMutex
	pendingOrder *UpgradeOrder
	orderHandler func(order *UpgradeOrder)
)

// SetUpgradeOrderHandler is called by the upgrade package, which starts an upgrade by itself when it may
func SetUpgradeOrderHandler(handler func(order *UpgradeOrder)) {
	orderMu.Lock()
	orderHandler = handler
	orderMu.Unlock()
}

// ParseUpgradeOrder verifies an upgrade order and checks that it is meant for this instance and its license
func ParseUpgradeOrder(text string) (*UpgradeOrder, error) {
	parts := strings.Split(strings.TrimSpace(text), ".")
	if len(parts) != 2 {
		return nil, fmt.Errorf("the upgrade order is malformed")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, fmt.Errorf("the upgrade order is malformed")
	}
	signature, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil || !ed25519.Verify(publicKey, []byte(parts[0]), signature) {
		return nil, fmt.Errorf("the upgrade order signature is invalid")
	}

	order := &UpgradeOrder{}
	if err = json.Unmarshal(payload, order); err != nil || order.Type != upgradeOrderType || order.Version == "" {
		return nil, fmt.Errorf("the upgrade order is malformed")
	}
	if order.ExpireTime != "" {
		expireTime, err := time.Parse(time.RFC3339, order.ExpireTime)
		if err != nil || time.Now().After(expireTime) {
			return nil, fmt.Errorf("the upgrade order has expired")
		}
	}

	stateMu.RLock()
	license := current
	stateMu.RUnlock()
	if license == nil || license.Id != order.License {
		return nil, fmt.Errorf("the upgrade order is for the license %s, not the license of this Casdoor", order.License)
	}
	if order.InstanceId != "" && order.InstanceId != getInstanceId() {
		return nil, fmt.Errorf("the upgrade order is for another Casdoor instance")
	}
	return order, nil
}

// GetUpgradeOrder returns the upgrade the license server asks for, if any
func GetUpgradeOrder() *UpgradeOrder {
	orderMu.RLock()
	defer orderMu.RUnlock()
	return pendingOrder
}

func setUpgradeOrder(order *UpgradeOrder) {
	orderMu.Lock()
	changed := (order == nil) != (pendingOrder == nil) || (order != nil && order.Id != pendingOrder.Id)
	pendingOrder = order
	handler := orderHandler
	orderMu.Unlock()

	if changed && order != nil && handler != nil {
		handler(order)
	}
}
