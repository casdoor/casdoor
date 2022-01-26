// Copyright 2021 The casbin Authors. All Rights Reserved.
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
	"xorm.io/core"
)

type Role struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`
	DisplayName string `xorm:"varchar(100)" json:"displayName"`

	Users     []string `xorm:"mediumtext" json:"users"`
	Roles     []string `xorm:"mediumtext" json:"roles"`
	IsEnabled bool     `json:"isEnabled"`
}

func GetRoleCount(owner, field, value string) int {
	session := GetSession(owner, -1, -1, field, value, "", "")
	count, err := session.Count(&Role{})
	if err != nil {
		panic(err)
	}

	return int(count)
}

func GetRoles(owner string) []*Role {
	roles := []*Role{}
	err := adapter.Engine.Desc("created_time").Find(&roles, &Role{Owner: owner})
	if err != nil {
		panic(err)
	}

	return roles
}

func GetPaginationRoles(owner string, offset, limit int, field, value, sortField, sortOrder string) []*Role {
	roles := []*Role{}
	session := GetSession(owner, offset, limit, field, value, sortField, sortOrder)
	err := session.Find(&roles)
	if err != nil {
		panic(err)
	}

	return roles
}

func getRole(owner string, name string) *Role {
	if owner == "" || name == "" {
		return nil
	}

	role := Role{Owner: owner, Name: name}
	existed, err := adapter.Engine.Get(&role)
	if err != nil {
		panic(err)
	}

	if existed {
		return &role
	} else {
		return nil
	}
}

func GetRole(id string) *Role {
	owner, name := util.GetOwnerAndNameFromId(id)
	return getRole(owner, name)
}

func UpdateRole(id string, role *Role) bool {
	owner, name := util.GetOwnerAndNameFromId(id)
	if getRole(owner, name) == nil {
		return false
	}

	affected, err := adapter.Engine.ID(core.PK{owner, name}).AllCols().Update(role)
	if err != nil {
		panic(err)
	}

	return affected != 0
}

func AddRole(role *Role) bool {
	affected, err := adapter.Engine.Insert(role)
	if err != nil {
		panic(err)
	}

	return affected != 0
}

func DeleteRole(role *Role) bool {
	affected, err := adapter.Engine.ID(core.PK{role.Owner, role.Name}).Delete(&Role{})
	if err != nil {
		panic(err)
	}

	return affected != 0
}

func (role *Role) GetId() string {
	return fmt.Sprintf("%s/%s", role.Owner, role.Name)
}
