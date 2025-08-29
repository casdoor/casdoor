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
	"github.com/casdoor/casdoor/util"
	xormadapter "github.com/casdoor/xorm-adapter/v3"
)

func getPermissionEnforcer(p *Permission, permissionIDs ...string) (*casbin.Enforcer, error) {
	// Init an enforcer instance without specifying a model or adapter.
	// If you specify an adapter, it will load all policies, which is a
	// heavy process that can slow down the application.
	enforcer, err := casbin.NewEnforcer(&log.DefaultLogger{}, false)
	if err != nil {
		return nil, err
	}

	err = p.setEnforcerModel(enforcer)
	if err != nil {
		return nil, err
	}

	err = p.setEnforcerAdapter(enforcer)
	if err != nil {
		return nil, err
	}

	policyFilterV5 := []string{p.GetId()}
	if len(permissionIDs) != 0 {
		policyFilterV5 = permissionIDs
	}

	policyFilter := xormadapter.Filter{
		V5: policyFilterV5,
	}

	if !HasRoleDefinition(enforcer.GetModel()) {
		policyFilter.Ptype = []string{"p"}
	}

	err = enforcer.LoadFilteredPolicy(policyFilter)
	if err != nil {
		return nil, err
	}

	return enforcer, nil
}

func (p *Permission) setEnforcerAdapter(enforcer *casbin.Enforcer) error {
	tableName := "permission_rule"
	if len(p.Adapter) != 0 {
		adapterObj, err := getAdapter(p.Owner, p.Adapter)
		if err != nil {
			return err
		}

		if adapterObj != nil && adapterObj.Table != "" {
			tableName = adapterObj.Table
		}
	}
	tableNamePrefix := conf.GetConfigString("tableNamePrefix")
	adapter, err := xormadapter.NewAdapterByEngineWithTableName(ormer.Engine, tableName, tableNamePrefix)
	if err != nil {
		return err
	}

	enforcer.SetAdapter(adapter)
	return nil
}

func (p *Permission) setEnforcerModel(enforcer *casbin.Enforcer) error {
	permissionModel, err := getModel(p.Owner, p.Model)
	if err != nil {
		return err
	}

	// TODO: return error if permissionModel is nil.
	m := model.Model{}
	if permissionModel != nil {
		m, err = GetBuiltInModel(permissionModel.ModelText)
	} else {
		m, err = GetBuiltInModel("")
	}
	if err != nil {
		return err
	}

	err = enforcer.InitWithModelAndAdapter(m, nil)
	if err != nil {
		return err
	}
	return nil
}

func getPolicies(permission *Permission) [][]string {
	var policies [][]string

	permissionId := permission.GetId()
	domainExist := len(permission.Domains) > 0

	usersAndRoles := append(permission.Users, permission.Roles...)
	for _, userOrRole := range usersAndRoles {
		for _, resource := range permission.Resources {
			for _, action := range permission.Actions {
				if domainExist {
					for _, domain := range permission.Domains {
						policies = append(policies, []string{userOrRole, domain, resource, action, strings.ToLower(permission.Effect), permissionId})
					}
				} else {
					policies = append(policies, []string{userOrRole, resource, action, strings.ToLower(permission.Effect), "", permissionId})
				}
			}
		}
	}

	return policies
}

func getRolesInRole(roleId string, visited map[string]struct{}) ([]*Role, error) {
	roleOwner, roleName := util.GetOwnerAndNameFromId(roleId)
	if roleName == "*" {
		roles, err := GetRoles(roleOwner)
		if err != nil {
			return []*Role{}, err
		}

		return roles, nil
	}

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

func getGroupingPolicies(permission *Permission) ([][]string, error) {
	var groupingPolicies [][]string

	domainExist := len(permission.Domains) > 0
	permissionId := permission.GetId()

	for _, roleId := range permission.Roles {
		visited := map[string]struct{}{}

		if roleId == "*" {
			roleId = util.GetId(permission.Owner, "*")
		}

		rolesInRole, err := getRolesInRole(roleId, visited)
		if err != nil {
			return nil, err
		}

		for _, role := range rolesInRole {
			roleId = role.GetId()
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

	return groupingPolicies, nil
}

func addPolicies(permission *Permission) error {
	enforcer, err := getPermissionEnforcer(permission)
	if err != nil {
		return err
	}

	policies := getPolicies(permission)

	_, err = enforcer.AddPolicies(policies)
	return err
}

func removePolicies(permission *Permission) error {
	enforcer, err := getPermissionEnforcer(permission)
	if err != nil {
		return err
	}

	policies := getPolicies(permission)

	_, err = enforcer.RemovePolicies(policies)
	return err
}

func addGroupingPolicies(permission *Permission) error {
	enforcer, err := getPermissionEnforcer(permission)
	if err != nil {
		return err
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

func removeGroupingPolicies(permission *Permission) error {
	enforcer, err := getPermissionEnforcer(permission)
	if err != nil {
		return err
	}

	groupingPolicies, err := getGroupingPolicies(permission)
	if err != nil {
		return err
	}

	if len(groupingPolicies) > 0 {
		_, err = enforcer.RemoveGroupingPolicies(groupingPolicies)
		if err != nil {
			return err
		}
	}

	return nil
}

func Enforce(permission *Permission, request []string, permissionIds ...string) (bool, error) {
	enforcer, err := getPermissionEnforcer(permission, permissionIds...)
	if err != nil {
		return false, err
	}

	// type transformation
	interfaceRequest := util.StringToInterfaceArray(request)

	return enforcer.Enforce(interfaceRequest...)
}

func BatchEnforce(permission *Permission, requests [][]string, permissionIds ...string) ([]bool, error) {
	enforcer, err := getPermissionEnforcer(permission, permissionIds...)
	if err != nil {
		return nil, err
	}

	// type transformation
	interfaceRequests := util.StringToInterfaceArray2d(requests)

	return enforcer.BatchEnforce(interfaceRequests)
}

func getEnforcers(userId string) ([]*casbin.Enforcer, error) {
	permissions, _, err := getPermissionsAndRolesByUser(userId)
	if err != nil {
		return nil, err
	}

	allRoles, err := GetAllRoles(userId)
	if err != nil {
		return nil, err
	}

	for _, role := range allRoles {
		var permissionsByRole []*Permission
		permissionsByRole, err = GetPermissionsByRole(role)
		if err != nil {
			return nil, err
		}

		permissions = append(permissions, permissionsByRole...)
	}

	var enforcers []*casbin.Enforcer
	for _, permission := range permissions {
		var enforcer *casbin.Enforcer
		enforcer, err = getPermissionEnforcer(permission)
		if err != nil {
			return nil, err
		}

		enforcers = append(enforcers, enforcer)
	}
	return enforcers, nil
}

func GetAllObjects(userId string) ([]string, error) {
	enforcers, err := getEnforcers(userId)
	if err != nil {
		return nil, err
	}

	res := []string{}
	for _, enforcer := range enforcers {
		items := enforcer.GetAllObjects()
		res = append(res, items...)
	}
	return res, nil
}

func GetAllActions(userId string) ([]string, error) {
	enforcers, err := getEnforcers(userId)
	if err != nil {
		return nil, err
	}

	res := []string{}
	for _, enforcer := range enforcers {
		items := enforcer.GetAllActions()
		res = append(res, items...)
	}
	return res, nil
}

func GetAllRoles(userId string) ([]string, error) {
	roles, err := getRolesByUser(userId)
	if err != nil {
		return nil, err
	}

	res := []string{}
	for _, role := range roles {
		res = append(res, role.Name)
	}
	return res, nil
}

func GetBuiltInModel(modelText string) (model.Model, error) {
	if modelText == "" {
		modelText = `[request_definition]
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
		if fieldsNum > builtInMaxFields {
			return nil, fmt.Errorf("the maximum policy_definition field number cannot exceed %d, got %d", builtInMaxFields, fieldsNum)
		}

		// filled empty field with "" and V5 with "permissionId"
		if fieldsNum == builtInMaxFields {
			sixthField := strings.TrimSpace(policyDefinition[builtInMaxFields-1])
			if sixthField != "permissionId" {
				return nil, fmt.Errorf("when adding policies with permissions, the sixth field of policy_definition must be permissionId, got %s", policyDefinition[builtInMaxFields-1])
			}
		} else {
			needFill := builtInMaxFields - fieldsNum
			for i := 0; i < needFill-1; i++ {
				policyDefinition = append(policyDefinition, "")
			}
			policyDefinition = append(policyDefinition, "permissionId")
		}

		m, err := model.NewModelFromString(modelText)
		if err != nil {
			return nil, err
		}

		m.AddDef("p", "p", strings.Join(policyDefinition, ","))

		return m, err
	}
}
