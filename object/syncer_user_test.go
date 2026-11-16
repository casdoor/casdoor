// Copyright 2021 The Casdoor Authors. All Rights Reserved.
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
	"testing"

	"github.com/casdoor/casdoor/util"
)

func TestGetUsers(t *testing.T) {
	InitConfig()
	syncers, _ := GetSyncers("admin")
	syncer := syncers[0]
	syncer.initAdapter()
	users, _ := syncer.getOriginalUsers()
	for _, user := range users {
		fmt.Printf("%v\n", user)
	}
}

func TestSyncUsers(t *testing.T) {
	InitConfig()

	RunSyncUsersJob()
}

func TestGetUpdateColumns(t *testing.T) {
	newSyncer := func(syncerType string, tableColumns []*TableColumn) *Syncer {
		return &Syncer{
			Type:         syncerType,
			TableColumns: tableColumns,
		}
	}

	// Providers that derive User.IsForbidden from the upstream account state must
	// persist it on update too, otherwise it is only written at creation time.
	for _, syncerType := range []string{
		"Lark", "DingTalk", "WeCom", "Okta", "SCIM",
		"Azure AD", "Google Workspace", "Active Directory", "AWS IAM", "Plugin",
	} {
		columns := newSyncer(syncerType, []*TableColumn{{Name: "enterprise_email", CasdoorName: "Email"}}).getUpdateColumns(&OriginalUser{})
		if !util.InSlice(columns, "is_forbidden") {
			t.Errorf("%s syncer: is_forbidden must be updated, got columns %v", syncerType, columns)
		}
	}

	// Database-backed syncers only write it when the admin maps it explicitly.
	for _, syncerType := range []string{"Database", "Keycloak"} {
		columns := newSyncer(syncerType, []*TableColumn{{Name: "email", CasdoorName: "Email"}}).getUpdateColumns(&OriginalUser{})
		if util.InSlice(columns, "is_forbidden") {
			t.Errorf("%s syncer: is_forbidden must not be updated unless mapped, got columns %v", syncerType, columns)
		}

		columns = newSyncer(syncerType, []*TableColumn{{Name: "enabled", CasdoorName: "IsForbidden"}}).getUpdateColumns(&OriginalUser{})
		if !util.InSlice(columns, "is_forbidden") {
			t.Errorf("%s syncer: a mapped IsForbidden column must be updated, got columns %v", syncerType, columns)
		}
	}
}
