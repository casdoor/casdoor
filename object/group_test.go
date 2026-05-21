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
	"path/filepath"
	"runtime"
	"testing"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	xormadapter "github.com/casdoor/xorm-adapter/v3"
)

func initGroupUserTestStore(t *testing.T) {
	t.Helper()

	oldOrmer := ormer
	oldUserEnforcer := userEnforcer

	testOrmer, err := NewAdapter("sqlite3", filepath.Join(t.TempDir(), "casdoor-test.db"), "")
	if err != nil {
		t.Fatalf("NewAdapter() error = %v", err)
	}
	ormer = testOrmer
	t.Cleanup(func() {
		runtime.SetFinalizer(testOrmer, nil)
		ormer.close()
		ormer = oldOrmer
		userEnforcer = oldUserEnforcer
	})

	if err = ormer.Engine.Sync2(new(User), new(Group)); err != nil {
		t.Fatalf("Sync2() error = %v", err)
	}

	m, err := model.NewModelFromString(`
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act
`)
	if err != nil {
		t.Fatalf("NewModelFromString() error = %v", err)
	}

	adapter, err := xormadapter.NewAdapterByEngineWithTableName(ormer.Engine, "casbin_user_rule", "")
	if err != nil {
		t.Fatalf("NewAdapterByEngineWithTableName() error = %v", err)
	}
	enforcer, err := casbin.NewEnforcer(m, adapter)
	if err != nil {
		t.Fatalf("NewEnforcer() error = %v", err)
	}
	userEnforcer = NewUserGroupEnforcer(enforcer)
}

func TestGetGroupUsersIncludesUsersAssignedByUserGroups(t *testing.T) {
	initGroupUserTestStore(t)

	groupId := "test/group1"
	user := &User{
		Owner:  "test",
		Name:   "alice",
		Groups: []string{groupId},
	}
	group := &Group{
		Owner:       "test",
		Name:        "group1",
		DisplayName: "Group 1",
	}

	if _, err := ormer.Engine.Insert(group, user); err != nil {
		t.Fatalf("Insert() error = %v", err)
	}

	users, err := GetGroupUsers(groupId)
	if err != nil {
		t.Fatalf("GetGroupUsers() error = %v", err)
	}

	if len(users) != 1 || users[0].GetId() != user.GetId() {
		t.Fatalf("GetGroupUsers(%q) = %#v, want user %q from User.Groups", groupId, users, user.GetId())
	}
}

func TestExtendGroupWithUsersIncludesUsersAssignedByUserGroups(t *testing.T) {
	initGroupUserTestStore(t)

	group := &Group{
		Owner:       "test",
		Name:        "group1",
		DisplayName: "Group 1",
	}
	user := &User{
		Owner:  "test",
		Name:   "alice",
		Groups: []string{group.GetId()},
	}

	if _, err := ormer.Engine.Insert(group, user); err != nil {
		t.Fatalf("Insert() error = %v", err)
	}

	if err := ExtendGroupWithUsers(group); err != nil {
		t.Fatalf("ExtendGroupWithUsers() error = %v", err)
	}

	if len(group.Users) != 1 || group.Users[0] != user.GetId() {
		t.Fatalf("ExtendGroupWithUsers() set group.Users = %#v, want [%q]", group.Users, user.GetId())
	}
}
