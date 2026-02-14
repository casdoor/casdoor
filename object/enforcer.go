// Copyright 2023 The Casdoor Authors. All Rights Reserved.
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
	"slices"

	"github.com/casbin/casbin/v2"
	"github.com/casdoor/casdoor/util"
	xormadapter "github.com/casdoor/xorm-adapter/v3"
	"github.com/xorm-io/core"
)

type Enforcer struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`
	UpdatedTime string `xorm:"varchar(100) updated" json:"updatedTime"`
	DisplayName string `xorm:"varchar(100)" json:"displayName"`
	Description string `xorm:"varchar(100)" json:"description"`

	Model   string `xorm:"varchar(100)" json:"model"`
	Adapter string `xorm:"varchar(100)" json:"adapter"`

	ModelCfg map[string]string `xorm:"-" json:"modelCfg"`
	*casbin.Enforcer
}

func GetEnforcerCount(owner, field, value string) (int64, error) {
	session := GetSession(owner, -1, -1, field, value, "", "")
	return session.Count(&Enforcer{})
}

func GetEnforcers(owner string) ([]*Enforcer, error) {
	enforcers := []*Enforcer{}
	err := ormer.Engine.Desc("created_time").Find(&enforcers, &Enforcer{Owner: owner})
	if err != nil {
		return enforcers, err
	}

	return enforcers, nil
}

func GetPaginationEnforcers(owner string, offset, limit int, field, value, sortField, sortOrder string) ([]*Enforcer, error) {
	enforcers := []*Enforcer{}
	session := GetSession(owner, offset, limit, field, value, sortField, sortOrder)
	err := session.Find(&enforcers)
	if err != nil {
		return enforcers, err
	}

	return enforcers, nil
}

func getEnforcer(owner string, name string) (*Enforcer, error) {
	if owner == "" || name == "" {
		return nil, nil
	}

	enforcer := Enforcer{Owner: owner, Name: name}
	existed, err := ormer.Engine.Get(&enforcer)
	if err != nil {
		return &enforcer, err
	}

	if existed {
		return &enforcer, nil
	} else {
		return nil, nil
	}
}

func GetEnforcer(id string) (*Enforcer, error) {
	owner, name, err := util.GetOwnerAndNameFromIdWithError(id)
	if err != nil {
		return nil, err
	}
	return getEnforcer(owner, name)
}

func UpdateEnforcer(id string, enforcer *Enforcer) (bool, error) {
	owner, name, err := util.GetOwnerAndNameFromIdWithError(id)
	if err != nil {
		return false, err
	}
	if oldEnforcer, err := getEnforcer(owner, name); err != nil {
		return false, err
	} else if oldEnforcer == nil {
		return false, nil
	}

	affected, err := ormer.Engine.ID(core.PK{owner, name}).AllCols().Update(enforcer)
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func AddEnforcer(enforcer *Enforcer) (bool, error) {
	affected, err := ormer.Engine.Insert(enforcer)
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func DeleteEnforcer(enforcer *Enforcer) (bool, error) {
	affected, err := ormer.Engine.ID(core.PK{enforcer.Owner, enforcer.Name}).Delete(&Enforcer{})
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func (enforcer *Enforcer) GetId() string {
	return fmt.Sprintf("%s/%s", enforcer.Owner, enforcer.Name)
}

func (enforcer *Enforcer) GetModelAndAdapter() string {
	return util.GetId(enforcer.Model, enforcer.Adapter)
}

func (enforcer *Enforcer) InitEnforcer() error {
	if enforcer.Enforcer != nil {
		return nil
	}

	if enforcer.Model == "" {
		return fmt.Errorf("the model for enforcer: %s should not be empty", enforcer.GetId())
	}
	if enforcer.Adapter == "" {
		return fmt.Errorf("the adapter for enforcer: %s should not be empty", enforcer.GetId())
	}

	m, err := GetModel(enforcer.Model)
	if err != nil {
		return err
	} else if m == nil {
		return fmt.Errorf("the model: %s for enforcer: %s is not found", enforcer.Model, enforcer.GetId())
	}

	a, err := GetAdapter(enforcer.Adapter)
	if err != nil {
		return err
	} else if a == nil {
		return fmt.Errorf("the adapter: %s for enforcer: %s is not found", enforcer.Adapter, enforcer.GetId())
	}

	err = m.initModel()
	if err != nil {
		return err
	}
	err = a.InitAdapter()
	if err != nil {
		return err
	}

	casbinEnforcer, err := casbin.NewEnforcer(m.Model, a.Adapter)
	if err != nil {
		return err
	}

	enforcer.Enforcer = casbinEnforcer
	return nil
}

func GetInitializedEnforcer(enforcerId string) (*Enforcer, error) {
	enforcer, err := GetEnforcer(enforcerId)
	if err != nil {
		return nil, err
	} else if enforcer == nil {
		return nil, fmt.Errorf("the enforcer: %s is not found", enforcerId)
	}

	err = enforcer.InitEnforcer()
	if err != nil {
		return nil, err
	}
	return enforcer, nil
}

func GetPolicies(id string) ([]*xormadapter.CasbinRule, error) {
	enforcer, err := GetInitializedEnforcer(id)
	if err != nil {
		return nil, err
	}

	pRules := enforcer.GetPolicy()
	res := util.MatrixToCasbinRules("p", pRules)

	if enforcer.GetModel()["g"] != nil {
		gRules := enforcer.GetGroupingPolicy()
		res2 := util.MatrixToCasbinRules("g", gRules)
		res = append(res, res2...)
	}

	return res, nil
}

// Filter represents filter criteria with optional policy type
type Filter struct {
	Ptype       string   `json:"ptype,omitempty"`
	FieldIndex  *int     `json:"fieldIndex,omitempty"`
	FieldValues []string `json:"fieldValues"`
}

func GetFilteredPolicies(id string, ptype string, fieldIndex int, fieldValues ...string) ([]*xormadapter.CasbinRule, error) {
	enforcer, err := GetInitializedEnforcer(id)
	if err != nil {
		return nil, err
	}

	var allRules [][]string

	if len(fieldValues) == 0 {
		if ptype == "g" {
			allRules = enforcer.GetFilteredGroupingPolicy(fieldIndex)
		} else {
			allRules = enforcer.GetFilteredPolicy(fieldIndex)
		}
	} else {
		for _, value := range fieldValues {
			if ptype == "g" {
				rules := enforcer.GetFilteredGroupingPolicy(fieldIndex, value)
				allRules = append(allRules, rules...)
			} else {
				rules := enforcer.GetFilteredPolicy(fieldIndex, value)
				allRules = append(allRules, rules...)
			}
		}
	}

	res := util.MatrixToCasbinRules(ptype, allRules)
	return res, nil
}

// GetFilteredPoliciesMulti applies multiple filters to policies
// Doing this in our loop is more efficient than using GetFilteredGroupingPolicy / GetFilteredPolicy which
// iterates over all policies again and again
func GetFilteredPoliciesMulti(id string, filters []Filter) ([]*xormadapter.CasbinRule, error) {
	// Get all policies first
	allPolicies, err := GetPolicies(id)
	if err != nil {
		return nil, err
	}

	// Filter policies based on multiple criteria
	var filteredPolicies []*xormadapter.CasbinRule
	if len(filters) == 0 {
		// No filters, return all policies
		return allPolicies, nil
	} else {
		for _, policy := range allPolicies {
			matchesAllFilters := true
			for _, filter := range filters {
				// Default policy type if unspecified
				if filter.Ptype == "" {
					filter.Ptype = "p"
				}
				// Always check policy type
				if policy.Ptype != filter.Ptype {
					matchesAllFilters = false
					break
				}

				// If FieldIndex is nil, only filter via ptype (skip field-value checks)
				if filter.FieldIndex == nil {
					continue
				}

				fieldIndex := *filter.FieldIndex
				// If FieldIndex is out of range, also only filter via ptype
				if fieldIndex < 0 || fieldIndex > 5 {
					continue
				}

				var fieldValue string
				switch fieldIndex {
				case 0:
					fieldValue = policy.V0
				case 1:
					fieldValue = policy.V1
				case 2:
					fieldValue = policy.V2
				case 3:
					fieldValue = policy.V3
				case 4:
					fieldValue = policy.V4
				case 5:
					fieldValue = policy.V5
				}

				// When FieldIndex is provided and valid, enforce FieldValues (if any)
				if len(filter.FieldValues) > 0 && !slices.Contains(filter.FieldValues, fieldValue) {
					matchesAllFilters = false
					break
				}
			}

			if matchesAllFilters {
				filteredPolicies = append(filteredPolicies, policy)
			}
		}
	}

	return filteredPolicies, nil
}

func UpdatePolicy(id string, ptype string, oldPolicy []string, newPolicy []string) (bool, error) {
	enforcer, err := GetInitializedEnforcer(id)
	if err != nil {
		return false, err
	}

	var affected bool
	if ptype == "p" {
		affected, err = enforcer.UpdatePolicy(oldPolicy, newPolicy)
	} else {
		affected, err = enforcer.UpdateGroupingPolicy(oldPolicy, newPolicy)
	}

	if err == nil && affected {
		// Notify other pods about the policy change
		_ = publishPolicyChange(id, "update")
	}

	return affected, err
}

func AddPolicy(id string, ptype string, policy []string) (bool, error) {
	enforcer, err := GetInitializedEnforcer(id)
	if err != nil {
		return false, err
	}

	var affected bool
	if ptype == "p" {
		affected, err = enforcer.AddPolicy(policy)
	} else {
		affected, err = enforcer.AddGroupingPolicy(policy)
	}

	if err == nil && affected {
		// Notify other pods about the policy change
		_ = publishPolicyChange(id, "add")
	}

	return affected, err
}

func RemovePolicy(id string, ptype string, policy []string) (bool, error) {
	enforcer, err := GetInitializedEnforcer(id)
	if err != nil {
		return false, err
	}

	var affected bool
	if ptype == "p" {
		affected, err = enforcer.RemovePolicy(policy)
	} else {
		affected, err = enforcer.RemoveGroupingPolicy(policy)
	}

	if err == nil && affected {
		// Notify other pods about the policy change
		_ = publishPolicyChange(id, "remove")
	}

	return affected, err
}

func (enforcer *Enforcer) LoadModelCfg() error {
	if enforcer.ModelCfg != nil {
		return nil
	}

	model, err := getModelEx(enforcer.Model)
	if err != nil {
		return err
	} else if model == nil {
		return fmt.Errorf("the model: %s for enforcer: %s is not found", enforcer.Model, enforcer.GetId())
	}

	enforcer.ModelCfg, err = getModelCfg(model)
	if err != nil {
		return err
	}

	return nil
}
