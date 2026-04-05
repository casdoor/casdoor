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

package object

import (
	"fmt"
	"strings"

	"github.com/casdoor/casdoor/util"
	"github.com/xorm-io/core"
)

type NodeItem struct {
	Name     string `json:"name"`
	Version  string `json:"version"`
	Diff     string `json:"diff"`
	Pid      int    `json:"pid"`
	Status   string `json:"status"`
	Message  string `json:"message"`
	Provider string `json:"provider"`
}

type Site struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`
	UpdatedTime string `xorm:"varchar(100)" json:"updatedTime"`
	DisplayName string `xorm:"varchar(100)" json:"displayName"`

	Tag            string      `xorm:"varchar(100)" json:"tag"`
	Domain         string      `xorm:"varchar(100)" json:"domain"`
	OtherDomains   []string    `xorm:"varchar(500)" json:"otherDomains"`
	NeedRedirect   bool        `json:"needRedirect"`
	DisableVerbose bool        `json:"disableVerbose"`
	Rules          []string    `xorm:"varchar(500)" json:"rules"`
	EnableAlert    bool        `json:"enableAlert"`
	AlertInterval  int         `json:"alertInterval"`
	AlertTryTimes  int         `json:"alertTryTimes"`
	AlertProviders []string    `xorm:"varchar(500)" json:"alertProviders"`
	Challenges     []string    `xorm:"mediumtext" json:"challenges"`
	Host           string      `xorm:"varchar(100)" json:"host"`
	Port           int         `json:"port"`
	Hosts          []string    `xorm:"varchar(1000)" json:"hosts"`
	SslMode        string      `xorm:"varchar(100)" json:"sslMode"`
	SslCert        string      `xorm:"-" json:"sslCert"`
	PublicIp       string      `xorm:"varchar(100)" json:"publicIp"`
	Node           string      `xorm:"varchar(100)" json:"node"`
	IsSelf         bool        `json:"isSelf"`
	Status         string      `xorm:"varchar(100)" json:"status"`
	Nodes          []*NodeItem `xorm:"mediumtext" json:"nodes"`

	CasdoorApplication string       `xorm:"varchar(100)" json:"casdoorApplication"`
	ApplicationObj     *Application `xorm:"-" json:"applicationObj"`
}

func GetGlobalSites() ([]*Site, error) {
	sites := []*Site{}
	err := ormer.Engine.Desc("created_time").Find(&sites)
	if err != nil {
		return nil, err
	}

	return sites, nil
}

func GetSites(owner string) ([]*Site, error) {
	sites := []*Site{}
	err := ormer.Engine.Asc("tag").Asc("port").Desc("created_time").Find(&sites, &Site{Owner: owner})
	if err != nil {
		return nil, err
	}

	for _, site := range sites {
		err = site.populateCert()
		if err != nil {
			return nil, err
		}
	}

	return sites, nil
}

func getSite(owner string, name string) (*Site, error) {
	site := Site{Owner: owner, Name: name}
	existed, err := ormer.Engine.Get(&site)
	if err != nil {
		return nil, err
	}

	if existed {
		return &site, nil
	}
	return nil, nil
}

func GetSite(id string) (*Site, error) {
	owner, name := util.GetOwnerAndNameFromIdNoCheck(id)
	site, err := getSite(owner, name)
	if err != nil {
		return nil, err
	}

	if site != nil {
		err = site.populateCert()
		if err != nil {
			return nil, err
		}
	}

	return site, nil
}

func GetMaskedSite(site *Site, node string) *Site {
	if site == nil {
		return nil
	}

	if site.PublicIp == "(empty)" {
		site.PublicIp = ""
	}

	site.IsSelf = false
	if site.Node == node {
		site.IsSelf = true
	}

	return site
}

func GetMaskedSites(sites []*Site, node string) []*Site {
	for _, site := range sites {
		site = GetMaskedSite(site, node)
	}
	return sites
}

func UpdateSite(id string, site *Site) (bool, error) {
	owner, name := util.GetOwnerAndNameFromIdNoCheck(id)
	if s, err := getSite(owner, name); err != nil {
		return false, err
	} else if s == nil {
		return false, nil
	}

	site.UpdatedTime = util.GetCurrentTime()

	_, err := ormer.Engine.ID(core.PK{owner, name}).AllCols().Update(site)
	if err != nil {
		return false, err
	}

	err = refreshSiteMap()
	if err != nil {
		return false, err
	}

	return true, nil
}

func UpdateSiteNoRefresh(id string, site *Site) (bool, error) {
	owner, name := util.GetOwnerAndNameFromIdNoCheck(id)
	if s, err := getSite(owner, name); err != nil {
		return false, err
	} else if s == nil {
		return false, nil
	}

	_, err := ormer.Engine.ID(core.PK{owner, name}).AllCols().Update(site)
	if err != nil {
		return false, err
	}

	return true, nil
}

func AddSite(site *Site) (bool, error) {
	affected, err := ormer.Engine.Insert(site)
	if err != nil {
		return false, err
	}

	if affected != 0 {
		err = refreshSiteMap()
		if err != nil {
			return false, err
		}
	}

	return affected != 0, nil
}

func DeleteSite(site *Site) (bool, error) {
	affected, err := ormer.Engine.ID(core.PK{site.Owner, site.Name}).Delete(&Site{})
	if err != nil {
		return false, err
	}

	if affected != 0 {
		err = refreshSiteMap()
		if err != nil {
			return false, err
		}
	}

	return affected != 0, nil
}

func (site *Site) GetId() string {
	return fmt.Sprintf("%s/%s", site.Owner, site.Name)
}

func (site *Site) GetChallengeMap() map[string]string {
	m := map[string]string{}
	for _, challenge := range site.Challenges {
		tokens := strings.Split(challenge, ":")
		m[tokens[0]] = tokens[1]
	}
	return m
}

func (site *Site) GetHost() string {
	if len(site.Hosts) != 0 {
		return site.Hosts[util.RandomIntn(len(site.Hosts))]
	}

	if site.Host != "" {
		return site.Host
	}

	if site.Port == 0 {
		return ""
	}

	res := fmt.Sprintf("http://localhost:%d", site.Port)
	return res
}

func addErrorToMsg(msg string, function string, err error) string {
	fmt.Printf("%s(): %s\n", function, err.Error())
	if msg == "" {
		return fmt.Sprintf("%s(): %s", function, err.Error())
	} else {
		return fmt.Sprintf("%s || %s(): %s", msg, function, err.Error())
	}
}

func GetSiteCount(owner, field, value string) (int64, error) {
	session := GetSession(owner, -1, -1, field, value, "", "")
	return session.Count(&Site{})
}

func GetPaginationSites(owner string, offset, limit int, field, value, sortField, sortOrder string) ([]*Site, error) {
	sites := []*Site{}
	session := GetSession(owner, offset, limit, field, value, sortField, sortOrder)
	err := session.Where("owner = ? or owner = ?", "admin", owner).Find(&sites)
	if err != nil {
		return sites, err
	}

	return sites, nil
}
