// Copyright 2021 The casbin Authors. All Rights Reserved.
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
	"testing"

	"github.com/casdoor/casdoor/util"
	"xorm.io/core"
)

func updateUserColumn(column string, user *User) bool {
	affected, err := adapter.Engine.ID(core.PK{user.Owner, user.Name}).Cols(column).Update(user)
	if err != nil {
		panic(err)
	}

	return affected != 0
}

func TestSyncAvatarsFromGitHub(t *testing.T) {
	InitConfig()

	users := GetGlobalUsers()
	for _, user := range users {
		if user.Github == "" {
			continue
		}

		user.Avatar = fmt.Sprintf("https://avatars.githubusercontent.com/%s", user.Github)
		updateUserColumn("avatar", user)
	}
}

func TestSyncIds(t *testing.T) {
	InitConfig()

	users := GetGlobalUsers()
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

	users := GetGlobalUsers()
	for _, user := range users {
		if user.Hash != "" {
			continue
		}

		user.UpdateUserHash()
		updateUserColumn("hash", user)
	}
}

func TestGetSaltedPassword(t *testing.T) {
	password := "123456"
	salt := "123"
	fmt.Printf("%s -> %s\n", password, getSaltedPassword(password, salt))
}
