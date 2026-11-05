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

// RadiusClient is a NAS allowed to talk to the built-in RADIUS server with its own shared secret.
type RadiusClient struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`
	UpdatedTime string `xorm:"varchar(100)" json:"updatedTime"`
	DisplayName string `xorm:"varchar(100)" json:"displayName"`

	Organization string `xorm:"varchar(100)" json:"organization"`
	IpRanges     string `xorm:"varchar(1000)" json:"ipRanges"`
	Secret       string `xorm:"varchar(100)" json:"secret"`
	State        string `xorm:"varchar(100)" json:"state"`
}

type radiusClientRange struct {
	ipNet  *net.IPNet
	client *RadiusClient
}

const radiusClientCacheTtl = time.Minute

var (
	radiusClientRanges   []radiusClientRange
	radiusClientLoadTime time.Time
	radiusClientLock     sync.Mutex
)

func GetRadiusClientCount(owner, field, value string) (int64, error) {
	session := GetSession(owner, -1, -1, field, value, "", "")
	return session.Count(&RadiusClient{})
}

func GetRadiusClients(owner string) ([]*RadiusClient, error) {
	clients := []*RadiusClient{}
	err := ormer.Engine.Desc("created_time").Find(&clients, &RadiusClient{Owner: owner})
	return clients, err
}

func GetPaginationRadiusClients(owner string, offset, limit int, field, value, sortField, sortOrder string) ([]*RadiusClient, error) {
	clients := []*RadiusClient{}
	session := GetSession(owner, offset, limit, field, value, sortField, sortOrder)
	err := session.Find(&clients)
	return clients, err
}

func getRadiusClient(owner, name string) (*RadiusClient, error) {
	if owner == "" || name == "" {
		return nil, nil
	}

	client := RadiusClient{Owner: owner, Name: name}
	existed, err := ormer.Engine.Get(&client)
	if err != nil {
		return nil, err
	}
	if existed {
		return &client, nil
	}
	return nil, nil
}

func GetRadiusClient(id string) (*RadiusClient, error) {
	owner, name, err := util.GetOwnerAndNameFromIdWithError(id)
	if err != nil {
		return nil, err
	}
	return getRadiusClient(owner, name)
}

func UpdateRadiusClient(id string, client *RadiusClient) (bool, error) {
	owner, name, err := util.GetOwnerAndNameFromIdWithError(id)
	if err != nil {
		return false, err
	}
	if oldClient, err := getRadiusClient(owner, name); err != nil {
		return false, err
	} else if oldClient == nil {
		return false, nil
	}

	client.UpdatedTime = util.GetCurrentTime()
	affected, err := ormer.Engine.ID(core.PK{owner, name}).AllCols().Update(client)
	if err != nil {
		return false, err
	}
	resetRadiusClientCache()
	return affected != 0, nil
}

func AddRadiusClient(client *RadiusClient) (bool, error) {
	affected, err := ormer.Engine.Insert(client)
	if err != nil {
		return false, err
	}
	resetRadiusClientCache()
	return affected != 0, nil
}

func DeleteRadiusClient(client *RadiusClient) (bool, error) {
	affected, err := ormer.Engine.ID(core.PK{client.Owner, client.Name}).Delete(&RadiusClient{})
	if err != nil {
		return false, err
	}
	resetRadiusClientCache()
	return affected != 0, nil
}

func (client *RadiusClient) GetId() string {
	return fmt.Sprintf("%s/%s", client.Owner, client.Name)
}

func ParseRadiusClientIpRanges(ipRanges string) ([]*net.IPNet, error) {
	res := []*net.IPNet{}
	for _, ipRange := range strings.Split(ipRanges, ",") {
		ipRange = strings.TrimSpace(ipRange)
		if ipRange == "" {
			continue
		}

		if !strings.Contains(ipRange, "/") {
			ip := net.ParseIP(ipRange)
			if ip == nil {
				return nil, fmt.Errorf("invalid IP address: %s", ipRange)
			}
			if ip.To4() != nil {
				ipRange += "/32"
			} else {
				ipRange += "/128"
			}
		}

		_, ipNet, err := net.ParseCIDR(ipRange)
		if err != nil {
			return nil, fmt.Errorf("invalid IP range: %s", ipRange)
		}
		res = append(res, ipNet)
	}

	if len(res) == 0 {
		return nil, fmt.Errorf("IP ranges should not be empty")
	}
	return res, nil
}

func resetRadiusClientCache() {
	radiusClientLock.Lock()
	defer radiusClientLock.Unlock()
	radiusClientLoadTime = time.Time{}
}

func getRadiusClientRanges() ([]radiusClientRange, error) {
	radiusClientLock.Lock()
	defer radiusClientLock.Unlock()

	if time.Since(radiusClientLoadTime) < radiusClientCacheTtl {
		return radiusClientRanges, nil
	}

	clients := []*RadiusClient{}
	err := ormer.Engine.Where("state = ?", "Active").Find(&clients)
	if err != nil {
		return nil, err
	}

	ranges := []radiusClientRange{}
	for _, client := range clients {
		ipNets, err := ParseRadiusClientIpRanges(client.IpRanges)
		if err != nil {
			continue
		}
		for _, ipNet := range ipNets {
			ranges = append(ranges, radiusClientRange{ipNet: ipNet, client: client})
		}
	}

	radiusClientRanges = ranges
	radiusClientLoadTime = time.Now()
	return radiusClientRanges, nil
}

// GetRadiusClientByIp returns the active RADIUS client whose IP ranges contain the IP, the most specific range wins.
func GetRadiusClientByIp(ip net.IP) (*RadiusClient, error) {
	if ip == nil {
		return nil, nil
	}

	ranges, err := getRadiusClientRanges()
	if err != nil {
		return nil, err
	}

	var res *RadiusClient
	bestBits := -1
	for _, r := range ranges {
		if !r.ipNet.Contains(ip) {
			continue
		}
		bits, _ := r.ipNet.Mask.Size()
		if bits > bestBits {
			bestBits = bits
			res = r.client
		}
	}
	return res, nil
}
