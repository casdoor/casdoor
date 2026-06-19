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
	"testing"

	"github.com/casbin/casbin/v2/model"
	xormadapter "github.com/casdoor/xorm-adapter/v3"
	"github.com/xorm-io/xorm"
	_ "modernc.org/sqlite"
)

func TestSafeAdapterSavePolicyReplacesRulesWithoutDroppingTable(t *testing.T) {
	engine, err := xorm.NewEngine("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("NewEngine() error = %v", err)
	}
	defer engine.Close()

	const tableName = "casbin_api_rule"

	err = engine.Sync2(&xormadapter.CasbinRule{})
	if err != nil {
		t.Fatalf("Sync2() error = %v", err)
	}
	_, err = engine.Exec("ALTER TABLE casbin_rule RENAME TO " + tableName)
	if err != nil {
		t.Fatalf("rename table error = %v", err)
	}
	_, err = engine.Exec("CREATE INDEX casbin_api_rule_marker_idx ON " + tableName + " (ptype)")
	if err != nil {
		t.Fatalf("create marker index error = %v", err)
	}

	safeAdapter := &SafeAdapter{
		engine:    engine,
		tableName: tableName,
	}

	err = safeAdapter.SavePolicy(newPolicyModel(t, [][]string{
		{"built-in", "*", "*", "*", "*", "*"},
	}))
	if err != nil {
		t.Fatalf("first SavePolicy() error = %v", err)
	}

	rules := getSafeAdapterRules(t, safeAdapter)
	if len(rules) != 1 {
		t.Fatalf("len(rules) = %d, want 1", len(rules))
	}

	err = safeAdapter.SavePolicy(newPolicyModel(t, [][]string{
		{"app", "*", "GET", "/api/get-account", "*", "*"},
		{"*", "*", "POST", "/api/login", "*", "*"},
	}))
	if err != nil {
		t.Fatalf("second SavePolicy() error = %v", err)
	}

	rules = getSafeAdapterRules(t, safeAdapter)
	if len(rules) != 2 {
		t.Fatalf("len(rules) = %d, want 2", len(rules))
	}

	err = safeAdapter.UpdatePolicy("p", "p",
		[]string{"app", "*", "GET", "/api/get-account", "*", "*"},
		[]string{"app", "*", "GET", "/api/get-user", "*", "*"})
	if err != nil {
		t.Fatalf("UpdatePolicy() error = %v", err)
	}

	rules = getSafeAdapterRules(t, safeAdapter)
	if len(rules) != 2 {
		t.Fatalf("len(rules) after update = %d, want 2", len(rules))
	}
	assertPolicyExists(t, rules, []string{"app", "*", "GET", "/api/get-user", "*", "*"})

	err = safeAdapter.RemovePolicy("p", "p", []string{"*", "*", "POST", "/api/login", "*", "*"})
	if err != nil {
		t.Fatalf("RemovePolicy() error = %v", err)
	}

	rules = getSafeAdapterRules(t, safeAdapter)
	if len(rules) != 1 {
		t.Fatalf("len(rules) after remove = %d, want 1", len(rules))
	}

	var markerIndexCount int64
	markerIndexCount, err = engine.SQL("SELECT COUNT(*) FROM sqlite_master WHERE type = 'index' AND tbl_name = ? AND name = ?", tableName, "casbin_api_rule_marker_idx").Count()
	if err != nil {
		t.Fatalf("marker index query error = %v", err)
	}
	if markerIndexCount != 1 {
		t.Fatal("SavePolicy() dropped and recreated the policy table")
	}
}

func newPolicyModel(t *testing.T, policies [][]string) model.Model {
	t.Helper()

	m, err := model.NewModelFromString(`[request_definition]
r = subOwner, subName, method, urlPath, objOwner, objName

[policy_definition]
p = subOwner, subName, method, urlPath, objOwner, objName

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = r.subOwner == p.subOwner && r.subName == p.subName && r.method == p.method && r.urlPath == p.urlPath && r.objOwner == p.objOwner && r.objName == p.objName`)
	if err != nil {
		t.Fatalf("NewModelFromString() error = %v", err)
	}

	for _, policy := range policies {
		if len(policy) != 6 {
			t.Fatalf("policy has %d fields, want 6", len(policy))
		}
		m["p"]["p"].Policy = append(m["p"]["p"].Policy, policy)
	}

	return m
}

func getSafeAdapterRules(t *testing.T, safeAdapter *SafeAdapter) []*xormadapter.CasbinRule {
	t.Helper()

	rules, err := safeAdapter.GetRules()
	if err != nil {
		t.Fatalf("GetRules() error = %v", err)
	}
	return rules
}

func assertPolicyExists(t *testing.T, rules []*xormadapter.CasbinRule, policy []string) {
	t.Helper()

	for _, rule := range rules {
		if rule.Ptype == "p" &&
			rule.V0 == policy[0] &&
			rule.V1 == policy[1] &&
			rule.V2 == policy[2] &&
			rule.V3 == policy[3] &&
			rule.V4 == policy[4] &&
			rule.V5 == policy[5] {
			return
		}
	}

	t.Fatalf("policy %v was not found in rules %#v", policy, rules)
}
