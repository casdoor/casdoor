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

	"github.com/casdoor/casdoor/util"
	"github.com/xorm-io/core"
)

type Permission struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`
	DisplayName string `xorm:"varchar(100)" json:"displayName"`

	Users   []string `xorm:"mediumtext" json:"users"`
	Roles   []string `xorm:"mediumtext" json:"roles"`
	Domains []string `xorm:"mediumtext" json:"domains"`

	Model        string   `xorm:"varchar(100)" json:"model"`
	Adapter      string   `xorm:"varchar(100)" json:"adapter"`
	ResourceType string   `xorm:"varchar(100)" json:"resourceType"`
	Resources    []string `xorm:"mediumtext" json:"resources"`
	Actions      []string `xorm:"mediumtext" json:"actions"`
	Effect       string   `xorm:"varchar(100)" json:"effect"`
	IsEnabled    bool     `json:"isEnabled"`

	Submitter   string `xorm:"varchar(100)" json:"submitter"`
	Approver    string `xorm:"varchar(100)" json:"approver"`
	ApproveTime string `xorm:"varchar(100)" json:"approveTime"`
	State       string `xorm:"varchar(100)" json:"state"`
}

type PermissionRule struct {
	Ptype string `xorm:"varchar(100) index not null default ''" json:"ptype"`
	V0    string `xorm:"varchar(100) index not null default ''" json:"v0"`
	V1    string `xorm:"varchar(100) index not null default ''" json:"v1"`
	V2    string `xorm:"varchar(100) index not null default ''" json:"v2"`
	V3    string `xorm:"varchar(100) index not null default ''" json:"v3"`
	V4    string `xorm:"varchar(100) index not null default ''" json:"v4"`
	V5    string `xorm:"varchar(100) index not null default ''" json:"v5"`
	Id    string `xorm:"varchar(100) index not null default ''" json:"id"`
}

const (
	builtInAvailableField = 5 // Casdoor built-in adapter, use V5 to filter permission, so has 5 available field
	builtInAdapter        = "permission_rule"
)

func (p *Permission) GetId() string {
	return util.GetId(p.Owner, p.Name)
}

func (p *PermissionRule) GetRequest(adapterName string, permissionId string) ([]interface{}, error) {
	request := []interface{}{p.V0, p.V1, p.V2}

	if p.V3 != "" {
		request = append(request, p.V3)
	}

	if p.V4 != "" {
		request = append(request, p.V4)
	}

	if adapterName == builtInAdapter {
		if p.V5 != "" {
			return nil, fmt.Errorf("too many parameters. The maximum parameter number cannot exceed %d", builtInAvailableField)
		}
		request = append(request, permissionId)
		return request, nil
	} else {
		if p.V5 != "" {
			request = append(request, p.V5)
		}
		return request, nil
	}
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

// checkPermissionValid verifies if the permission is valid
func checkPermissionValid(permission *Permission) {
	enforcer := getEnforcer(permission)
	enforcer.EnableAutoSave(false)

	policies := getPolicies(permission)
	_, err := enforcer.AddPolicies(policies)
	if err != nil {
		panic(err)
	}

	if !HasRoleDefinition(enforcer.GetModel()) {
		permission.Roles = []string{}
		return
	}

	groupingPolicies := getGroupingPolicies(permission)
	if len(groupingPolicies) > 0 {
		_, err := enforcer.AddGroupingPolicies(groupingPolicies)
		if err != nil {
			panic(err)
		}
	}
}

func UpdatePermission(id string, permission *Permission) bool {
	checkPermissionValid(permission)
	owner, name := util.GetOwnerAndNameFromId(id)
	oldPermission := getPermission(owner, name)
	if oldPermission == nil {
		return false
	}

	affected, err := adapter.Engine.ID(core.PK{owner, name}).AllCols().Update(permission)
	if err != nil {
		panic(err)
	}

	if affected != 0 {
		removeGroupingPolicies(oldPermission)
		removePolicies(oldPermission)
		if oldPermission.Adapter != "" && oldPermission.Adapter != permission.Adapter {
			isEmpty, _ := adapter.Engine.IsTableEmpty(oldPermission.Adapter)
			if isEmpty {
				err = adapter.Engine.DropTables(oldPermission.Adapter)
				if err != nil {
					panic(err)
				}
			}
		}
		addGroupingPolicies(permission)
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
		addGroupingPolicies(permission)
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
		removeGroupingPolicies(permission)
		removePolicies(permission)
		if permission.Adapter != "" && permission.Adapter != "permission_rule" {
			isEmpty, _ := adapter.Engine.IsTableEmpty(permission.Adapter)
			if isEmpty {
				err = adapter.Engine.DropTables(permission.Adapter)
				if err != nil {
					panic(err)
				}
			}
		}
	}

	return affected != 0
}

func GetPermissionsByUser(userId string) []*Permission {
	permissions := []*Permission{}
	err := adapter.Engine.Where("users like ?", "%"+userId+"%").Find(&permissions)
	if err != nil {
		panic(err)
	}

	for i := range permissions {
		permissions[i].Users = nil
	}

	return permissions
}

func GetPermissionsByRole(roleId string) []*Permission {
	permissions := []*Permission{}
	err := adapter.Engine.Where("roles like ?", "%"+roleId+"%").Find(&permissions)
	if err != nil {
		panic(err)
	}

	return permissions
}

func GetPermissionsBySubmitter(owner string, submitter string) []*Permission {
	permissions := []*Permission{}
	err := adapter.Engine.Desc("created_time").Find(&permissions, &Permission{Owner: owner, Submitter: submitter})
	if err != nil {
		panic(err)
	}

	return permissions
}

func ContainsAsterisk(userId string, users []string) bool {
	containsAsterisk := false
	group, _ := util.GetOwnerAndNameFromId(userId)
	for _, user := range users {
		permissionGroup, permissionUserName := util.GetOwnerAndNameFromId(user)
		if permissionGroup == group && permissionUserName == "*" {
			containsAsterisk = true
			break
		}
	}

	return containsAsterisk
}

func GetMaskedPermissions(permissions []*Permission) []*Permission {
	for _, permission := range permissions {
		permission.Users = nil
		permission.Submitter = ""
	}

	return permissions
}
