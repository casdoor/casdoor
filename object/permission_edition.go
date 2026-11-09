// Copyright 2026 The Casdoor Authors. All Rights Reserved.
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
	"regexp"
	"strings"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"github.com/casdoor/casdoor/edition"
	"github.com/casdoor/casdoor/util"
)

// Casbin escapes "r.sub" in matchers to "r_sub", so both forms are matched
var abacMatcherRegex = regexp.MustCompile(`\b[rp][0-9]*[._][A-Za-z_]\w*\.[A-Za-z_]|\beval\(`)

func isRbacAllowed() bool {
	return edition.Allows(edition.Commercial, edition.FeatureRbac)
}

func isAbacAllowed() bool {
	return edition.Allows(edition.Commercial, edition.FeatureAbac)
}

func requireRbac() error {
	return edition.Require(edition.Commercial, edition.FeatureRbac)
}

func isAbacModelText(modelText string) bool {
	return abacMatcherRegex.MatchString(modelText)
}

func isAbacModel(m model.Model) bool {
	for _, assertion := range m["m"] {
		if abacMatcherRegex.MatchString(assertion.Value) {
			return true
		}
	}
	return false
}

func isAbacRequest(request []interface{}) bool {
	for _, elem := range request {
		switch elem.(type) {
		case map[string]interface{}:
			return true
		case string:
			if _, ok := util.InterfaceToEnforceValue(elem).(string); !ok {
				return true
			}
		}
	}
	return false
}

func hasGroupingPolicies(m model.Model) bool {
	for _, assertion := range m["g"] {
		if len(assertion.Policy) > 0 {
			return true
		}
	}
	return false
}

func checkAbacEdition(m model.Model, requests [][]interface{}) error {
	if isAbacAllowed() {
		return nil
	}

	isAbac := isAbacModel(m)
	for _, request := range requests {
		isAbac = isAbac || isAbacRequest(request)
	}
	if isAbac {
		return edition.Require(edition.Commercial, edition.FeatureAbac)
	}
	return nil
}

// checkPermissionsEdition refuses to enforce permissions that grant by groups or roles without a license,
// instead of leaving out the users who would get them through a group or role
func checkPermissionsEdition(permission *Permission, permissionIds ...string) error {
	if isRbacAllowed() {
		return nil
	}

	targets, err := getPermissionEnforcerTargets(permission, permissionIds...)
	if err != nil {
		return err
	}
	for _, target := range targets {
		if len(target.Groups) > 0 || len(target.Roles) > 0 {
			return requireRbac()
		}
	}
	return nil
}

// CheckEnforcerEdition refuses the RBAC (grouping policies) and ABAC enforcement of an enforcer without a license
func CheckEnforcerEdition(enforcer *casbin.Enforcer, requests [][]interface{}) error {
	m := enforcer.GetModel()
	if !isRbacAllowed() && hasGroupingPolicies(m) {
		return requireRbac()
	}
	return checkAbacEdition(m, requests)
}

// CheckPermissionEdition is checked when a permission is saved
func CheckPermissionEdition(permission *Permission) error {
	if permission == nil || isRbacAllowed() {
		return nil
	}
	if len(permission.Groups) > 0 || len(permission.Roles) > 0 {
		return requireRbac()
	}
	return nil
}

// CheckModelEdition is checked when a model is saved
func CheckModelEdition(m *Model) error {
	if m == nil || isAbacAllowed() || !isAbacModelText(m.ModelText) {
		return nil
	}
	return edition.Require(edition.Commercial, edition.FeatureAbac)
}

// CheckPolicyEdition is checked when rules are added to an enforcer: "g" rules are role inheritance
func CheckPolicyEdition(ptypes ...string) error {
	if isRbacAllowed() {
		return nil
	}
	for _, ptype := range ptypes {
		if strings.HasPrefix(ptype, "g") {
			return requireRbac()
		}
	}
	return nil
}
