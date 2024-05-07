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

	"github.com/casdoor/casdoor/conf"

	"github.com/casdoor/casdoor/util"
	"github.com/xorm-io/core"
)

type Role struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`
	DisplayName string `xorm:"varchar(100)" json:"displayName"`
	Description string `xorm:"varchar(100)" json:"description"`

	Users     []string `xorm:"mediumtext" json:"users"`
	Groups    []string `xorm:"mediumtext" json:"groups"`
	Roles     []string `xorm:"mediumtext" json:"roles"`
	Domains   []string `xorm:"mediumtext" json:"domains"`
	IsEnabled bool     `json:"isEnabled"`
}

func GetRoleCount(owner, field, value string) (int64, error) {
	session := GetSession(owner, -1, -1, field, value, "", "")
	return session.Count(&Role{})
}

func GetRoles(owner string) ([]*Role, error) {
	roles := []*Role{}
	err := ormer.Engine.Desc("created_time").Find(&roles, &Role{Owner: owner})
	if err != nil {
		return roles, err
	}

	return roles, nil
}

func GetPaginationRoles(owner string, offset, limit int, field, value, sortField, sortOrder string) ([]*Role, error) {
	roles := []*Role{}
	session := GetSession(owner, offset, limit, field, value, sortField, sortOrder)
	err := session.Find(&roles)
	if err != nil {
		return roles, err
	}

	return roles, nil
}

func getRole(owner string, name string) (*Role, error) {
	if owner == "" || name == "" {
		return nil, nil
	}

	role := Role{Owner: owner, Name: name}
	existed, err := ormer.Engine.Get(&role)
	if err != nil {
		return &role, err
	}

	if existed {
		return &role, nil
	} else {
		return nil, nil
	}
}

func GetRole(id string) (*Role, error) {
	owner, name := util.GetOwnerAndNameFromIdNoCheck(id)
	return getRole(owner, name)
}

func UpdateRole(id string, role *Role) (bool, error) {
	owner, name := util.GetOwnerAndNameFromIdNoCheck(id)
	oldRole, err := getRole(owner, name)
	if err != nil {
		return false, err
	}

	if oldRole == nil {
		return false, nil
	}

	visited := map[string]struct{}{}

	permissions, err := GetPermissionsByRole(id)
	if err != nil {
		return false, err
	}

	for _, permission := range permissions {
		removeGroupingPolicies(permission)
		removePolicies(permission)
		visited[permission.GetId()] = struct{}{}
	}

	ancestorRoles, err := GetAncestorRoles(id)
	if err != nil {
		return false, err
	}

	for _, r := range ancestorRoles {
		permissions, err := GetPermissionsByRole(r.GetId())
		if err != nil {
			return false, err
		}

		for _, permission := range permissions {
			permissionId := permission.GetId()
			if _, ok := visited[permissionId]; !ok {
				removeGroupingPolicies(permission)
				visited[permissionId] = struct{}{}
			}
		}
	}

	if name != role.Name {
		err := roleChangeTrigger(name, role.Name)
		if err != nil {
			return false, err
		}
	}

	affected, err := ormer.Engine.ID(core.PK{owner, name}).AllCols().Update(role)
	if err != nil {
		return false, err
	}

	visited = map[string]struct{}{}
	newRoleID := role.GetId()
	permissions, err = GetPermissionsByRole(newRoleID)
	if err != nil {
		return false, err
	}

	for _, permission := range permissions {
		err = addGroupingPolicies(permission)
		if err != nil {
			return false, err
		}

		err = addPolicies(permission)
		if err != nil {
			return false, err
		}

		visited[permission.GetId()] = struct{}{}
	}

	ancestorRoles, err = GetAncestorRoles(newRoleID)
	if err != nil {
		return false, err
	}

	for _, r := range ancestorRoles {
		permissions, err := GetPermissionsByRole(r.GetId())
		if err != nil {
			return false, err
		}

		for _, permission := range permissions {
			permissionId := permission.GetId()
			if _, ok := visited[permissionId]; !ok {
				err = addGroupingPolicies(permission)
				if err != nil {
					return false, err
				}

				visited[permissionId] = struct{}{}
			}
		}
	}

	return affected != 0, nil
}

func AddRole(role *Role) (bool, error) {
	affected, err := ormer.Engine.Insert(role)
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func AddRoles(roles []*Role) bool {
	if len(roles) == 0 {
		return false
	}
	affected, err := ormer.Engine.Insert(roles)
	if err != nil {
		if !strings.Contains(err.Error(), "Duplicate entry") {
			panic(err)
		}
	}
	return affected != 0
}

func AddRolesInBatch(roles []*Role) bool {
	batchSize := conf.GetConfigBatchSize()

	if len(roles) == 0 {
		return false
	}

	affected := false
	for i := 0; i < len(roles); i += batchSize {
		start := i
		end := i + batchSize
		if end > len(roles) {
			end = len(roles)
		}

		tmp := roles[start:end]
		fmt.Printf("The syncer adds roles: [%d - %d]\n", start, end)
		if AddRoles(tmp) {
			affected = true
		}
	}

	return affected
}

func deleteRole(role *Role) (bool, error) {
	affected, err := ormer.Engine.ID(core.PK{role.Owner, role.Name}).Delete(&Role{})
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func DeleteRole(role *Role) (bool, error) {
	roleId := role.GetId()
	permissions, err := GetPermissionsByRole(roleId)
	if err != nil {
		return false, err
	}

	for _, permission := range permissions {
		permission.Roles = util.DeleteVal(permission.Roles, roleId)
		_, err := UpdatePermission(permission.GetId(), permission)
		if err != nil {
			return false, err
		}
	}

	return deleteRole(role)
}

func (role *Role) GetId() string {
	return fmt.Sprintf("%s/%s", role.Owner, role.Name)
}

func getRolesByUserInternal(userId string) ([]*Role, error) {
	user, err := GetUser(userId)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, fmt.Errorf("The user: %s doesn't exist", userId)
	}

	query := ormer.Engine.Alias("r").Where("r.users like ?", fmt.Sprintf("%%%s%%", userId))
	for _, group := range user.Groups {
		query = query.Or("r.groups like ?", fmt.Sprintf("%%%s%%", group))
	}

	roles := []*Role{}
	err = query.Find(&roles)
	if err != nil {
		return nil, err
	}

	res := []*Role{}
	for _, role := range roles {
		if util.InSlice(role.Users, userId) || util.HaveIntersection(role.Groups, user.Groups) {
			res = append(res, role)
		}
	}
	return res, nil
}

func getRolesByUser(userId string) ([]*Role, error) {
	roles, err := getRolesByUserInternal(userId)
	if err != nil {
		return nil, err
	}

	allRolesIds := []string{}
	for _, role := range roles {
		allRolesIds = append(allRolesIds, role.GetId())
	}

	allRoles, err := GetAncestorRoles(allRolesIds...)
	if err != nil {
		return nil, err
	}

	for i := range allRoles {
		allRoles[i].Users = nil
	}

	return allRoles, nil
}

func roleChangeTrigger(oldName string, newName string) error {
	session := ormer.Engine.NewSession()
	defer session.Close()

	err := session.Begin()
	if err != nil {
		return err
	}

	var roles []*Role
	err = ormer.Engine.Find(&roles)
	if err != nil {
		return err
	}

	for _, role := range roles {
		for j, u := range role.Roles {
			owner, name := util.GetOwnerAndNameFromId(u)
			if name == oldName {
				role.Roles[j] = util.GetId(owner, newName)
			}
		}
		_, err = session.Where("name=?", role.Name).And("owner=?", role.Owner).Update(role)
		if err != nil {
			return err
		}
	}

	var permissions []*Permission
	err = ormer.Engine.Find(&permissions)
	if err != nil {
		return err
	}

	for _, permission := range permissions {
		for j, u := range permission.Roles {
			// u = organization/username
			owner, name := util.GetOwnerAndNameFromId(u)
			if name == oldName {
				permission.Roles[j] = util.GetId(owner, newName)
			}
		}
		_, err = session.Where("name=?", permission.Name).And("owner=?", permission.Owner).Update(permission)
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

// GetAncestorRoles returns a list of roles that contain the given roleIds
func GetAncestorRoles(roleIds ...string) ([]*Role, error) {
	if len(roleIds) == 0 {
		return []*Role{}, nil
	}

	visited := map[string]bool{}
	for _, roleId := range roleIds {
		visited[roleId] = true
	}

	owner, _ := util.GetOwnerAndNameFromIdNoCheck(roleIds[0])

	allRoles, err := GetRoles(owner)
	if err != nil {
		return nil, err
	}

	roleMap := map[string]*Role{}
	for _, r := range allRoles {
		roleMap[r.GetId()] = r
	}

	// find all the roles that contain father roles
	res := []*Role{}
	for _, r := range allRoles {
		isContain, ok := visited[r.GetId()]
		if isContain {
			res = append(res, r)
		} else if !ok {
			rId := r.GetId()
			visited[rId] = containsRole(r, roleMap, visited, roleIds...)
			if visited[rId] {
				res = append(res, r)
			}
		}
	}
	return res, nil
}

// containsRole is a helper function to check if a roles is related to any role in the given list roles
func containsRole(role *Role, roleMap map[string]*Role, visited map[string]bool, roleIds ...string) bool {
	roleId := role.GetId()
	if isContain, ok := visited[roleId]; ok {
		return isContain
	}

	visited[role.GetId()] = false

	for _, subRole := range role.Roles {
		if util.HasString(roleIds, subRole) {
			return true
		}

		r, ok := roleMap[subRole]
		if ok && containsRole(r, roleMap, visited, roleIds...) {
			return true
		}
	}

	return false
}
