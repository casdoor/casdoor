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
	"github.com/casbin/casbin/v2/config"
	"github.com/casbin/casbin/v2/log"
	"github.com/casbin/casbin/v2/model"
	"github.com/casdoor/casdoor/conf"
	xormadapter "github.com/casdoor/xorm-adapter/v3"
)

func getEnforcer(permission *Permission, permissionIDs ...string) *casbin.Enforcer {
	tableName := "permission_rule"
	if len(permission.Adapter) != 0 {
		adapterObj, err := getCasbinAdapter(permission.Owner, permission.Adapter)
		if err != nil {
			panic(err)
		}

		if adapterObj != nil && adapterObj.Table != "" {
			tableName = adapterObj.Table
		}
	}
	tableNamePrefix := conf.GetConfigString("tableNamePrefix")
	driverName := conf.GetConfigString("driverName")
	dataSourceName := conf.GetConfigRealDataSourceName(driverName)
	adapter, err := xormadapter.NewAdapterWithTableName(driverName, dataSourceName, tableName, tableNamePrefix, true)
	if err != nil {
		panic(err)
	}

	permissionModel, err := getModel(permission.Owner, permission.Model)
	if err != nil {
		panic(err)
	}

	m := model.Model{}
	if permissionModel != nil {
		m, err = GetBuiltInModel(permissionModel.ModelText)
	} else {
		m, err = GetBuiltInModel("")
	}

	if err != nil {
		panic(err)
	}

	// Init an enforcer instance without specifying a model or adapter.
	// If you specify an adapter, it will load all policies, which is a
	// heavy process that can slow down the application.
	enforcer, err := casbin.NewEnforcer(&log.DefaultLogger{}, false)
	if err != nil {
		panic(err)
	}

	err = enforcer.InitWithModelAndAdapter(m, nil)
	if err != nil {
		panic(err)
	}

	enforcer.SetAdapter(adapter)

	policyFilterV5 := []string{permission.GetId()}
	if len(permissionIDs) != 0 {
		policyFilterV5 = permissionIDs
	}

	policyFilter := xormadapter.Filter{
		V5: policyFilterV5,
	}

	if !HasRoleDefinition(m) {
		policyFilter.Ptype = []string{"p"}
	}

	err = enforcer.LoadFilteredPolicy(policyFilter)
	if err != nil {
		panic(err)
	}

	return enforcer
}

func getPolicies(permission *Permission) [][]string {
	var policies [][]string

	permissionId := permission.GetId()
	domainExist := len(permission.Domains) > 0

	for _, user := range permission.Users {
		for _, resource := range permission.Resources {
			for _, action := range permission.Actions {
				if domainExist {
					for _, domain := range permission.Domains {
						policies = append(policies, []string{user, domain, resource, strings.ToLower(action), "", permissionId})
					}
				} else {
					policies = append(policies, []string{user, resource, strings.ToLower(action), "", "", permissionId})
				}
			}
		}
	}

	for _, role := range permission.Roles {
		for _, resource := range permission.Resources {
			for _, action := range permission.Actions {
				if domainExist {
					for _, domain := range permission.Domains {
						policies = append(policies, []string{role, domain, resource, strings.ToLower(action), "", permissionId})
					}
				} else {
					policies = append(policies, []string{role, resource, strings.ToLower(action), "", "", permissionId})
				}
			}
		}
	}

	return policies
}

func getRolesInRole(roleId string, visited map[string]struct{}) ([]*Role, error) {
	role, err := GetRole(roleId)
	if err != nil {
		return []*Role{}, err
	}

	if role == nil {
		return []*Role{}, nil
	}
	visited[roleId] = struct{}{}

	roles := []*Role{role}
	for _, subRole := range role.Roles {
		if _, ok := visited[subRole]; !ok {
			r, err := getRolesInRole(subRole, visited)
			if err != nil {
				return []*Role{}, err
			}

			roles = append(roles, r...)
		}
	}

	return roles, nil
}

func getGroupingPolicies(permission *Permission) [][]string {
	var groupingPolicies [][]string

	domainExist := len(permission.Domains) > 0
	permissionId := permission.GetId()

	for _, roleId := range permission.Roles {
		visited := map[string]struct{}{}
		rolesInRole, err := getRolesInRole(roleId, visited)
		if err != nil {
			panic(err)
		}
		for _, role := range rolesInRole {
			roleId := role.GetId()
			for _, subUser := range role.Users {
				if domainExist {
					for _, domain := range permission.Domains {
						groupingPolicies = append(groupingPolicies, []string{subUser, roleId, domain, "", "", permissionId})
					}
				} else {
					groupingPolicies = append(groupingPolicies, []string{subUser, roleId, "", "", "", permissionId})
				}
			}

			for _, subRole := range role.Roles {
				if domainExist {
					for _, domain := range permission.Domains {
						groupingPolicies = append(groupingPolicies, []string{subRole, roleId, domain, "", "", permissionId})
					}
				} else {
					groupingPolicies = append(groupingPolicies, []string{subRole, roleId, "", "", "", permissionId})
				}
			}
		}
	}

	return groupingPolicies
}

func addPolicies(permission *Permission) {
	enforcer := getEnforcer(permission)
	policies := getPolicies(permission)

	_, err := enforcer.AddPolicies(policies)
	if err != nil {
		panic(err)
	}
}

func addGroupingPolicies(permission *Permission) {
	enforcer := getEnforcer(permission)
	groupingPolicies := getGroupingPolicies(permission)

	if len(groupingPolicies) > 0 {
		_, err := enforcer.AddGroupingPolicies(groupingPolicies)
		if err != nil {
			panic(err)
		}
	}
}

func removeGroupingPolicies(permission *Permission) {
	enforcer := getEnforcer(permission)
	groupingPolicies := getGroupingPolicies(permission)

	if len(groupingPolicies) > 0 {
		_, err := enforcer.RemoveGroupingPolicies(groupingPolicies)
		if err != nil {
			panic(err)
		}
	}
}

func removePolicies(permission *Permission) {
	enforcer := getEnforcer(permission)
	policies := getPolicies(permission)

	_, err := enforcer.RemovePolicies(policies)
	if err != nil {
		panic(err)
	}
}

type CasbinRequest = []interface{}

func Enforce(permissionId string, request *CasbinRequest) (bool, error) {
	permission, err := GetPermission(permissionId)
	if err != nil {
		return false, err
	}

	enforcer := getEnforcer(permission)
	return enforcer.Enforce(*request...)
}

func BatchEnforce(permissionId string, requests *[]CasbinRequest, permissionIds ...string) ([]bool, error) {
	permission, err := GetPermission(permissionId)
	if err != nil {
		res := []bool{}
		for i := 0; i < len(*requests); i++ {
			res = append(res, false)
		}

		return res, err
	}

	enforcer := getEnforcer(permission, permissionIds...)
	return enforcer.BatchEnforce(*requests)
}

func getAllValues(userId string, fn func(enforcer *casbin.Enforcer) []string) []string {
	permissions, _, err := GetPermissionsAndRolesByUser(userId)
	if err != nil {
		panic(err)
	}

	for _, role := range GetAllRoles(userId) {
		permissionsByRole, err := GetPermissionsByRole(role)
		if err != nil {
			panic(err)
		}

		permissions = append(permissions, permissionsByRole...)
	}

	var values []string
	for _, permission := range permissions {
		enforcer := getEnforcer(permission)
		values = append(values, fn(enforcer)...)
	}
	return values
}

func GetAllObjects(userId string) []string {
	return getAllValues(userId, func(enforcer *casbin.Enforcer) []string {
		return enforcer.GetAllObjects()
	})
}

func GetAllActions(userId string) []string {
	return getAllValues(userId, func(enforcer *casbin.Enforcer) []string {
		return enforcer.GetAllActions()
	})
}

func GetAllRoles(userId string) []string {
	roles, err := GetRolesByUser(userId)
	if err != nil {
		panic(err)
	}

	var res []string
	for _, role := range roles {
		res = append(res, role.Name)
	}
	return res
}

func GetBuiltInModel(modelText string) (model.Model, error) {
	if modelText == "" {
		modelText = `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act, "", "", permissionId

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act`
		return model.NewModelFromString(modelText)
	} else {
		cfg, err := config.NewConfigFromText(modelText)
		if err != nil {
			return nil, err
		}

		// load [policy_definition]
		policyDefinition := strings.Split(cfg.String("policy_definition::p"), ",")
		fieldsNum := len(policyDefinition)
		if fieldsNum > builtInAvailableField {
			panic(fmt.Errorf("the maximum policy_definition field number cannot exceed %d", builtInAvailableField))
		}
		// filled empty field with "" and V5 with "permissionId"
		for i := builtInAvailableField - fieldsNum; i > 0; i-- {
			policyDefinition = append(policyDefinition, "")
		}
		policyDefinition = append(policyDefinition, "permissionId")

		m, _ := model.NewModelFromString(modelText)
		m.AddDef("p", "p", strings.Join(policyDefinition, ","))

		return m, err
	}
}
