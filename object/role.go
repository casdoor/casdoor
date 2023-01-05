// Copyright 2021 The Casdoor Authors. All Rights Reserved.
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
	"xorm.io/core"
)

type Role struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`
	DisplayName string `xorm:"varchar(100)" json:"displayName"`

	Users     []string `xorm:"mediumtext" json:"users"`
	Roles     []string `xorm:"mediumtext" json:"roles"`
	Domains   []string `xorm:"mediumtext" json:"domains"`
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
	oldRole := getRole(owner, name)
	if oldRole == nil {
		return false
	}

	if name != role.Name {
		err := roleChangeTrigger(name, role.Name)
		if err != nil {
			return false
		}
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

func GetRolesByUser(userId string) []*Role {
	roles := []*Role{}
	err := adapter.Engine.Where("users like ?", "%"+userId+"%").Find(&roles)
	if err != nil {
		panic(err)
	}

	return roles
}

func roleChangeTrigger(oldName string, newName string) error {
	session := adapter.Engine.NewSession()
	defer session.Close()

	err := session.Begin()
	if err != nil {
		return err
	}

	var roles []*Role
	err = adapter.Engine.Find(&roles)
	if err != nil {
		return err
	}
	for _, role := range roles {
		for j, u := range role.Roles {
			split := strings.Split(u, "/")
			if split[1] == oldName {
				split[1] = newName
				role.Roles[j] = split[0] + "/" + split[1]
			}
		}
		_, err = session.Where("name=?", role.Name).Update(role)
		if err != nil {
			return err
		}
	}

	var permissions []*Permission
	err = adapter.Engine.Find(&permissions)
	if err != nil {
		return err
	}
	for _, permission := range permissions {
		for j, u := range permission.Roles {
			// u = organization/username
			split := strings.Split(u, "/")
			if split[1] == oldName {
				split[1] = newName
				permission.Roles[j] = split[0] + "/" + split[1]
			}
		}
		_, err = session.Where("name=?", permission.Name).Update(permission)
		if err != nil {
			return err
		}
	}

	return session.Commit()
}

func GetMaskedRoles(roles []*Role) []*Role {
	for _, role := range roles {
		role.Users = nil
	}

	return roles
}
