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

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	xormadapter "github.com/casbin/xorm-adapter/v2"
	"github.com/casdoor/casdoor/conf"
	"github.com/casdoor/casdoor/util"
	"xorm.io/core"
)

type Permission struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`
	DisplayName string `xorm:"varchar(100)" json:"displayName"`

	Users []string `xorm:"mediumtext" json:"users"`
	Roles []string `xorm:"mediumtext" json:"roles"`

	Model        string   `xorm:"varchar(100)" json:"model"`
	ResourceType string   `xorm:"varchar(100)" json:"resourceType"`
	Resources    []string `xorm:"mediumtext" json:"resources"`
	Actions      []string `xorm:"mediumtext" json:"actions"`
	Effect       string   `xorm:"varchar(100)" json:"effect"`

	IsEnabled bool `json:"isEnabled"`
}

type PermissionRule struct {
	PType      string `xorm:"varchar(100) index not null default ''"`
	V0         string `xorm:"varchar(100) index not null default ''"`
	V1         string `xorm:"varchar(100) index not null default ''"`
	V2         string `xorm:"varchar(100) index not null default ''"`
	V3         string `xorm:"varchar(100) index not null default ''"`
	V4         string `xorm:"varchar(100) index not null default ''"`
	V5         string `xorm:"varchar(100) index not null default ''"`
}

func GetPermissionCount(owner, field, value string) int {
	session := GetSession(owner, -1, -1, field, value, "", "")
	count, err := session.Count(&Permission{})
	if err != nil {
		panic(err)
	}

	return int(count)
}

func GetPermissions(owner string) []*Permission {
	permissions := []*Permission{}
	err := adapter.Engine.Desc("created_time").Find(&permissions, &Permission{Owner: owner})
	if err != nil {
		panic(err)
	}

	return permissions
}

func GetPaginationPermissions(owner string, offset, limit int, field, value, sortField, sortOrder string) []*Permission {
	permissions := []*Permission{}
	session := GetSession(owner, offset, limit, field, value, sortField, sortOrder)
	err := session.Find(&permissions)
	if err != nil {
		panic(err)
	}

	return permissions
}

func getPermission(owner string, name string) *Permission {
	if owner == "" || name == "" {
		return nil
	}

	permission := Permission{Owner: owner, Name: name}
	existed, err := adapter.Engine.Get(&permission)
	if err != nil {
		panic(err)
	}

	if existed {
		return &permission
	} else {
		return nil
	}
}

func GetPermission(id string) *Permission {
	owner, name := util.GetOwnerAndNameFromId(id)
	return getPermission(owner, name)
}

func UpdatePermission(id string, permission *Permission) bool {
	owner, name := util.GetOwnerAndNameFromId(id)
	oldPermission := getPermission(owner, name)
	if oldPermission == nil {
		return false
	} else {
		removePolicies(oldPermission)
	}

	affected, err := adapter.Engine.ID(core.PK{owner, name}).AllCols().Update(permission)
	if err != nil {
		panic(err)
	}

	if affected != 0 {
		addPolicies(permission)
	}

	return affected != 0
}

func AddPermission(permission *Permission) bool {
	affected, err := adapter.Engine.Insert(permission)
	if err != nil {
		panic(err)
	}

	if affected != 0 {
		addPolicies(permission)
	}

	return affected != 0
}

func DeletePermission(permission *Permission) bool {
	affected, err := adapter.Engine.ID(core.PK{permission.Owner, permission.Name}).Delete(&Permission{})
	if err != nil {
		panic(err)
	}

	if affected != 0 {
		removePolicies(permission)
	}

	return affected != 0
}

func (permission *Permission) GetId() string {
	return fmt.Sprintf("%s/%s", permission.Owner, permission.Name)
}

func getEnforcer(permission *Permission) *casbin.Enforcer {
	tableNamePrefix := conf.GetConfigString("tableNamePrefix")
	adapter, err := xormadapter.NewAdapterWithTableName(conf.GetConfigString("driverName"), conf.GetBeegoConfDataSourceName()+conf.GetConfigString("dbName"), "permission_rule", tableNamePrefix, true)
	if err != nil {
		panic(err)
	}

	modelText := `
[request_definition]
r = sub, obj, act

[policy_definition]
p = permission, sub, obj, act

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = r.sub == p.sub && r.obj == p.obj && r.act == p.act`
	permissionModel := getModel(permission.Owner, permission.Model)
	if permissionModel != nil {
		modelText = permissionModel.ModelText
	}
	m, err := model.NewModelFromString(modelText)
	if err != nil {
		panic(err)
	}

	enforcer, err := casbin.NewEnforcer(m, adapter)
	if err != nil {
		panic(err)
	}

	err = enforcer.LoadFilteredPolicy(xormadapter.Filter{V0: []string{permission.GetId()}})
	if err != nil {
		panic(err)
	}

	return enforcer
}

func getPolicies(permission *Permission) [][]string {
	var policies [][]string
	for _, user := range permission.Users {
		for _, resource := range permission.Resources {
			for _, action := range permission.Actions {
				policies = append(policies, []string{permission.GetId(), user, resource, strings.ToLower(action)})
			}
		}
	}
	return policies
}

func addPolicies(permission *Permission) {
	enforcer := getEnforcer(permission)
	policies := getPolicies(permission)

	_, err := enforcer.AddPolicies(policies)
	if err != nil {
		panic(err)
	}
}

func removePolicies(permission *Permission) {
	enforcer := getEnforcer(permission)

	_, err := enforcer.RemoveFilteredPolicy(0, permission.GetId())
	if err != nil {
		panic(err)
	}
}