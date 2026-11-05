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
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/casdoor/casdoor/util"
	"github.com/xorm-io/core"
)

// RadiusClient is a RADIUS client (NAS) with its own shared secret.
// Clients are matched by the source address of the RADIUS request.
// When no client matches, the server falls back to the global
// "radiusSecret" from app.conf, so existing deployments keep working.
type RadiusClient struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`

	DisplayName string `xorm:"varchar(100)" json:"displayName"`
	Secret      string `xorm:"varchar(200)" json:"secret"`
	IpCidr      string `xorm:"varchar(500)" json:"ipCidr"`
	Description string `xorm:"varchar(200)" json:"description"`
	State       string `xorm:"varchar(100)" json:"state"`
}

const (
	// RadiusClientCacheTtl bounds how long a stale client table can be
	// used if all cache-reloading paths fail. Mutations also reload
	// the cache proactively, so in practice this is only a safety net.
	RadiusClientCacheTtl = 5 * time.Minute
)

var (
	radiusClientCache     []*RadiusClient
	radiusClientCacheTime time.Time
	radiusClientCacheLock sync.RWMutex
)

func GetRadiusClientCount(owner, field, value string) (int64, error) {
	session := GetSession(owner, -1, -1, field, value, "", "")
	return session.Count(&RadiusClient{})
}

func GetRadiusClients(owner string) ([]*RadiusClient, error) {
	radiusClients := []*RadiusClient{}
	err := ormer.Engine.Desc("created_time").Find(&radiusClients, &RadiusClient{Owner: owner})
	if err != nil {
		return radiusClients, err
	}
	return radiusClients, nil
}

func GetPaginationRadiusClients(owner string, offset, limit int, field, value, sortField, sortOrder string) ([]*RadiusClient, error) {
	radiusClients := []*RadiusClient{}
	session := GetSession(owner, offset, limit, field, value, sortField, sortOrder)
	err := session.Find(&radiusClients)
	if err != nil {
		return radiusClients, err
	}
	return radiusClients, nil
}

func getRadiusClient(owner, name string) (*RadiusClient, error) {
	if owner == "" || name == "" {
		return nil, nil
	}

	radiusClient := RadiusClient{Owner: owner, Name: name}
	existed, err := ormer.Engine.Get(&radiusClient)
	if err != nil {
		return nil, err
	}

	if existed {
		return &radiusClient, nil
	} else {
		return nil, nil
	}
}

func GetRadiusClient(id string) (*RadiusClient, error) {
	owner, name, err := util.GetOwnerAndNameFromIdWithError(id)
	if err != nil {
		return nil, err
	}
	return getRadiusClient(owner, name)
}

func UpdateRadiusClient(id string, radiusClient *RadiusClient) (bool, error) {
	owner, name, err := util.GetOwnerAndNameFromIdWithError(id)
	if err != nil {
		return false, err
	}
	if existing, err := getRadiusClient(owner, name); existing == nil {
		return false, err
	}

	if err := radiusClient.check(); err != nil {
		return false, err
	}

	affected, err := ormer.Engine.ID(core.PK{owner, name}).AllCols().Update(radiusClient)
	if err != nil {
		return false, err
	}

	_ = LoadRadiusClientCache()
	return affected != 0, nil
}

func AddRadiusClient(radiusClient *RadiusClient) (bool, error) {
	if err := radiusClient.check(); err != nil {
		return false, err
	}

	affected, err := ormer.Engine.Insert(radiusClient)
	if err != nil {
		return false, err
	}

	_ = LoadRadiusClientCache()
	return affected != 0, nil
}

func DeleteRadiusClient(radiusClient *RadiusClient) (bool, error) {
	affected, err := ormer.Engine.ID(core.PK{radiusClient.Owner, radiusClient.Name}).Delete(&RadiusClient{})
	if err != nil {
		return false, err
	}

	_ = LoadRadiusClientCache()
	return affected != 0, nil
}

func (radiusClient *RadiusClient) GetId() string {
	return fmt.Sprintf("%s/%s", radiusClient.Owner, radiusClient.Name)
}

// check validates the fields that cannot be fixed at request time.
func (radiusClient *RadiusClient) check() error {
	if radiusClient.Owner == "" || radiusClient.Name == "" {
		return fmt.Errorf("owner and name cannot be empty")
	}
	if len(radiusClient.Secret) < 8 {
		return fmt.Errorf("secret must be at least 8 characters long")
	}
	for _, entry := range strings.Split(radiusClient.IpCidr, ",") {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		if strings.Contains(entry, "/") {
			if _, _, err := net.ParseCIDR(entry); err != nil {
				return fmt.Errorf("invalid CIDR entry %q: %w", entry, err)
			}
		} else if net.ParseIP(entry) == nil {
			return fmt.Errorf("invalid IP entry %q", entry)
		}
	}
	return nil
}

// LoadRadiusClientCache refreshes the in-memory client table used by the
// RADIUS server. Called lazily on first use and proactively after mutations.
func LoadRadiusClientCache() error {
	radiusClients, err := GetRadiusClients("")
	if err != nil {
		return err
	}
	radiusClientCacheLock.Lock()
	defer radiusClientCacheLock.Unlock()
	radiusClientCache = radiusClients
	radiusClientCacheTime = time.Now()
	return nil
}

func getRadiusClientCache() []*RadiusClient {
	radiusClientCacheLock.RLock()
	clients := radiusClientCache
	fresh := time.Since(radiusClientCacheTime) < RadiusClientCacheTtl
	radiusClientCacheLock.RUnlock()

	if fresh {
		return clients
	}
	// Cache missing or stale: try to refresh, fall back to what we have.
	if err := LoadRadiusClientCache(); err != nil {
		return clients
	}
	radiusClientCacheLock.RLock()
	defer radiusClientCacheLock.RUnlock()
	return radiusClientCache
}

// SetRadiusClientCacheForTest replaces the in-memory client cache without
// touching the database. For tests only.
func SetRadiusClientCacheForTest(clients []*RadiusClient) {
	radiusClientCacheLock.Lock()
	defer radiusClientCacheLock.Unlock()
	radiusClientCache = clients
	radiusClientCacheTime = time.Now()
}

// GetRadiusClientSecret returns the shared secret configured for the RADIUS
// client at ip, or "" when no enabled client matches (the caller falls back
// to the global radiusSecret).
func GetRadiusClientSecret(ip net.IP) string {
	if ip == nil {
		return ""
	}
	for _, radiusClient := range getRadiusClientCache() {
		if radiusClient.State != "Active" {
			continue
		}
		if radiusClientMatchesIp(radiusClient, ip) {
			return radiusClient.Secret
		}
	}
	return ""
}

func radiusClientMatchesIp(radiusClient *RadiusClient, ip net.IP) bool {
	for _, entry := range strings.Split(radiusClient.IpCidr, ",") {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		if strings.Contains(entry, "/") {
			if _, ipnet, err := net.ParseCIDR(entry); err == nil && ipnet.Contains(ip) {
				return true
			}
		} else if parsed := net.ParseIP(entry); parsed != nil && parsed.Equal(ip) {
			return true
		}
	}
	return false
}
