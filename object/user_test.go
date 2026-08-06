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
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/casdoor/casdoor/conf"
	"github.com/casdoor/casdoor/util"
	xormadapter "github.com/casdoor/xorm-adapter/v3"
	"github.com/xorm-io/core"
)

func updateUserColumn(column string, user *User) bool {
	affected, err := ormer.Engine.ID(core.PK{user.Owner, user.Name}).Cols(column).Update(user)
	if err != nil {
		panic(err)
	}

	return affected != 0
}

func TestFaceIdUsesLowerCamelImageUrlJsonField(t *testing.T) {
	var faceId FaceId
	err := json.Unmarshal([]byte(`{"name":"face","imageUrl":"http://example.com/face.jpg","faceIdData":[]}`), &faceId)
	if err != nil {
		t.Fatal(err)
	}

	if faceId.ImageUrl != "http://example.com/face.jpg" {
		t.Fatalf("ImageUrl = %q, want %q", faceId.ImageUrl, "http://example.com/face.jpg")
	}

	data, err := json.Marshal(faceId)
	if err != nil {
		t.Fatal(err)
	}

	var fields map[string]interface{}
	if err := json.Unmarshal(data, &fields); err != nil {
		t.Fatal(err)
	}

	if _, ok := fields["imageUrl"]; !ok {
		t.Fatalf("marshaled FaceId does not contain imageUrl: %s", string(data))
	}
	if _, ok := fields["ImageUrl"]; ok {
		t.Fatalf("marshaled FaceId unexpectedly contains ImageUrl: %s", string(data))
	}
}

func TestSyncAvatarsFromGitHub(t *testing.T) {
	InitConfig()

	users, _ := GetGlobalUsers()
	for _, user := range users {
		if user.GitHub == "" {
			continue
		}

		user.Avatar = fmt.Sprintf("https://avatars.githubusercontent.com/%s", user.GitHub)
		updateUserColumn("avatar", user)
	}
}

func TestSyncIds(t *testing.T) {
	InitConfig()

	users, _ := GetGlobalUsers()
	for _, user := range users {
		if user.Id != "" {
			continue
		}

		user.Id = util.GenerateId()
		updateUserColumn("id", user)
	}
}

func TestSyncHashes(t *testing.T) {
	InitConfig()

	users, _ := GetGlobalUsers()
	for _, user := range users {
		if user.Hash != "" {
			continue
		}

		err := user.UpdateUserHash()
		if err != nil {
			panic(err)
		}
		updateUserColumn("hash", user)
	}
}

func TestGetMaskedUsers(t *testing.T) {
	type args struct {
		users []*User
	}
	tests := []struct {
		name string
		args args
		want []*User
	}{
		{
			name: "1",
			args: args{users: []*User{{Password: "casdoor"}, {Password: "casbin"}}},
			want: []*User{{Password: "***"}, {Password: "***"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := GetMaskedUsers(tt.args.users); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetMaskedUsers() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetUserByField(t *testing.T) {
	InitConfig()

	user, _ := GetUserByField("built-in", "DingTalk", "test")
	if user != nil {
		t.Logf("%+v", user)
	} else {
		t.Log("no user found")
	}
}

func TestGetEmailsForUsers(t *testing.T) {
	InitConfig()

	emailMap := map[string]int{}
	emails := []string{}
	users, _ := GetUsers("built-in")
	for _, user := range users {
		if user.Email == "" {
			continue
		}

		if _, ok := emailMap[user.Email]; !ok {
			emailMap[user.Email] = 1
			emails = append(emails, user.Email)
		}
	}

	text := strings.Join(emails, "\n")
	println(text)
}

func TestUserChangeTriggerRenamesCasbinRule(t *testing.T) {
	InitConfig()

	const (
		owner   = "built-in"
		oldName = "test-rename-old"
		newName = "test-rename-new"
		group   = "group:test-rename-group"
	)
	oldId := util.GetId(owner, oldName)
	newId := util.GetId(owner, newName)
	tableName := conf.GetConfigString("tableNamePrefix") + "casbin_user_rule"

	if _, err := xormadapter.NewAdapterByEngineWithTableName(ormer.Engine, "casbin_user_rule", conf.GetConfigString("tableNamePrefix")); err != nil {
		t.Fatalf("create casbin user rule table: %v", err)
	}

	rule := &xormadapter.CasbinRule{Ptype: "g", V0: oldId, V1: group}
	if _, err := ormer.Engine.Table(tableName).Insert(rule); err != nil {
		t.Fatalf("seed g rule: %v", err)
	}
	t.Cleanup(func() {
		_, _ = ormer.Engine.Table(tableName).
			Where("ptype = ? AND v0 = ? AND v1 = ?", "g", newId, group).
			Delete(&xormadapter.CasbinRule{})
	})

	if err := userChangeTrigger(owner, oldName, newName); err != nil {
		t.Fatalf("userChangeTrigger: %v", err)
	}

	var rows []*xormadapter.CasbinRule
	if err := ormer.Engine.Table(tableName).
		Where("ptype = ? AND v1 = ?", "g", group).
		Find(&rows); err != nil {
		t.Fatalf("query g rules: %v", err)
	}

	var foundNew, foundOld bool
	for _, r := range rows {
		switch r.V0 {
		case newId:
			foundNew = true
		case oldId:
			foundOld = true
		}
	}

	if !foundNew {
		t.Errorf("expected g rule with v0=%q, none found (rows=%+v)", newId, rows)
	}
	if foundOld {
		t.Errorf("expected old g rule with v0=%q to be renamed away, still present (rows=%+v)", oldId, rows)
	}
}
