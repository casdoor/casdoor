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
	"reflect"
	"testing"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	xormadapter "github.com/casdoor/xorm-adapter/v3"
	"github.com/xorm-io/xorm"
)

func TestUserSignInStatus(t *testing.T) {
	t.Setenv("tableNamePrefix", "")
	engine, err := xorm.NewEngine("sqlite", filepath.Join(t.TempDir(), "users.db"))
	if err != nil {
		t.Fatal(err)
	}
	previousOrmer, previousEnforcer := ormer, userEnforcer
	ormer = &Ormer{Engine: engine}
	t.Cleanup(func() {
		ormer, userEnforcer = previousOrmer, previousEnforcer
		_ = engine.Close()
	})
	if err := engine.Sync2(new(User)); err != nil {
		t.Fatal(err)
	}
	rbac, err := model.NewModelFromString(`[request_definition]
r = sub, obj, act
[policy_definition]
p = sub, obj, act
[role_definition]
g = _, _
[policy_effect]
e = some(where (p.eft == allow))
[matchers]
m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act`)
	if err != nil {
		t.Fatal(err)
	}
	adapter, err := xormadapter.NewAdapterByEngineWithTableName(engine, "casbin_user_rule", "")
	if err != nil {
		t.Fatal(err)
	}
	enforcer, err := casbin.NewSyncedEnforcer(rbac, adapter)
	if err != nil {
		t.Fatal(err)
	}
	userEnforcer = NewUserGroupEnforcer(enforcer)
	users := []*User{
		{Owner: "example", Name: "a-enabled"},
		{Owner: "example", Name: "b-forbidden", IsForbidden: true},
		{Owner: "example", Name: "c-deleted", IsDeleted: true},
		{Owner: "example", Name: "d-both", IsForbidden: true, IsDeleted: true},
		{Owner: "example", Name: "e-legacy"},
		{Owner: "other", Name: "foreign"},
	}
	for _, user := range users {
		if _, err := engine.Insert(user); err != nil {
			t.Fatal(err)
		}
		if user.Owner == "example" {
			if _, err := userEnforcer.AddGroupForUser(user.GetId(), "example/team"); err != nil {
				t.Fatal(err)
			}
		}
	}
	if _, err := engine.Exec(`UPDATE "user" SET is_forbidden = NULL, is_deleted = NULL WHERE name = ?`, "e-legacy"); err != nil {
		t.Fatal(err)
	}

	for _, group := range []string{"", "team"} {
		for _, status := range []string{"enabled", "disabled", ""} {
			t.Run(group+"/"+status, func(t *testing.T) {
				want := []string{"a-enabled", "e-legacy"}
				if status == "disabled" {
					want = []string{"b-forbidden", "c-deleted", "d-both"}
				} else if status == "" {
					want = []string{"a-enabled", "b-forbidden", "c-deleted", "d-both", "e-legacy"}
				}
				count, err := GetUserCount("example", "", "", group, status)
				if err != nil || count != int64(len(want)) {
					t.Fatalf("count = %d, want %d: %v", count, len(want), err)
				}
				var got []string
				for offset := range want {
					page, err := GetPaginationUsers("example", offset, 1, "", "", "name", "ascend", group, status)
					if err != nil || len(page) != 1 {
						t.Fatalf("page %d: %v, %v", offset, page, err)
					}
					got = append(got, page[0].Name)
				}
				if !reflect.DeepEqual(got, want) {
					t.Fatalf("pages = %v, want %v", got, want)
				}
				all, err := GetPaginationUsers("example", -1, -1, "", "", "name", "ascend", group, status)
				if err != nil || len(all) != len(want) {
					t.Fatalf("unpaginated length = %d, want %d: %v", len(all), len(want), err)
				}
				count, err = GetUserCount("example", "name", want[0], group, status)
				if err != nil || count != 1 {
					t.Fatalf("matching search count = %d, want 1: %v", count, err)
				}
				page, err := GetPaginationUsers("example", 0, 1, "name", want[0], "name", "ascend", group, status)
				if err != nil || len(page) != 1 || page[0].Name != want[0] {
					t.Fatalf("matching search page = %v: %v", page, err)
				}
				count, err = GetUserCount("example", "name", "missing", group, status)
				if err != nil || count != 0 {
					t.Fatalf("empty search count = %d: %v", count, err)
				}
			})
		}
	}
	count, err := GetGlobalUserCount("", "", "enabled")
	if err != nil || count != 3 {
		t.Fatalf("global count = %d, want 3: %v", count, err)
	}
	page, err := GetPaginationGlobalUsers(0, 100, "", "", "name", "ascend", "enabled")
	if err != nil || len(page) != 3 {
		t.Fatalf("global page length = %d, want 3: %v", len(page), err)
	}
	count, err = GetUserCount("example", "", "", "")
	if err != nil || count != 5 {
		t.Fatalf("omitted status count = %d, want 5: %v", count, err)
	}
}
