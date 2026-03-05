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
)

var ruleMap = map[string]*Rule{}

func InitRuleMap() {
	err := refreshRuleMap()
	if err != nil {
		panic(err)
	}
}

func refreshRuleMap() error {
	newRuleMap := map[string]*Rule{}
	rules, err := GetGlobalRules()
	if err != nil {
		return err
	}

	for _, rule := range rules {
		newRuleMap[util.GetId(rule.Owner, rule.Name)] = rule
	}

	ruleMap = newRuleMap
	return nil
}

func GetRulesByRuleIds(ids []string) ([]*Rule, error) {
	var res []*Rule
	for _, id := range ids {
		rule, ok := ruleMap[id]
		if !ok {
			return nil, fmt.Errorf("rule: %s not found", id)
		}
		res = append(res, rule)
	}
	return res, nil
}
