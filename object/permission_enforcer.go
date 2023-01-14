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
	xormadapter "github.com/casbin/xorm-adapter/v3"
	"github.com/casdoor/casdoor/conf"
)

func getEnforcer(permission *Permission) *casbin.Enforcer {
	tableName := "permission_rule"
	if len(permission.Adapter) != 0 {
		tableName = permission.Adapter
	}
	tableNamePrefix := conf.GetConfigString("tableNamePrefix")
	driverName := conf.GetConfigString("driverName")
	dataSourceName := conf.GetConfigRealDataSourceName(driverName)
	adapter, err := xormadapter.NewAdapterWithTableName(driverName, dataSourceName, tableName, tableNamePrefix, true)
	if err != nil {
		panic(err)
	}

	modelText := `
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

	// load Policy with a specific Permission
	enforcer.LoadFilteredPolicy(xormadapter.Filter{
		V5: []string{permission.Owner + "/" + permission.Name},
	})
	return enforcer
}

func getPolicies(permission *Permission) [][]string {
	var policies [][]string
	permissionId := permission.Owner + "/" + permission.Name
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

func getGroupingPolicies(permission *Permission) [][]string {
	var groupingPolicies [][]string
	domainExist := len(permission.Domains) > 0

	for _, role := range permission.Roles {
		roleObj := GetRole(role)
		if roleObj != nil {
			for _, subUser := range roleObj.Users {
				if domainExist {
					for _, domain := range permission.Domains {
						groupingPolicies = append(groupingPolicies, []string{subUser, domain, role})
					}
				} else {
					groupingPolicies = append(groupingPolicies, []string{subUser, role})
				}
			}
			for _, subRole := range roleObj.Roles {
				if domainExist {
					for _, domain := range permission.Domains {
						groupingPolicies = append(groupingPolicies, []string{subRole, domain, role})
					}
				} else {
					groupingPolicies = append(groupingPolicies, []string{subRole, role})
				}
			}
		}

	}
	return groupingPolicies
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

func addPolicies(permission *Permission) {
	enforcer := getEnforcer(permission)
	policies := getPolicies(permission)

	_, err := enforcer.AddPolicies(policies)
	if err != nil {
		panic(err)
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

func Enforce(permissionRule *PermissionRule) bool {
	permission := GetPermission(permissionRule.Id)
	enforcer := getEnforcer(permission)

	request := []interface{}{permissionRule.V0, permissionRule.V1, permissionRule.V2}
	if permissionRule.V3 != "" {
		request = append(request, permissionRule.V3)
	}
	allow, err := enforcer.Enforce(request...)
	if err != nil {
		panic(err)
	}
	return allow
}

func BatchEnforce(permissionRules []PermissionRule) []bool {
	var requests [][]interface{}
	for _, permissionRule := range permissionRules {
		if permissionRule.V3 != "" {
			requests = append(requests, []interface{}{permissionRule.V0, permissionRule.V1, permissionRule.V2, permissionRule.V3})
		} else {
			requests = append(requests, []interface{}{permissionRule.V0, permissionRule.V1, permissionRule.V2})
		}
	}
	permission := GetPermission(permissionRules[0].Id)
	enforcer := getEnforcer(permission)
	allow, err := enforcer.BatchEnforce(requests)
	if err != nil {
		panic(err)
	}
	return allow
}

func getAllValues(userId string, fn func(enforcer *casbin.Enforcer) []string) []string {
	permissions := GetPermissionsByUser(userId)
	for _, role := range GetAllRoles(userId) {
		permissions = append(permissions, GetPermissionsByRole(role)...)
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
	roles := GetRolesByUser(userId)
	var res []string
	for _, role := range roles {
		res = append(res, role.Name)
	}
	return res
}

func getGroupingPoliciesByPermissions(column []string, role *Role, permissions []*Permission) map[string][][]string {
	var groupingPolicies = make(map[string][][]string, len(permissions))
	for _, p := range permissions {
		domainExist := len(p.Domains) > 0
		key := p.Adapter + "/" + strings.Join(p.Domains, ",")
		if _, ok := groupingPolicies[key]; ok {
			continue
		}
		for _, v := range column {
			if domainExist {
				for _, domain := range p.Domains {
					groupingPolicies[key] = append(groupingPolicies[key], []string{v, domain, role.Owner + "/" + role.Name})
				}
			} else {
				groupingPolicies[key] = append(groupingPolicies[key], []string{v, role.Owner + "/" + role.Name})
			}
		}
	}

	return groupingPolicies
}

func getPoliciesByPermissions(column []string, permissions []*Permission) map[string][][]string {
	var policies = make(map[string][][]string, len(permissions))
	for _, p := range permissions {
		permissionId := p.Owner + "/" + p.Name
		domainExist := len(p.Domains) > 0
		key := p.Adapter + "/" + strings.Join(p.Domains, ",")
		//if _, ok := policies[key]; ok {
		//	continue
		//}
		for _, v := range column {
			for _, resource := range p.Resources {
				for _, action := range p.Actions {
					if domainExist {
						for _, domain := range p.Domains {
							policies[key] = append(policies[key], []string{v, domain, resource, strings.ToLower(action), "", permissionId})
						}
					} else {
						policies[key] = append(policies[key], []string{v, resource, strings.ToLower(action), "", "", permissionId})
					}
				}
			}
		}
	}
	return policies
}

//func operateGroupingPoliciesByPermission(permission *Permission, enforcer *casbin.Enforcer, isAdd bool) {
//	var err error
//	domainExist := len(permission.Domains) > 0
//	for _, role := range permission.Roles {
//		roleObj := GetRole(role)
//		for _, user := range roleObj.Users {
//			if domainExist {
//				for _, domain := range permission.Domains {
//					if isAdd {
//						_, err = enforcer.AddNamedGroupingPolicy("g", user, domain, roleObj.Owner+"/"+roleObj.Name)
//					} else {
//						_, err = enforcer.RemoveNamedGroupingPolicy("g", user, domain, roleObj.Owner+"/"+roleObj.Name)
//					}
//					if err != nil {
//						panic(err)
//					}
//				}
//			} else {
//				if isAdd {
//					_, err = enforcer.AddNamedGroupingPolicy("g", user, roleObj.Owner+"/"+roleObj.Name)
//				} else {
//					_, err = enforcer.RemoveNamedGroupingPolicy("g", user, roleObj.Owner+"/"+roleObj.Name)
//				}
//				if err != nil {
//					panic(err)
//				}
//			}
//		}
//	}
//}
//
//func operatePoliciesByPermission(permission *Permission, enforcer *casbin.Enforcer, isAdd bool, isUser bool) {
//	var err error
//	column := permission.Roles
//	if isUser {
//		column = permission.Users
//	}
//	domainExist := len(permission.Domains) > 0
//	for _, v := range column {
//		for _, resource := range permission.Resources {
//			for _, action := range permission.Actions {
//				if domainExist {
//					for _, domain := range permission.Domains {
//						if isAdd {
//							_, err = enforcer.AddNamedPolicy("p", v, domain, resource, action)
//						} else {
//							_, err = enforcer.RemoveNamedPolicy("p", v, domain, resource, action)
//						}
//						if err != nil {
//							panic(err)
//						}
//					}
//				} else {
//					if isAdd {
//						_, err = enforcer.AddNamedPolicy("p", v, resource, action)
//					} else {
//						_, err = enforcer.RemoveNamedPolicy("p", v, resource, action)
//					}
//					if err != nil {
//						panic(err)
//					}
//				}
//			}
//		}
//	}
//}
