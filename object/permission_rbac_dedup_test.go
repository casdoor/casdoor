// Copyright 2024 The Casdoor Authors. All Rights Reserved.
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

//go:build !skipCi

package object

import (
	"fmt"
	"sync"
	"testing"

	"github.com/casdoor/casdoor/util"
)

type permissionRuleRecord struct {
	Id    int64  `xorm:"pk autoincr"`
	Ptype string `xorm:"varchar(100) index not null default ''"`
	V0    string `xorm:"varchar(100) index not null default ''"`
	V1    string `xorm:"varchar(100) index not null default ''"`
	V2    string `xorm:"varchar(100) index not null default ''"`
	V3    string `xorm:"varchar(100) index not null default ''"`
	V4    string `xorm:"varchar(100) index not null default ''"`
	V5    string `xorm:"varchar(100) index not null default ''"`
}

func (permissionRuleRecord) TableName() string {
	return "permission_rule"
}

var permissionRbacTestInit sync.Once

func initPermissionRbacTestDb(t *testing.T) {
	t.Helper()

	permissionRbacTestInit.Do(func() {
		oldCreateDatabase := createDatabase
		createDatabase = false
		InitConfig()
		createDatabase = oldCreateDatabase
	})
}

func newPermissionRbacTestOwner(t *testing.T) string {
	t.Helper()

	initPermissionRbacTestDb(t)

	owner := "rbac-dedup-" + util.GenerateId()

	t.Cleanup(func() {
		_, err := ormer.Engine.Where("v5 like ?", owner+"/%").Delete(&permissionRuleRecord{})
		if err != nil {
			t.Fatalf("failed to delete permission rules for owner %s: %v", owner, err)
		}

		_, err = ormer.Engine.Where("owner = ?", owner).Delete(&Permission{})
		if err != nil {
			t.Fatalf("failed to delete permissions for owner %s: %v", owner, err)
		}

		_, err = ormer.Engine.Where("owner = ?", owner).Delete(&Role{})
		if err != nil {
			t.Fatalf("failed to delete roles for owner %s: %v", owner, err)
		}
	})

	return owner
}

func newTestPermission(owner string, name string, roleIDs ...string) *Permission {
	return &Permission{
		Owner:     owner,
		Name:      name,
		Roles:     roleIDs,
		Resources: []string{"data1"},
		Actions:   []string{"read"},
		Effect:    "Allow",
	}
}

func getPermissionRulesByPermissionID(t *testing.T, permissionID string) []permissionRuleRecord {
	t.Helper()

	rules := make([]permissionRuleRecord, 0)
	err := ormer.Engine.Where("v5 = ?", permissionID).Asc("id").Find(&rules)
	if err != nil {
		t.Fatalf("failed to query permission rules for %s: %v", permissionID, err)
	}

	return rules
}

func TestPermissionRuntimeGroupingIgnoresPersistedG(t *testing.T) {
	owner := newPermissionRbacTestOwner(t)

	role := &Role{
		Owner: owner,
		Name:  "reader",
		Users: []string{owner + "/alice"},
	}
	affected, err := AddRole(role)
	if err != nil {
		t.Fatalf("AddRole() error: %v", err)
	}
	if !affected {
		t.Fatalf("expected AddRole to affect rows")
	}

	permission := newTestPermission(owner, "perm-reader", role.GetId())
	affected, err = AddPermission(permission)
	if err != nil {
		t.Fatalf("AddPermission() error: %v", err)
	}
	if !affected {
		t.Fatalf("expected AddPermission to affect rows")
	}

	rules := getPermissionRulesByPermissionID(t, permission.GetId())
	if len(rules) != 1 || rules[0].Ptype != "p" {
		t.Fatalf("expected exactly one persisted p rule, got %+v", rules)
	}

	allowed, err := Enforce(permission, []string{owner + "/alice", "data1", "read"})
	if err != nil {
		t.Fatalf("Enforce() for alice error: %v", err)
	}
	if !allowed {
		t.Fatalf("expected alice to be allowed")
	}

	_, err = ormer.Engine.Insert(&permissionRuleRecord{
		Ptype: "g",
		V0:    owner + "/mallory",
		V1:    role.GetId(),
		V5:    permission.GetId(),
	})
	if err != nil {
		t.Fatalf("failed to insert legacy g rule: %v", err)
	}

	allowed, err = Enforce(permission, []string{owner + "/mallory", "data1", "read"})
	if err != nil {
		t.Fatalf("Enforce() for mallory error: %v", err)
	}
	if allowed {
		t.Fatalf("expected legacy persisted g rule to be ignored")
	}
}

func TestUpdateRoleUsesRuntimeGroupingAndOnlyRenameRewritesP(t *testing.T) {
	owner := newPermissionRbacTestOwner(t)

	role := &Role{
		Owner: owner,
		Name:  "reader-old",
		Users: []string{owner + "/alice"},
	}
	affected, err := AddRole(role)
	if err != nil {
		t.Fatalf("AddRole() error: %v", err)
	}
	if !affected {
		t.Fatalf("expected AddRole to affect rows")
	}

	permission := newTestPermission(owner, "perm-reader", role.GetId())
	affected, err = AddPermission(permission)
	if err != nil {
		t.Fatalf("AddPermission() error: %v", err)
	}
	if !affected {
		t.Fatalf("expected AddPermission to affect rows")
	}

	rulesBefore := getPermissionRulesByPermissionID(t, permission.GetId())

	updatedRole := *role
	updatedRole.Users = []string{owner + "/bob"}
	affected, err = UpdateRole(role.GetId(), &updatedRole)
	if err != nil {
		t.Fatalf("UpdateRole() for membership change error: %v", err)
	}
	if !affected {
		t.Fatalf("expected UpdateRole membership change to affect rows")
	}

	rulesAfterMembershipChange := getPermissionRulesByPermissionID(t, permission.GetId())
	if fmt.Sprintf("%#v", rulesBefore) != fmt.Sprintf("%#v", rulesAfterMembershipChange) {
		t.Fatalf("expected membership change to keep persisted permission rules unchanged")
	}

	allowed, err := Enforce(permission, []string{owner + "/alice", "data1", "read"})
	if err != nil {
		t.Fatalf("Enforce() for alice after membership change error: %v", err)
	}
	if allowed {
		t.Fatalf("expected alice to lose permission after membership change")
	}

	allowed, err = Enforce(permission, []string{owner + "/bob", "data1", "read"})
	if err != nil {
		t.Fatalf("Enforce() for bob after membership change error: %v", err)
	}
	if !allowed {
		t.Fatalf("expected bob to gain permission after membership change")
	}

	renamedRole := updatedRole
	renamedRole.Name = "reader-new"
	affected, err = UpdateRole(updatedRole.GetId(), &renamedRole)
	if err != nil {
		t.Fatalf("UpdateRole() for rename error: %v", err)
	}
	if !affected {
		t.Fatalf("expected UpdateRole rename to affect rows")
	}

	updatedPermission, err := GetPermission(permission.GetId())
	if err != nil {
		t.Fatalf("GetPermission() error: %v", err)
	}
	if len(updatedPermission.Roles) != 1 || updatedPermission.Roles[0] != renamedRole.GetId() {
		t.Fatalf("expected permission role reference to be renamed")
	}

	rulesAfterRename := getPermissionRulesByPermissionID(t, permission.GetId())
	if len(rulesAfterRename) != 1 || rulesAfterRename[0].Ptype != "p" || rulesAfterRename[0].V0 != renamedRole.GetId() {
		t.Fatalf("expected rename to rebuild persisted p rule with new role id, got %+v", rulesAfterRename)
	}

	allowed, err = Enforce(updatedPermission, []string{owner + "/bob", "data1", "read"})
	if err != nil {
		t.Fatalf("Enforce() for bob after rename error: %v", err)
	}
	if !allowed {
		t.Fatalf("expected bob to stay allowed after role rename")
	}
}

// issue 5346
func TestPermissionEnforcerDeduplicatesRuntimeGroupingPoliciesAcross1000Permissions(t *testing.T) {
	owner := newPermissionRbacTestOwner(t)

	const (
		permissionCount = 1000
		userCount       = 1000
	)

	users := make([]string, 0, userCount)
	for i := range userCount {
		users = append(users, fmt.Sprintf("%s/user-%04d", owner, i))
	}

	role := &Role{
		Owner: owner,
		Name:  "shared-role",
		Users: users,
	}
	affected, err := AddRole(role)
	if err != nil {
		t.Fatalf("AddRole() error: %v", err)
	}
	if !affected {
		t.Fatalf("expected AddRole to affect rows")
	}

	permissions := make([]*Permission, 0, permissionCount)
	permissionIDs := make([]string, 0, permissionCount)
	for i := 0; i < permissionCount; i++ {
		permission := newTestPermission(owner, fmt.Sprintf("perm-%04d", i), role.GetId())
		permissions = append(permissions, permission)
		permissionIDs = append(permissionIDs, permission.GetId())
	}

	affected, err = AddPermissions(permissions)
	if err != nil {
		t.Fatalf("AddPermissions() error: %v", err)
	}
	if !affected {
		t.Fatalf("expected AddPermissions to affect rows")
	}

	enforcer, err := getPermissionEnforcer(permissions[0], permissionIDs...)
	if err != nil {
		t.Fatalf("getPermissionEnforcer() error: %v", err)
	}

	if len(enforcer.GetPolicy()) != permissionCount {
		t.Fatalf("expected %d p rules in merged enforcer, got %d", permissionCount, len(enforcer.GetPolicy()))
	}
	if len(enforcer.GetGroupingPolicy()) != userCount {
		t.Fatalf("expected deduplicated runtime g rules to stay at %d, got %d", userCount, len(enforcer.GetGroupingPolicy()))
	}

	allowed, err := enforcer.Enforce(users[userCount-1], "data1", "read")
	if err != nil {
		t.Fatalf("Enforce() in 1000x1000 scenario error: %v", err)
	}
	if !allowed {
		t.Fatalf("expected last user to be allowed in 1000x1000 scenario")
	}
}
