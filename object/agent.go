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

	"github.com/casdoor/casdoor/util"
	"github.com/xorm-io/core"
)

type Agent struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`
	UpdatedTime string `xorm:"varchar(100)" json:"updatedTime"`
	DisplayName string `xorm:"varchar(100)" json:"displayName"`

	Url         string `xorm:"varchar(500)" json:"url"`
	Token       string `xorm:"varchar(500)" json:"token"`
	Application string `xorm:"varchar(100)" json:"application"`
}

func GetAgents(owner string) ([]*Agent, error) {
	agents := []*Agent{}
	err := ormer.Engine.Desc("created_time").Find(&agents, &Agent{Owner: owner})
	if err != nil {
		return nil, err
	}

	return agents, nil
}

func getAgent(owner string, name string) (*Agent, error) {
	agent := Agent{Owner: owner, Name: name}
	existed, err := ormer.Engine.Get(&agent)
	if err != nil {
		return nil, err
	}

	if existed {
		return &agent, nil
	}
	return nil, nil
}

func GetAgent(id string) (*Agent, error) {
	owner, name := util.GetOwnerAndNameFromIdNoCheck(id)
	return getAgent(owner, name)
}

func UpdateAgent(id string, agent *Agent) (bool, error) {
	owner, name := util.GetOwnerAndNameFromIdNoCheck(id)
	if a, err := getAgent(owner, name); err != nil {
		return false, err
	} else if a == nil {
		return false, nil
	}

	agent.UpdatedTime = util.GetCurrentTime()

	_, err := ormer.Engine.ID(core.PK{owner, name}).AllCols().Update(agent)
	if err != nil {
		return false, err
	}

	return true, nil
}

func AddAgent(agent *Agent) (bool, error) {
	affected, err := ormer.Engine.Insert(agent)
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func DeleteAgent(agent *Agent) (bool, error) {
	affected, err := ormer.Engine.ID(core.PK{agent.Owner, agent.Name}).Delete(&Agent{})
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func (agent *Agent) GetId() string {
	return fmt.Sprintf("%s/%s", agent.Owner, agent.Name)
}

func GetAgentCount(owner, field, value string) (int64, error) {
	session := GetSession(owner, -1, -1, field, value, "", "")
	return session.Count(&Agent{})
}

func GetPaginationAgents(owner string, offset, limit int, field, value, sortField, sortOrder string) ([]*Agent, error) {
	agents := []*Agent{}
	session := GetSession(owner, offset, limit, field, value, sortField, sortOrder)
	err := session.Find(&agents)
	if err != nil {
		return agents, err
	}

	return agents, nil
}
