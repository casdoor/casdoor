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
	"strings"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	xormadapter "github.com/casbin/xorm-adapter/v2"
	"github.com/casdoor/casdoor/conf"
)

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

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act`
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
	for _, role := range permission.Roles {
		for _, resource := range permission.Resources {
			for _, action := range permission.Actions {
				policies = append(policies, []string{permission.GetId(), role, resource, strings.ToLower(action)})
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

func getGroupingPolicies(role *Role) [][]string {
	var groupingPolicies [][]string
	for _, subUser := range role.Users {
		groupingPolicies = append(groupingPolicies, []string{subUser, role.GetId()})
	}
	for _, subRole := range role.Roles {
		groupingPolicies = append(groupingPolicies, []string{subRole, role.GetId()})
	}
	return groupingPolicies
}

func addGroupingPolicies(role *Role) {
	enforcer := getEnforcer(&Permission{})
	groupingPolicies := getGroupingPolicies(role)

	_, err := enforcer.AddGroupingPolicies(groupingPolicies)
	if err != nil {
		panic(err)
	}
}

func removeGroupingPolicies(role *Role) {
	enforcer := getEnforcer(&Permission{})
	groupingPolicies := getGroupingPolicies(role)

	_, err := enforcer.RemoveGroupingPolicies(groupingPolicies)
	if err != nil {
		panic(err)
	}
}

func Enforce(userId string, permissionRule *PermissionRule) bool {
	permission := GetPermission(permissionRule.V0)
	enforcer := getEnforcer(permission)
	allow, err := enforcer.Enforce(userId, permissionRule.V2, permissionRule.V3)
	if err != nil {
		panic(err)
	}
	return allow
}

func BatchEnforce(userId string, permissionRules []PermissionRule) []bool {
	var requests [][]interface{}
	for _, permissionRule := range permissionRules {
		requests = append(requests, []interface{}{userId, permissionRule.V2, permissionRule.V3})
	}
	permission := GetPermission(permissionRules[0].V0)
	enforcer := getEnforcer(permission)
	allow, err := enforcer.BatchEnforce(requests)
	if err != nil {
		panic(err)
	}
	return allow
}

func getAllValues(userId string, sec string, fieldIndex int) []string {
	permissions := GetPermissionsByUser(userId)
	var values []string
	for _, permission := range permissions {
		enforcer := getEnforcer(permission)
		enforcer.ClearPolicy()
		err := enforcer.LoadFilteredPolicy(xormadapter.Filter{V0: []string{permission.GetId()}, V1: []string{userId}})
		if err != nil {
			return nil
		}

		for _, value := range enforcer.GetModel().GetValuesForFieldInPolicyAllTypes(sec, fieldIndex) {
			values = append(values, value)
		}
	}
	return values
}

func GetAllObjects(userId string) []string {
	return getAllValues(userId, "p", 2)
}

func GetAllActions(userId string) []string {
	return getAllValues(userId, "p", 3)
}

func GetAllRoles(userId string) []string {
	roles := GetRolesByUser(userId)
	var res []string
	for _, role := range roles {
		res = append(res, role.Name)
	}
	return res
}
