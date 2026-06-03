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
	"slices"
	"strings"
	"testing"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"github.com/casdoor/casdoor/util"
	xormadapter "github.com/casdoor/xorm-adapter/v3"
)

func TestSyncLdapUsersAssignsMemberOfGroupsExistUser(t *testing.T) {
	initLdapConnTestDb(t)

	fixture := createLdapSyncFixture(t)
	userName := "ldap-user-" + fixture.suffix

	tests := []struct {
		name           string
		initialGroups  []string
		ldapGroups     []string
		expectedGroups []string
	}{
		{
			"test_one_group_excluded",
			[]string{fixture.orgName + "/Groups_Engineering", fixture.orgName + "/Groups_Admins"},
			[]string{"cn=Engineering,ou=Groups,dc=example,dc=org"},
			[]string{fixture.orgName + "/Groups_Engineering"},
		},
		{
			"test_no_changes",
			[]string{fixture.orgName + "/Groups_Engineering", fixture.orgName + "/Groups_Admins"},
			[]string{
				"cn=Engineering,ou=Groups,dc=example,dc=org",
				"cn=Admins,ou=Groups,dc=example,dc=org",
			},
			[]string{fixture.orgName + "/Groups_Engineering", fixture.orgName + "/Groups_Admins"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			for _, groupName := range tc.initialGroups {
				origGroupName := strings.Split(groupName, "/")[1]
				group := &Group{
					Owner: fixture.orgName,
					Name:  origGroupName,
				}
				_, err := AddGroup(group)
				if err != nil {
					t.Fatalf("AddGroup() error = %v", err)
				}
				t.Cleanup(func() {
					_, _ = DeleteGroup(group)
				})
			}

			user := &User{
				Owner:       fixture.orgName,
				Name:        userName,
				CreatedTime: util.GetCurrentTime(),
				Groups:      tc.initialGroups,
				Ldap:        fixture.ldapId,
			}
			_, err := AddUser(user, "en")
			if err != nil {
				t.Fatalf("AddUser() error = %v", err)
			}
			t.Cleanup(func() {
				_, _ = DeleteUser(user)
			})

			ldapUsers := []LdapUser{
				{
					Uid:         userName,
					UidNumber:   "1001",
					Cn:          "LDAP Test User",
					Uuid:        fixture.ldapId,
					DisplayName: "LDAP Test User",
					Email:       "ldap-user@example.org",
					MemberOf:    tc.ldapGroups,
				},
			}

			existingUsers, failedUsers, err := SyncLdapUsers(fixture.orgName, ldapUsers, fixture.ldapId)
			if err != nil {
				t.Fatalf("SyncLdapUsers() error = %v", err)
			}
			if len(existingUsers) != 1 {
				t.Fatalf("SyncLdapUsers() existingUsers length = %d, want 1", len(existingUsers))
			}
			if len(failedUsers) != 0 {
				t.Fatalf("SyncLdapUsers() failedUsers length = %d, want 0", len(failedUsers))
			}

			userUpdated, err := GetUserNoCheck(fixture.orgName + "/" + userName)
			if err != nil {
				t.Fatalf("GetUserNoCheck() error = %v", err)
			}
			if userUpdated == nil {
				t.Fatalf("GetUserNoCheck() returned nil, want synced LDAP user")
			}
			if userUpdated.Ldap != fixture.ldapId {
				t.Fatalf("user.Ldap = %q, want %q", user.Ldap, fixture.ldapId)
			}

			if !sameStringSet(userUpdated.Groups, tc.expectedGroups) {
				t.Fatalf("user.Groups = %#v, want %#v", userUpdated.Groups, tc.expectedGroups)
			}

			enforcerGroups, err := userEnforcer.GetGroupsForUser(user.GetId())
			if err != nil {
				t.Fatalf("GetGroupsForUser() error = %v", err)
			}
			if !sameStringSet(enforcerGroups, tc.expectedGroups) {
				t.Fatalf("enforcer groups = %#v, want %#v", enforcerGroups, tc.expectedGroups)
			}
		})
	}
}

func TestSyncLdapUsersAssignsMemberOfGroupsNewUser(t *testing.T) {
	initLdapConnTestDb(t)

	fixture := createLdapSyncFixture(t)
	userName := "ldap-user-" + fixture.suffix

	ldapUsers := []LdapUser{
		{
			Uid:         userName,
			UidNumber:   "1001",
			Cn:          "LDAP Test User",
			Uuid:        fixture.ldapId,
			DisplayName: "LDAP Test User",
			Email:       "ldap-user@example.org",
			MemberOf: []string{
				"cn=Engineering,ou=Groups,dc=example,dc=org",
				"cn=Admins,ou=Groups,dc=example,dc=org",
			},
		},
	}

	existingUsers, failedUsers, err := SyncLdapUsers(fixture.orgName, ldapUsers, fixture.ldapId)
	if err != nil {
		t.Fatalf("SyncLdapUsers() error = %v", err)
	}
	if len(existingUsers) != 0 {
		t.Fatalf("SyncLdapUsers() existingUsers length = %d, want 0", len(existingUsers))
	}
	if len(failedUsers) != 0 {
		t.Fatalf("SyncLdapUsers() failedUsers length = %d, want 0", len(failedUsers))
	}
	t.Cleanup(func() {
		user, _ := GetUserNoCheck(fixture.orgName + "/" + userName)
		if user != nil {
			_, _ = DeleteUser(user)
		}
	})

	user, err := GetUserNoCheck(fixture.orgName + "/" + userName)
	if err != nil {
		t.Fatalf("GetUserNoCheck() error = %v", err)
	}
	if user == nil {
		t.Fatalf("GetUserNoCheck() returned nil, want synced LDAP user")
	}
	if user.Ldap != fixture.ldapId {
		t.Fatalf("user.Ldap = %q, want %q", user.Ldap, fixture.ldapId)
	}

	wantGroups := []string{fixture.orgName + "/Groups_Engineering", fixture.orgName + "/Groups_Admins"}
	if !sameStringSet(user.Groups, wantGroups) {
		t.Fatalf("user.Groups = %#v, want %#v", user.Groups, wantGroups)
	}

	enforcerGroups, err := userEnforcer.GetGroupsForUser(user.GetId())
	if err != nil {
		t.Fatalf("GetGroupsForUser() error = %v", err)
	}
	if !sameStringSet(enforcerGroups, wantGroups) {
		t.Fatalf("enforcer groups = %#v, want %#v", enforcerGroups, wantGroups)
	}
}

type ldapSyncFixture struct {
	suffix  string
	orgName string
	ldapId  string
}

func createLdapSyncFixture(t *testing.T) ldapSyncFixture {
	t.Helper()

	suffix := util.GenerateId()
	orgName := "ldap-sync-test-" + suffix
	appName := "app-" + orgName
	ldapId := "ldap-" + suffix

	organization := &Organization{
		Owner:              "admin",
		Name:               orgName,
		CreatedTime:        util.GetCurrentTime(),
		DisplayName:        "LDAP sync test organization",
		PasswordType:       "plain",
		DefaultApplication: appName,
		InitScore:          2000,
		AccountItems:       GetDefaultAccountItems(),
	}
	_, err := AddOrganization(organization)
	if err != nil {
		t.Fatalf("AddOrganization() error = %v", err)
	}
	t.Cleanup(func() {
		_, _ = DeleteOrganization(organization)
	})

	application := &Application{
		Owner:        "admin",
		Name:         appName,
		CreatedTime:  util.GetCurrentTime(),
		DisplayName:  "LDAP sync test application",
		Organization: orgName,
		EnableSignUp: true,
	}
	_, err = AddApplication(application)
	if err != nil {
		t.Fatalf("AddApplication() error = %v", err)
	}
	t.Cleanup(func() {
		_, _ = DeleteApplication(application)
	})

	ldapServer := &Ldap{
		Id:       ldapId,
		Owner:    orgName,
		Username: "cn=admin,ou=People,dc=example,dc=org",
		BaseDn:   "dc=example,dc=org",
		Filter:   "(objectClass=person)",
	}
	_, err = AddLdap(ldapServer)
	if err != nil {
		t.Fatalf("AddLdap() error = %v", err)
	}
	t.Cleanup(func() {
		_, _ = DeleteLdap(ldapServer)
	})

	return ldapSyncFixture{
		suffix:  suffix,
		orgName: orgName,
		ldapId:  ldapId,
	}
}

func sameStringSet(actual []string, expected []string) bool {
	if len(actual) != len(expected) {
		return false
	}

	actualCopy := slices.Clone(actual)
	expectedCopy := slices.Clone(expected)
	slices.Sort(actualCopy)
	slices.Sort(expectedCopy)

	return slices.Equal(actualCopy, expectedCopy)
}

func initLdapConnTestDb(t *testing.T) {
	t.Helper()

	adapter, err := NewAdapter("sqlite", filepath.Join(t.TempDir(), "casdoor-test.db"), "")
	if err != nil {
		t.Fatalf("NewAdapter() error = %v", err)
	}

	previousOrmer := ormer
	previousUserEnforcer := userEnforcer
	previousCreateDatabase := createDatabase
	ormer = adapter
	createDatabase = false
	t.Cleanup(func() {
		ormer = previousOrmer
		userEnforcer = previousUserEnforcer
		createDatabase = previousCreateDatabase
		adapter.close()
	})

	CreateTables()
	initLdapConnTestUserEnforcer(t, adapter)
}

func initLdapConnTestUserEnforcer(t *testing.T, adapter *Ormer) {
	t.Helper()

	m, err := model.NewModelFromString(`[request_definition]
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
		t.Fatalf("NewModelFromString() error = %v", err)
	}

	xa, err := xormadapter.NewAdapterByEngineWithTableName(adapter.Engine, "casbin_user_rule", "")
	if err != nil {
		t.Fatalf("NewAdapterByEngineWithTableName() error = %v", err)
	}

	enforcer, err := casbin.NewEnforcer(m, xa)
	if err != nil {
		t.Fatalf("casbin.NewEnforcer() error = %v", err)
	}

	userEnforcer = NewUserGroupEnforcer(enforcer)
}
