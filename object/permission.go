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
	"github.com/casdoor/casdoor/util"
	"xorm.io/core"
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

//// checkPermissionValid verifies if the permission is valid
//func checkPermissionValid(permission *Permission) {
//	enforcer := getEnforcer(permission)
//	enforcer.EnableAutoSave(false)
//	policies, groupingPolicies := getPolicies(permission)
//
//	if len(groupingPolicies) > 0 {
//		_, err := enforcer.AddGroupingPolicies(groupingPolicies)
//		if err != nil {
//			panic(err)
//		}
//	}
//
//	_, err := enforcer.AddPolicies(policies)
//	if err != nil {
//		panic(err)
//	}
//}

func UpdatePermission(id string, permission *Permission) bool {
	//checkPermissionValid(permission)
	owner, name := util.GetOwnerAndNameFromId(id)
	oldPermission := getPermission(owner, name)
	if oldPermission == nil {
		return false
	}

	oldEnforcer := getEnforcer(oldPermission)
	oldIndex := 1
	if len(oldPermission.Domains) > 0 {
		oldIndex = 2
	}

	newEnforcer := getEnforcer(permission)
	//newIndex := 1
	//if len(permission.Domains) > 0 {
	//	newIndex = 2
	//}

	//If the adapter is modified, move the data to the new adapter
	if oldPermission.Adapter != permission.Adapter {
		permissions := GetPermissionsByAdapterAndDomainsAndRole(oldPermission.Adapter, oldPermission.Domains, "")
		//If only one permission uses the adapter, remove the GroupingPolicy directly.
		if len(permissions) == 1 {
			for _, role := range oldPermission.Roles {
				RemoveGroupingPolicyByDomains(oldEnforcer, oldPermission.Domains, oldIndex, role)
			}
		} else {
			//If there are multiple permissions using the adapter, determine whether the elements in oldPermission.Roles are referenced in other permissions, and if so, do not delete them.
			judgeRepeatRole(oldEnforcer, oldPermission.Roles, oldPermission.Domains, oldIndex, permissions)
		}

		for _, resource := range oldPermission.Resources {
			_, err := oldEnforcer.RemoveFilteredNamedPolicy("p", oldIndex, resource)
			if err != nil {
				panic(err)
			}
		}

		addGroupingPolicies(permission)
		addPolicies(permission)

		affected, err := adapter.Engine.ID(core.PK{owner, name}).AllCols().Update(permission)
		if err != nil {
			panic(err)
		}

		return affected != 0
	}

	usersAdded, usersDeleted := util.Arrcmp(oldPermission.Users, permission.Users)
	rolesAdded, rolesDeleted := util.Arrcmp(oldPermission.Roles, permission.Roles)
	domainsAdded, domainsDeleted := util.Arrcmp(oldPermission.Domains, permission.Domains)
	resourcesAdded, resourcesDeleted := util.Arrcmp(oldPermission.Resources, permission.Resources)
	actionsAdded, actionsDeleted := util.Arrcmp(oldPermission.Actions, permission.Actions)

	if len(domainsDeleted) > 0 {
		permissions := GetPermissionsByAdapterAndDomainsAndRole(oldPermission.Adapter, domainsDeleted, "")
		if len(permissions) == 1 {
			for _, role := range oldPermission.Roles {
				RemoveGroupingPolicyByDomains(oldEnforcer, domainsDeleted, oldIndex, role)
			}
		} else {
			judgeRepeatRole(oldEnforcer, oldPermission.Roles, domainsDeleted, oldIndex, permissions)
		}

		for _, domain := range domainsDeleted {
			for _, resource := range oldPermission.Resources {
				_, err := oldEnforcer.RemoveFilteredNamedPolicy("p", oldIndex-1, domain, resource)
				if err != nil {
					panic(err)
				}
			}
		}

		//If permissions are modified and Domains are [] regenerate GroupingPolicies and Policies
		if len(permission.Domains) == 0 {
			addGroupingPolicies(permission)
			addPolicies(permission)
		}

	}

	if len(domainsAdded) > 0 {
		//If oldPermission.Domains was originally [], delete the original GroupingPolicy and Policy after adding the new domain.
		if len(oldPermission.Domains) == 0 {
			permissions := GetPermissionsByAdapterAndDomainsAndRole(oldPermission.Adapter, []string{}, "")
			if len(permissions) == 1 {
				for _, role := range oldPermission.Roles {
					RemoveGroupingPolicyByDomains(oldEnforcer, []string{}, oldIndex, role)
				}
			} else {
				judgeRepeatRole(oldEnforcer, oldPermission.Roles, domainsAdded, oldIndex, permissions)
			}

			for _, resource := range oldPermission.Resources {
				_, err := oldEnforcer.RemoveFilteredNamedPolicy("p", oldIndex, resource)
				if err != nil {
					panic(err)
				}
			}
		}

		permissionMock := &Permission{
			Owner:     permission.Owner,
			Name:      permission.Name,
			Users:     permission.Users,
			Roles:     permission.Roles,
			Domains:   domainsAdded,
			Resources: permission.Resources,
			Actions:   permission.Actions,
		}
		operateGroupingPoliciesByPermission(permissionMock, newEnforcer, true)

		operatePoliciesByPermission(permissionMock, newEnforcer, true, false)
		operatePoliciesByPermission(permissionMock, newEnforcer, true, true)
	}

	if len(usersDeleted) > 0 {
		permissionMock := &Permission{
			Owner:     oldPermission.Owner,
			Name:      oldPermission.Name,
			Users:     usersDeleted,
			Roles:     oldPermission.Roles,
			Resources: oldPermission.Resources,
			Actions:   oldPermission.Actions,
			Domains:   oldPermission.Domains,
		}
		operatePoliciesByPermission(permissionMock, oldEnforcer, false, true)

	}

	if len(usersAdded) > 0 {
		permissionMock := &Permission{
			Owner:     permission.Owner,
			Name:      permission.Name,
			Users:     usersAdded,
			Roles:     permission.Roles,
			Resources: permission.Resources,
			Actions:   permission.Actions,
			Domains:   permission.Domains,
		}
		operatePoliciesByPermission(permissionMock, newEnforcer, true, true)
	}

	if len(rolesDeleted) > 0 {

		for _, role := range rolesDeleted {
			permissions := GetPermissionsByAdapterAndDomainsAndRole(oldPermission.Adapter, oldPermission.Domains, role)
			var num int
			for _, p := range permissions {
				if ok, _ := util.InArray(role, p.Roles); ok {
					num++
					if num > 1 {
						break
					}
				}
			}
			if num <= 1 {
				RemoveGroupingPolicyByDomains(oldEnforcer, oldPermission.Domains, oldIndex, role)
			}
		}

		permissionMock := &Permission{
			Owner:     oldPermission.Owner,
			Name:      oldPermission.Name,
			Users:     oldPermission.Users,
			Roles:     rolesDeleted,
			Resources: oldPermission.Resources,
			Actions:   oldPermission.Actions,
			Domains:   oldPermission.Domains,
		}
		operatePoliciesByPermission(permissionMock, oldEnforcer, false, false)

	}

	if len(rolesAdded) > 0 {

		permissionMock := &Permission{
			Owner:     permission.Owner,
			Name:      permission.Name,
			Roles:     rolesAdded,
			Domains:   permission.Domains,
			Resources: permission.Resources,
			Actions:   permission.Actions,
		}

		operateGroupingPoliciesByPermission(permissionMock, newEnforcer, true)

		operatePoliciesByPermission(permissionMock, newEnforcer, true, false)

	}

	if len(resourcesDeleted) > 0 {
		for _, resource := range resourcesDeleted {
			_, err := oldEnforcer.RemoveFilteredNamedPolicy("p", oldIndex, resource)
			if err != nil {
				panic(err)
			}
		}
	}

	if len(resourcesAdded) > 0 {
		permissionMock := &Permission{
			Owner:     permission.Owner,
			Name:      permission.Name,
			Users:     permission.Users,
			Roles:     permission.Roles,
			Domains:   permission.Domains,
			Resources: resourcesAdded,
			Actions:   permission.Actions,
		}
		operatePoliciesByPermission(permissionMock, newEnforcer, true, false)
		operatePoliciesByPermission(permissionMock, newEnforcer, true, true)

	}

	if len(actionsDeleted) > 0 {
		for _, resource := range oldPermission.Resources {
			for _, action := range actionsDeleted {
				_, err := oldEnforcer.RemoveFilteredNamedPolicy("p", oldIndex, resource, action)
				if err != nil {
					panic(err)
				}
			}
		}

	}

	if len(actionsAdded) > 0 {

		permissionMock := &Permission{
			Owner:     permission.Owner,
			Name:      permission.Name,
			Users:     permission.Users,
			Roles:     permission.Roles,
			Domains:   permission.Domains,
			Resources: permission.Resources,
			Actions:   actionsAdded,
		}
		operatePoliciesByPermission(permissionMock, newEnforcer, true, false)
		operatePoliciesByPermission(permissionMock, newEnforcer, true, true)
	}

	affected, err := adapter.Engine.ID(core.PK{owner, name}).AllCols().Update(permission)
	if err != nil {
		panic(err)
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

func (permission *Permission) GetId() string {
	return fmt.Sprintf("%s/%s", permission.Owner, permission.Name)
}

func GetPermissionsByUser(userId string) []*Permission {
	permissions := []*Permission{}
	err := adapter.Engine.Where("users like ?", "%"+userId+"%").Find(&permissions)
	if err != nil {
		panic(err)
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

func GetPermissionsByAdapterAndDomainsAndRole(table string, domains []string, role string) []*Permission {
	permissions := []*Permission{}
	where := "adapter = " + "'" + table + "'"

	if l := len(domains); l > 0 {
		domainsWhere := make([]string, l)
		for k, v := range domains {
			domainsWhere[k] = "domains like " + "'%" + v + "%'"
		}
		orWhere := "(" + strings.Join(domainsWhere, " or ") + ")"
		where += " and " + orWhere
	} else {
		orWhere := "domains = ''"
		where += " and " + orWhere
	}

	if role != "" {
		orWhere := " roles like " + "'%" + role + "%'"
		where += " and " + orWhere
	}

	err := adapter.Engine.Where(where).Find(&permissions)
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

func MigratePermissionRule() {
	models := []*Model{}
	err := adapter.Engine.Find(&models, &Model{})
	if err != nil {
		panic(err)
	}

	isHit := false
	for _, model := range models {
		if strings.Contains(model.ModelText, "permission") {
			// update model table
			model.ModelText = strings.Replace(model.ModelText, "permission,", "", -1)
			UpdateModel(model.GetId(), model)
			isHit = true
		}
	}

	if isHit {
		// update permission_rule table
		sql := "UPDATE `permission_rule`SET V0 = V1, V1 = V2, V2 = V3, V3 = V4, V4 = V5 WHERE V0 IN (SELECT CONCAT(owner, '/', name) AS permission_id FROM `permission`)"
		_, err = adapter.Engine.Exec(sql)
		if err != nil {
			return
		}
	}
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

func RemoveGroupingPolicyByDomains(enforcer *casbin.Enforcer, domains []string, index int, roleName string) {
	if len(domains) > 0 {
		for _, domain := range domains {
			_, err := enforcer.RemoveFilteredGroupingPolicy(index-1, domain, roleName)
			if err != nil {
				panic(err)
			}
		}
	} else {
		_, err := enforcer.RemoveFilteredGroupingPolicy(index, roleName)
		if err != nil {
			panic(err)
		}
	}
}

func judgeRepeatRole(enforcer *casbin.Enforcer, roles []string, domains []string, index int, permissions []*Permission) {
	for _, role := range roles {
		var num int
		for _, p := range permissions {
			if ok, _ := util.InArray(role, p.Roles); ok {
				num++
				if num > 1 {
					break
				}
			}
		}
		if num <= 1 {
			RemoveGroupingPolicyByDomains(enforcer, domains, index, role)
		}
	}
}

func operateGroupingPoliciesByPermission(permission *Permission, enforcer *casbin.Enforcer, isAdd bool) {
	var err error
	domainExist := len(permission.Domains) > 0
	permissionId := permission.Owner + "/" + permission.Name
	for _, role := range permission.Roles {
		roleObj := GetRole(role)
		for _, user := range roleObj.Users {
			if domainExist {
				for _, domain := range permission.Domains {
					if isAdd {
						_, err = enforcer.AddNamedGroupingPolicy("g", user, domain, roleObj.Owner+"/"+roleObj.Name, "", "", permissionId)
					} else {
						_, err = enforcer.RemoveNamedGroupingPolicy("g", user, domain, roleObj.Owner+"/"+roleObj.Name, "", "", permissionId)
					}
					if err != nil {
						panic(err)
					}
				}
			} else {
				if isAdd {
					_, err = enforcer.AddNamedGroupingPolicy("g", user, roleObj.Owner+"/"+roleObj.Name, "", "", "", permissionId)
				} else {
					_, err = enforcer.RemoveNamedGroupingPolicy("g", user, roleObj.Owner+"/"+roleObj.Name, "", "", "", permissionId)
				}
				if err != nil {
					panic(err)
				}
			}
		}
	}
}

func operatePoliciesByPermission(permission *Permission, enforcer *casbin.Enforcer, isAdd bool, isUser bool) {
	var err error
	permissionId := permission.Owner + "/" + permission.Name
	column := permission.Roles
	if isUser {
		column = permission.Users
	}
	domainExist := len(permission.Domains) > 0
	for _, v := range column {
		for _, resource := range permission.Resources {
			for _, action := range permission.Actions {
				if domainExist {
					for _, domain := range permission.Domains {
						if isAdd {
							_, err = enforcer.AddNamedPolicy("p", v, domain, resource, strings.ToLower(action), "", permissionId)
						} else {
							_, err = enforcer.RemoveNamedPolicy("p", v, domain, resource, strings.ToLower(action), "", permissionId)
						}
						if err != nil {
							panic(err)
						}
					}
				} else {
					if isAdd {
						_, err = enforcer.AddNamedPolicy("p", v, resource, action, "", "", permissionId)
					} else {
						_, err = enforcer.RemoveNamedPolicy("p", v, resource, action, "", "", permissionId)
					}
					if err != nil {
						panic(err)
					}
				}
			}
		}
	}
}

func GetMaskedPermissions(permissions []*Permission) []*Permission {
	for _, permission := range permissions {
		permission.Users = nil
		permission.Submitter = ""
	}

	return permissions

}
