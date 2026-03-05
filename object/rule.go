// Copyright 2023 The casbin Authors. All Rights Reserved.
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

	"github.com/casdoor/casdoor/util"
	"github.com/xorm-io/core"
)

type Expression struct {
	Name     string `json:"name"`
	Operator string `json:"operator"`
	Value    string `json:"value"`
}

type Rule struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100) notnull" json:"createdTime"`
	UpdatedTime string `xorm:"varchar(100) notnull" json:"updatedTime"`

	Type        string        `xorm:"varchar(100) notnull" json:"type"`
	Expressions []*Expression `xorm:"mediumtext" json:"expressions"`
	Action      string        `xorm:"varchar(100) notnull" json:"action"`
	StatusCode  int           `xorm:"int notnull" json:"statusCode"`
	Reason      string        `xorm:"varchar(100) notnull" json:"reason"`
	IsVerbose   bool          `xorm:"bool" json:"isVerbose"`
}

func GetGlobalRules() ([]*Rule, error) {
	rules := []*Rule{}
	err := ormer.Engine.Asc("owner").Desc("created_time").Find(&rules)
	return rules, err
}

func GetRules(owner string) ([]*Rule, error) {
	rules := []*Rule{}
	err := ormer.Engine.Desc("updated_time").Find(&rules, &Rule{Owner: owner})
	return rules, err
}

func getRule(owner string, name string) (*Rule, error) {
	rule := Rule{Owner: owner, Name: name}
	existed, err := ormer.Engine.Get(&rule)
	if err != nil {
		return nil, err
	}
	if existed {
		return &rule, nil
	} else {
		return nil, nil
	}
}

func GetRule(id string) (*Rule, error) {
	owner, name := util.GetOwnerAndNameFromIdNoCheck(id)
	return getRule(owner, name)
}

func UpdateRule(id string, rule *Rule) (bool, error) {
	owner, name := util.GetOwnerAndNameFromIdNoCheck(id)
	if s, err := getRule(owner, name); err != nil {
		return false, err
	} else if s == nil {
		return false, nil
	}
	rule.UpdatedTime = util.GetCurrentTime()
	_, err := ormer.Engine.ID(core.PK{owner, name}).AllCols().Update(rule)
	if err != nil {
		return false, err
	}
	err = refreshRuleMap()
	if err != nil {
		return false, err
	}
	return true, nil
}

func AddRule(rule *Rule) (bool, error) {
	affected, err := ormer.Engine.Insert(rule)
	if err != nil {
		return false, err
	}
	if affected != 0 {
		err = refreshRuleMap()
		if err != nil {
			return false, err
		}
	}
	return affected != 0, nil
}

func DeleteRule(rule *Rule) (bool, error) {
	affected, err := ormer.Engine.ID(core.PK{rule.Owner, rule.Name}).Delete(&Rule{})
	if err != nil {
		return false, err
	}
	if affected != 0 {
		err = refreshRuleMap()
		if err != nil {
			return false, err
		}
	}
	return affected != 0, nil
}

func (rule *Rule) GetId() string {
	return fmt.Sprintf("%s/%s", rule.Owner, rule.Name)
}

func GetRuleCount(owner, field, value string) (int64, error) {
	session := GetSession(owner, -1, -1, field, value, "", "")
	return session.Count(&Rule{})
}

func GetPaginationRules(owner string, offset, limit int, field, value, sortField, sortOrder string) ([]*Rule, error) {
	rules := []*Rule{}
	session := GetSession(owner, offset, limit, field, value, sortField, sortOrder)
	err := session.Where("owner = ? or owner = ?", "admin", owner).Find(&rules)
	if err != nil {
		return rules, err
	}

	return rules, nil
}
