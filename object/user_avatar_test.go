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
	"fmt"
	"testing"

	"github.com/casdoor/casdoor/proxy"
)

func TestSyncPermanentAvatars(t *testing.T) {
	InitConfig()
	InitDefaultStorageProvider()
	proxy.InitHttpClient()

	users, err := GetGlobalUsers()
	if err != nil {
		panic(err)
	}

	for i, user := range users {
		if user.Avatar == "" {
			continue
		}

		user.PermanentAvatar, err = getPermanentAvatarUrl(user.Owner, user.Name, user.Avatar, true)
		if err != nil {
			panic(err)
		}

		updateUserColumn("permanent_avatar", user)
		fmt.Printf("[%d/%d]: Update user: [%s]'s permanent avatar: %s\n", i, len(users), user.GetId(), user.PermanentAvatar)
	}
}

func TestUpdateAvatars(t *testing.T) {
	InitConfig()
	InitDefaultStorageProvider()
	proxy.InitHttpClient()

	users, err := GetUsers("casdoor")
	if err != nil {
		panic(err)
	}

	for _, user := range users {
		//if strings.HasPrefix(user.Avatar, "http") {
		//	continue
		//}

		if user.AvatarType != "Favicon" {
			continue
		}

		updated, err := user.refreshAvatar()
		if err != nil {
			panic(err)
		}

		if updated {
			user.PermanentAvatar = "*"
			_, err = UpdateUser(user.GetId(), user, []string{"avatar", "avatar_type"}, true)
			if err != nil {
				panic(err)
			}
		}
	}
}
