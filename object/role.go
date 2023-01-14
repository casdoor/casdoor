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
	"github.com/casbin/casbin/v2"
	"github.com/casdoor/casdoor/util"
	"strings"
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

func SetRoles(userId string, roles []string) bool {

	owner, name := util.GetOwnerAndNameFromId(userId)
	user := getUser(owner, name)
	if user == nil {
		return false
	}

	emap := make(map[string]*casbin.Enforcer)
	for _, roleId := range roles {
		owner, name = util.GetOwnerAndNameFromId(roleId)
		role := getRole(owner, name)
		if role == nil {
			continue
		}

		permissions := GetPermissionsByRole(roleId)
		for _, p := range permissions {
			key := p.Adapter + "/" + strings.Join(p.Domains, ",")
			if _, ok := emap[key]; !ok {
				emap[key] = getEnforcer(p)
			}
		}

		usersAddedGroupingPolicies := getGroupingPoliciesByPermissions([]string{userId}, role, permissions)
		for k, v := range usersAddedGroupingPolicies {
			enforcer := emap[k]
			_, err := enforcer.AddGroupingPolicies(v)
			if err != nil {
				panic(err)
			}
		}

		ok, _ := util.InArray(userId, role.Users)
		if !ok {
			role.Users = append(role.Users, userId)
		}
		affected, err := adapter.Engine.ID(core.PK{owner, name}).AllCols().Update(role)
		if err != nil {
			panic(err)
		}
		if affected == 0 {
			return false
		}
	}
	return true

}

// 1.这里需要将permission的信息也要update
// 2.修改角色，资源，动作对照表
// 3.如果同个适配器 会出现重复删的情况
//func UpdateRole(id string, role *Role) bool {
//	owner, name := util.GetOwnerAndNameFromId(id)
//	oldRole := getRole(owner, name)
//	if oldRole == nil {
//		return false
//	}
//
//	permissions := GetPermissionsByRole(id)
//	for _, p := range permissions {
//		removeGroupingPolicies(p)
//		removePolicies(p)
//
//		//判断role name或owner字段有没被改变,删除掉permission中roles的元素 重新添加进去
//		if owner != role.Owner || name != role.Name {
//			for k, v := range p.Roles {
//				if v == id {
//					p.Roles = append(p.Roles[:k], p.Roles[k+1:]...)
//					p.Roles = append(p.Roles, role.Owner+"/"+role.Name)
//					break
//				}
//			}
//		}
//
//		_, err := adapter.Engine.ID(core.PK{p.Owner, p.Name}).AllCols().Update(p)
//		if err != nil {
//			panic(err)
//		}
//	}
//
//	affected, err := adapter.Engine.ID(core.PK{owner, name}).AllCols().Update(role)
//	if err != nil {
//		panic(err)
//	}
//
//	for _, p := range permissions {
//		addGroupingPolicies(p)
//		addPolicies(p)
//	}
//
//	return affected != 0
//}

//func UpdateRole(id string, role *Role) bool {
//	owner, name := util.GetOwnerAndNameFromId(id)
//	oldRole := getRole(owner, name)
//	if oldRole == nil {
//		return false
//	}
//
//	permissions := GetPermissionsByRole(id)
//	//if len(permissions) == 0 {
//	//	return affected != 0
//	//}
//
//	//删除全部的p和g
//	//if id != role.Owner+"/"+role.Name || len(domainsAdded) > 0 || len(domainsDeleted) > 0 {
//	if id != role.Owner+"/"+role.Name {
//		for _, p := range permissions {
//			removeGroupingPolicies(p)
//			removePolicies(p)
//			for k, v := range p.Roles {
//				if v == id {
//					p.Roles = append(p.Roles[:k], p.Roles[k+1:]...)
//					p.Roles = append(p.Roles, role.Owner+"/"+role.Name)
//					break
//				}
//			}
//			_, err := adapter.Engine.ID(core.PK{p.Owner, p.Name}).AllCols().Update(p)
//			if err != nil {
//				panic(err)
//			}
//		}
//
//		affected, err := adapter.Engine.ID(core.PK{owner, name}).AllCols().Update(role)
//		if err != nil {
//			panic(err)
//		}
//
//		for _, p := range permissions {
//			addGroupingPolicies(p)
//			addPolicies(p)
//		}
//
//		return affected != 0
//
//	}
//
//	usersAdded, usersDeleted := util.Arrcmp(oldRole.Users, role.Users)
//	rolesAdded, rolesDeleted := util.Arrcmp(oldRole.Roles, role.Roles)
//
//	//emap := make(map[string]*casbin.Enforcer, len(permissions))
//	pmap := make(map[string]*Permission, len(permissions))
//	for _, p := range permissions {
//		//emap[p.Owner+"/"+p.Name] = getEnforcer(p)
//		pmap[p.Owner+"/"+p.Name] = p
//	}
//
//	if len(usersDeleted) > 0 {
//		for _, u := range usersDeleted {
//			for _, p := range pmap {
//				enforcer := getEnforcer(p)
//				enforcer.RemoveNamedGroupingPolicy("g", u)
//			}
//		}
//	}
//
//	if len(usersAdded) > 0 {
//		usersAddedGroupingPolicies := getGroupingPoliciesByColumn(usersAdded, role, permissions)
//		for k, v := range usersAddedGroupingPolicies {
//			enforcer := getEnforcer(pmap[k])
//			_, err := enforcer.AddGroupingPolicies(v)
//			if err != nil {
//				panic(err)
//			}
//		}
//	}
//
//	if len(rolesDeleted) > 0 {
//		for _, u := range rolesDeleted {
//			for _, p := range pmap {
//				enforcer := getEnforcer(p)
//				enforcer.RemoveNamedGroupingPolicy("g", u)
//			}
//		}
//	}
//
//	if len(rolesAdded) > 0 {
//		rolesAddedGroupingPolicies := getGroupingPoliciesByColumn(rolesAdded, role, permissions)
//		for k, v := range rolesAddedGroupingPolicies {
//			enforcer := getEnforcer(pmap[k])
//			_, err := enforcer.AddGroupingPolicies(v)
//			if err != nil {
//				panic(err)
//			}
//		}
//	}
//
//	affected, err := adapter.Engine.ID(core.PK{owner, name}).AllCols().Update(role)
//	if err != nil {
//		panic(err)
//	}
//
//	return affected != 0
//}

//func UpdateRole(id string, role *Role) bool {
//	owner, name := util.GetOwnerAndNameFromId(id)
//	oldRole := getRole(owner, name)
//	if oldRole == nil {
//		return false
//	}
//
//	permissions := GetPermissionsByRole(id)
//	//if len(permissions) == 0 {
//	//	return affected != 0
//	//}
//
//	//删除全部的p和g
//	//if id != role.Owner+"/"+role.Name || len(domainsAdded) > 0 || len(domainsDeleted) > 0 {
//	if id != role.Owner+"/"+role.Name {
//		for _, p := range permissions {
//			removeGroupingPolicies(p)
//			removePolicies(p)
//			for k, v := range p.Roles {
//				if v == id {
//					p.Roles = append(p.Roles[:k], p.Roles[k+1:]...)
//					p.Roles = append(p.Roles, role.Owner+"/"+role.Name)
//					break
//				}
//			}
//			_, err := adapter.Engine.ID(core.PK{p.Owner, p.Name}).AllCols().Update(p)
//			if err != nil {
//				panic(err)
//			}
//		}
//
//		affected, err := adapter.Engine.ID(core.PK{owner, name}).AllCols().Update(role)
//		if err != nil {
//			panic(err)
//		}
//
//		for _, p := range permissions {
//			addGroupingPolicies(p)
//			addPolicies(p)
//		}
//
//		return affected != 0
//
//	}
//
//	usersAdded, usersDeleted := util.Arrcmp(oldRole.Users, role.Users)
//	rolesAdded, rolesDeleted := util.Arrcmp(oldRole.Roles, role.Roles)
//
//	emap := make(map[string]*casbin.Enforcer, len(permissions))
//	for _, p := range permissions {
//		key := p.Adapter + "/" + strings.Join(p.Domains, ",")
//		if _, ok := emap[key]; !ok {
//			emap[key] = getEnforcer(p)
//		}
//	}
//
//	if len(usersDeleted) > 0 {
//		usersDeletedGroupingPolicies := getGroupingPoliciesByColumn(usersDeleted, role, permissions)
//		for k, v := range usersDeletedGroupingPolicies {
//			enforcer := emap[k]
//			_, err := enforcer.RemoveGroupingPolicies(v)
//			if err != nil {
//				panic(err)
//			}
//		}
//	}
//
//	if len(usersAdded) > 0 {
//		usersAddedGroupingPolicies := getGroupingPoliciesByColumn(usersAdded, role, permissions)
//		for k, v := range usersAddedGroupingPolicies {
//			enforcer := emap[k]
//			_, err := enforcer.AddGroupingPolicies(v)
//			if err != nil {
//				panic(err)
//			}
//		}
//	}
//
//	if len(rolesDeleted) > 0 {
//		rolesDeletedGroupingPolicies := getGroupingPoliciesByColumn(rolesDeleted, role, permissions)
//		for k, v := range rolesDeletedGroupingPolicies {
//			enforcer := emap[k]
//			_, err := enforcer.RemoveGroupingPolicies(v)
//			if err != nil {
//				panic(err)
//			}
//		}
//	}
//
//	if len(rolesAdded) > 0 {
//		rolesAddedGroupingPolicies := getGroupingPoliciesByColumn(rolesAdded, role, permissions)
//		for k, v := range rolesAddedGroupingPolicies {
//			enforcer := emap[k]
//			_, err := enforcer.AddGroupingPolicies(v)
//			if err != nil {
//				panic(err)
//			}
//		}
//	}
//
//	affected, err := adapter.Engine.ID(core.PK{owner, name}).AllCols().Update(role)
//	if err != nil {
//		panic(err)
//	}
//
//	return affected != 0
//}

func UpdateRole(id string, role *Role) bool {
	owner, name := util.GetOwnerAndNameFromId(id)
	oldRole := getRole(owner, name)
	if oldRole == nil {
		return false
	}

	permissions := GetPermissionsByRole(id)

	emap := make(map[string]*casbin.Enforcer, len(permissions))
	for _, p := range permissions {
		key := p.Adapter + "/" + strings.Join(p.Domains, ",")
		if _, ok := emap[key]; !ok {
			emap[key] = getEnforcer(p)
		}
	}

	if id != role.Owner+"/"+role.Name {
		groupingPolicies := getGroupingPoliciesByPermissions(oldRole.Users, oldRole, permissions)

		for k, e := range emap {
			res := strings.Split(k, "/")
			index := 1
			if res[1] != "" {
				index = 2
			}

			for _, beforeGroupingPolicy := range groupingPolicies[k] {
				var afterGroupingPolicy []string = make([]string, len(beforeGroupingPolicy))
				copy(afterGroupingPolicy, beforeGroupingPolicy)
				afterGroupingPolicy[index] = role.Owner + "/" + role.Name
				_, err := e.UpdateGroupingPolicy(beforeGroupingPolicy, afterGroupingPolicy)
				if err != nil {
					panic(err)
				}
			}

			beforePolicy := []string{id}
			afterPolicy := []string{role.Owner + "/" + role.Name}
			_, err := e.UpdatePolicy(beforePolicy, afterPolicy)
			if err != nil {
				panic(err)
			}
		}

		for _, p := range permissions {
			for k, v := range p.Roles {
				if v == id {
					p.Roles = append(p.Roles[:k], p.Roles[k+1:]...)
					p.Roles = append(p.Roles, role.Owner+"/"+role.Name)
					break
				}
			}
			_, err := adapter.Engine.ID(core.PK{p.Owner, p.Name}).AllCols().Update(p)
			if err != nil {
				panic(err)
			}
		}

	}

	usersAdded, usersDeleted := util.Arrcmp(oldRole.Users, role.Users)
	rolesAdded, rolesDeleted := util.Arrcmp(oldRole.Roles, role.Roles)

	if len(usersDeleted) > 0 {
		usersDeletedGroupingPolicies := getGroupingPoliciesByPermissions(usersDeleted, role, permissions)
		for k, v := range usersDeletedGroupingPolicies {
			enforcer := emap[k]
			_, err := enforcer.RemoveGroupingPolicies(v)
			if err != nil {
				panic(err)
			}
		}
	}

	if len(usersAdded) > 0 {
		usersAddedGroupingPolicies := getGroupingPoliciesByPermissions(usersAdded, role, permissions)
		for k, v := range usersAddedGroupingPolicies {
			enforcer := emap[k]
			_, err := enforcer.AddGroupingPolicies(v)
			if err != nil {
				panic(err)
			}
		}
	}

	if len(rolesDeleted) > 0 {
		rolesDeletedGroupingPolicies := getGroupingPoliciesByPermissions(rolesDeleted, role, permissions)
		for k, v := range rolesDeletedGroupingPolicies {
			enforcer := emap[k]
			_, err := enforcer.RemoveGroupingPolicies(v)
			if err != nil {
				panic(err)
			}
		}
	}

	if len(rolesAdded) > 0 {
		rolesAddedGroupingPolicies := getGroupingPoliciesByPermissions(rolesAdded, role, permissions)
		for k, v := range rolesAddedGroupingPolicies {
			enforcer := emap[k]
			_, err := enforcer.AddGroupingPolicies(v)
			if err != nil {
				panic(err)
			}
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
	permissions := GetPermissionsByRole(role.GetId())

	emap := make(map[string]*casbin.Enforcer, len(permissions))
	for _, p := range permissions {
		key := p.Adapter + "/" + strings.Join(p.Domains, ",")
		if _, ok := emap[key]; !ok {
			emap[key] = getEnforcer(p)
		}
	}

	for k, e := range emap {
		res := strings.Split(k, "/")
		index := 1
		if res[1] != "" {
			index = 2
		}
		_, err := e.RemoveFilteredGroupingPolicy(index, role.Owner+"/"+role.Name)
		if err != nil {
			panic(err)
		}

		_, err = e.RemoveFilteredNamedPolicy("p", 0, role.Owner+"/"+role.Name)
		if err != nil {
			panic(err)
		}
	}

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
