// Copyright 2025 The Casdoor Authors. All Rights Reserved.
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
	xormadapter "github.com/casdoor/xorm-adapter/v3"
	"github.com/xorm-io/xorm"
)

// safeAdapter wraps xormadapter.Adapter to fix the RemovePolicy bug
// where xorm's Delete ignores zero-value (empty string) fields in the
// WHERE clause, causing all rules of a ptype to be deleted instead of
// just the target rule.
type safeAdapter struct {
	*xormadapter.Adapter
	engine    *xorm.Engine
	tableName string
}

func newSafeAdapter(adapter *xormadapter.Adapter, engine *xorm.Engine, tableName string) *safeAdapter {
	return &safeAdapter{
		Adapter:   adapter,
		engine:    engine,
		tableName: tableName,
	}
}

// RemovePolicy removes a policy rule from the storage.
// Uses AllCols() to include empty string fields in the WHERE clause.
func (a *safeAdapter) RemovePolicy(sec string, ptype string, rule []string) error {
	line := buildPolicyLine(ptype, rule)
	_, err := a.engine.Table(a.tableName).AllCols().Delete(line)
	return err
}

// RemovePolicies removes multiple policy rules from the storage.
// Uses AllCols() to include empty string fields in the WHERE clause.
func (a *safeAdapter) RemovePolicies(sec string, ptype string, rules [][]string) error {
	_, err := a.engine.Transaction(func(tx *xorm.Session) (interface{}, error) {
		for _, rule := range rules {
			line := buildPolicyLine(ptype, rule)
			_, err := tx.Table(a.tableName).AllCols().Delete(line)
			if err != nil {
				return nil, err
			}
		}
		return nil, nil
	})
	return err
}

func buildPolicyLine(ptype string, rule []string) *xormadapter.CasbinRule {
	line := &xormadapter.CasbinRule{Ptype: ptype}

	l := len(rule)
	if l > 0 {
		line.V0 = rule[0]
	}
	if l > 1 {
		line.V1 = rule[1]
	}
	if l > 2 {
		line.V2 = rule[2]
	}
	if l > 3 {
		line.V3 = rule[3]
	}
	if l > 4 {
		line.V4 = rule[4]
	}
	if l > 5 {
		line.V5 = rule[5]
	}

	return line
}
