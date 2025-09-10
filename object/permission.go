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

type Permission struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`
	DisplayName string `xorm:"varchar(100)" json:"displayName"`
	Description string `xorm:"varchar(100)" json:"description"`

	Users   []string `xorm:"mediumtext" json:"users"`
	Groups  []string `xorm:"mediumtext" json:"groups"`
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

const builtInMaxFields = 6 // Casdoor built-in adapter, use V5 to filter permission, so has 6 max field

func GetPermissionCount(owner, field, value string) (int64, error) {
	session := GetSession(owner, -1, -1, field, value, "", "")
	return session.Count(&Permission{})
}

func GetPermissions(owner string) ([]*Permission, error) {
	permissions := []*Permission{}
	err := ormer.Engine.Desc("created_time").Find(&permissions, &Permission{Owner: owner})
	if err != nil {
		return permissions, err
	}

	return permissions, nil
}

func GetPaginationPermissions(owner string, offset, limit int, field, value, sortField, sortOrder string) ([]*Permission, error) {
	permissions := []*Permission{}
	session := GetSession(owner, offset, limit, field, value, sortField, sortOrder)
	err := session.Find(&permissions)
	if err != nil {
		return permissions, err
	}

	return permissions, nil
}

func getPermission(owner string, name string) (*Permission, error) {
	if owner == "" || name == "" {
		return nil, nil
	}

	permission := Permission{Owner: owner, Name: name}
	existed, err := ormer.Engine.Get(&permission)
	if err != nil {
		return &permission, err
	}

	if existed {
		return &permission, nil
	} else {
		return nil, nil
	}
}

func GetPermission(id string) (*Permission, error) {
	owner, name := util.GetOwnerAndNameFromIdNoCheck(id)
	return getPermission(owner, name)
}

// checkPermissionValid verifies if the permission is valid
func checkPermissionValid(permission *Permission) error {
	enforcer, err := getPermissionEnforcer(permission)
	if err != nil {
		return err
	}

	enforcer.EnableAutoSave(false)

	policies := getPolicies(permission)
	_, err = enforcer.AddPolicies(policies)
	if err != nil {
		return err
	}

	if !HasRoleDefinition(enforcer.GetModel()) {
		permission.Roles = []string{}
		return nil
	}

	groupingPolicies, err := getGroupingPolicies(permission)
	if err != nil {
		return err
	}

	if len(groupingPolicies) > 0 {
		_, err = enforcer.AddGroupingPolicies(groupingPolicies)
		if err != nil {
			return err
		}
	}

	return nil
}

func UpdatePermission(id string, permission *Permission) (bool, error) {
	err := checkPermissionValid(permission)
	if err != nil {
		return false, err
	}

	owner, name := util.GetOwnerAndNameFromIdNoCheck(id)
	oldPermission, err := getPermission(owner, name)
	if oldPermission == nil {
		return false, nil
	}

	if permission.ResourceType == "Application" && permission.Model != "" {
		model, err := GetModelEx(util.GetId(permission.Owner, permission.Model))
		if err != nil {
			return false, err
		} else if model == nil {
			return false, fmt.Errorf("the model: %s for permission: %s is not found", permission.Model, permission.GetId())
		}

		modelCfg, err := getModelCfg(model)
		if err != nil {
			return false, err
		}

		if len(strings.Split(modelCfg["p"], ",")) != 3 {
			return false, fmt.Errorf("the model: %s for permission: %s is not valid, Casbin model's [policy_defination] section should have 3 elements", permission.Model, permission.GetId())
		}
	}

	affected, err := ormer.Engine.ID(core.PK{owner, name}).AllCols().Update(permission)
	if err != nil {
		return false, err
	}

	if affected != 0 {
		err = removeGroupingPolicies(oldPermission)
		if err != nil {
			return false, err
		}

		err = removePolicies(oldPermission)
		if err != nil {
			return false, err
		}

		// if oldPermission.Adapter != "" && oldPermission.Adapter != permission.Adapter {
		// 	isEmpty, _ := ormer.Engine.IsTableEmpty(oldPermission.Adapter)
		// 	if isEmpty {
		// 		err = ormer.Engine.DropTables(oldPermission.Adapter)
		// 		if err != nil {
		// 			return false, err
		// 		}
		// 	}
		// }

		err = addGroupingPolicies(permission)
		if err != nil {
			return false, err
		}

		err = addPolicies(permission)
		if err != nil {
			return false, err
		}
	}

	return affected != 0, nil
}

func AddPermission(permission *Permission) (bool, error) {
	affected, err := ormer.Engine.Insert(permission)
	if err != nil {
		return false, err
	}

	if affected != 0 {
		err = addGroupingPolicies(permission)
		if err != nil {
			return false, err
		}

		err = addPolicies(permission)
		if err != nil {
			return false, err
		}
	}

	return affected != 0, nil
}

func AddPermissions(permissions []*Permission) (bool, error) {
	if len(permissions) == 0 {
		return false, nil
	}

	affected, err := ormer.Engine.Insert(permissions)
	if err != nil {
		if !strings.Contains(err.Error(), "Duplicate entry") {
			return false, err
		}
	}

	for _, permission := range permissions {
		// add using for loop
		if affected != 0 {
			err = addGroupingPolicies(permission)
			if err != nil {
				return false, err
			}

			err = addPolicies(permission)
			if err != nil {
				return false, err
			}
		}
	}
	return affected != 0, nil
}

func AddPermissionsInBatch(permissions []*Permission) (bool, error) {
	batchSize := conf.GetConfigBatchSize()

	if len(permissions) == 0 {
		return false, nil
	}

	affected := false
	for i := 0; i < len(permissions); i += batchSize {
		start := i
		end := i + batchSize
		if end > len(permissions) {
			end = len(permissions)
		}

		tmp := permissions[start:end]
		fmt.Printf("The syncer adds permissions: [%d - %d]\n", start, end)

		b, err := AddPermissions(tmp)
		if err != nil {
			return false, err
		}

		if b {
			affected = true
		}
	}

	return affected, nil
}

func deletePermission(permission *Permission) (bool, error) {
	affected, err := ormer.Engine.ID(core.PK{permission.Owner, permission.Name}).Delete(&Permission{})
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func DeletePermission(permission *Permission) (bool, error) {
	affected, err := deletePermission(permission)
	if err != nil {
		return false, err
	}

	if affected {
		err = removeGroupingPolicies(permission)
		if err != nil {
			return false, err
		}

		err = removePolicies(permission)
		if err != nil {
			return false, err
		}

		// if permission.Adapter != "" && permission.Adapter != "permission_rule" {
		// 	isEmpty, _ := ormer.Engine.IsTableEmpty(permission.Adapter)
		// 	if isEmpty {
		// 		err = ormer.Engine.DropTables(permission.Adapter)
		// 		if err != nil {
		// 			return false, err
		// 		}
		// 	}
		// }
	}

	return affected, nil
}

func getPermissionsByUser(userId string) ([]*Permission, error) {
	permissions := []*Permission{}
	err := ormer.Engine.Where("users like ?", "%"+userId+"\"%").Find(&permissions)
	if err != nil {
		return permissions, err
	}

	res := []*Permission{}
	for _, permission := range permissions {
		if util.InSlice(permission.Users, userId) {
			res = append(res, permission)
		}
	}

	return res, nil
}

func GetPermissionsByRole(roleId string) ([]*Permission, error) {
	permissions := []*Permission{}
	err := ormer.Engine.Where("roles like ?", "%"+roleId+"\"%").Find(&permissions)
	if err != nil {
		return permissions, err
	}

	res := []*Permission{}
	for _, permission := range permissions {
		if util.InSlice(permission.Roles, roleId) {
			res = append(res, permission)
		}
	}

	return res, nil
}

func GetPermissionsByResource(resourceId string) ([]*Permission, error) {
	permissions := []*Permission{}
	err := ormer.Engine.Where("resources like ?", "%"+resourceId+"\"%").Find(&permissions)
	if err != nil {
		return permissions, err
	}

	res := []*Permission{}
	for _, permission := range permissions {
		if util.InSlice(permission.Resources, resourceId) {
			res = append(res, permission)
		}
	}

	return res, nil
}

func getPermissionsAndRolesByUser(userId string) ([]*Permission, []*Role, error) {
	permissions, err := getPermissionsByUser(userId)
	if err != nil {
		return nil, nil, err
	}

	existedPerms := map[string]struct{}{}

	for _, perm := range permissions {
		perm.Users = nil

		if _, ok := existedPerms[perm.Name]; !ok {
			existedPerms[perm.Name] = struct{}{}
		}
	}

	permFromRoles := []*Permission{}

	roles, err := getRolesByUser(userId)
	if err != nil {
		return nil, nil, err
	}

	for _, role := range roles {
		perms, err := GetPermissionsByRole(role.GetId())
		if err != nil {
			return nil, nil, err
		}
		permFromRoles = append(permFromRoles, perms...)
	}

	for _, perm := range permFromRoles {
		perm.Users = nil
		if _, ok := existedPerms[perm.Name]; !ok {
			existedPerms[perm.Name] = struct{}{}
			permissions = append(permissions, perm)
		}
	}

	return permissions, roles, nil
}

func GetPermissionsBySubmitter(owner string, submitter string) ([]*Permission, error) {
	permissions := []*Permission{}
	err := ormer.Engine.Desc("created_time").Find(&permissions, &Permission{Owner: owner, Submitter: submitter})
	if err != nil {
		return permissions, err
	}

	return permissions, nil
}

func GetPermissionsByModel(owner string, model string) ([]*Permission, error) {
	permissions := []*Permission{}
	err := ormer.Engine.Desc("created_time").Find(&permissions, &Permission{Owner: owner, Model: model})
	if err != nil {
		return permissions, err
	}

	return permissions, nil
}

func GetMaskedPermissions(permissions []*Permission) []*Permission {
	for _, permission := range permissions {
		permission.Users = nil
		permission.Submitter = ""
	}

	return permissions
}

// GroupPermissionsByModelAdapter group permissions by model and adapter.
// Every model and adapter will be a key, and the value is a list of permission ids.
// With each list of permission ids have the same key, we just need to init the
// enforcer and do the enforce/batch-enforce once (with list of permission ids
// as the policyFilter when the enforcer load policy).
func GroupPermissionsByModelAdapter(permissions []*Permission) map[string][]string {
	m := make(map[string][]string)
	for _, permission := range permissions {
		key := permission.GetModelAndAdapter()
		permissionIds, ok := m[key]
		if !ok {
			m[key] = []string{permission.GetId()}
		} else {
			m[key] = append(permissionIds, permission.GetId())
		}
	}

	return m
}

func (p *Permission) GetId() string {
	return util.GetId(p.Owner, p.Name)
}

func (p *Permission) GetModelAndAdapter() string {
	return util.GetId(p.Model, p.Adapter)
}

func (p *Permission) isUserHit(name string) bool {
	targetOrg, targetName := util.GetOwnerAndNameFromId(name)
	for _, user := range p.Users {
		if user == "*" {
			return true
		}

		userOrg, userName := util.GetOwnerAndNameFromId(user)
		if userOrg == targetOrg && (userName == "*" || userName == targetName) {
			return true
		}
	}
	return false
}

func (p *Permission) isRoleHit(userId string) bool {
	targetRoles, err := getRolesByUser(userId)
	if err != nil {
		return false
	}

	for _, role := range p.Roles {
		if role == "*" {
			return true
		}

		for _, targetRole := range targetRoles {
			if role == targetRole.GetId() {
				return true
			}
		}
	}
	return false
}

func (p *Permission) isResourceHit(name string) bool {
	for _, resource := range p.Resources {
		if resource == "*" || resource == name {
			return true
		}
	}
	return false
}
