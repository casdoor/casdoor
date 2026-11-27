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
	"reflect"
	"slices"
	"testing"

	"github.com/casbin/casbin/v2"
	stringadapter "github.com/qiangmzsx/string-adapter/v2"
	"github.com/xorm-io/core"
	"github.com/xorm-io/xorm"
)

func TestDisabledRolesDoNotGrantAccess(t *testing.T) {
	engine, err := xorm.NewEngine("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	engine.SetMaxOpenConns(1)
	oldOrmer := ormer
	ormer = &Ormer{driverName: "sqlite", Engine: engine}
	t.Cleanup(func() {
		ormer = oldOrmer
		if err := engine.Close(); err != nil {
			t.Error(err)
		}
	})

	if err := engine.Sync2(&User{}, &ThirdPartyLink{}, &Role{}, &Permission{}); err != nil {
		t.Fatal(err)
	}

	user := &User{Owner: "org", Name: "alice", Groups: []string{"org/team"}}
	roles := []*Role{
		{Owner: "org", Name: "member", IsEnabled: true, Users: []string{"org/alice"}},
		{Owner: "org", Name: "disabled_direct", Users: []string{"org/alice"}},
		{Owner: "org", Name: "group", IsEnabled: true, Groups: []string{"org/team"}},
		{Owner: "org", Name: "disabled_group", Groups: []string{"org/team"}},
		{Owner: "org", Name: "parent_enabled", IsEnabled: true, Roles: []string{"org/member"}},
		{Owner: "org", Name: "parent_disabled", Roles: []string{"org/member"}},
		{Owner: "org", Name: "beyond_disabled", IsEnabled: true, Roles: []string{"org/parent_disabled"}},
		{Owner: "org", Name: "parent_of_disabled_direct", IsEnabled: true, Roles: []string{"org/disabled_direct"}},
		{Owner: "org", Name: "parallel_route", IsEnabled: true, Roles: []string{"org/parent_disabled", "org/group"}},
		{Owner: "org", Name: "different_user", IsEnabled: true, Users: []string{"org/alice2"}},
	}
	permissions := []*Permission{
		{Owner: "org", Name: "enabled_permission", IsEnabled: true, Roles: []string{"org/member"}},
		{Owner: "org", Name: "disabled_role_permission", IsEnabled: true, Roles: []string{"org/disabled_direct"}},
		{Owner: "org", Name: "disabled_path_permission", IsEnabled: true, Roles: []string{"org/beyond_disabled"}},
	}
	if _, err := engine.Insert(user); err != nil {
		t.Fatal(err)
	}
	if _, err := engine.Insert(&roles); err != nil {
		t.Fatal(err)
	}
	if _, err := engine.Insert(&permissions); err != nil {
		t.Fatal(err)
	}

	wantRoles := []string{"group", "member", "parallel_route", "parent_enabled"}
	t.Run("roles", func(t *testing.T) {
		got, err := GetRolesByUser(user.GetId())
		if err != nil {
			t.Fatal(err)
		}
		var names []string
		wantSources := map[string][]string{
			"group":          {"org/team"},
			"parallel_route": {"org/team"},
		}
		for _, role := range got {
			names = append(names, role.Name)
			if !slices.Equal(role.SourceGroups, wantSources[role.Name]) {
				t.Errorf("role %s source groups = %v, want %v", role.Name, role.SourceGroups, wantSources[role.Name])
			}
		}
		slices.Sort(names)
		if !reflect.DeepEqual(names, wantRoles) {
			t.Errorf("GetRolesByUser() = %v, want %v", names, wantRoles)
		}
	})

	t.Run("disabled role has no effective ancestors", func(t *testing.T) {
		got, err := GetAncestorRoles("org/disabled_direct")
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 0 {
			t.Errorf("GetAncestorRoles() = %v, want no roles", got)
		}
	})

	t.Run("permissions", func(t *testing.T) {
		got, _, err := getPermissionsAndRolesByUser(user.GetId())
		if err != nil {
			t.Fatal(err)
		}
		var names []string
		for _, permission := range got {
			names = append(names, permission.Name)
		}
		slices.Sort(names)
		if want := []string{"enabled_permission"}; !reflect.DeepEqual(names, want) {
			t.Errorf("permissions = %v, want %v", names, want)
		}
	})

	t.Run("saml attributes", func(t *testing.T) {
		application := &Application{SamlAttributes: []*SamlItem{{Name: "CustomRoles", Value: "$user.roles"}}}
		response, err := NewSamlResponse(application, user, "https://idp.example", "", "https://sp.example", "sp", "", "session", nil)
		if err != nil {
			t.Fatal(err)
		}
		assertion := response.SelectElement("Assertion")
		if assertion == nil || assertion.SelectElement("AttributeStatement") == nil {
			t.Fatal("SAML response does not contain an attribute statement")
		}
		attributes := assertion.SelectElement("AttributeStatement")
		values := map[string][]string{}
		for _, attribute := range attributes.SelectElements("Attribute") {
			name := attribute.SelectAttrValue("Name", "")
			for _, value := range attribute.SelectElements("AttributeValue") {
				values[name] = append(values[name], value.Text())
			}
		}
		for _, name := range []string{"Roles", "CustomRoles"} {
			slices.Sort(values[name])
			if !reflect.DeepEqual(values[name], wantRoles) {
				t.Errorf("SAML %s = %v, want %v", name, values[name], wantRoles)
			}
		}
	})

	t.Run("casbin enforcement", func(t *testing.T) {
		groupModel, err := GetBuiltInModel("")
		if err != nil {
			t.Fatal(err)
		}
		groupEnforcer, err := casbin.NewSyncedEnforcer(groupModel, stringadapter.NewAdapter("g, org/alice, group:org/team"))
		if err != nil {
			t.Fatal(err)
		}
		oldUserEnforcer := userEnforcer
		userEnforcer = NewUserGroupEnforcer(groupEnforcer)
		t.Cleanup(func() { userEnforcer = oldUserEnforcer })

		tests := []struct {
			role       string
			subject    string
			directUser bool
			want       bool
		}{
			{"member", "org/alice", false, true},
			{"group", "org/alice", false, true},
			{"disabled_group", "org/alice", false, false},
			{"parallel_route", "org/alice", false, true},
			{"parent_enabled", "org/alice", false, true},
			{"disabled_direct", "org/alice", false, false},
			{"disabled_direct", "org/disabled_direct", false, false},
			{"disabled_direct", "org/disabled_direct", true, true},
			{"parent_disabled", "org/alice", false, false},
			{"beyond_disabled", "org/alice", false, false},
			{"parent_of_disabled_direct", "org/alice", false, false},
			{"beyond_disabled", "org/parent_disabled", false, false},
		}
		for _, tt := range tests {
			t.Run(tt.role+"/"+tt.subject, func(t *testing.T) {
				permission := &Permission{
					Owner: "org", Name: "permission", Roles: []string{"org/" + tt.role},
					Resources: []string{"resource"}, Actions: []string{"read"}, Effect: "Allow",
				}
				if tt.directUser {
					permission.Users = []string{tt.subject}
				}
				model, err := GetBuiltInModel("")
				if err != nil {
					t.Fatal(err)
				}
				enforcer, err := casbin.NewEnforcer(model)
				if err != nil {
					t.Fatal(err)
				}
				if _, err := enforcer.AddPolicies(getPolicies(permission)); err != nil {
					t.Fatal(err)
				}
				if err := loadRuntimeGroupingPolicies(enforcer, permission); err != nil {
					t.Fatal(err)
				}
				if err := removeDisabledRolePolicies(enforcer, permission); err != nil {
					t.Fatal(err)
				}
				got, err := enforcer.Enforce(tt.subject, "resource", "read")
				if err != nil {
					t.Fatal(err)
				}
				if got != tt.want {
					t.Errorf("Enforce(%s) = %v, want %v", tt.subject, got, tt.want)
				}
			})
		}
	})

	t.Run("wildcard role resolution", func(t *testing.T) {
		for _, wildcard := range []string{"*", "org/*"} {
			policies, err := getRuntimeGroupingPolicies([]*Permission{{Owner: "org", Roles: []string{wildcard}}})
			if err != nil {
				t.Fatal(err)
			}
			for _, policy := range policies {
				for _, roleId := range []string{"org/disabled_direct", "org/disabled_group", "org/parent_disabled"} {
					if policy[0] == roleId || policy[1] == roleId {
						t.Errorf("wildcard %s produced grouping policy through disabled role: %v", wildcard, policy)
					}
				}
			}
		}
	})

	t.Run("domain enforcement", func(t *testing.T) {
		modelText := `[request_definition]
r = sub, dom, obj, act
[policy_definition]
p = sub, dom, obj, act, eft, permissionId
[role_definition]
g = _, _, _
[policy_effect]
e = some(where (p.eft == allow))
[matchers]
m = g(r.sub, p.sub, r.dom) && r.dom == p.dom && r.obj == p.obj && r.act == p.act`
		for _, tt := range []struct {
			role string
			want bool
		}{{"member", true}, {"disabled_direct", false}, {"beyond_disabled", false}} {
			t.Run(tt.role, func(t *testing.T) {
				permission := &Permission{
					Owner: "org", Name: "domain_permission", Roles: []string{"org/" + tt.role}, Domains: []string{"domain"},
					Resources: []string{"resource"}, Actions: []string{"read"}, Effect: "Allow",
				}
				model, err := GetBuiltInModel(modelText)
				if err != nil {
					t.Fatal(err)
				}
				enforcer, err := casbin.NewEnforcer(model)
				if err != nil {
					t.Fatal(err)
				}
				if _, err := enforcer.AddPolicies(getPolicies(permission)); err != nil {
					t.Fatal(err)
				}
				if err := loadRuntimeGroupingPolicies(enforcer, permission); err != nil {
					t.Fatal(err)
				}
				if err := removeDisabledRolePolicies(enforcer, permission); err != nil {
					t.Fatal(err)
				}
				got, err := enforcer.Enforce("org/alice", "domain", "resource", "read")
				if err != nil {
					t.Fatal(err)
				}
				if got != tt.want {
					t.Errorf("domain enforcement = %v, want %v", got, tt.want)
				}
			})
		}
	})

	t.Run("filtering one permission preserves another", func(t *testing.T) {
		rolePermission := &Permission{
			Owner: "org", Name: "role_only", Roles: []string{"org/disabled_direct"},
			Resources: []string{"resource"}, Actions: []string{"read"}, Effect: "Allow",
		}
		userPermission := &Permission{
			Owner: "org", Name: "user_only", Users: []string{"org/disabled_direct"},
			Resources: []string{"resource"}, Actions: []string{"read"}, Effect: "Allow",
		}
		if _, err := engine.Insert(rolePermission, userPermission); err != nil {
			t.Fatal(err)
		}
		model, err := GetBuiltInModel("")
		if err != nil {
			t.Fatal(err)
		}
		enforcer, err := casbin.NewEnforcer(model)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := enforcer.AddPolicies(append(getPolicies(rolePermission), getPolicies(userPermission)...)); err != nil {
			t.Fatal(err)
		}
		if err := removeDisabledRolePolicies(enforcer, rolePermission, rolePermission.GetId(), userPermission.GetId()); err != nil {
			t.Fatal(err)
		}
		policies := enforcer.GetPolicy()
		if len(policies) != 1 || policies[0][5] != userPermission.GetId() {
			t.Errorf("remaining policies = %v, want only the direct user permission", policies)
		}
		got, err := enforcer.Enforce("org/disabled_direct", "resource", "read")
		if err != nil {
			t.Fatal(err)
		}
		if !got {
			t.Error("a separate direct user grant was removed")
		}
	})

	t.Run("persisted policies follow role state", func(t *testing.T) {
		t.Setenv("tableNamePrefix", "")
		permission := &Permission{
			Owner: "org", Name: "persisted_permission", Roles: []string{"org/disabled_direct"},
			Resources: []string{"resource"}, Actions: []string{"read"}, Effect: "Allow",
		}
		if err := addPolicies(permission); err != nil {
			t.Fatal(err)
		}
		for _, enabled := range []bool{false, true, false} {
			_, err := engine.ID(core.PK{"org", "disabled_direct"}).Cols("is_enabled").Update(&Role{IsEnabled: enabled})
			if err != nil {
				t.Fatal(err)
			}
			request := []interface{}{"org/alice", "resource", "read"}
			got, err := Enforce(permission, request)
			if err != nil {
				t.Fatal(err)
			}
			if got != enabled {
				t.Errorf("Enforce() = %v with isEnabled = %v", got, enabled)
			}
			batch, err := BatchEnforce(permission, [][]interface{}{request, {"org/disabled_direct", "resource", "read"}})
			if err != nil {
				t.Fatal(err)
			}
			if want := []bool{enabled, enabled}; !reflect.DeepEqual(batch, want) {
				t.Errorf("BatchEnforce() = %v, want %v", batch, want)
			}
		}
		if err := removePolicies(permission); err != nil {
			t.Fatal(err)
		}
		enforcer, err := getPermissionEnforcer(permission)
		if err != nil {
			t.Fatal(err)
		}
		if got := enforcer.GetPolicy(); len(got) != 0 {
			t.Errorf("policies after removal = %v, want none", got)
		}
	})

	t.Run("wildcard application login", func(t *testing.T) {
		t.Setenv("tableNamePrefix", "")
		user := &User{Owner: "org", Name: "bob"}
		role := &Role{Owner: "org", Name: "disabled_bob", Users: []string{user.GetId()}}
		permission := &Permission{
			Owner: "org", Name: "application_wildcard", Roles: []string{"*"},
			ResourceType: "Application", Resources: []string{"app"}, Actions: []string{"Read"},
			Effect: "Allow", IsEnabled: true, State: "Approved",
		}
		for _, bean := range []interface{}{user, role, permission} {
			if _, err := engine.Insert(bean); err != nil {
				t.Fatal(err)
			}
		}
		if err := addPolicies(permission); err != nil {
			t.Fatal(err)
		}
		application := &Application{Name: "app", Organization: "org"}
		got, err := CheckLoginPermission(user.GetId(), application)
		if err != nil {
			t.Fatal(err)
		}
		if got {
			t.Error("CheckLoginPermission() allowed a user with only a disabled role")
		}
		got, err = CheckLoginPermission("org/alice", application)
		if err != nil {
			t.Fatal(err)
		}
		if !got {
			t.Error("CheckLoginPermission() rejected a user with enabled roles")
		}
	})

	t.Run("reenabling and disabling a role", func(t *testing.T) {
		for _, enabled := range []bool{true, false} {
			_, err := engine.ID(core.PK{"org", "disabled_direct"}).Cols("is_enabled").Update(&Role{IsEnabled: enabled})
			if err != nil {
				t.Fatal(err)
			}
			got, err := GetRolesByUser(user.GetId())
			if err != nil {
				t.Fatal(err)
			}
			names := map[string]bool{}
			for _, role := range got {
				names[role.Name] = true
			}
			for _, name := range []string{"disabled_direct", "parent_of_disabled_direct"} {
				if names[name] != enabled {
					t.Errorf("role %s present = %v after setting isEnabled to %v", name, names[name], enabled)
				}
			}
		}
	})
}
