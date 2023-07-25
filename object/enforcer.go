// Copyright 2023 The Casdoor Authors. All Rights Reserved.
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
	"github.com/casbin/casbin/v2"
	"github.com/casdoor/casdoor/util"
	"github.com/xorm-io/core"
)

type Enforcer struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`
	UpdatedTime string `xorm:"varchar(100) updated" json:"updatedTime"`
	DisplayName string `xorm:"varchar(100)" json:"displayName"`
	Description string `xorm:"varchar(100)" json:"description"`

	Model     string `xorm:"varchar(100)" json:"model"`
	Adapter   string `xorm:"varchar(100)" json:"adapter"`
	IsEnabled bool   `json:"isEnabled"`

	*casbin.Enforcer
}

func GetEnforcerCount(owner, field, value string) (int64, error) {
	session := GetSession(owner, -1, -1, field, value, "", "")
	return session.Count(&Enforcer{})
}

func GetEnforcers(owner string) ([]*Enforcer, error) {
	enforcers := []*Enforcer{}
	err := adapter.Engine.Desc("created_time").Find(&enforcers, &Enforcer{Owner: owner})
	if err != nil {
		return enforcers, err
	}

	return enforcers, nil
}

func GetPaginationEnforcers(owner string, offset, limit int, field, value, sortField, sortOrder string) ([]*Enforcer, error) {
	enforcers := []*Enforcer{}
	session := GetSession(owner, offset, limit, field, value, sortField, sortOrder)
	err := session.Find(&enforcers)
	if err != nil {
		return enforcers, err
	}

	return enforcers, nil
}

func getEnforcer(owner string, name string) (*Enforcer, error) {
	if owner == "" || name == "" {
		return nil, nil
	}

	enforcer := Enforcer{Owner: owner, Name: name}
	existed, err := adapter.Engine.Get(&enforcer)
	if err != nil {
		return &enforcer, err
	}

	if existed {
		return &enforcer, nil
	} else {
		return nil, nil
	}
}

func GetEnforcer(id string) (*Enforcer, error) {
	owner, name := util.GetOwnerAndNameFromId(id)
	return getEnforcer(owner, name)
}

func UpdateEnforcer(id string, enforcer *Enforcer) (bool, error) {
	owner, name := util.GetOwnerAndNameFromId(id)
	if oldEnforcer, err := getEnforcer(owner, name); err != nil {
		return false, err
	} else if oldEnforcer == nil {
		return false, nil
	}

	affected, err := adapter.Engine.ID(core.PK{owner, name}).AllCols().Update(enforcer)
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func AddEnforcer(enforcer *Enforcer) (bool, error) {
	affected, err := adapter.Engine.Insert(enforcer)
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func DeleteEnforcer(enforcer *Enforcer) (bool, error) {
	affected, err := adapter.Engine.ID(core.PK{enforcer.Owner, enforcer.Name}).Delete(&Enforcer{})
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}
