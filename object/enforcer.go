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
	owner, name := util.GetOwnerAndNameFromId(id)
	return getEnforcer(owner, name)
}

func UpdateEnforcer(id string, enforcer *Enforcer) (bool, error) {
	owner, name := util.GetOwnerAndNameFromId(id)
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

func UpdatePolicy(id string, ptype string, oldPolicy []string, newPolicy []string) (bool, error) {
	enforcer, err := GetInitializedEnforcer(id)
	if err != nil {
		return false, err
	}

	if ptype == "p" {
		return enforcer.UpdatePolicy(oldPolicy, newPolicy)
	} else {
		return enforcer.UpdateGroupingPolicy(oldPolicy, newPolicy)
	}
}

func AddPolicy(id string, ptype string, policy []string) (bool, error) {
	enforcer, err := GetInitializedEnforcer(id)
	if err != nil {
		return false, err
	}

	if ptype == "p" {
		return enforcer.AddPolicy(policy)
	} else {
		return enforcer.AddGroupingPolicy(policy)
	}
}

func RemovePolicy(id string, ptype string, policy []string) (bool, error) {
	enforcer, err := GetInitializedEnforcer(id)
	if err != nil {
		return false, err
	}

	if ptype == "p" {
		return enforcer.RemovePolicy(policy)
	} else {
		return enforcer.RemoveGroupingPolicy(policy)
	}
}

func (enforcer *Enforcer) LoadModelCfg() error {
	if enforcer.ModelCfg != nil {
		return nil
	}

	model, err := GetModelEx(enforcer.Model)
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
